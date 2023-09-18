package static

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nkgAPI "github.com/nginxinc/nginx-kubernetes-gateway/apis/v1alpha1"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state"
	staticConds "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/conditions"
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
	// eventRecorder records events for Kubernetes resources.
	eventRecorder record.EventRecorder
	// logLevelSetter is used to update the logging level.
	logLevelSetter ZapLogLevelSetter
	// healthChecker sets the health of the Pod to Ready once we've written out our initial config
	healthChecker *healthChecker
	// controlConfigNSName is the NamespacedName of the NginxGateway config for this controller.
	controlConfigNSName types.NamespacedName
	// logger is the logger to be used by the EventHandler.
	logger logr.Logger
	// version is the current version number of the nginx config.
	version int
}

// eventHandlerImpl implements EventHandler.
// eventHandlerImpl is responsible for:
// (1) Reconciling the Gateway API and Kubernetes built-in resources with the NGINX configuration.
// (2) Keeping the statuses of the Gateway API resources updated.
// (3) Updating control plane configuration.
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
			if cfg, ok := e.Resource.(*nkgAPI.NginxGateway); ok {
				h.updateControlPlaneAndSetStatus(ctx, cfg)
			} else {
				h.cfg.processor.CaptureUpsertChange(e.Resource)
			}
		case *events.DeleteEvent:
			if _, ok := e.Type.(*nkgAPI.NginxGateway); ok {
				h.updateControlPlaneAndSetStatus(ctx, nil)
			} else {
				h.cfg.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
			}
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}

	changed, graph := h.cfg.processor.Process()
	if !changed {
		h.cfg.logger.Info("Handling events didn't result into NGINX configuration changes")
		if !h.cfg.healthChecker.ready && h.cfg.healthChecker.firstBatchError == nil {
			h.cfg.healthChecker.setAsReady()
		}
		return
	}

	var nginxReloadRes nginxReloadResult
	h.cfg.version++
	if err := h.updateNginx(
		ctx,
		dataplane.BuildConfiguration(ctx, graph, h.cfg.serviceResolver, h.cfg.version),
	); err != nil {
		h.cfg.logger.Error(err, "Failed to update NGINX configuration")
		nginxReloadRes.error = err
		if !h.cfg.healthChecker.ready {
			h.cfg.healthChecker.firstBatchError = err
		}
	} else {
		h.cfg.logger.Info("NGINX configuration was successfully updated")
		if !h.cfg.healthChecker.ready {
			h.cfg.healthChecker.setAsReady()
		}
	}

	h.cfg.statusUpdater.Update(ctx, buildGatewayAPIStatuses(graph, nginxReloadRes))
}

func (h *eventHandlerImpl) updateNginx(ctx context.Context, conf dataplane.Configuration) error {
	files := h.cfg.generator.Generate(conf)

	if err := h.cfg.nginxFileMgr.ReplaceFiles(files); err != nil {
		return fmt.Errorf("failed to replace NGINX configuration files: %w", err)
	}

	if err := h.cfg.nginxRuntimeMgr.Reload(ctx, conf.Version); err != nil {
		return fmt.Errorf("failed to reload NGINX: %w", err)
	}

	return nil
}

// updateControlPlaneAndSetStatus updates the control plane configuration and then sets the status
// based on the outcome
func (h *eventHandlerImpl) updateControlPlaneAndSetStatus(ctx context.Context, cfg *nkgAPI.NginxGateway) {
	var cond []conditions.Condition
	if err := updateControlPlane(
		cfg,
		h.cfg.logger,
		h.cfg.eventRecorder,
		h.cfg.controlConfigNSName,
		h.cfg.logLevelSetter,
	); err != nil {
		msg := "Failed to update control plane configuration"
		h.cfg.logger.Error(err, msg)
		h.cfg.eventRecorder.Eventf(
			cfg,
			apiv1.EventTypeWarning,
			"UpdateFailed",
			msg+": %s",
			err.Error(),
		)
		cond = []conditions.Condition{staticConds.NewNginxGatewayInvalid(fmt.Sprintf("%s: %v", msg, err))}
	} else {
		cond = []conditions.Condition{staticConds.NewNginxGatewayValid()}
	}

	if cfg != nil {

		NginxGatewayStatus := &status.NginxGatewayStatus{
			NsName:             client.ObjectKeyFromObject(cfg),
			Conditions:         cond,
			ObservedGeneration: cfg.Generation,
		}

		h.cfg.statusUpdater.Update(ctx, NginxGatewayStatus)
		h.cfg.logger.Info("Reconfigured control plane.")
	}
}
