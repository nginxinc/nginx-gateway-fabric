package usage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
)

func CreateUsageJobWorker(
	logger logr.Logger,
	k8sClient client.Reader,
	reporter Reporter,
	cfg config.Config,
) func(ctx context.Context) {
	return func(ctx context.Context) {
		nodeCount, err := CollectNodeCount(ctx, k8sClient)
		if err != nil {
			logger.Error(err, "Failed to collect node count")
			return
		}

		podCount, err := GetTotalNGFPodCount(ctx, k8sClient)
		if err != nil {
			logger.Error(err, "Failed to collect replica count")
			return
		}

		clusterUID, err := telemetry.CollectClusterID(ctx, k8sClient)
		if err != nil {
			logger.Error(err, "Failed to collect cluster UID")
			return
		}

		clusterDetails := ClusterDetails{
			Metadata: Metadata{
				DisplayName: cfg.UsageReportConfig.ClusterDisplayName,
				UID:         clusterUID,
			},
			NodeCount: int64(nodeCount),
			PodDetails: PodDetails{
				CurrentPodCounts: CurrentPodsCount{
					DosCount: int64(0),
					PodCount: int64(podCount),
					WafCount: int64(0),
				},
			},
		}

		if err := reporter.Report(ctx, clusterDetails); err != nil {
			logger.Error(err, "Failed to report NGINX Plus usage")
		}
	}
}

// GetTotalNGFPodCount returns the total count of NGF Pods in the cluster.
// Uses the "app.kubernetes.io/name" label with either value of "nginx-gateway" or "nginx-gateway-fabric".
func GetTotalNGFPodCount(ctx context.Context, k8sClient client.Reader) (int, error) {
	labelKey := "app.kubernetes.io/name"
	labelVals := map[string]struct{}{
		"nginx-gateway-fabric": {},
		"nginx-gateway":        {},
	}

	var rsList appsv1.ReplicaSetList
	if err := k8sClient.List(ctx, &rsList, client.HasLabels{labelKey}); err != nil {
		return 0, fmt.Errorf("failed to list replicasets: %w", err)
	}

	var count int
	for _, rs := range rsList.Items {
		val := rs.Labels[labelKey]
		if _, ok := labelVals[val]; ok && rs.Spec.Replicas != nil {
			count += int(*rs.Spec.Replicas)
		}
	}

	return count, nil
}

// CollectNodeCount returns the number of nodes in the cluster.
func CollectNodeCount(ctx context.Context, k8sClient client.Reader) (int, error) {
	var nodes v1.NodeList
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return 0, fmt.Errorf("failed to get NodeList: %w", err)
	}

	return len(nodes.Items), nil
}
