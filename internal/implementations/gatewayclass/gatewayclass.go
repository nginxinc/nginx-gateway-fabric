package implementation

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type gatewayClassImplementation struct {
	logger           logr.Logger
	gatewayClassName string
	eventCh          chan<- interface{}
}

func NewGatewayClassImplementation(conf config.Config, eventCh chan<- interface{}) sdk.GatewayClassImpl {
	return &gatewayClassImplementation{
		logger:           conf.Logger,
		gatewayClassName: conf.GatewayClassName,
		eventCh:          eventCh,
	}
}

func (impl *gatewayClassImplementation) Upsert(gc *v1alpha2.GatewayClass) {
	if gc.Name != impl.gatewayClassName {
		msg := fmt.Sprintf("GatewayClass was upserted but ignored because this controller only supports the GatewayClass %s", impl.gatewayClassName)
		impl.logger.Info(msg,
			"name", gc.Name,
		)
		return
	}

	impl.eventCh <- &events.UpsertEvent{
		Resource: gc,
	}

	impl.logger.Info("GatewayClass was upserted",
		"name", gc.Name)
}

func (impl *gatewayClassImplementation) Remove(nsname types.NamespacedName) {
	// GatewayClass is a cluster scoped resource - no namespace.

	if nsname.Name != impl.gatewayClassName {
		msg := fmt.Sprintf("GatewayClass was removed but ignored because this controller only supports the GatewayClass %s", impl.gatewayClassName)
		impl.logger.Info(msg,
			"name", nsname.Name,
		)
		return
	}

	impl.logger.Info("GatewayClass was removed",
		"name", nsname.Name)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &v1alpha2.GatewayClass{},
	}
}
