package telemetry

import (
	"context"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

// kubeSystem indicates the name of kube-system namespace.
const kubeSystem = "kube-system"

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
	// Endpoints include the total count of Endpoints(IP:port) across all referenced services.
	Endpoints int
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
	ClusterID         string
	NodeCount         int
	NGFResourceCounts NGFResourceCounts
	NGFReplicaCount   int
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
	// PodNSName is the NamespacedName of the NGF Pod.
	PodNSName types.NamespacedName
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

	ngfReplicaCount, err := collectNGFReplicaCount(ctx, c.cfg.K8sClientReader, c.cfg.PodNSName)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect NGF replica count: %w", err)
	}

	var clusterID string
	if clusterID, err = collectClusterID(ctx, c.cfg.K8sClientReader); err != nil {
		return Data{}, fmt.Errorf("failed to collect clusterID: %w", err)
	}

	data := Data{
		NodeCount:         nodeCount,
		NGFResourceCounts: graphResourceCount,
		ProjectMetadata: ProjectMetadata{
			Name:    "NGF",
			Version: c.cfg.Version,
		},
		NGFReplicaCount: ngfReplicaCount,
		ClusterID:       clusterID,
	}

	return data, nil
}

func collectNodeCount(ctx context.Context, k8sClient client.Reader) (int, error) {
	var nodes v1.NodeList
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return 0, fmt.Errorf("failed to get NodeList: %w", err)
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

func collectNGFReplicaCount(ctx context.Context, k8sClient client.Reader, podNSName types.NamespacedName) (int, error) {
	var pod v1.Pod
	if err := k8sClient.Get(
		ctx,
		types.NamespacedName{Namespace: podNSName.Namespace, Name: podNSName.Name},
		&pod,
	); err != nil {
		return 0, fmt.Errorf("failed to get NGF Pod: %w", err)
	}

	podOwnerRefs := pod.GetOwnerReferences()
	if len(podOwnerRefs) != 1 {
		return 0, fmt.Errorf("expected one owner reference of the NGF Pod, got %d", len(podOwnerRefs))
	}

	if podOwnerRefs[0].Kind != "ReplicaSet" {
		return 0, fmt.Errorf("expected pod owner reference to be ReplicaSet, got %s", podOwnerRefs[0].Kind)
	}

	var replicaSet appsv1.ReplicaSet
	if err := k8sClient.Get(
		ctx,
		types.NamespacedName{Namespace: podNSName.Namespace, Name: podOwnerRefs[0].Name},
		&replicaSet,
	); err != nil {
		return 0, fmt.Errorf("failed to get NGF Pod's ReplicaSet: %w", err)
	}

	if replicaSet.Spec.Replicas == nil {
		return 0, errors.New("replica set replicas was nil")
	}

	return int(*replicaSet.Spec.Replicas), nil
}

func collectClusterID(ctx context.Context, k8sClient client.Reader) (string, error) {
	key := types.NamespacedName{
		Name: kubeSystem,
	}
	var kubeNamespace v1.Namespace
	if err := k8sClient.Get(ctx, key, &kubeNamespace); err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %w", err)
	}
	return string(kubeNamespace.GetUID()), nil
}
