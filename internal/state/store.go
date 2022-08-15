package state

import (
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// store contains the resources that represent the state of the Gateway.
type store struct {
	gc         *v1alpha2.GatewayClass
	gateways   map[types.NamespacedName]*v1alpha2.Gateway
	httpRoutes map[types.NamespacedName]*v1alpha2.HTTPRoute

	// services maps services to the set of http routes that reference the service.
	services map[types.NamespacedName]map[types.NamespacedName]struct{}
	// endpointSlices is the set of endpoint slices that belong to the services we are tracking.
	endpointSlices map[types.NamespacedName]*discoveryV1.EndpointSlice
}

func newStore() *store {
	return &store{
		gateways:       make(map[types.NamespacedName]*v1alpha2.Gateway),
		httpRoutes:     make(map[types.NamespacedName]*v1alpha2.HTTPRoute),
		services:       make(map[types.NamespacedName]map[types.NamespacedName]struct{}),
		endpointSlices: make(map[types.NamespacedName]*discoveryV1.EndpointSlice),
	}
}
