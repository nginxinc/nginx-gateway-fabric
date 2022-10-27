package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ChangeProcessor

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
	Process(ctx context.Context) (changed bool, conf Configuration, statuses Statuses)
}

// ChangeProcessorConfig holds configuration parameters for ChangeProcessorImpl.
type ChangeProcessorConfig struct {
	// GatewayCtlrName is the name of the Gateway controller.
	GatewayCtlrName string
	// GatewayClassName is the name of the GatewayClass resource.
	GatewayClassName string
	// SecretMemoryManager is the secret memory manager.
	SecretMemoryManager SecretDiskMemoryManager
	// ServiceResolver resolves Services to Endpoints.
	ServiceResolver resolver.ServiceResolver
	// RelationshipCapturer captures relationships between Kubernetes API resources and Gateway API resources.
	RelationshipCapturer relationship.Capturer
	// Logger is the logger for this Change Processor.
	Logger logr.Logger
}

// ChangeProcessorImpl is an implementation of ChangeProcessor.
type ChangeProcessorImpl struct {
	store *store
	cfg   ChangeProcessorConfig

	// changed is true if any changes that were captured require an update to nginx.
	// It is true if the store changed, or if a Kubernetes resource (e.g.
	// Service, EndpointSlice) that is related to a Gateway API resource (e.g. Gateway, HTTPRoute) changed.
	// It is reset to false after Process is called.
	changed bool

	lock sync.Mutex
}

// NewChangeProcessorImpl creates a new ChangeProcessorImpl for the Gateway resource with the configured namespace name.
func NewChangeProcessorImpl(cfg ChangeProcessorConfig) *ChangeProcessorImpl {
	return &ChangeProcessorImpl{
		store: newStore(),
		cfg:   cfg,
	}
}

// FIXME(pleshakov)
// Currently, changes (upserts/delete) trigger rebuilding of the configuration, even if the change doesn't change
// the configuration or the statuses of the resources. For example, a change in a Gateway resource that doesn't
// belong to the NGINX Gateway or an HTTPRoute that doesn't belong to any of the Gateways of the NGINX Gateway.
// Find a way to ignore changes that don't affect the configuration and/or statuses of the resources.

func (c *ChangeProcessorImpl) CaptureUpsertChange(obj client.Object) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cfg.RelationshipCapturer.Capture(obj)

	switch o := obj.(type) {
	case *v1beta1.GatewayClass:
		c.store.captureGatewayClassChange(o, c.cfg.GatewayClassName)
	case *v1beta1.Gateway:
		c.store.captureGatewayChange(o)
	case *v1beta1.HTTPRoute:
		c.store.captureHTTPRouteChange(o)
	case *v1.Service:
		c.store.captureServiceChange(o)
	case *discoveryV1.EndpointSlice:
		break
	default:
		panic(fmt.Errorf("ChangeProcessor doesn't support %T", obj))
	}

	c.changed = c.changed || c.store.changed || c.cfg.RelationshipCapturer.Exists(obj, client.ObjectKeyFromObject(obj))
}

func (c *ChangeProcessorImpl) CaptureDeleteChange(
	resourceType client.Object,
	nsname types.NamespacedName,
) {
	c.lock.Lock()
	defer c.lock.Unlock()

	switch resourceType.(type) {
	case *v1beta1.GatewayClass:
		if nsname.Name != c.cfg.GatewayClassName {
			panic(fmt.Errorf("gatewayclass resource must be %s, got %s", c.cfg.GatewayClassName, nsname.Name))
		}
		c.store.gc = nil
		c.store.changed = true
	case *v1beta1.Gateway:
		delete(c.store.gateways, nsname)
		c.store.changed = true
	case *v1beta1.HTTPRoute:
		delete(c.store.httpRoutes, nsname)
		c.store.changed = true
	case *v1.Service:
		delete(c.store.services, nsname)
	case *discoveryV1.EndpointSlice:
		break
	default:
		panic(fmt.Errorf("ChangeProcessor doesn't support %T", resourceType))
	}

	c.changed = c.changed || c.store.changed || c.cfg.RelationshipCapturer.Exists(resourceType, nsname)

	c.cfg.RelationshipCapturer.Remove(resourceType, nsname)
}

func (c *ChangeProcessorImpl) Process(
	ctx context.Context,
) (changed bool, conf Configuration, statuses Statuses) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.changed {
		return false, conf, statuses
	}

	c.store.changed = false
	c.changed = false

	graph := buildGraph(
		c.store,
		c.cfg.GatewayCtlrName,
		c.cfg.GatewayClassName,
		c.cfg.SecretMemoryManager,
	)

	var warnings Warnings
	conf, warnings = buildConfiguration(ctx, graph, c.cfg.ServiceResolver)

	for obj, objWarnings := range warnings {
		for _, w := range objWarnings {
			// FIXME(pleshakov): report warnings via Object status
			c.cfg.Logger.Info("Got warning while building graph",
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
				"namespace", obj.GetNamespace(),
				"name", obj.GetName(),
				"warning", w)
		}
	}

	statuses = buildStatuses(graph)

	return true, conf, statuses
}
