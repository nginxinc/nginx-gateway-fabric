package telemetry

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GraphGetter

// GraphGetter gets the latest Graph.
type GraphGetter interface {
	GetLatestGraph() *graph.Graph
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ConfigurationGetter

// ConfigurationGetter gets the latest Configuration.
type ConfigurationGetter interface {
	GetLatestConfiguration() *dataplane.Configuration
}

// NGFResourceCounts stores the counts of all relevant resources that NGF processes and generates configuration from.
type NGFResourceCounts struct {
	Gateways       int
	GatewayClasses int
	HTTPRoutes     int
	Secrets        int
	Services       int
	Endpoints      int
}

// ProjectMetadata stores the name of the project and the current version.
type ProjectMetadata struct {
	Name    string
	Version string
}

// Data is telemetry data.
// Note: this type might change once https://github.com/nginxinc/nginx-gateway-fabric/issues/1318 is implemented.
type Data struct {
	ProjectMetadata   ProjectMetadata
	NodeCount         int
	NGFResourceCounts NGFResourceCounts
}

// DataCollectorConfig holds configuration parameters for DataCollectorImpl.
type DataCollectorConfig struct {
	// K8sClientReader is a Kubernetes API client Reader.
	K8sClientReader client.Reader
	// GraphGetter allows us to get the Graph.
	GraphGetter GraphGetter
	// ConfigurationGetter allows us to get the Configuration.
	ConfigurationGetter ConfigurationGetter
	// Version is the NGF version.
	Version string
}

// DataCollectorImpl is am implementation of DataCollector.
type DataCollectorImpl struct {
	cfg DataCollectorConfig
}

// NewDataCollectorImpl creates a new DataCollectorImpl for a telemetry Job.
func NewDataCollectorImpl(
	cfg DataCollectorConfig,
) *DataCollectorImpl {
	return &DataCollectorImpl{
		cfg: cfg,
	}
}

// Collect collects and returns telemetry Data.
func (c DataCollectorImpl) Collect(ctx context.Context) (Data, error) {
	nodeCount, err := collectNodeCount(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect node count: %w", err)
	}

	graphResourceCount, err := collectGraphResourceCount(c.cfg.GraphGetter, c.cfg.ConfigurationGetter)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect NGF resource counts: %w", err)
	}

	data := Data{
		NodeCount:         nodeCount,
		NGFResourceCounts: graphResourceCount,
		ProjectMetadata: ProjectMetadata{
			Name:    "NGF",
			Version: c.cfg.Version,
		},
	}

	return data, nil
}

func collectNodeCount(ctx context.Context, k8sClient client.Reader) (int, error) {
	var nodes v1.NodeList
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return 0, err
	}

	return len(nodes.Items), nil
}

func collectGraphResourceCount(
	graphGetter GraphGetter,
	configurationGetter ConfigurationGetter,
) (NGFResourceCounts, error) {
	ngfResourceCounts := NGFResourceCounts{}
	g := graphGetter.GetLatestGraph()
	cfg := configurationGetter.GetLatestConfiguration()

	if g == nil {
		return ngfResourceCounts, errors.New("latest graph cannot be nil")
	}
	if cfg == nil {
		return ngfResourceCounts, errors.New("latest configuration cannot be nil")
	}

	ngfResourceCounts.GatewayClasses = len(g.IgnoredGatewayClasses)
	if g.GatewayClass != nil {
		ngfResourceCounts.GatewayClasses++
	}

	ngfResourceCounts.Gateways = len(g.IgnoredGateways)
	if g.Gateway != nil {
		ngfResourceCounts.Gateways++
	}

	ngfResourceCounts.HTTPRoutes = len(g.Routes)
	ngfResourceCounts.Secrets = len(g.ReferencedSecrets)
	ngfResourceCounts.Services = len(g.ReferencedServices)

	for _, upstream := range cfg.Upstreams {
		if upstream.ErrorMsg == "" {
			ngfResourceCounts.Endpoints += len(upstream.Endpoints)
		}
	}

	return ngfResourceCounts, nil
}
