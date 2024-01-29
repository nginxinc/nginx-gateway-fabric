package telemetry

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GraphGetter

// GraphGetter gets the current Graph.
type GraphGetter interface {
	GetLatestGraph() *graph.Graph
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ConfigurationGetter

// ConfigurationGetter gets the current Configuration.
type ConfigurationGetter interface {
	GetLatestConfiguration() *dataplane.Configuration
}

// NGFResourceCounts stores the counts of all relevant Graph resources.
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

type DataCollectorImpl struct {
	cfg DataCollectorConfig
}

func NewDataCollector(
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
		return Data{}, err
	}

	graphResourceCount, err := collectGraphResourceCount(c.cfg.GraphGetter, c.cfg.ConfigurationGetter)
	if err != nil {
		return Data{}, err
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
		return NGFResourceCounts{}, errors.New("latest graph cannot be nil")
	}
	if cfg == nil {
		return NGFResourceCounts{}, errors.New("latest configuration cannot be nil")
	}

	if g.GatewayClass != nil {
		ngfResourceCounts.GatewayClasses = 1
	}
	if g.Gateway != nil {
		ngfResourceCounts.Gateways = 1
	}
	ngfResourceCounts.HTTPRoutes = len(g.Routes)
	ngfResourceCounts.Secrets = countReferencedResources(g.ReferencedSecrets)
	ngfResourceCounts.Services = countReferencedResources(g.ReferencedServices)

	for _, upstream := range cfg.Upstreams {
		if upstream.ErrorMsg == "" {
			ngfResourceCounts.Endpoints += len(upstream.Endpoints)
		}
	}

	return ngfResourceCounts, nil
}

// countReferencedResources counts the amount of non-nil resources.
func countReferencedResources[T comparable](referencedMap map[types.NamespacedName]T) int {
	var count int
	// because we can't compare T to nil, we need to use the zeroValue of T
	var zeroValue T

	for name := range referencedMap {
		if referencedMap[name] != zeroValue {
			count++
		}
	}
	return count
}
