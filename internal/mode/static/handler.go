package static

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/resolver"
)

// EventHandlerConfig holds configuration parameters for EventHandlerImpl.
type EventHandlerConfig struct {
	// Processor is the state ChangeProcessor.
	Processor state.ChangeProcessor
	// ServiceResolver resolves Services to Endpoints.
	ServiceResolver resolver.ServiceResolver
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

func (h *EventHandlerImpl) HandleEventBatch(ctx context.Context, batch events.EventBatch) {
	for _, event := range batch {
		switch e := event.(type) {
		case *events.UpsertEvent:
			h.cfg.Processor.CaptureUpsertChange(e.Resource)
		case *events.DeleteEvent:
			h.cfg.Processor.CaptureDeleteChange(e.Type, e.NamespacedName)
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}

	changed, graph := h.cfg.Processor.Process()
	if !changed {
		h.cfg.Logger.Info("Handling events didn't result into NGINX configuration changes")
		return
	}

	var nginxReloadRes nginxReloadResult
	err := h.updateNginx(ctx, dataplane.BuildConfiguration(ctx, graph, h.cfg.ServiceResolver))
	if err != nil {
		h.cfg.Logger.Error(err, "Failed to update NGINX configuration")
		nginxReloadRes.error = err
	} else {
		h.cfg.Logger.Info("NGINX configuration was successfully updated")
	}

	h.cfg.StatusUpdater.Update(ctx, buildStatuses(graph, nginxReloadRes))
}

func (h *EventHandlerImpl) updateNginx(ctx context.Context, conf dataplane.Configuration) error {
	files := h.cfg.Generator.Generate(conf)

	if err := h.cfg.NginxFileMgr.ReplaceFiles(files); err != nil {
		return fmt.Errorf("failed to replace NGINX configuration files: %w", err)
	}

	if err := h.cfg.NginxRuntimeMgr.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload NGINX: %w", err)
	}

	return nil
}
