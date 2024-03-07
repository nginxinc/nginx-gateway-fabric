package telemetry

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	tel "github.com/nginxinc/telemetry-exporter/pkg/telemetry"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
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

// Data is telemetry data.
//
//go:generate go run -tags generator github.com/nginxinc/telemetry-exporter/cmd/generator -type=Data -scheme -scheme-protocol=NGFProductTelemetry -scheme-df-datatype=ngf-product-telemetry
type Data struct {
	// ImageSource tells whether the image was built by GitHub or locally (values are 'gha', 'local', or 'unknown')
	ImageSource string
	tel.Data    // embedding is required by the generator.
	// FlagNames contains the command-line flag names.
	FlagNames []string
	// FlagValues contains the values of the command-line flags, where each value corresponds to the flag from FlagNames
	// at the same index.
	// Each value is either 'true' or 'false' for boolean flags and 'default' or 'user-defined' for non-boolean flags.
	FlagValues        []string
	NGFResourceCounts // embedding is required by the generator.
	// NGFReplicaCount is the number of replicas of the NGF Pod.
	NGFReplicaCount int64
}

// NGFResourceCounts stores the counts of all relevant resources that NGF processes and generates configuration from.
//
//go:generate go run -tags generator github.com/nginxinc/telemetry-exporter/cmd/generator -type=NGFResourceCounts
type NGFResourceCounts struct {
	// GatewayCount is the number of relevant Gateways.
	GatewayCount int64
	// GatewayClassCount is the number of relevant GatewayClasses.
	GatewayClassCount int64
	// HTTPRouteCount is the number of relevant HTTPRoutes.
	HTTPRouteCount int64
	// SecretCount is the number of relevant Secrets.
	SecretCount int64
	// ServiceCount is the number of relevant Services.
	ServiceCount int64
	// EndpointCount include the total count of Endpoints(IP:port) across all referenced services.
	EndpointCount int64
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
	// ImageSource is the source of the NGF image.
	ImageSource string
	// Flags contains the command-line NGF flag keys and values.
	Flags config.Flags
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

// notImplemented is a value for string field, for which collection is not implemented yet.
const notImplemented = "not-implemented"

// Collect collects and returns telemetry Data.
func (c DataCollectorImpl) Collect(ctx context.Context) (Data, error) {
	nodeCount, err := CollectNodeCount(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect node count: %w", err)
	}

	nodes, err := CollectNodeList(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return Data{}, err
	}

	node := nodes.Items[0]
	k8sVersion := node.Status.NodeInfo.KubeletVersion
	k8sPlatform := strings.Split(node.Spec.ProviderID, "://")[0]

	graphResourceCount, err := collectGraphResourceCount(c.cfg.GraphGetter, c.cfg.ConfigurationGetter)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect NGF resource counts: %w", err)
	}

	replicaSet, err := getPodReplicaSet(ctx, c.cfg.K8sClientReader, c.cfg.PodNSName)
	if err != nil {
		return Data{}, fmt.Errorf("failed to get replica set for pod %s: %w", c.cfg.PodNSName, err)
	}

	replicaCount, err := getReplicas(replicaSet)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect NGF replica count: %w", err)
	}

	deploymentID, err := getDeploymentID(replicaSet)
	if err != nil {
		return Data{}, fmt.Errorf("failed to get NGF deploymentID: %w", err)
	}

	var clusterID string
	if clusterID, err = CollectClusterID(ctx, c.cfg.K8sClientReader); err != nil {
		return Data{}, fmt.Errorf("failed to collect clusterID: %w", err)
	}

	data := Data{
		Data: tel.Data{
			ProjectName:         "NGF",
			ProjectVersion:      c.cfg.Version,
			ProjectArchitecture: runtime.GOARCH,
			ClusterID:           clusterID,
			ClusterVersion:      k8sVersion,
			ClusterPlatform:     k8sPlatform,
			InstallationID:      deploymentID,
			ClusterNodeCount:    int64(nodeCount),
		},
		NGFResourceCounts: graphResourceCount,
		ImageSource:       c.cfg.ImageSource,
		FlagNames:         c.cfg.Flags.Names,
		FlagValues:        c.cfg.Flags.Values,
		NGFReplicaCount:   int64(replicaCount),
	}

	return data, nil
}

// CollectNodeCount returns the number of nodes in the cluster.
func CollectNodeCount(ctx context.Context, k8sClient client.Reader) (int, error) {
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

	ngfResourceCounts.GatewayClassCount = int64(len(g.IgnoredGatewayClasses))
	if g.GatewayClass != nil {
		ngfResourceCounts.GatewayClassCount++
	}

	ngfResourceCounts.GatewayCount = int64(len(g.IgnoredGateways))
	if g.Gateway != nil {
		ngfResourceCounts.GatewayCount++
	}

	ngfResourceCounts.HTTPRouteCount = int64(len(g.Routes))
	ngfResourceCounts.SecretCount = int64(len(g.ReferencedSecrets))
	ngfResourceCounts.ServiceCount = int64(len(g.ReferencedServices))

	for _, upstream := range cfg.Upstreams {
		if upstream.ErrorMsg == "" {
			ngfResourceCounts.EndpointCount += int64(len(upstream.Endpoints))
		}
	}

	return ngfResourceCounts, nil
}

func getPodReplicaSet(
	ctx context.Context,
	k8sClient client.Reader,
	podNSName types.NamespacedName,
) (*appsv1.ReplicaSet, error) {
	var pod v1.Pod
	if err := k8sClient.Get(
		ctx,
		types.NamespacedName{Namespace: podNSName.Namespace, Name: podNSName.Name},
		&pod,
	); err != nil {
		return nil, fmt.Errorf("failed to get NGF Pod: %w", err)
	}

	podOwnerRefs := pod.GetOwnerReferences()
	if len(podOwnerRefs) != 1 {
		return nil, fmt.Errorf("expected one owner reference of the NGF Pod, got %d", len(podOwnerRefs))
	}

	if podOwnerRefs[0].Kind != "ReplicaSet" {
		return nil, fmt.Errorf("expected pod owner reference to be ReplicaSet, got %s", podOwnerRefs[0].Kind)
	}

	var replicaSet appsv1.ReplicaSet
	if err := k8sClient.Get(
		ctx,
		types.NamespacedName{Namespace: podNSName.Namespace, Name: podOwnerRefs[0].Name},
		&replicaSet,
	); err != nil {
		return nil, fmt.Errorf("failed to get NGF Pod's ReplicaSet: %w", err)
	}

	return &replicaSet, nil
}

func getReplicas(replicaSet *appsv1.ReplicaSet) (int, error) {
	if replicaSet.Spec.Replicas == nil {
		return 0, errors.New("replica set replicas was nil")
	}

	return int(*replicaSet.Spec.Replicas), nil
}

func getDeploymentID(replicaSet *appsv1.ReplicaSet) (string, error) {
	replicaOwnerRefs := replicaSet.GetOwnerReferences()
	if len(replicaOwnerRefs) != 1 {
		return "", fmt.Errorf("expected one owner reference of the NGF ReplicaSet, got %d", len(replicaOwnerRefs))
	}

	if replicaOwnerRefs[0].Kind != "Deployment" {
		return "", fmt.Errorf("expected replicaSet owner reference to be Deployment, got %s", replicaOwnerRefs[0].Kind)
	}

	if replicaOwnerRefs[0].UID == "" {
		return "", fmt.Errorf("expected replicaSet owner reference to have a UID")
	}

	return string(replicaOwnerRefs[0].UID), nil
}

// CollectClusterID gets the UID of the kube-system namespace.
func CollectClusterID(ctx context.Context, k8sClient client.Reader) (string, error) {
	key := types.NamespacedName{
		Name: metav1.NamespaceSystem,
	}
	var kubeNamespace v1.Namespace
	if err := k8sClient.Get(ctx, key, &kubeNamespace); err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %w", err)
	}
	return string(kubeNamespace.GetUID()), nil
}

// CollectNodeList returns a NodeList of all the Nodes in the cluster.
func CollectNodeList(ctx context.Context, k8sClient client.Reader) (v1.NodeList, error) {
	var nodes v1.NodeList
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return nodes, fmt.Errorf("failed to get NodeList: %w", err)
	}

	return nodes, nil
}
