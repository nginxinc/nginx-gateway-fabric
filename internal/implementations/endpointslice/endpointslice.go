package implementation

import (
	"github.com/go-logr/logr"
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
)

type endpointSliceImplementation struct {
	conf    config.Config
	eventCh chan<- interface{}
}

// NewEndpointSliceImplementation creates a new EndpointSliceImplementation.
func NewEndpointSliceImplementation(cfg config.Config, eventCh chan<- interface{}) *endpointSliceImplementation {
	return &endpointSliceImplementation{
		conf:    cfg,
		eventCh: eventCh,
	}
}

func (impl *endpointSliceImplementation) Logger() logr.Logger {
	return impl.conf.Logger
}

func (impl *endpointSliceImplementation) Upsert(endpSlice *discoveryV1.EndpointSlice) {
	impl.Logger().Info("Endpoint Slice was upserted",
		"namespace", endpSlice.Namespace, "name", endpSlice.Name,
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: endpSlice,
	}
}

func (impl *endpointSliceImplementation) Remove(nsname types.NamespacedName) {
	impl.Logger().Info("Endpoint Slice resource was removed",
		"namespace", nsname.Namespace, "name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           &discoveryV1.EndpointSlice{},
	}
}
