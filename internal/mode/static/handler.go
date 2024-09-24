package static

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/v2/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	frameworkStatus "github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"

	ngfConfig "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	ngxConfig "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/status"
)

type handlerMetricsCollector interface {
	ObserveLastEventBatchProcessTime(time.Duration)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . secretStorer

// secretStorer should store the usage Secret that contains the credentials for NGINX Plus usage reporting.
type secretStorer interface {
	// Set stores the updated Secret.
	Set(*v1.Secret)
	// Delete nullifies the Secret value.
	Delete()
}

// eventHandlerConfig holds configuration parameters for eventHandlerImpl.
type eventHandlerConfig struct {
	// nginxFileMgr is the file Manager for nginx.
	nginxFileMgr file.Manager
	// metricsCollector collects metrics for this controller.
	metricsCollector handlerMetricsCollector
	// nginxRuntimeMgr manages nginx runtime.
	nginxRuntimeMgr runtime.Manager
	// statusUpdater updates statuses on Kubernetes resources.
	statusUpdater frameworkStatus.GroupUpdater
	// usageSecret contains the Secret for the NGINX Plus reporting credentials.
	usageSecret secretStorer
	// processor is the state ChangeProcessor.
	processor state.ChangeProcessor
	// serviceResolver resolves Services to Endpoints.
	serviceResolver resolver.ServiceResolver
	// generator is the nginx config generator.
	generator ngxConfig.Generator
	// k8sClient is a Kubernetes API client
	k8sClient client.Client
	// logLevelSetter is used to update the logging level.
	logLevelSetter logLevelSetter
	// eventRecorder records events for Kubernetes resources.
	eventRecorder record.EventRecorder
	// usageReportConfig contains the configuration for NGINX Plus usage reporting.
	usageReportConfig *ngfConfig.UsageReportConfig
	// nginxConfiguredOnStartChecker sets the health of the Pod to Ready once we've written out our initial config.
	nginxConfiguredOnStartChecker *nginxConfiguredOnStartChecker
	// gatewayPodConfig contains information about this Pod.
	gatewayPodConfig ngfConfig.GatewayPodConfig
	// controlConfigNSName is the NamespacedName of the NginxGateway config for this controller.
	controlConfigNSName types.NamespacedName
	// gatewayCtlrName is the name of the NGF controller.
	gatewayCtlrName string
	// updateGatewayClassStatus enables updating the status of the GatewayClass resource.
	updateGatewayClassStatus bool
}

const (
	// groups for GroupStatusUpdater.
	groupAllExceptGateways = "all-graphs-except-gateways"
	groupGateways          = "gateways"
	groupControlPlane      = "control-plane"
)

// filterKey is the `kind_namespace_name" of an object being filtered.
type filterKey string

// objectFilter contains callbacks for an object that should be treated differently by the handler instead of
// just using the typical Capture() call.
type objectFilter struct {
	upsert               func(context.Context, logr.Logger, client.Object)
	delete               func(context.Context, logr.Logger, types.NamespacedName)
	captureChangeInGraph bool
}

// eventHandlerImpl implements EventHandler.
// eventHandlerImpl is responsible for:
// (1) Reconciling the Gateway API and Kubernetes built-in resources with the NGINX configuration.
// (2) Keeping the statuses of the Gateway API resources updated.
// (3) Updating control plane configuration.
// (4) Tracks the NGINX Plus usage reporting Secret (if applicable).
type eventHandlerImpl struct {
	// latestConfiguration is the latest Configuration generation.
	latestConfiguration *dataplane.Configuration

	// objectFilters contains all created objectFilters, with the key being a filterKey
	objectFilters map[filterKey]objectFilter

	latestReloadResult status.NginxReloadResult

	cfg  eventHandlerConfig
	lock sync.Mutex

	// version is the current version number of the nginx config.
	version int
}

// newEventHandlerImpl creates a new eventHandlerImpl.
func newEventHandlerImpl(cfg eventHandlerConfig) *eventHandlerImpl {
	handler := &eventHandlerImpl{
		cfg: cfg,
	}

	handler.objectFilters = map[filterKey]objectFilter{
		// NginxGateway CRD
		objectFilterKey(&ngfAPI.NginxGateway{}, handler.cfg.controlConfigNSName): {
			upsert:               handler.nginxGatewayCRDUpsert,
			delete:               handler.nginxGatewayCRDDelete,
			captureChangeInGraph: false,
		},
		// NGF-fronting Service
		objectFilterKey(
			&v1.Service{},
			types.NamespacedName{
				Name:      handler.cfg.gatewayPodConfig.ServiceName,
				Namespace: handler.cfg.gatewayPodConfig.Namespace,
			},
		): {
			upsert:               handler.nginxGatewayServiceUpsert,
			delete:               handler.nginxGatewayServiceDelete,
			captureChangeInGraph: true,
		},
	}

	if handler.cfg.usageReportConfig != nil {
		// N+ usage reporting Secret
		nsName := handler.cfg.usageReportConfig.SecretNsName
		handler.objectFilters[objectFilterKey(&v1.Secret{}, nsName)] = objectFilter{
			upsert:               handler.nginxPlusUsageSecretUpsert,
			delete:               handler.nginxPlusUsageSecretDelete,
			captureChangeInGraph: true,
		}
	}

	return handler
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
		h.parseAndCaptureEvent(ctx, logger, event)
	}

	changeType, gr := h.cfg.processor.Process()

	var err error
	switch changeType {
	case state.NoChange:
		logger.Info("Handling events didn't result into NGINX configuration changes")
		if !h.cfg.nginxConfiguredOnStartChecker.ready && h.cfg.nginxConfiguredOnStartChecker.firstBatchError == nil {
			h.cfg.nginxConfiguredOnStartChecker.setAsReady()
		}
		return
	case state.EndpointsOnlyChange:
		h.version++
		cfg := dataplane.BuildConfiguration(ctx, gr, h.cfg.serviceResolver, h.version)

		h.setLatestConfiguration(&cfg)

		err = h.updateUpstreamServers(
			ctx,
			logger,
			cfg,
		)
	case state.ClusterStateChange:
		h.version++
		cfg := dataplane.BuildConfiguration(ctx, gr, h.cfg.serviceResolver, h.version)

		h.setLatestConfiguration(&cfg)

		err = h.updateNginxConf(
			ctx,
			cfg,
		)
	}

	var nginxReloadRes status.NginxReloadResult
	if err != nil {
		logger.Error(err, "Failed to update NGINX configuration")
		nginxReloadRes.Error = err
		if !h.cfg.nginxConfiguredOnStartChecker.ready {
			h.cfg.nginxConfiguredOnStartChecker.firstBatchError = err
		}
	} else {
		logger.Info("NGINX configuration was successfully updated")
		if !h.cfg.nginxConfiguredOnStartChecker.ready {
			h.cfg.nginxConfiguredOnStartChecker.setAsReady()
		}
	}

	h.latestReloadResult = nginxReloadRes

	h.updateStatuses(ctx, logger, gr)
}

func (h *eventHandlerImpl) updateStatuses(ctx context.Context, logger logr.Logger, gr *graph.Graph) {
	gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, nil, h.cfg.gatewayPodConfig)
	if err != nil {
		logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
	}

	transitionTime := metav1.Now()

	var gcReqs []frameworkStatus.UpdateRequest
	if h.cfg.updateGatewayClassStatus {
		gcReqs = status.PrepareGatewayClassRequests(gr.GatewayClass, gr.IgnoredGatewayClasses, transitionTime)
	}
	routeReqs := status.PrepareRouteRequests(
		gr.L4Routes,
		gr.Routes,
		transitionTime,
		h.latestReloadResult,
		h.cfg.gatewayCtlrName,
	)

	polReqs := status.PrepareBackendTLSPolicyRequests(gr.BackendTLSPolicies, transitionTime, h.cfg.gatewayCtlrName)
	ngfPolReqs := status.PrepareNGFPolicyRequests(gr.NGFPolicies, transitionTime, h.cfg.gatewayCtlrName)

	reqs := make([]frameworkStatus.UpdateRequest, 0, len(gcReqs)+len(routeReqs)+len(polReqs)+len(ngfPolReqs))
	reqs = append(reqs, gcReqs...)
	reqs = append(reqs, routeReqs...)
	reqs = append(reqs, polReqs...)
	reqs = append(reqs, ngfPolReqs...)

	h.cfg.statusUpdater.UpdateGroup(ctx, groupAllExceptGateways, reqs...)

	// We put Gateway status updates separately from the rest of the statuses because we want to be able
	// to update them separately from the rest of the graph whenever the public IP of NGF changes.
	gwReqs := status.PrepareGatewayRequests(
		gr.Gateway,
		gr.IgnoredGateways,
		transitionTime,
		gwAddresses,
		h.latestReloadResult,
	)
	h.cfg.statusUpdater.UpdateGroup(ctx, groupGateways, gwReqs...)
}

func (h *eventHandlerImpl) parseAndCaptureEvent(ctx context.Context, logger logr.Logger, event interface{}) {
	switch e := event.(type) {
	case *events.UpsertEvent:
		upFilterKey := objectFilterKey(e.Resource, client.ObjectKeyFromObject(e.Resource))

		if filter, ok := h.objectFilters[upFilterKey]; ok {
			filter.upsert(ctx, logger, e.Resource)
			if !filter.captureChangeInGraph {
				return
			}
		}

		h.cfg.processor.CaptureUpsertChange(e.Resource)
	case *events.DeleteEvent:
		delFilterKey := objectFilterKey(e.Type, e.NamespacedName)

		if filter, ok := h.objectFilters[delFilterKey]; ok {
			filter.delete(ctx, logger, e.NamespacedName)
			if !filter.captureChangeInGraph {
				return
			}
		}

		h.cfg.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	default:
		panic(fmt.Errorf("unknown event type %T", e))
	}
}

// updateNginxConf updates nginx conf files and reloads nginx.
func (h *eventHandlerImpl) updateNginxConf(ctx context.Context, conf dataplane.Configuration) error {
	files := h.cfg.generator.Generate(conf)
	if err := h.cfg.nginxFileMgr.ReplaceFiles(files); err != nil {
		return fmt.Errorf("failed to replace NGINX configuration files: %w", err)
	}

	if err := h.cfg.nginxRuntimeMgr.Reload(ctx, conf.Version); err != nil {
		return fmt.Errorf("failed to reload NGINX: %w", err)
	}

	return nil
}

// updateUpstreamServers is called only when endpoints have changed. It updates nginx conf files and then:
// - if using NGINX Plus, determines which servers have changed and uses the N+ API to update them;
// - otherwise if not using NGINX Plus, or an error was returned from the API, reloads nginx.
func (h *eventHandlerImpl) updateUpstreamServers(
	ctx context.Context,
	logger logr.Logger,
	conf dataplane.Configuration,
) error {
	isPlus := h.cfg.nginxRuntimeMgr.IsPlus()

	files := h.cfg.generator.Generate(conf)
	if err := h.cfg.nginxFileMgr.ReplaceFiles(files); err != nil {
		return fmt.Errorf("failed to replace NGINX configuration files: %w", err)
	}

	reload := func() error {
		if err := h.cfg.nginxRuntimeMgr.Reload(ctx, conf.Version); err != nil {
			return fmt.Errorf("failed to reload NGINX: %w", err)
		}

		return nil
	}

	if isPlus {
		type upstream struct {
			name    string
			servers []ngxclient.UpstreamServer
		}
		var upstreams []upstream

		prevUpstreams, err := h.cfg.nginxRuntimeMgr.GetUpstreams()
		if err != nil {
			logger.Error(err, "failed to get upstreams from API, reloading configuration instead")
			return reload()
		}

		for _, u := range conf.Upstreams {
			confUpstream := upstream{
				name:    u.Name,
				servers: ngxConfig.ConvertEndpoints(u.Endpoints),
			}

			if u, ok := prevUpstreams[confUpstream.name]; ok {
				if !serversEqual(confUpstream.servers, u.Peers) {
					upstreams = append(upstreams, confUpstream)
				}
			}
		}

		var reloadPlus bool
		for _, upstream := range upstreams {
			if err := h.cfg.nginxRuntimeMgr.UpdateHTTPServers(upstream.name, upstream.servers); err != nil {
				logger.Error(
					err, "couldn't update upstream via the API, reloading configuration instead",
					"upstreamName", upstream.name,
				)
				reloadPlus = true
			}
		}

		if !reloadPlus {
			return nil
		}
	}

	return reload()
}

func serversEqual(newServers []ngxclient.UpstreamServer, oldServers []ngxclient.Peer) bool {
	if len(newServers) != len(oldServers) {
		return false
	}

	diff := make(map[string]struct{}, len(newServers))
	for _, s := range newServers {
		diff[s.Server] = struct{}{}
	}

	for _, s := range oldServers {
		if _, ok := diff[s.Server]; !ok {
			return false
		}
	}

	return true
}

// updateControlPlaneAndSetStatus updates the control plane configuration and then sets the status
// based on the outcome.
func (h *eventHandlerImpl) updateControlPlaneAndSetStatus(
	ctx context.Context,
	logger logr.Logger,
	cfg *ngfAPI.NginxGateway,
) {
	var cpUpdateRes status.ControlPlaneUpdateResult

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
			v1.EventTypeWarning,
			"UpdateFailed",
			msg+": %s",
			err.Error(),
		)
		cpUpdateRes.Error = err
	}

	var reqs []frameworkStatus.UpdateRequest

	req := status.PrepareNginxGatewayStatus(cfg, metav1.Now(), cpUpdateRes)
	if req != nil {
		reqs = append(reqs, *req)
	}

	h.cfg.statusUpdater.UpdateGroup(ctx, groupControlPlane, reqs...)

	logger.Info("Reconfigured control plane.")
}

// getGatewayAddresses gets the addresses for the Gateway.
func getGatewayAddresses(
	ctx context.Context,
	k8sClient client.Client,
	svc *v1.Service,
	podConfig ngfConfig.GatewayPodConfig,
) ([]gatewayv1.GatewayStatusAddress, error) {
	podAddress := []gatewayv1.GatewayStatusAddress{
		{
			Type:  helpers.GetPointer(gatewayv1.IPAddressType),
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
	if gwSvc.Spec.Type == v1.ServiceTypeLoadBalancer {
		for _, ingress := range gwSvc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				addresses = append(addresses, ingress.IP)
			} else if ingress.Hostname != "" {
				hostnames = append(hostnames, ingress.Hostname)
			}
		}
	}

	gwAddresses := make([]gatewayv1.GatewayStatusAddress, 0, len(addresses)+len(hostnames))
	for _, addr := range addresses {
		statusAddr := gatewayv1.GatewayStatusAddress{
			Type:  helpers.GetPointer(gatewayv1.IPAddressType),
			Value: addr,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	for _, hostname := range hostnames {
		statusAddr := gatewayv1.GatewayStatusAddress{
			Type:  helpers.GetPointer(gatewayv1.HostnameAddressType),
			Value: hostname,
		}
		gwAddresses = append(gwAddresses, statusAddr)
	}

	return gwAddresses, nil
}

// GetLatestConfiguration gets the latest configuration.
func (h *eventHandlerImpl) GetLatestConfiguration() *dataplane.Configuration {
	h.lock.Lock()
	defer h.lock.Unlock()

	return h.latestConfiguration
}

// setLatestConfiguration sets the latest configuration.
func (h *eventHandlerImpl) setLatestConfiguration(cfg *dataplane.Configuration) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.latestConfiguration = cfg
}

func objectFilterKey(obj client.Object, nsName types.NamespacedName) filterKey {
	return filterKey(fmt.Sprintf("%T_%s_%s", obj, nsName.Namespace, nsName.Name))
}

/*

Handler Callback functions

These functions are provided as callbacks to the handler. They are for objects that need special
treatment other than the typical Capture() call that leads to generating nginx config.

*/

func (h *eventHandlerImpl) nginxGatewayCRDUpsert(ctx context.Context, logger logr.Logger, obj client.Object) {
	cfg, ok := obj.(*ngfAPI.NginxGateway)
	if !ok {
		panic(fmt.Errorf("obj type mismatch: got %T, expected %T", obj, &ngfAPI.NginxGateway{}))
	}

	h.updateControlPlaneAndSetStatus(ctx, logger, cfg)
}

func (h *eventHandlerImpl) nginxGatewayCRDDelete(
	ctx context.Context,
	logger logr.Logger,
	_ types.NamespacedName,
) {
	h.updateControlPlaneAndSetStatus(ctx, logger, nil)
}

func (h *eventHandlerImpl) nginxGatewayServiceUpsert(ctx context.Context, logger logr.Logger, obj client.Object) {
	svc, ok := obj.(*v1.Service)
	if !ok {
		panic(fmt.Errorf("obj type mismatch: got %T, expected %T", svc, &v1.Service{}))
	}

	gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, svc, h.cfg.gatewayPodConfig)
	if err != nil {
		logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
	}

	gr := h.cfg.processor.GetLatestGraph()
	if gr == nil {
		return
	}

	transitionTime := metav1.Now()
	gatewayStatuses := status.PrepareGatewayRequests(
		gr.Gateway,
		gr.IgnoredGateways,
		transitionTime,
		gwAddresses,
		h.latestReloadResult,
	)
	h.cfg.statusUpdater.UpdateGroup(ctx, groupGateways, gatewayStatuses...)
}

func (h *eventHandlerImpl) nginxGatewayServiceDelete(
	ctx context.Context,
	logger logr.Logger,
	_ types.NamespacedName,
) {
	gwAddresses, err := getGatewayAddresses(ctx, h.cfg.k8sClient, nil, h.cfg.gatewayPodConfig)
	if err != nil {
		logger.Error(err, "Setting GatewayStatusAddress to Pod IP Address")
	}

	gr := h.cfg.processor.GetLatestGraph()
	if gr == nil {
		return
	}

	transitionTime := metav1.Now()
	gatewayStatuses := status.PrepareGatewayRequests(
		gr.Gateway,
		gr.IgnoredGateways,
		transitionTime,
		gwAddresses,
		h.latestReloadResult,
	)
	h.cfg.statusUpdater.UpdateGroup(ctx, groupGateways, gatewayStatuses...)
}

func (h *eventHandlerImpl) nginxPlusUsageSecretUpsert(_ context.Context, _ logr.Logger, obj client.Object) {
	secret, ok := obj.(*v1.Secret)
	if !ok {
		panic(fmt.Errorf("obj type mismatch: got %T, expected %T", obj, &v1.Secret{}))
	}

	h.cfg.usageSecret.Set(secret)
}

func (h *eventHandlerImpl) nginxPlusUsageSecretDelete(
	_ context.Context,
	_ logr.Logger,
	_ types.NamespacedName,
) {
	h.cfg.usageSecret.Delete()
}
