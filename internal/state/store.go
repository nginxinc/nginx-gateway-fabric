package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// store contains the resources that represent the state of the Gateway.
type store struct {
	gc         *v1alpha2.GatewayClass
	gateways   map[types.NamespacedName]*v1alpha2.Gateway
	httpRoutes map[types.NamespacedName]*v1alpha2.HTTPRoute
}

func newStore() *store {
	return &store{
		gateways:   make(map[types.NamespacedName]*v1alpha2.Gateway),
		httpRoutes: make(map[types.NamespacedName]*v1alpha2.HTTPRoute),
	}
}
