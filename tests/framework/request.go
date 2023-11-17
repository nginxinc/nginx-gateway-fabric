package framework

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// GET sends a GET request to the specified url.
// It resolves to localhost (where the NGF port forward is running) instead of using DNS.
// The body of the response is returned, or an error.
func GET(url string, timeout time.Duration) (string, error) {
	dialer := &net.Dialer{}

	http.DefaultTransport.(*http.Transport).DialContext = func(
		ctx context.Context,
		network,
		addr string,
	) (net.Conn, error) {
		split := strings.Split(addr, ":")
		port := split[len(split)-1]
		return dialer.DialContext(ctx, network, fmt.Sprintf("127.0.0.1:%s", port))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body := new(bytes.Buffer)
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}
