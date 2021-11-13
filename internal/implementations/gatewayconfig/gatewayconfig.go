package gatewayconfig

import (
	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"

	nginxgwv1alpha1 "github.com/nginxinc/nginx-gateway-kubernetes/pkg/apis/v1alpha1"
)

type gatewayConfigImplementation struct {
	conf config.Config
}

func NewGatewayConfigImplementation(conf config.Config) sdk.GatewayConfigImpl {
	return &gatewayConfigImplementation{
		conf: conf,
	}
}

func (impl *gatewayConfigImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl *gatewayConfigImplementation) Upsert(gcfg *nginxgwv1alpha1.GatewayConfig) {
	impl.Logger().Info("Processing GatewayConfig",
		"name", gcfg.Name,
	)

	if gcfg.Spec.Worker != nil && gcfg.Spec.Worker.Processes != nil {
		impl.Logger().Info("Worker config",
			"processes", gcfg.Spec.Worker.Processes)
	}

	if gcfg.Spec.HTTP != nil {
		for _, l := range gcfg.Spec.HTTP.AccessLogs {
			impl.Logger().Info("AccessLog config",
				"format", l.Format,
				"destination", l.Destination)
		}
	}
}

func (impl *gatewayConfigImplementation) Remove(name string) {
	impl.Logger().Info("Removing GatewayConfig",
		"name", name,
	)
}
