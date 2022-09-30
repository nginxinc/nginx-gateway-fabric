package sdk

import (
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

type GatewayClassImpl interface {
	Upsert(gc *v1beta1.GatewayClass)
	Remove(nsname types.NamespacedName)
}

type GatewayImpl interface {
	Upsert(*v1beta1.Gateway)
	Remove(types.NamespacedName)
}

type HTTPRouteImpl interface {
	Upsert(config *v1beta1.HTTPRoute)
	// FIXME(pleshakov): change other interfaces to use types.NamespacedName
	Remove(types.NamespacedName)
}

type ServiceImpl interface {
	Upsert(svc *apiv1.Service)
	Remove(nsname types.NamespacedName)
}

type SecretImpl interface {
	Upsert(secret *apiv1.Secret)
	Remove(name types.NamespacedName)
}

type EndpointSliceImpl interface {
	Upsert(endpSlice *v1.EndpointSlice)
	Remove(nsname types.NamespacedName)
}
