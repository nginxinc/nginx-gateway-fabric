package state

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// store contains the resources that represent the state of the Gateway.
type store struct {
	gc         *v1beta1.GatewayClass
	gateways   map[types.NamespacedName]*v1beta1.Gateway
	httpRoutes map[types.NamespacedName]*v1beta1.HTTPRoute
	services   map[types.NamespacedName]*v1.Service

	// changed tells if the store is changed.
	// The store is considered changed if:
	// (1) Any of its resources was deleted.
	// (2) A new resource was upserted.
	// (3) An existing resource with the updated Generation was upserted.
	changed bool
}

func newStore() *store {
	return &store{
		gateways:   make(map[types.NamespacedName]*v1beta1.Gateway),
		httpRoutes: make(map[types.NamespacedName]*v1beta1.HTTPRoute),
		services:   make(map[types.NamespacedName]*v1.Service),
	}
}

func (s *store) captureGatewayClassChange(gc *v1beta1.GatewayClass, gwClassName string) {
	resourceChanged := true

	if gc.Name != gwClassName {
		panic(fmt.Errorf("gatewayclass resource must be %s, got %s", gwClassName, gc.Name))
	}

	// if the resource spec hasn't changed (its generation is the same), ignore the upsert
	if s.gc != nil && s.gc.Generation == gc.Generation {
		resourceChanged = false
	}

	s.gc = gc

	s.changed = s.changed || resourceChanged
}

func (s *store) captureGatewayChange(gw *v1beta1.Gateway) {
	resourceChanged := true

	// if the resource spec hasn't changed (its generation is the same), ignore the upsert
	prev, exist := s.gateways[client.ObjectKeyFromObject(gw)]
	if exist && gw.Generation == prev.Generation {
		resourceChanged = false
	}

	s.gateways[client.ObjectKeyFromObject(gw)] = gw

	s.changed = s.changed || resourceChanged
}

func (s *store) captureHTTPRouteChange(hr *v1beta1.HTTPRoute) {
	resourceChanged := true
	// if the resource spec hasn't changed (its generation is the same), ignore the upsert
	prev, exist := s.httpRoutes[client.ObjectKeyFromObject(hr)]
	if exist && hr.Generation == prev.Generation {
		resourceChanged = false
	}
	s.httpRoutes[client.ObjectKeyFromObject(hr)] = hr

	s.changed = s.changed || resourceChanged
}

// Service changes are treated differently than Gateway API resource changes in the following ways:
// (1) We don't check generation here because services do not use generation, and Service Controller filters upsert
// events based on the Service ports. This means we will only receive upsert events for Services with port changes.
// (2) We don't set the store's changed value to true because we don't want to trigger a reload on every Service change.
// We will rely on the relationship.Capturer to trigger a reload when necessary.
func (s *store) captureServiceChange(svc *v1.Service) {
	s.services[client.ObjectKeyFromObject(svc)] = svc
}
