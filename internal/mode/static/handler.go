package static

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	ngfConfig "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	ngxConfig "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

type handlerMetricsCollector interface {
	ObserveLastEventBatchProcessTime(time.Duration)
}

// eventHandlerConfig holds configuration parameters for eventHandlerImpl.
type eventHandlerConfig struct {
	// k8sClient is a Kubernetes API client
	k8sClient client.Client
	// gatewayPodConfig contains information about this Pod.
	gatewayPodConfig ngfConfig.GatewayPodConfig
	// processor is the state ChangeProcessor.
	processor state.ChangeProcessor
	// serviceResolver resolves Services to Endpoints.
	serviceResolver resolver.ServiceResolver
	// generator is the nginx config generator.
	generator ngxConfig.Generator
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
	// metricsCollector collects metrics for this controller.
	metricsCollector handlerMetricsCollector
	// healthChecker sets the health of the Pod to Ready once we've written out our initial config
	healthChecker *healthChecker
	// controlConfigNSName is the NamespacedName of the NginxGateway config for this controller.
	controlConfigNSName types.NamespacedName
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

func (h *eventHandlerImpl) HandleEventBatch(ctx context.Context, logger logr.Logger, batch events.EventBatch) {
	start := time.Now()
	logger.V(1).Info("Started processing event batch")

	defer func() {
		duration := time.Since(start)
		logger.V(1).Info(
			"Finished processing event batch",
			"duration", duration.String(),
		)
		h.cfg.metricsCollector.ObserveLastEventBatchProcessTime(duration)
	}()

	for _, event := range batch {
		h.handleEvent(ctx, logger, event)
	}

	changed, graph := h.cfg.processor.Process()
	if !changed {
		logger.Info("Handling events didn't result into NGINX configuration changes")
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
		logger.Error(err, "Failed to update NGINX configuration")
		nginxReloadRes.error = err
		if !h.cfg.healthChecker.ready {
			h.cfg.healthChecker.firstBatchError = err
		}
	} else {
		logger.Info("NGINX configuration was successfully updated")
		if !h.cfg.healthChecker.ready {
			h.cfg.healthChecker.setAsReady()
		}
	}

	gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, nil, h.cfg.gatewayPodConfig)
	if err != nil {
		logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
	}

	h.cfg.statusUpdater.Update(ctx, buildGatewayAPIStatuses(graph, gwAddresses, nginxReloadRes))
	h.cfg.statusUpdater.Update(ctx, buildNginxProxyStatus(graph.NginxProxy, nginxReloadRes))
}

func (h *eventHandlerImpl) handleEvent(ctx context.Context, logger logr.Logger, event interface{}) {
	switch e := event.(type) {
	case *events.UpsertEvent:
		switch obj := e.Resource.(type) {
		case *ngfAPI.NginxGateway:
			h.updateControlPlaneAndSetStatus(ctx, logger, obj)
		case *apiv1.Service:
			podConfig := h.cfg.gatewayPodConfig
			if obj.Name == podConfig.ServiceName && obj.Namespace == podConfig.Namespace {
				gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, obj, h.cfg.gatewayPodConfig)
				if err != nil {
					logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
				}
				h.cfg.statusUpdater.UpdateAddresses(ctx, gwAddresses)
			}
			h.cfg.processor.CaptureUpsertChange(e.Resource)
		default:
			h.cfg.processor.CaptureUpsertChange(e.Resource)
		}
	case *events.DeleteEvent:
		switch e.Type.(type) {
		case *ngfAPI.NginxGateway:
			h.updateControlPlaneAndSetStatus(ctx, logger, nil)
		case *apiv1.Service:
			podConfig := h.cfg.gatewayPodConfig
			if e.NamespacedName.Name == podConfig.ServiceName && e.NamespacedName.Namespace == podConfig.Namespace {
				gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, nil, h.cfg.gatewayPodConfig)
				if err != nil {
					logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
				}
				h.cfg.statusUpdater.UpdateAddresses(ctx, gwAddresses)
			}
			h.cfg.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
		default:
			h.cfg.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
		}
	default:
		panic(fmt.Errorf("unknown event type %T", e))
	}
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
func (h *eventHandlerImpl) updateControlPlaneAndSetStatus(
	ctx context.Context,
	logger logr.Logger,
	cfg *ngfAPI.NginxGateway,
) {
	var cond []conditions.Condition
	if err := updateControlPlane(
		cfg,
		logger,
		h.cfg.eventRecorder,
		h.cfg.controlConfigNSName,
		h.cfg.logLevelSetter,
	); err != nil {
		msg := "Failed to update control plane configuration"
		logger.Error(err, msg)
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
		nginxGatewayStatus := &status.NginxGatewayStatus{
			NsName:             client.ObjectKeyFromObject(cfg),
			Conditions:         cond,
			ObservedGeneration: cfg.Generation,
		}

		h.cfg.statusUpdater.Update(ctx, nginxGatewayStatus)
		logger.Info("Reconfigured control plane.")
	}
}

// getGatewayAddresses gets the addresses for the Gateway.
func getGatewayAddresses(
	ctx context.Context,
	k8sClient client.Client,
	svc *v1.Service,
	podConfig config.GatewayPodConfig,
) ([]v1beta1.GatewayStatusAddress, error) {
	podAddress := []v1beta1.GatewayStatusAddress{
		{
			Type:  helpers.GetPointer(v1beta1.IPAddressType),
			Value: podConfig.PodIP,
		},
	}

	var gwSvc v1.Service
	if svc == nil {
		key := types.NamespacedName{Name: podConfig.ServiceName, Namespace: podConfig.Namespace}
		if err := k8sClient.Get(ctx, key, &gwSvc); err != nil {
			return podAddress, fmt.Errorf("error finding Service for Gateway: %w", err)
		}
	} else {
		gwSvc = *svc
	}

	var addresses, hostnames []string
	switch gwSvc.Spec.Type {
	case v1.ServiceTypeLoadBalancer:
		for _, ingress := range gwSvc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				addresses = append(addresses, ingress.IP)
			} else if ingress.Hostname != "" {
				hostnames = append(hostnames, ingress.Hostname)
			}
		}
	}

	gwAddresses := make([]v1beta1.GatewayStatusAddress, 0, len(addresses)+len(hostnames))
	for _, addr := range addresses {
		statusAddr := v1beta1.GatewayStatusAddress{
			Type:  helpers.GetPointer(v1beta1.IPAddressType),
			Value: addr,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	for _, hostname := range hostnames {
		statusAddr := v1beta1.GatewayStatusAddress{
			Type:  helpers.GetPointer(v1beta1.HostnameAddressType),
			Value: hostname,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	return gwAddresses, nil
}
