package sdk

import (
	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/v1alpha1"
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
