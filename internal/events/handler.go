package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . EventHandler

// EventHandler handle events.
type EventHandler interface {
	// HandleEventBatch handles a batch of events.
	// EventBatch can include duplicated events.
	HandleEventBatch(ctx context.Context, batch EventBatch)
}

type ConfigUpdater interface {
	Update(cfg observer.VersionedConfig)
}

type VersionedConfigAdapter interface {
	VersionedConfig(cfg dataplane.Configuration) (observer.VersionedConfig, error)
}

// EventHandlerConfig holds configuration parameters for EventHandlerImpl.
type EventHandlerConfig struct {
	// Processor is the state ChangeProcessor.
	Processor state.ChangeProcessor
	// SecretStore is the state SecretStore.
	SecretStore secrets.SecretStore
	// ConfigAdapter adapts dataplane.Configuration to a dataplane-specific versioned configuration.
	ConfigAdapter VersionedConfigAdapter
	// ConfigUpdater updates configuration.
	ConfigUpdater ConfigUpdater
	// StatusUpdater updates statuses on Kubernetes resources.
	StatusUpdater status.Updater
	// Logger is the logger to be used by the EventHandler.
	Logger logr.Logger
}

// EventHandlerImpl implements EventHandler.
// EventHandlerImpl is responsible for:
// (1) Reconciling the Gateway API and Kubernetes built-in resources with the NGINX configuration.
// (2) Keeping the statuses of the Gateway API resources updated.
type EventHandlerImpl struct {
	cfg EventHandlerConfig
	// changeCounter keeps track of the number of dataplane configuration changes (used as config generation).
	changeCounter int
}

// NewEventHandlerImpl creates a new EventHandlerImpl.
func NewEventHandlerImpl(cfg EventHandlerConfig) *EventHandlerImpl {
	return &EventHandlerImpl{
		cfg: cfg,
	}
}

func (h *EventHandlerImpl) HandleEventBatch(ctx context.Context, batch EventBatch) {
	for _, event := range batch {
		switch e := event.(type) {
		case *UpsertEvent:
			h.propagateUpsert(e)
		case *DeleteEvent:
			h.propagateDelete(e)
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}

	changed, conf, statuses := h.cfg.Processor.Process(ctx)
	if !changed {
		h.cfg.Logger.Info("Handling events didn't result into NGINX configuration changes")
		return
	}

	h.changeCounter++
	conf.Version = h.changeCounter

	// TODO should I pass in the version to the adapter?
	vc, err := h.cfg.ConfigAdapter.VersionedConfig(conf)
	if err != nil {
		h.cfg.Logger.Error(err, "error adapting dataplane configuration to a versioned configuration")
		return
	}

	h.cfg.ConfigUpdater.Update(vc)

	h.cfg.StatusUpdater.Update(ctx, statuses)
}

func (h *EventHandlerImpl) propagateUpsert(e *UpsertEvent) {
	switch r := e.Resource.(type) {
	case *v1beta1.GatewayClass:
		h.cfg.Processor.CaptureUpsertChange(r)
	case *v1beta1.Gateway:
		h.cfg.Processor.CaptureUpsertChange(r)
	case *v1beta1.HTTPRoute:
		h.cfg.Processor.CaptureUpsertChange(r)
	case *apiv1.Service:
		h.cfg.Processor.CaptureUpsertChange(r)
	case *apiv1.Secret:
		// FIXME(kate-osborn): need to handle certificate rotation
		h.cfg.SecretStore.Upsert(r)
	case *discoveryV1.EndpointSlice:
		h.cfg.Processor.CaptureUpsertChange(r)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Resource))
	}
}

func (h *EventHandlerImpl) propagateDelete(e *DeleteEvent) {
	switch e.Type.(type) {
	case *v1beta1.GatewayClass:
		h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *v1beta1.Gateway:
		h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *v1beta1.HTTPRoute:
		h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *apiv1.Service:
		h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *apiv1.Secret:
		// FIXME(kate-osborn): make sure that affected servers are updated
		h.cfg.SecretStore.Delete(e.NamespacedName)
	case *discoveryV1.EndpointSlice:
		h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Type))
	}
}
