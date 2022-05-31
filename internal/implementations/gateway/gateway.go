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

type gatewayImplementation struct {
	conf    config.Config
	eventCh chan<- interface{}
}

func NewGatewayImplementation(conf config.Config, eventCh chan<- interface{}) sdk.GatewayImpl {
	return &gatewayImplementation{
		conf:    conf,
		eventCh: eventCh,
	}
}

func (impl *gatewayImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl *gatewayImplementation) Upsert(gw *v1alpha2.Gateway) {
	if gw.Namespace != impl.conf.GatewayNsName.Namespace || gw.Name != impl.conf.GatewayNsName.Name {
		msg := fmt.Sprintf("Gateway was upserted but ignored because this controller only supports the Gateway %s", impl.conf.GatewayNsName)
		impl.Logger().Info(msg,
			"namespace", gw.Namespace,
			"name", gw.Name,
		)
		return
	}

	impl.Logger().Info("Gateway was upserted",
		"namespace", gw.Namespace,
		"name", gw.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: gw,
	}
}

func (impl *gatewayImplementation) Remove(nsname types.NamespacedName) {
	if nsname != impl.conf.GatewayNsName {
		msg := fmt.Sprintf("Gateway was removed but ignored because this controller only supports the Gateway %s", impl.conf.GatewayNsName)
		impl.Logger().Info(msg,
			"namespace", nsname.Namespace,
			"name", nsname.Name,
		)
		return
	}

	impl.Logger().Info("Gateway was removed",
		"namespace", nsname.Namespace,
		"name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &v1alpha2.Gateway{},
	}
}
