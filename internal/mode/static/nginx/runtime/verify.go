package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const configVersionURI = "/var/run/nginx/nginx-config-version.sock"

var noNewWorkersErrFmt = "reload unsuccessful: no new NGINX worker processes started for config version %d." +
	" Please check the NGINX container logs for possible configuration issues: %w"

// verifyClient is a client for verifying the config version.
type verifyClient struct {
	client  *http.Client
	timeout time.Duration
}

// newVerifyClient returns a new client pointed at the config version socket.
func newVerifyClient(timeout time.Duration) *verifyClient {
	return &verifyClient{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", configVersionURI)
				},
			},
		},
		timeout: timeout,
	}
}

// getConfigVersion gets the version number that we put in the nginx config to verify that we're using
// the correct config.
func (c *verifyClient) getConfigVersion() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://config-version/version", nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error getting client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-200 response: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read the response body: %w", err)
	}
	v, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, fmt.Errorf("error converting string to int: %w", err)
	}
	return v, nil
}

// waitForCorrectVersion first ensures any new worker processes have been started, and then calls the config version
// endpoint until it gets the expectedVersion, which ensures that a new worker process has been started for that config
// version.
func (c *verifyClient) waitForCorrectVersion(
	ctx context.Context,
	expectedVersion int,
	childProcFile string,
	previousChildProcesses []byte,
	readFile readFileFunc,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := ensureNewNginxWorkers(
		ctx,
		childProcFile,
		previousChildProcesses,
		readFile,
	); err != nil {
		return fmt.Errorf(noNewWorkersErrFmt, expectedVersion, err)
	}

	if err := c.ensureConfigVersion(ctx, expectedVersion); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf(
				"config version check didn't return expected version %d within the deadline",
				expectedVersion,
			)
		}
		return fmt.Errorf("could not get expected config version %d: %w", expectedVersion, err)
	}
	return nil
}

func (c *verifyClient) ensureConfigVersion(ctx context.Context, expectedVersion int) error {
	return wait.PollUntilContextCancel(
		ctx,
		25*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			version, err := c.getConfigVersion()
			return version == expectedVersion, err
		},
	)
}

func ensureNewNginxWorkers(
	ctx context.Context,
	childProcFile string,
	previousContents []byte,
	readFile readFileFunc,
) error {
	return wait.PollUntilContextCancel(
		ctx,
		25*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			content, err := readFile(childProcFile)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(previousContents, content) {
				return true, nil
			}
			return false, nil
		},
	)
}
