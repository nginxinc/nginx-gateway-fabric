package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
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

// EventHandlerConfig holds configuration parameters for EventHandlerImpl.
type EventHandlerConfig struct {
	// Processor is the state ChangeProcessor.
	Processor state.ChangeProcessor
	// SecretStore is the state SecretStore.
	SecretStore secrets.SecretStore
	// SecretMemoryManager is the state SecretMemoryManager.
	SecretMemoryManager secrets.SecretDiskMemoryManager
	// Generator is the nginx config Generator.
	Generator config.Generator
	// NginxFileMgr is the file Manager for nginx.
	NginxFileMgr file.Manager
	// NginxRuntimeMgr manages nginx runtime.
	NginxRuntimeMgr runtime.Manager
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
}

// NewEventHandlerImpl creates a new EventHandlerImpl.
func NewEventHandlerImpl(cfg EventHandlerConfig) *EventHandlerImpl {
	return &EventHandlerImpl{
		cfg: cfg,
	}
}

func (h *EventHandlerImpl) HandleEventBatch(ctx context.Context, batch EventBatch) {
	change := state.NewChange()

	for _, event := range batch {
		switch e := event.(type) {
		case *UpsertEvent:
			// FIXME(pleshakov): Move secret processing to ChangeProcessor
			secret, ok := e.Resource.(*apiv1.Secret)
			if ok {
				// FIXME(kate-osborn): need to handle certificate rotation
				h.cfg.SecretStore.Upsert(secret)
				continue
			}

			change.IncludeUpsert(e.Resource)
		case *DeleteEvent:
			// FIXME(pleshakov): Move secret processing to ChangeProcessor
			_, ok := e.Type.(*apiv1.Secret)
			if ok {
				// FIXME(kate-osborn): make sure that affected servers are updated
				h.cfg.SecretStore.Delete(e.NamespacedName)
				continue
			}

			change.IncludeDelete(e.Type, e.NamespacedName)
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}

	changed, conf, statuses := h.cfg.Processor.Process(ctx, change)
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

func (h *EventHandlerImpl) updateNginx(ctx context.Context, conf dataplane.Configuration) error {
	// Write all secrets (nuke and pave).
	// This will remove all secrets in the secrets directory before writing the requested secrets.
	// FIXME(kate-osborn): We may want to rethink this approach in the future and write and remove secrets individually.
	err := h.cfg.SecretMemoryManager.WriteAllRequestedSecrets()
	if err != nil {
		return err
	}

	cfg := h.cfg.Generator.Generate(conf)

	// For now, we keep all http servers and upstreams in one config file.
	// We might rethink that. For example, we can write each server to its file
	// or group servers in some way.
	err = h.cfg.NginxFileMgr.WriteHTTPConfig("http", cfg)
	if err != nil {
		return err
	}

	return h.cfg.NginxRuntimeMgr.Reload(ctx)
}