package licensing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
)

const integrationID = "ngf"

// DeploymentContextCollectorConfig contains the configuration for the DeploymentContextCollector.
type DeploymentContextCollectorConfig struct {
	// K8sClientReader is a Kubernetes API client Reader.
	K8sClientReader client.Reader
	// PodNSName is the NamespacedName of the NGF Pod.
	PodNSName types.NamespacedName
}

// DeploymentContextCollector collects the deployment context information needed for N+ licensing.
type DeploymentContextCollector struct {
	cfg DeploymentContextCollectorConfig
}

// NewDeploymentContextCollector returns a new instance of DeploymentContextCollector.
func NewDeploymentContextCollector(
	cfg DeploymentContextCollectorConfig,
) *DeploymentContextCollector {
	return &DeploymentContextCollector{
		cfg: cfg,
	}
}

// Collect collects all the information needed to create the deployment context for N+ licensing.
func (c *DeploymentContextCollector) Collect(
	ctx context.Context,
	logger logr.Logger,
) (dataplane.DeploymentContext, error) {
	clusterInfo, err := telemetry.CollectClusterInformation(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return dataplane.DeploymentContext{}, fmt.Errorf("error getting cluster information: %w", err)
	}

	var installationID string

	// InstallationID is not required by the usage API, so if we can't get it, don't return an error
	replicaSet, err := telemetry.GetPodReplicaSet(ctx, c.cfg.K8sClientReader, c.cfg.PodNSName)
	if err != nil {
		logger.Error(err, "failed to get NGF installationID")
	} else {
		installationID, err = telemetry.GetDeploymentID(replicaSet)
		if err != nil {
			logger.Error(err, "failed to get NGF installationID")
		}
	}

	depCtx := dataplane.DeploymentContext{
		Integration:      integrationID,
		ClusterID:        clusterInfo.ClusterID,
		ClusterNodeCount: clusterInfo.NodeCount,
		InstallationID:   installationID,
	}

	return depCtx, nil
}
