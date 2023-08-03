package static

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// atomicLevel is used for updating the logger's log level.
	atomicLevel zap.AtomicLevel
	// logger is the logger to be used by the EventHandler.
	logger logr.Logger
	// controlConfigNSName is the NamespacedName of the NginxGateway config for this controller.
	controlConfigNSName types.NamespacedName
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
				h.updateControlPlane(ctx, cfg)
			} else {
				h.cfg.processor.CaptureUpsertChange(e.Resource)
			}
		case *events.DeleteEvent:
			if _, ok := e.Type.(*nkgAPI.NginxGateway); ok {
				h.updateControlPlane(ctx, nil)
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

// updateControlPlane updates the control plane configuration with the given user spec.
// If any fields are not set within the user spec, the default configuration values are used.
func (h *eventHandlerImpl) updateControlPlane(ctx context.Context, cfg *nkgAPI.NginxGateway) {
	// build up default configuration
	defaultLogLevel := nkgAPI.ControllerLogLevelInfo
	controlConfig := nkgAPI.NginxGatewaySpec{
		Logging: &nkgAPI.Logging{
			Level: &defaultLogLevel,
		},
	}

	updateControlPlane := func() error {
		// by marshaling the user config and then unmarshaling on top of the default config,
		// we ensure that any unset user values are set with the default values
		if cfg != nil {
			cfgBytes, err := json.Marshal(cfg.Spec)
			if err != nil {
				return fmt.Errorf("error marshaling control config: %w", err)
			}

			if err := json.Unmarshal(cfgBytes, &controlConfig); err != nil {
				return fmt.Errorf("error unmarshaling control config: %w", err)
			}
		} else {
			msg := "NginxGateway configuration was deleted; using defaults"
			h.cfg.logger.Error(nil, msg)
			h.cfg.eventRecorder.Event(
				&nkgAPI.NginxGateway{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: h.cfg.controlConfigNSName.Namespace,
						Name:      h.cfg.controlConfigNSName.Name,
					},
				},
				apiv1.EventTypeWarning,
				"ResourceDeleted",
				msg,
			)
		}

		// set the log level
		level, err := zapcore.ParseLevel(string(*controlConfig.Logging.Level))
		if err != nil {
			return fmt.Errorf("error parsing log level string: %w", err)
		}
		h.cfg.atomicLevel.SetLevel(level)

		return nil
	}

	var cond []conditions.Condition
	if err := updateControlPlane(); err != nil {
		msg := "Failed to update control plane configuration"
		h.cfg.logger.Error(err, msg)
		h.cfg.eventRecorder.Eventf(
			cfg,
			apiv1.EventTypeWarning,
			"FailedUpdate",
			"%s; "+msg,
			err.Error(),
		)
		cond = []conditions.Condition{staticConds.NewNginxGatewayInvalid(fmt.Sprintf("%s: %v", msg, err))}
	} else {
		cond = []conditions.Condition{staticConds.NewNginxGatewayValid()}
	}

	if cfg != nil {
		statuses := status.Statuses{
			NginxGatewayStatus: status.NginxGatewayStatus{
				NSName:             client.ObjectKeyFromObject(cfg),
				Conditions:         cond,
				ObservedGeneration: cfg.Generation,
			},
		}

		h.cfg.statusUpdater.Update(ctx, statuses)
	}
}
