package metrics

import (
	"context"
	"net"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	prometheusClient "github.com/nginxinc/nginx-prometheus-exporter/client"
	nginxCollector "github.com/nginxinc/nginx-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	nginxStatusSock        = "/var/run/nginx/nginx-status.sock"
	nginxStatusURI         = "http://config-status/stub_status"
	nginxStatusSockTimeout = 10 * time.Second
)

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxMetricsCollector(ctx context.Context, constLabels map[string]string) (prometheus.Collector, error) {
	httpClient := getSocketClient(nginxStatusSock)
	err := ensureStatusSockActive(ctx, httpClient, nginxStatusSockTimeout)
	if err != nil {
		return nil, err
	}
	client, err := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)
	if err != nil {
		return nil, err
	}
	return nginxCollector.NewNginxCollector(client, "nginx_kubernetes_gateway", constLabels), nil
}

// getSocketClient gets an http.Client with a unix socket transport.
func getSocketClient(sockPath string) http.Client {
	return http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}
}

// ensureStatusSockActive waits until NGINX is serving metrics
func ensureStatusSockActive(ctx context.Context, httpClient http.Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nginxStatusURI, nil)
	if err != nil {
		return err
	}

	err = wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			resp, err := httpClient.Do(req)
			if err != nil {
				//lint:ignore nilerr reason
				return false, nil
			}
			resp.Body.Close()
			return true, nil
		})
	if err != nil {
		return err
	}

	return nil
}
