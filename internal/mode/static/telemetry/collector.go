package telemetry

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector interface {
	CollectData() Data
}

type DataCollector struct {
	ctx       context.Context
	k8sClient client.Client
}

func NewDataCollector(ctx context.Context, k8sClient client.Client) *DataCollector {
	return &DataCollector{
		ctx:       ctx,
		k8sClient: k8sClient,
	}
}

func (c DataCollector) CollectData() Data {
	nodeCount := collectNodeCount(c.ctx, c.k8sClient)

	data := Data{NodeCount: nodeCount}

	return data
}

func collectNodeCount(ctx context.Context, k8sClient client.Client) int {
	nodes := &v1.NodeList{}
	_ = k8sClient.List(ctx, nodes)
	return len(nodes.Items)
}
