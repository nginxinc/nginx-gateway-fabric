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
	upstreamServerVariableLabels               = []string{}
	upstreamServerPeerVariableLabelNames       = []string{}
	streamUpstreamServerVariableLabels         = []string{}
	streamUpstreamServerPeerVariableLabelNames = []string{}
	serverZoneVariableLabels                   = []string{}
	streamServerZoneVariableLabels             = []string{}
)

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxMetricsCollector(constLabels map[string]string) (prometheus.Collector, error) {
	httpClient := getSocketClient(nginxStatusSock)

	client, err := prometheusClient.NewNginxClient(&httpClient, nginxStatusURI)
	if err != nil {
		return nil, err
	}
	return nginxCollector.NewNginxCollector(client, metrics.Namespace, constLabels), nil
}

// NewNginxMetricsCollector creates an NginxCollector which fetches stats from NGINX over a unix socket
func NewNginxPlusMetricsCollector(constLabels map[string]string) (prometheus.Collector, error) {
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
		return nil, fmt.Errorf("failed to create NginxClient for Plus: %w", err)
	}
	return plusClient, nil
}
