package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	gwapivalidation "sigs.k8s.io/gateway-api/apis/v1beta1/validation"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
)

const (
	webhookValidationErrorLogMsg = "the resource failed webhook validation, however the Gateway API webhook " +
		"failed to reject it with the error; make sure the webhook is installed and running correctly"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ChangeProcessor

type extractGVKFunc func(obj client.Object) schema.GroupVersionKind

// ChangeProcessor processes the changes to resources producing the internal representation
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
	// Process processes any captured changes and produces an internal representation of the Gateway configuration and
	// the status information about the processed resources.
	// If no changes were captured, the changed return argument will be false and both the configuration and statuses
	// will be empty.
	Process(ctx context.Context) (changed bool, conf dataplane.Configuration, statuses Statuses)
}

// ChangeProcessorConfig holds configuration parameters for ChangeProcessorImpl.
type ChangeProcessorConfig struct {
	// SecretMemoryManager is the secret memory manager.
	SecretMemoryManager secrets.SecretDiskMemoryManager
	// ServiceResolver resolves Services to Endpoints.
	ServiceResolver resolver.ServiceResolver
	// RelationshipCapturer captures relationships between Kubernetes API resources and Gateway API resources.
	RelationshipCapturer relationship.Capturer
	// Validators validate resources according to data-plane specific rules.
	Validators validation.Validators
	// Logger is the logger for this Change Processor.
	Logger logr.Logger
	// EventRecorder records events for Kubernetes resources.
	EventRecorder record.EventRecorder
	// Scheme is the a Kubernetes scheme.
	Scheme *runtime.Scheme
	// GatewayCtlrName is the name of the Gateway controller.
	GatewayCtlrName string
	// GatewayClassName is the name of the GatewayClass resource.
	GatewayClassName string
}

// ChangeProcessorImpl is an implementation of ChangeProcessor.
type ChangeProcessorImpl struct {
	// clusterState holds the current state of the cluster
	clusterState graph.ClusterState
	// updater acts upon the cluster state.
	updater Updater
	// getAndResetClusterStateChanged tells if the cluster state has changed.
	getAndResetClusterStateChanged func() bool

	cfg ChangeProcessorConfig

	lock sync.Mutex
}

// NewChangeProcessorImpl creates a new ChangeProcessorImpl for the Gateway resource with the configured namespace name.
func NewChangeProcessorImpl(cfg ChangeProcessorConfig) *ChangeProcessorImpl {
	clusterStore := graph.ClusterState{
		GatewayClasses: make(map[types.NamespacedName]*v1beta1.GatewayClass),
		Gateways:       make(map[types.NamespacedName]*v1beta1.Gateway),
		HTTPRoutes:     make(map[types.NamespacedName]*v1beta1.HTTPRoute),
		Services:       make(map[types.NamespacedName]*apiv1.Service),
	}

	extractGVK := func(obj client.Object) schema.GroupVersionKind {
		gvk, err := apiutil.GVKForObject(obj, cfg.Scheme)
		if err != nil {
			panic(fmt.Errorf("failed to get GVK for object %T: %w", obj, err))
		}
		return gvk
	}

	trackingUpdater := newChangeTrackingUpdater(
		cfg.RelationshipCapturer,
		extractGVK,
		[]changeTrackingUpdaterObjectTypeCfg{
			{
				gvk:               extractGVK(&v1beta1.GatewayClass{}),
				store:             newObjectStoreMapAdapter(clusterStore.GatewayClasses),
				trackUpsertDelete: true,
			},
			{
				gvk:               extractGVK(&v1beta1.Gateway{}),
				store:             newObjectStoreMapAdapter(clusterStore.Gateways),
				trackUpsertDelete: true,
			},
			{
				gvk:               extractGVK(&v1beta1.HTTPRoute{}),
				store:             newObjectStoreMapAdapter(clusterStore.HTTPRoutes),
				trackUpsertDelete: true,
			},
			{
				gvk:               extractGVK(&apiv1.Service{}),
				store:             newObjectStoreMapAdapter(clusterStore.Services),
				trackUpsertDelete: false,
			},
			{
				gvk:               extractGVK(&discoveryV1.EndpointSlice{}),
				store:             nil,
				trackUpsertDelete: false,
			},
		},
	)

	updater := newValidatingUpsertUpdater(
		trackingUpdater,
		cfg.EventRecorder,
		func(obj client.Object) error {
			// Add the validation for Gateway API resources which the webhook validates

			var err error
			switch o := obj.(type) {
			// We don't validate GatewayClass, because as of 0.7.0, the webhook doesn't validate it (it only
			// validates an update that requires the previous version of the resource,
			// which NKG cannot reliably provide - for example, after NKG restarts).
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.7.0/apis/v1beta1/validation/gatewayclass.go#L28
			case *v1beta1.Gateway:
				err = gwapivalidation.ValidateGateway(o).ToAggregate()
			case *v1beta1.HTTPRoute:
				err = gwapivalidation.ValidateHTTPRoute(o).ToAggregate()
			}

			if err != nil {
				return fmt.Errorf(webhookValidationErrorLogMsg+"; validation error: %w", err)
			}

			return nil
		},
	)

	return &ChangeProcessorImpl{
		cfg:                            cfg,
		getAndResetClusterStateChanged: trackingUpdater.getAndResetChangedStatus,
		updater:                        updater,
		clusterState:                   clusterStore,
	}
}

// Currently, changes (upserts/delete) trigger rebuilding of the configuration, even if the change doesn't change
// the configuration or the statuses of the resources. For example, a change in a Gateway resource that doesn't
// belong to the NGINX Gateway or an HTTPRoute that doesn't belong to any of the Gateways of the NGINX Gateway.
// Find a way to ignore changes that don't affect the configuration and/or statuses of the resources.

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

func (c *ChangeProcessorImpl) Process(
	ctx context.Context,
) (changed bool, conf dataplane.Configuration, statuses Statuses) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.getAndResetClusterStateChanged() {
		return false, conf, statuses
	}

	g := graph.BuildGraph(
		c.clusterState,
		c.cfg.GatewayCtlrName,
		c.cfg.GatewayClassName,
		c.cfg.SecretMemoryManager,
		c.cfg.Validators,
	)

	conf = dataplane.BuildConfiguration(ctx, g, c.cfg.ServiceResolver)
	statuses = buildStatuses(g)

	return true, conf, statuses
}
