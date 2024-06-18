package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/nginxinc/nginx-plus-go-client/client"
)

const (
	nginxPlusAPISock = "/var/lib/nginx/nginx-plus-api.sock"
	nginxPlusAPIURI  = "http://nginx-plus-api/api"
)

// CreatePlusClient returns a client for communicating with the NGINX Plus API.
func CreatePlusClient() (*client.NginxClient, error) {
	var plusClient *client.NginxClient
	var err error

	httpClient := GetSocketClient(nginxPlusAPISock)
	plusClient, err = client.NewNginxClient(nginxPlusAPIURI, client.WithHTTPClient(&httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create NginxClient for Plus: %w", err)
	}
	return plusClient, nil
}

// GetSocketClient gets an http.Client with a unix socket transport.
func GetSocketClient(sockPath string) http.Client {
	return http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}
}
