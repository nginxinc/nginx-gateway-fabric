package sdk

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	nginxgwv1alpha1 "github.com/nginxinc/nginx-kubernetes-gateway/pkg/apis/gateway/v1alpha1"
)

type GatewayClassImpl interface {
	Upsert(gc *v1beta1.GatewayClass)
	Remove(nsname types.NamespacedName)
}

type GatewayImpl interface {
	Upsert(*v1beta1.Gateway)
	Remove(types.NamespacedName)
}

type GatewayConfigImpl interface {
	Upsert(config *nginxgwv1alpha1.GatewayConfig)
	Remove(string)
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
