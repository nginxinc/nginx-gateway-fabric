package implementation

import (
	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/config"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	"github.com/nginxinc/nginx-gateway-kubernetes/pkg/sdk"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type httpRouteImplementation struct {
	conf    config.Config
	eventCh chan<- interface{}
}

// NewHTTPRouteImplementation creates a new HTTPRouteImplementation.
func NewHTTPRouteImplementation(cfg config.Config, eventCh chan<- interface{}) sdk.HTTPRouteImpl {
	return &httpRouteImplementation{
		conf:    cfg,
		eventCh: eventCh,
	}
}

func (impl *httpRouteImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl *httpRouteImplementation) ControllerName() string {
	return impl.conf.GatewayCtlrName
}

func (impl *httpRouteImplementation) Upsert(hr *v1alpha2.HTTPRoute) {
	impl.Logger().Info("HTTPRoute was upserted",
		"namespace", hr.Namespace, "name", hr.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: hr,
	}
}

func (impl *httpRouteImplementation) Remove(nsname types.NamespacedName) {
	impl.Logger().Info("HTTPRoute resource was removed",
		"namespace", nsname.Namespace, "name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &v1alpha2.HTTPRoute{},
	}
}
