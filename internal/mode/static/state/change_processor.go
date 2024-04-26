package state

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/gatewayclass"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// ChangeType is the type of change that occurred based on a k8s object event.
type ChangeType int

const (
	// NoChange means that nothing changed.
	NoChange ChangeType = iota
	// EndpointsOnlyChange means that only the endpoints changed.
	// If using NGINX Plus, this update can be done using the API without a reload.
	EndpointsOnlyChange
	// ClusterStateChange means that something other than endpoints changed. This requires an NGINX reload.
	ClusterStateChange
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ChangeProcessor

type extractGVKFunc func(obj client.Object) schema.GroupVersionKind

// ChangeProcessor processes the changes to resources and produces a graph-like representation
// of the Gateway configuration. It only supports one GatewayClass resource.
type ChangeProcessor interface {
	// CaptureUpsertChange captures an upsert change to a resource.
	// It panics if the resource is of unsupported type or if the passed Gateway is different from the one this
	// ChangeProcessor was created for.
	CaptureUpsertChange(obj client.Object)
	// CaptureDeleteChange captures a delete change to a resource.
	// The method panics if the resource is of unsupported type or if the passed Gateway is different from the one
	// this ChangeProcessor was created for.
	CaptureDeleteChange(resourceType client.Object, nsname types.NamespacedName)
	// Process produces a graph-like representation of GatewayAPI resources.
	// If no changes were captured, the changed return argument will be NoChange and graph will be empty.
	Process() (changeType ChangeType, graphCfg *graph.Graph)
	// GetLatestGraph returns the latest Graph.
	GetLatestGraph() *graph.Graph
}

// ChangeProcessorConfig holds configuration parameters for ChangeProcessorImpl.
type ChangeProcessorConfig struct {
	// Validators validate resources according to data-plane specific rules.
	Validators validation.Validators
	// EventRecorder records events for Kubernetes resources.
	EventRecorder record.EventRecorder
	// Scheme is the Kubernetes scheme.
	Scheme *runtime.Scheme
	// ProtectedPorts are the ports that may not be configured by a listener with a descriptive name of the ports.
	ProtectedPorts graph.ProtectedPorts
	// Logger is the logger for this Change Processor.
	Logger logr.Logger
	// GatewayCtlrName is the name of the Gateway controller.
	GatewayCtlrName string
	// GatewayClassName is the name of the GatewayClass resource.
	GatewayClassName string
}

// ChangeProcessorImpl is an implementation of ChangeProcessor.
type ChangeProcessorImpl struct {
	latestGraph *graph.Graph

	// clusterState holds the current state of the cluster
	clusterState graph.ClusterState
	// updater acts upon the cluster state.
	updater Updater
	// getAndResetClusterStateChanged tells if and how the cluster state has changed.
	getAndResetClusterStateChanged func() ChangeType

	cfg  ChangeProcessorConfig
	lock sync.Mutex
}

// NewChangeProcessorImpl creates a new ChangeProcessorImpl for the Gateway resource with the configured namespace name.
func NewChangeProcessorImpl(cfg ChangeProcessorConfig) *ChangeProcessorImpl {
	clusterStore := graph.ClusterState{
		GatewayClasses:     make(map[types.NamespacedName]*v1.GatewayClass),
		Gateways:           make(map[types.NamespacedName]*v1.Gateway),
		HTTPRoutes:         make(map[types.NamespacedName]*v1.HTTPRoute),
		Services:           make(map[types.NamespacedName]*apiv1.Service),
		Namespaces:         make(map[types.NamespacedName]*apiv1.Namespace),
		ReferenceGrants:    make(map[types.NamespacedName]*v1beta1.ReferenceGrant),
		Secrets:            make(map[types.NamespacedName]*apiv1.Secret),
		CRDMetadata:        make(map[types.NamespacedName]*metav1.PartialObjectMetadata),
		BackendTLSPolicies: make(map[types.NamespacedName]*v1alpha2.BackendTLSPolicy),
		ConfigMaps:         make(map[types.NamespacedName]*apiv1.ConfigMap),
		NginxProxies:       make(map[types.NamespacedName]*ngfAPI.NginxProxy),
	}

	extractGVK := func(obj client.Object) schema.GroupVersionKind {
		gvk, err := apiutil.GVKForObject(obj, cfg.Scheme)
		if err != nil {
			panic(fmt.Errorf("failed to get GVK for object %T: %w", obj, err))
		}
		return gvk
	}

	processor := &ChangeProcessorImpl{
		cfg:          cfg,
		clusterState: clusterStore,
	}

	isReferenced := func(obj client.Object, nsname types.NamespacedName) bool {
		return processor.latestGraph != nil && processor.latestGraph.IsReferenced(obj, nsname)
	}

	trackingUpdater := newChangeTrackingUpdater(
		extractGVK,
		[]changeTrackingUpdaterObjectTypeCfg{
			{
				gvk:       extractGVK(&v1.GatewayClass{}),
				store:     newObjectStoreMapAdapter(clusterStore.GatewayClasses),
				predicate: nil,
			},
			{
				gvk:       extractGVK(&v1.Gateway{}),
				store:     newObjectStoreMapAdapter(clusterStore.Gateways),
				predicate: nil,
			},
			{
				gvk:       extractGVK(&v1.HTTPRoute{}),
				store:     newObjectStoreMapAdapter(clusterStore.HTTPRoutes),
				predicate: nil,
			},
			{
				gvk:       extractGVK(&v1beta1.ReferenceGrant{}),
				store:     newObjectStoreMapAdapter(clusterStore.ReferenceGrants),
				predicate: nil,
			},
			{
				gvk:       extractGVK(&v1alpha2.BackendTLSPolicy{}),
				store:     newObjectStoreMapAdapter(clusterStore.BackendTLSPolicies),
				predicate: nil,
			},
			{
				gvk:       extractGVK(&apiv1.Namespace{}),
				store:     newObjectStoreMapAdapter(clusterStore.Namespaces),
				predicate: funcPredicate{stateChanged: isReferenced},
			},
			{
				gvk:       extractGVK(&apiv1.Service{}),
				store:     newObjectStoreMapAdapter(clusterStore.Services),
				predicate: funcPredicate{stateChanged: isReferenced},
			},
			{
				gvk:       extractGVK(&discoveryV1.EndpointSlice{}),
				store:     nil,
				predicate: funcPredicate{stateChanged: isReferenced},
			},
			{
				gvk:       extractGVK(&apiv1.Secret{}),
				store:     newObjectStoreMapAdapter(clusterStore.Secrets),
				predicate: funcPredicate{stateChanged: isReferenced},
			},
			{
				gvk:       extractGVK(&apiv1.ConfigMap{}),
				store:     newObjectStoreMapAdapter(clusterStore.ConfigMaps),
				predicate: funcPredicate{stateChanged: isReferenced},
			},
			{
				gvk:       extractGVK(&apiext.CustomResourceDefinition{}),
				store:     newObjectStoreMapAdapter(clusterStore.CRDMetadata),
				predicate: annotationChangedPredicate{annotation: gatewayclass.BundleVersionAnnotation},
			},
			{
				gvk:       extractGVK(&ngfAPI.NginxProxy{}),
				store:     newObjectStoreMapAdapter(clusterStore.NginxProxies),
				predicate: funcPredicate{stateChanged: isReferenced},
			},
		},
	)

	processor.getAndResetClusterStateChanged = trackingUpdater.getAndResetChangedStatus
	processor.updater = trackingUpdater

	return processor
}

// Currently, changes (upserts/delete) trigger rebuilding of the configuration, even if the change doesn't change
// the configuration or the statuses of the resources. For example, a change in a Gateway resource that doesn't
// belong to the NGINX Gateway Fabric or an HTTPRoute that doesn't belong to any of the Gateways of the
// NGINX Gateway Fabric. Find a way to ignore changes that don't affect the configuration and/or statuses of
// the resources.
// Tracking issues: https://github.com/nginxinc/nginx-gateway-fabric/issues/1123,
// https://github.com/nginxinc/nginx-gateway-fabric/issues/1124,
// https://github.com/nginxinc/nginx-gateway-fabric/issues/1577

// FIXME(pleshakov)
// Remove CaptureUpsertChange() and CaptureDeleteChange() from ChangeProcessor and pass all changes directly to
// Process() instead. As a result, the clients will only need to call Process(), which will simplify them.
// Now the clients make a combination of CaptureUpsertChange() and CaptureDeleteChange() calls followed by a call to
// Process().

func (c *ChangeProcessorImpl) CaptureUpsertChange(obj client.Object) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.updater.Upsert(obj)
}

func (c *ChangeProcessorImpl) CaptureDeleteChange(resourceType client.Object, nsname types.NamespacedName) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.updater.Delete(resourceType, nsname)
}

func (c *ChangeProcessorImpl) Process() (ChangeType, *graph.Graph) {
	c.lock.Lock()
	defer c.lock.Unlock()

	changeType := c.getAndResetClusterStateChanged()
	if changeType == NoChange {
		return NoChange, nil
	}

	c.latestGraph = graph.BuildGraph(
		c.clusterState,
		c.cfg.GatewayCtlrName,
		c.cfg.GatewayClassName,
		c.cfg.Validators,
		c.cfg.ProtectedPorts,
	)

	return changeType, c.latestGraph
}

func (c *ChangeProcessorImpl) GetLatestGraph() *graph.Graph {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.latestGraph
}
