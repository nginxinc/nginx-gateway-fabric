package implementation

import (
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/sdk"
	"go.uber.org/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type gatewayClassImplementation struct {
	logger *zap.SugaredLogger
}

func NewGatewayClassImplementation(logger *zap.SugaredLogger) sdk.GatewayClassImpl {
	return &gatewayClassImplementation{logger: logger}
}

func (impl *gatewayClassImplementation) Upsert(gc *v1alpha2.GatewayClass) {
	if gc.Spec.ControllerName != "k8s-gateway.nginx.org/gateway" {
		impl.logger.Errorw("Wrong ControllerName in the GatewayClass resource",
			"expected", "k8s-gateway.nginx.org/gateway",
			"got", "gc.Spec.ControllerName")
		return
	}

	impl.logger.Infow("Processing GatewayClass resource",
		"name", gc.Name)
}
func (impl *gatewayClassImplementation) Remove(key string) {
	impl.logger.Errorw("GatewayClass resource was removed",
		"name", key)
}
