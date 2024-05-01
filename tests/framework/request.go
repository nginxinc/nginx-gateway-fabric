package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
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
	resp, err := makeRequest(http.MethodGet, url, address, nil, timeout)
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

// Post sends a POST request to the specified url with the body as the payload.
// It resolves to the specified address instead of using DNS.
func Post(url, address string, body io.Reader, timeout time.Duration) (*http.Response, error) {
	return makeRequest(http.MethodPost, url, address, body, timeout)
}

func makeRequest(method, url, address string, body io.Reader, timeout time.Duration) (*http.Response, error) {
	dialer := &net.Dialer{}

	http.DefaultTransport.(*http.Transport).DialContext = func(
		ctx context.Context,
		network,
		addr string,
	) (net.Conn, error) {
		split := strings.Split(addr, ":")
		port := split[len(split)-1]
		return dialer.DialContext(ctx, network, fmt.Sprintf("%s:%s", address, port))
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if strings.HasPrefix(url, "https") {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		// similar to how in our examples with https requests we run our curl command
		// we turn off verification of the certificate, we do the same here
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // for https test traffic
		client := &http.Client{Transport: customTransport}
		resp, err = client.Do(req)
	} else {
		resp, err = http.DefaultClient.Do(req)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetWithRetry retries the Get function until it succeeds or the context times out.
func GetWithRetry(
	ctx context.Context,
	url,
	address string,
	requestTimeout time.Duration,
) (int, string, error) {
	var statusCode int
	var body string

	err := wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var getErr error
			statusCode, body, getErr = Get(url, address, requestTimeout)
			if getErr != nil {
				return false, getErr
			}

			if statusCode != 200 {
				log.Printf("got %d code instead of expected 200\n", statusCode)
				return false, nil
			}

			return true, nil
		},
	)

	return statusCode, body, err
}
