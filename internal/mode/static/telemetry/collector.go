package telemetry

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GraphGetter

type GraphGetter interface {
	GetLatestGraph() *graph.Graph
}

type NGFResourceCounts struct {
	Gateways       int
	GatewayClasses int
	HTTPRoutes     int
	Secrets        int
	Services       int
	EndpointSlices int
	Endpoints      int
}

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

func (c DataCollectorImpl) Collect(ctx context.Context) (Data, error) {
	nodeCount, err := collectNodeCount(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return Data{}, err
	}
	graphResourceCount := collectGraphResourceCount(c.cfg.GraphGetter)

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
	nodes := v1.NodeList{}
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return 0, err
	}

	return len(nodes.Items), nil
}

func collectGraphResourceCount(graphGetter GraphGetter) NGFResourceCounts {
	ngfResourceCounts := NGFResourceCounts{}
	g := graphGetter.GetLatestGraph()

	if g.GatewayClass != nil {
		ngfResourceCounts.GatewayClasses = 1
	}
	if g.Gateway != nil {
		ngfResourceCounts.Gateways = 1
	}
	ngfResourceCounts.HTTPRoutes = len(g.Routes)
	ngfResourceCounts.Secrets = countReferencedResources(g.ReferencedSecrets)
	ngfResourceCounts.Services = countReferencedResources(g.ReferencedServices)

	return ngfResourceCounts
}

func countReferencedResources[T comparable](referencedMap map[types.NamespacedName]T) int {
	var count int
	var zeroValue T
	for name := range referencedMap {
		if referencedMap[name] != zeroValue {
			count++
		}
	}
	return count
}
