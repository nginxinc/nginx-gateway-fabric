package collectors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nginxinc/nginx-plus-go-client/client"
	prometheusClient "github.com/nginxinc/nginx-prometheus-exporter/client"
	nginxCollector "github.com/nginxinc/nginx-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics"
)

const (
	nginxStatusSock        = "/var/run/nginx/nginx-status.sock"
	nginxStatusURI         = "http://config-status/stub_status"
	nginxPlusAPISock       = "/var/run/nginx/nginx-plus-api.sock"
	nginxPlusAPIURI        = "http://nginx-plus-api/api"
	nginxStatusSockTimeout = 10 * time.Second
)

var (
	upstreamServerVariableLabels = []string{
		"service", "resource_type", "resource_name", "resource_namespace",
	}
	upstreamServerPeerVariableLabelNames = []string{"pod_name"}
	streamUpstreamServerVariableLabels   = []string{
		"service", "resource_type", "resource_name", "resource_namespace",
	}
	streamUpstreamServerPeerVariableLabelNames = []string{"pod_name"}
	serverZoneVariableLabels                   = []string{"resource_type", "resource_name", "resource_namespace"}
	streamServerZoneVariableLabels             = []string{"resource_type", "resource_name", "resource_namespace"}
)

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxMetricsCollector(constLabels map[string]string, isPlus bool) (prometheus.Collector, error) {
	if isPlus {
		plusClient, err := createPlusClient()
		if err != nil {
			return nil, err
		}
		variableLabelNames := nginxCollector.NewVariableLabelNames(
			upstreamServerVariableLabels,
			serverZoneVariableLabels,
			upstreamServerPeerVariableLabelNames,
			streamUpstreamServerVariableLabels,
			streamServerZoneVariableLabels,
			streamUpstreamServerPeerVariableLabelNames,
		)
		return nginxCollector.NewNginxPlusCollector(plusClient, metrics.Namespace, variableLabelNames, constLabels), nil
	}
	httpClient := getSocketClient(nginxStatusSock)

	client, err := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)
	if err != nil {
		return nil, err
	}
	return nginxCollector.NewNginxCollector(client, metrics.Namespace, constLabels), nil
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
	plusClient, err = client.NewNginxClient(&httpClient, nginxPlusAPIURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to create NginxClient for Plus: %w", err)
	}
	return plusClient, nil
}
