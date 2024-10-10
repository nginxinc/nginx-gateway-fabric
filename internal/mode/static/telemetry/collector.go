package telemetry

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"

	tel "github.com/nginxinc/telemetry-exporter/pkg/telemetry"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sversion "k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

//counterfeiter:generate . GraphGetter

// GraphGetter gets the latest Graph.
type GraphGetter interface {
	GetLatestGraph() *graph.Graph
}

//counterfeiter:generate . ConfigurationGetter

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
	FlagValues []string
	// SnippetsFiltersContextDirectives contains the context-directive strings of all applied SnippetsFilters.
	// Both lists are ordered first by count, then by lexicographical order on the context-directive string.
	SnippetsFiltersContextDirectives []string
	// SnippetsFiltersContextDirectivesCount contains the count of the context-directive strings, where each count
	// corresponds to the string from SnippetsFiltersContextDirectives at the same index. Both lists are ordered
	// first by count, then by lexicographical order on the context-directive string.
	SnippetsFiltersContextDirectivesCount []int64
	NGFResourceCounts                     // embedding is required by the generator.
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
	// TLSRouteCount is the number of relevant TLSRoutes.
	TLSRouteCount int64
	// SecretCount is the number of relevant Secrets.
	SecretCount int64
	// ServiceCount is the number of relevant Services.
	ServiceCount int64
	// EndpointCount include the total count of Endpoints(IP:port) across all referenced services.
	EndpointCount int64
	// GRPCRouteCount is the number of relevant GRPCRoutes.
	GRPCRouteCount int64
	// BackendTLSPolicyCount is the number of relevant BackendTLSPolicies.
	BackendTLSPolicyCount int64
	// GatewayAttachedClientSettingsPolicyCount is the number of relevant ClientSettingsPolicies
	// attached at the Gateway level.
	GatewayAttachedClientSettingsPolicyCount int64
	// RouteAttachedClientSettingsPolicyCount is the number of relevant ClientSettingsPolicies attached at the Route level.
	RouteAttachedClientSettingsPolicyCount int64
	// ObservabilityPolicyCount is the number of relevant ObservabilityPolicies.
	ObservabilityPolicyCount int64
	// NginxProxyCount is the number of NginxProxies.
	NginxProxyCount int64
	// SnippetsFilterCount is the number of SnippetsFilters.
	SnippetsFilterCount int64
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

// Collect collects and returns telemetry Data.
func (c DataCollectorImpl) Collect(ctx context.Context) (Data, error) {
	clusterInfo, err := collectClusterInformation(ctx, c.cfg.K8sClientReader)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect cluster information: %w", err)
	}

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

	snippetsFiltersContextDirectives,
		snippetsFiltersContextDirectivesCount,
		err := collectSnippetsFilterSnippetsInfo(c.cfg.GraphGetter)
	if err != nil {
		return Data{}, fmt.Errorf("failed to collect snippet filter directive info: %w", err)
	}

	data := Data{
		Data: tel.Data{
			ProjectName:         "NGF",
			ProjectVersion:      c.cfg.Version,
			ProjectArchitecture: runtime.GOARCH,
			ClusterID:           clusterInfo.ClusterID,
			ClusterVersion:      clusterInfo.Version,
			ClusterPlatform:     clusterInfo.Platform,
			InstallationID:      deploymentID,
			ClusterNodeCount:    int64(clusterInfo.NodeCount),
		},
		NGFResourceCounts: graphResourceCount,
		ImageSource:       c.cfg.ImageSource,
		FlagNames:         c.cfg.Flags.Names,
		FlagValues:        c.cfg.Flags.Values,
		NGFReplicaCount:   int64(replicaCount),
		// maybe SnippetValues?
		SnippetsFiltersContextDirectives:      snippetsFiltersContextDirectives,
		SnippetsFiltersContextDirectivesCount: snippetsFiltersContextDirectivesCount,
	}

	return data, nil
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

	routeCounts := computeRouteCount(g.Routes, g.L4Routes)
	ngfResourceCounts.HTTPRouteCount = routeCounts.HTTPRouteCount
	ngfResourceCounts.GRPCRouteCount = routeCounts.GRPCRouteCount
	ngfResourceCounts.TLSRouteCount = routeCounts.TLSRouteCount

	ngfResourceCounts.SecretCount = int64(len(g.ReferencedSecrets))
	ngfResourceCounts.ServiceCount = int64(len(g.ReferencedServices))

	for _, upstream := range cfg.Upstreams {
		if upstream.ErrorMsg == "" {
			ngfResourceCounts.EndpointCount += int64(len(upstream.Endpoints))
		}
	}

	ngfResourceCounts.BackendTLSPolicyCount = int64(len(g.BackendTLSPolicies))

	for policyKey, policy := range g.NGFPolicies {
		switch policyKey.GVK.Kind {
		case kinds.ClientSettingsPolicy:
			if len(policy.TargetRefs) == 0 {
				continue
			}

			if policy.TargetRefs[0].Kind == kinds.Gateway {
				ngfResourceCounts.GatewayAttachedClientSettingsPolicyCount++
			} else {
				ngfResourceCounts.RouteAttachedClientSettingsPolicyCount++
			}
		case kinds.ObservabilityPolicy:
			ngfResourceCounts.ObservabilityPolicyCount++
		}
	}

	if g.NginxProxy != nil {
		ngfResourceCounts.NginxProxyCount = 1
	}

	ngfResourceCounts.SnippetsFilterCount = int64(len(g.SnippetsFilters))

	return ngfResourceCounts, nil
}

type RouteCounts struct {
	HTTPRouteCount int64
	GRPCRouteCount int64
	TLSRouteCount  int64
}

func computeRouteCount(
	routes map[graph.RouteKey]*graph.L7Route,
	l4routes map[graph.L4RouteKey]*graph.L4Route,
) RouteCounts {
	httpRouteCount := int64(0)
	grpcRouteCount := int64(0)

	for _, r := range routes {
		if r.RouteType == graph.RouteTypeHTTP {
			httpRouteCount++
		}
		if r.RouteType == graph.RouteTypeGRPC {
			grpcRouteCount++
		}
	}

	return RouteCounts{
		HTTPRouteCount: httpRouteCount,
		GRPCRouteCount: grpcRouteCount,
		TLSRouteCount:  int64(len(l4routes)),
	}
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

type clusterInformation struct {
	Platform  string
	Version   string
	ClusterID string
	NodeCount int
}

func collectClusterInformation(ctx context.Context, k8sClient client.Reader) (clusterInformation, error) {
	var clusterInfo clusterInformation

	var nodes v1.NodeList
	if err := k8sClient.List(ctx, &nodes); err != nil {
		return clusterInformation{}, fmt.Errorf("failed to get NodeList: %w", err)
	}

	nodeCount := len(nodes.Items)
	if nodeCount == 0 {
		return clusterInformation{}, errors.New("failed to collect cluster information: NodeList length is zero")
	}
	clusterInfo.NodeCount = nodeCount

	node := nodes.Items[0]

	kubeletVersion := node.Status.NodeInfo.KubeletVersion
	version, err := k8sversion.ParseGeneric(kubeletVersion)
	if err != nil {
		clusterInfo.Version = "unknown"
	} else {
		clusterInfo.Version = version.String()
	}

	var namespaces v1.NamespaceList
	if err = k8sClient.List(ctx, &namespaces); err != nil {
		return clusterInformation{}, fmt.Errorf("failed to collect cluster information: %w", err)
	}

	clusterInfo.Platform = getPlatform(node, namespaces)

	var clusterID string
	clusterID, err = CollectClusterID(ctx, k8sClient)
	if err != nil {
		return clusterInformation{}, fmt.Errorf("failed to collect cluster information: %w", err)
	}
	clusterInfo.ClusterID = clusterID

	return clusterInfo, nil
}

type sfContextDirective struct {
	context   string
	directive string
}

func collectSnippetsFilterSnippetsInfo(graphGetter GraphGetter) ([]string, []int64, error) {
	g := graphGetter.GetLatestGraph()
	if g == nil {
		return nil, nil, errors.New("latest graph cannot be nil")
	}

	contextDirectiveMap := make(map[sfContextDirective]int)

	for name := range g.SnippetsFilters {
		sf := g.SnippetsFilters[name]
		if sf == nil {
			continue
		}

		for nginxContext := range sf.Snippets {
			snippetValue := sf.Snippets[nginxContext]

			var parsedContext string
			switch nginxContext {
			case ngfAPI.NginxContextMain:
				parsedContext = "main"
			case ngfAPI.NginxContextHTTP:
				parsedContext = "http"
			case ngfAPI.NginxContextHTTPServer:
				parsedContext = "server"
			case ngfAPI.NginxContextHTTPServerLocation:
				parsedContext = "location"
			default:
				parsedContext = "unknown"
			}

			directives := parseSnippetValueIntoDirectives(snippetValue)

			for _, directive := range directives {
				contextDirective := sfContextDirective{
					context:   parsedContext,
					directive: directive,
				}

				contextDirectiveMap[contextDirective]++
			}
		}
	}

	contextDirectiveList, countList := parseContextDirectiveMapIntoLists(contextDirectiveMap)

	return contextDirectiveList, countList, nil
}

func parseSnippetValueIntoDirectives(snippetValue string) []string {
	separatedDirectives := strings.Split(snippetValue, ";")
	directives := make([]string, 0, len(separatedDirectives))

	for _, directive := range separatedDirectives {
		// the strings.TrimSpace is needed in the case of multi-line NGINX Snippet values
		directive = strings.Split(strings.TrimSpace(directive), " ")[0]

		// splitting on the delimiting character can result in a directive being empty or a space/newline character,
		// so we check here to ensure it's not
		if directive != "" {
			directives = append(directives, directive)
		}
	}

	return directives
}

// parseContextDirectiveMapIntoLists returns two same-length lists where the elements at each corresponding index
// are paired together.
// The first list contains strings which are the NGINX context and directive of a Snippet joined with a hyphen.
// The second list contains ints which are the count of total same context-directive values of the first list.
// Both lists are ordered based off of count first, then lexicographically on the context-directive string.
func parseContextDirectiveMapIntoLists(contextDirectiveMap map[sfContextDirective]int) ([]string, []int64) {
	type sfContextDirectiveCount struct {
		contextDirective string
		count            int64
	}

	kvPairs := make([]sfContextDirectiveCount, 0, len(contextDirectiveMap))

	for k, v := range contextDirectiveMap {
		kvPairs = append(kvPairs, sfContextDirectiveCount{k.context + "-" + k.directive, int64(v)})
	}

	sort.Slice(kvPairs, func(i, j int) bool {
		if kvPairs[i].count == kvPairs[j].count {
			return kvPairs[i].contextDirective < kvPairs[j].contextDirective
		}
		return kvPairs[i].count > kvPairs[j].count
	})

	contextDirectiveList := make([]string, len(kvPairs))
	countList := make([]int64, len(kvPairs))

	for i, pair := range kvPairs {
		contextDirectiveList[i] = pair.contextDirective
		countList[i] = pair.count
	}

	return contextDirectiveList, countList
}
