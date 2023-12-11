package collectors

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/log"
	"github.com/nginxinc/nginx-plus-go-client/client"
	prometheusClient "github.com/nginxinc/nginx-prometheus-exporter/client"
	nginxCollector "github.com/nginxinc/nginx-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics"
)

const (
	nginxStatusSock  = "/var/run/nginx/nginx-status.sock"
	nginxStatusURI   = "http://config-status/stub_status"
	nginxPlusAPISock = "/var/run/nginx/nginx-plus-api.sock"
	nginxPlusAPIURI  = "http://nginx-plus-api/api"
)

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxMetricsCollector(constLabels map[string]string) (prometheus.Collector, error) {
	httpClient := getSocketClient(nginxStatusSock)

	ngxClient := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)

	promLogger, err := newPromLogger()
	if err != nil {
		return nil, err
	}

	return nginxCollector.NewNginxCollector(ngxClient, metrics.Namespace, constLabels, promLogger), nil
}

// NewNginxPlusMetricsCollector creates an NginxCollector which fetches stats from NGINX Plus API over a unix socket
func NewNginxPlusMetricsCollector(constLabels map[string]string) (prometheus.Collector, error) {
	plusClient, err := createPlusClient()
	if err != nil {
		return nil, err
	}

	promLogger, err := newPromLogger()
	if err != nil {
		return nil, err
	}

	collector := nginxCollector.NewNginxPlusCollector(
		plusClient,
		metrics.Namespace,
		nginxCollector.VariableLabelNames{},
		constLabels,
		promLogger,
	)

	return collector, nil
}

// newPromLogger creates a Prometheus logger that implements to go-kit log.Logger that the prometheus exporter requires.
func newPromLogger() (log.Logger, error) {
	logFormat := &promlog.AllowedFormat{}

	if err := logFormat.Set("json"); err != nil {
		return nil, err
	}

	logConfig := &promlog.Config{Format: logFormat}

	return promlog.New(logConfig), nil
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

func createPlusClient() (*client.NginxClient, error) {
	var plusClient *client.NginxClient
	var err error

	httpClient := getSocketClient(nginxPlusAPISock)
	plusClient, err = client.NewNginxClient(nginxPlusAPIURI, client.WithHTTPClient(&httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create NginxClient for Plus: %w", err)
	}
	return plusClient, nil
}
