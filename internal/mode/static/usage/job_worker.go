package usage

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
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
		nodeCount, err := telemetry.CollectNodeCount(ctx, k8sClient)
		if err != nil {
			logger.Error(err, "Failed to collect node count")
		}

		podCount, err := telemetry.CollectNGFReplicaCount(
			ctx,
			k8sClient,
			types.NamespacedName{
				Namespace: cfg.GatewayPodConfig.Namespace,
				Name:      cfg.GatewayPodConfig.Name,
			},
		)
		if err != nil {
			logger.Error(err, "Failed to collect replica count")
		}

		clusterUID, err := telemetry.CollectClusterID(ctx, k8sClient)
		if err != nil {
			logger.Error(err, "Failed to collect cluster UID")
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
