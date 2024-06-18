package collectors

import (
	"github.com/go-kit/log"
	"github.com/nginxinc/nginx-plus-go-client/client"
	prometheusClient "github.com/nginxinc/nginx-prometheus-exporter/client"
	nginxCollector "github.com/nginxinc/nginx-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
)

const (
	nginxStatusSock = "/var/lib/nginx/nginx-status.sock"
	nginxStatusURI  = "http://config-status/stub_status"
)

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxMetricsCollector(constLabels map[string]string, logger log.Logger) prometheus.Collector {
	httpClient := runtime.GetSocketClient(nginxStatusSock)
	ngxClient := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)

	return nginxCollector.NewNginxCollector(ngxClient, metrics.Namespace, constLabels, logger)
}

// NewNginxPlusMetricsCollector creates an NginxCollector which fetches stats from NGINX Plus API over a unix socket
func NewNginxPlusMetricsCollector(
	plusClient *client.NginxClient,
	constLabels map[string]string,
	logger log.Logger,
) (prometheus.Collector, error) {
	collector := nginxCollector.NewNginxPlusCollector(
		plusClient,
		metrics.Namespace,
		nginxCollector.VariableLabelNames{},
		constLabels,
		logger,
	)

	return collector, nil
}
