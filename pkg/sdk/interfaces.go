package sdk

import (
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
