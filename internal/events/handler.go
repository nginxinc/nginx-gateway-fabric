package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . EventHandler

// EventHandler handle events.
type EventHandler interface {
	// HandleEventBatch handles a batch of events.
	// EventBatch can include duplicated events.
	HandleEventBatch(ctx context.Context, batch EventBatch)
}

// EventHandlerConfig holds configuration parameters for EventHandlerImpl.
type EventHandlerConfig struct {
	// Processor is the state ChangeProcessor.
	Processor state.ChangeProcessor
	// ServiceStore is the state ServiceStore.
	ServiceStore state.ServiceStore
	// SecretStore is the state SecretStore.
	SecretStore state.SecretStore
	// SecretMemoryManager is the state SecretMemoryManager.
	SecretMemoryManager state.SecretDiskMemoryManager
	// Generator is the nginx config Generator.
	Generator config.Generator
	// Logger is the logger to be used by the EventHandler.
	Logger logr.Logger
	// NginxFileMgr is the file Manager for nginx.
	NginxFileMgr file.Manager
	// NginxRuntimeMgr manages nginx runtime.
	NginxRuntimeMgr runtime.Manager
	// StatusUpdater updates statuses on Kubernetes resources.
	StatusUpdater status.Updater
}

// EventHandlerImpl implements EventHandler.
// EventHandlerImpl is responsible for:
// (1) Reconciling the Gateway API and Kubernetes built-in resources with the NGINX configuration.
// (2) Keeping the statuses of the Gateway API resources updated.
type EventHandlerImpl struct {
	cfg EventHandlerConfig
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

	changed, conf, statuses := h.cfg.Processor.Process()
	if !changed {
		h.cfg.Logger.Info("Handling events didn't result into NGINX configuration changes")
		return
	}

	err := h.updateNginx(ctx, conf)
	if err != nil {
		h.cfg.Logger.Error(err, "Failed to update NGINX configuration")
	} else {
		h.cfg.Logger.Info("NGINX configuration was successfully updated")
	}

	h.cfg.StatusUpdater.Update(ctx, statuses)
}

func (h *EventHandlerImpl) updateNginx(ctx context.Context, conf state.Configuration) error {
	// Write all secrets (nuke and pave).
	// This will remove all secrets in the secrets directory before writing the requested secrets.
	// FIXME(kate-osborn): We may want to rethink this approach in the future and write and remove secrets individually.
	err := h.cfg.SecretMemoryManager.WriteAllRequestedSecrets()
	if err != nil {
		return err
	}

	cfg, warnings := h.cfg.Generator.Generate(conf)

	// For now, we keep all http servers in one config
	// We might rethink that. For example, we can write each server to its file
	// or group servers in some way.
	err = h.cfg.NginxFileMgr.WriteHTTPServersConfig("http-servers", cfg)
	if err != nil {
		return err
	}

	for obj, objWarnings := range warnings {
		for _, w := range objWarnings {
			// FIXME(pleshakov): report warnings via Object status
			h.cfg.Logger.Info("Got warning while generating config",
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
				"namespace", obj.GetNamespace(),
				"name", obj.GetName(),
				"warning", w)
		}
	}

	return h.cfg.NginxRuntimeMgr.Reload(ctx)
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
		// FIXME(pleshakov): make sure the affected hosts are updated
		h.cfg.ServiceStore.Upsert(r)
	case *apiv1.Secret:
		// FIXME(kate-osborn): need to handle certificate rotation
		h.cfg.SecretStore.Upsert(r)
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
		// FIXME(pleshakov): make sure the affected hosts are updated
		h.cfg.ServiceStore.Delete(e.NamespacedName)
	case *apiv1.Secret:
		// FIXME(kate-osborn): make sure that affected servers are updated
		h.cfg.SecretStore.Delete(e.NamespacedName)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Type))
	}
}
