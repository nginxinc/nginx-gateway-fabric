package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Get sends a GET request to the specified url.
// It resolves to the specified address instead of using DNS.
// The status and body of the response is returned, or an error.
func Get(url, address string, timeout time.Duration) (int, string, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}

	var resp *http.Response
	if strings.HasPrefix(url, "https") {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		// similar to how in our examples with https requests we run our curl command
		// we turn off verification of the certificate, we do the same here
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		client := &http.Client{Transport: customTransport}
		resp, err = client.Do(req)
	} else {
		resp, err = http.DefaultClient.Do(req)
	}

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
