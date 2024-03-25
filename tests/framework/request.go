package framework

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Get sends a GET request to the specified url.
// It resolves to the specified address instead of using DNS.
// The status and body of the response is returned, or an error.
func Get(url, address string, timeout time.Duration) (int, string, error) {
	dialer := &net.Dialer{}

	transport := &http.Transport{
		DialContext: func(
			ctx context.Context,
			network,
			addr string,
		) (net.Conn, error) {
			split := strings.Split(addr, ":")
			port := split[len(split)-1]
			return dialer.DialContext(ctx, network, fmt.Sprintf("%s:%s", address, port))
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body := new(bytes.Buffer)
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, body.String(), nil
}

func WaitForResponseForHost(url, address string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	start := time.Now()

	err := wait.PollUntilContextCancel(
		ctx,
		200*time.Millisecond,
		true,
		func(ctx context.Context) (done bool, err error) {
			status, _, err := Get(url, address, timeout)
			if err != nil {
				return false, err
			}

			if status == http.StatusOK {
				return true, nil
			}

			return false, nil
		})

	return time.Since(start), err
}
