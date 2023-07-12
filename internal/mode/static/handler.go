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

// eventHandlerConfig holds configuration parameters for eventHandlerImpl.
type eventHandlerConfig struct {
	// processor is the state ChangeProcessor.
	processor state.ChangeProcessor
	// serviceResolver resolves Services to Endpoints.
	serviceResolver resolver.ServiceResolver
	// generator is the nginx config generator.
	generator config.Generator
	// nginxFileMgr is the file Manager for nginx.
	nginxFileMgr file.Manager
	// nginxRuntimeMgr manages nginx runtime.
	nginxRuntimeMgr runtime.Manager
	// statusUpdater updates statuses on Kubernetes resources.
	statusUpdater status.Updater
	// logger is the logger to be used by the EventHandler.
	logger logr.Logger
}

// eventHandlerImpl implements EventHandler.
// eventHandlerImpl is responsible for:
// (1) Reconciling the Gateway API and Kubernetes built-in resources with the NGINX configuration.
// (2) Keeping the statuses of the Gateway API resources updated.
type eventHandlerImpl struct {
	cfg eventHandlerConfig
}

// newEventHandlerImpl creates a new eventHandlerImpl.
func newEventHandlerImpl(cfg eventHandlerConfig) *eventHandlerImpl {
	return &eventHandlerImpl{
		cfg: cfg,
	}
}

func (h *eventHandlerImpl) HandleEventBatch(ctx context.Context, batch events.EventBatch) {
	for _, event := range batch {
		switch e := event.(type) {
		case *events.UpsertEvent:
			h.cfg.processor.CaptureUpsertChange(e.Resource)
		case *events.DeleteEvent:
			h.cfg.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}

	changed, graph := h.cfg.processor.Process()
	if !changed {
		h.cfg.logger.Info("Handling events didn't result into NGINX configuration changes")
		return
	}

	var nginxReloadRes nginxReloadResult
	err := h.updateNginx(ctx, dataplane.BuildConfiguration(ctx, graph, h.cfg.serviceResolver))
	if err != nil {
		h.cfg.logger.Error(err, "Failed to update NGINX configuration")
		nginxReloadRes.error = err
	} else {
		h.cfg.logger.Info("NGINX configuration was successfully updated")
	}

	h.cfg.statusUpdater.Update(ctx, buildStatuses(graph, nginxReloadRes))
}

func (h *eventHandlerImpl) updateNginx(ctx context.Context, conf dataplane.Configuration) error {
	files := h.cfg.generator.Generate(conf)

	if err := h.cfg.nginxFileMgr.ReplaceFiles(files); err != nil {
		return fmt.Errorf("failed to replace NGINX configuration files: %w", err)
	}

	if err := h.cfg.nginxRuntimeMgr.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload NGINX: %w", err)
	}

	return nil
}
