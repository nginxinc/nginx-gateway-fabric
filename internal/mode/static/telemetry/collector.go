package telemetry

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
)

type Collector interface {
	CollectData() Data
}

type DataCollector struct {
	ctx       context.Context
	k8sClient client.Client
	processor state.ChangeProcessor
}

type GraphResourceCount struct {
	Gateways           int
	GatewayClasses     int
	HTTPRoutes         int
	ReferencedSecrets  int
	ReferencedServices int
}

func NewDataCollector(
	ctx context.Context,
	k8sClient client.Client,
	processor state.ChangeProcessor,
) *DataCollector {
	return &DataCollector{
		ctx:       ctx,
		k8sClient: k8sClient,
		processor: processor,
	}
}

func (c DataCollector) CollectData() Data {
	nodeCount := collectNodeCount(c.ctx, c.k8sClient)
	graphResourceCount := collectGraphResourceCount(c.processor)

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

func collectGraphResourceCount(processor state.ChangeProcessor) GraphResourceCount {
	graphResourceCount := GraphResourceCount{}
	graph := processor.GetLatestGraph()

	if graph.GatewayClass != nil {
		graphResourceCount.GatewayClasses = 1
	}
	if graph.Gateway != nil {
		graphResourceCount.GatewayClasses = 1
	}
	if graph.Routes != nil {
		graphResourceCount.HTTPRoutes = len(graph.Routes)
	}
	if graph.ReferencedSecrets != nil {
		graphResourceCount.ReferencedSecrets = len(graph.ReferencedSecrets)
	}

	return graphResourceCount
}
