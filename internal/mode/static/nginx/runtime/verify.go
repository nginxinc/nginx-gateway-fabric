package runtime

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

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
					return net.Dial("unix", "/var/run/nginx/nginx-config-version.sock")
				},
			},
		},
		timeout: timeout,
	}
}

// GetConfigVersion gets the version number that we put in the nginx config to verify that we're using
// the correct config.
func (c *verifyClient) GetConfigVersion() (int, error) {
	ctx := context.Background()
	reqContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqContext, "GET", "http://config-version/configVersion", nil)
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

// WaitForCorrectVersion calls the config version endpoint until it gets the expectedVersion,
// which ensures that a new worker process has been started for that config version.
func (c *verifyClient) WaitForCorrectVersion(ctx context.Context, expectedVersion int) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := wait.PollUntilContextCancel(
		ctx,
		25*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			version, err := c.GetConfigVersion()
			if err != nil {
				return false, err
			}
			if version == expectedVersion {
				return true, nil
			}
			return false, nil
		})
	if err != nil {
		return fmt.Errorf("could not get expected version %v: %w", expectedVersion, err)
	}

	return nil
}
