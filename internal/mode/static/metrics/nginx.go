package metrics

import (
	"context"
	"net"
	"net/http"
	"time"

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
func NewNginxMetricsCollector(constLabels map[string]string) (prometheus.Collector, error) {
	httpClient := getSocketClient(nginxStatusSock)
	client, err := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)
	if err != nil {
		return nil, err
	}
	return nginxCollector.NewNginxCollector(client, metricsNamespace, constLabels), nil
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
