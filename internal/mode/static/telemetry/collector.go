package telemetry

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

type DataCollector interface {
	Collect(ctx context.Context) Data
}
type GraphGetter interface {
	GetLatestGraph() *graph.Graph
}

type GraphResourceCount struct {
	Gateways       int
	GatewayClasses int
	HTTPRoutes     int
	Secrets        int
	Services       int
}

// Data is telemetry data.
// Note: this type might change once https://github.com/nginxinc/nginx-gateway-fabric/issues/1318 is implemented.
type Data struct {
	NodeCount          int
	GraphResourceCount GraphResourceCount
}

type DataCollectorImpl struct {
	k8sClient   client.Client
	graphGetter GraphGetter
}

func NewDataCollector(
	k8sClient client.Client,
	processor state.ChangeProcessor,
) *DataCollectorImpl {
	return &DataCollectorImpl{
		k8sClient:   k8sClient,
		graphGetter: processor,
	}
}

func (c DataCollectorImpl) Collect(ctx context.Context) Data {
	nodeCount := collectNodeCount(ctx, c.k8sClient)
	graphResourceCount := collectGraphResourceCount(c.graphGetter)

	data := Data{
		NodeCount:          nodeCount,
		GraphResourceCount: graphResourceCount,
	}

	return data
}

func collectNodeCount(ctx context.Context, k8sClient client.Client) int {
	nodes := &v1.NodeList{}
	_ = k8sClient.List(ctx, nodes)
	return len(nodes.Items)
}

func collectGraphResourceCount(graphGetter GraphGetter) GraphResourceCount {
	graphResourceCount := GraphResourceCount{}
	g := graphGetter.GetLatestGraph()

	if g.GatewayClass != nil {
		graphResourceCount.GatewayClasses = 1
	}
	if g.Gateway != nil {
		graphResourceCount.GatewayClasses = 1
	}
	if g.Routes != nil {
		graphResourceCount.HTTPRoutes = len(g.Routes)
	}
	if g.ReferencedSecrets != nil {
		graphResourceCount.Secrets = len(g.ReferencedSecrets)
	}

	return graphResourceCount
}
