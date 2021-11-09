package implementation

import (
	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type gatewayClassImplementation struct {
	logger logr.Logger
}

func NewGatewayClassImplementation(logger logr.Logger) sdk.GatewayClassImpl {
	return &gatewayClassImplementation{logger: logger}
}

func (impl *gatewayClassImplementation) Upsert(gc *v1alpha2.GatewayClass) {
	if gc.Spec.ControllerName != "k8s-gateway.nginx.org/gateway" {
		impl.logger.Info("Wrong ControllerName in the GatewayClass resource",
			"expected", "k8s-gateway.nginx.org/gateway",
			"got", "gc.Spec.ControllerName")
		return
	}

	impl.logger.Info("Processing GatewayClass resource",
		"name", gc.Name)
}
func (impl *gatewayClassImplementation) Remove(key string) {
	impl.logger.Info("GatewayClass resource was removed",
		"name", key)
}
