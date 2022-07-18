package implementation

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type gatewayImplementation struct {
	logger  logr.Logger
	eventCh chan<- interface{}
}

func NewGatewayImplementation(conf config.Config, eventCh chan<- interface{}) sdk.GatewayImpl {
	return &gatewayImplementation{
		logger:  conf.Logger,
		eventCh: eventCh,
	}
}

// FIXME(pleshakov) All Implementations (Gateway, HTTPRoute, ...) look similar. Consider writing a general-purpose
// component to implement all implementations. This will avoid the duplication code and tests.

func (impl *gatewayImplementation) Upsert(gw *v1alpha2.Gateway) {
	impl.logger.Info("Gateway was upserted",
		"namespace", gw.Namespace,
		"name", gw.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: gw,
	}
}

func (impl *gatewayImplementation) Remove(nsname types.NamespacedName) {
	impl.logger.Info("Gateway was removed",
		"namespace", nsname.Namespace,
		"name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &v1alpha2.Gateway{},
	}
}
