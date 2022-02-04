package sdk

import (
	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/gateway/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type GatewayClassImpl interface {
	Upsert(gc *v1alpha2.GatewayClass)
	Remove(key string)
}

type GatewayImpl interface {
	Upsert(*v1alpha2.Gateway)
	Remove(string)
}

type GatewayConfigImpl interface {
	Upsert(config *nginxgwv1alpha1.GatewayConfig)
	Remove(string)
}

type HTTPRouteImpl interface {
	Upsert(config *v1alpha2.HTTPRoute)
	// TO-DO: change other interfaces to use types.NamespacedName
	Remove(types.NamespacedName)
}
