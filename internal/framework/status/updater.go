package status

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Updater

// Updater updates statuses of the Gateway API resources.
// Updater can be disabled. In this case, it will stop updating the statuses of resources, while
// always saving the statuses of the last Update call. This is used to support multiple replicas of
// control plane being able to run simultaneously where only the leader will update statuses.
type Updater interface {
	// Update updates the statuses of the resources.
	Update(context.Context, Status)
	// Enable enables status updates. The updater will update the statuses in Kubernetes API to ensure they match the
	// statuses of the last Update invocation.
	Enable(ctx context.Context)
	// Disable disables status updates.
	Disable()
}

// UpdaterConfig holds configuration parameters for Updater.
type UpdaterConfig struct {
	// Client is a Kubernetes API client.
	Client client.Client
	// Clock is used as a source of time for the LastTransitionTime field in Conditions in resource statuses.
	Clock Clock
	// Logger holds a logger to be used.
	Logger logr.Logger
	// GatewayCtlrName is the name of the Gateway controller.
	GatewayCtlrName string
	// GatewayClassName is the name of the GatewayClass resource.
	GatewayClassName string
	// PodIP is the IP address of this Pod.
	PodIP string
	// UpdateGatewayClassStatus enables updating the status of the GatewayClass resource.
	UpdateGatewayClassStatus bool
	// LeaderElectionEnabled indicates whether Leader Election is enabled.
	// If it is not enabled, the updater will always write statuses to the Kubernetes API.
	LeaderElectionEnabled bool
}

// UpdaterImpl updates statuses of the Gateway API resources.
//
// It has the following limitations:
//
// (1) It is not smart. It will update the status of a resource (make an API call) even if it hasn't changed.
//
// (2) It is synchronous, which means the status reporter can slow down the event loop.
// Consider the following cases:
// (a) Sometimes the Gateway will need to update statuses of all resources it handles, which could be ~1000. Making 1000
// status API calls sequentially will take time.
// (b) k8s API can become slow or even timeout. This will increase every update status API call.
// Making UpdaterImpl asynchronous will prevent it from adding variable delays to the event loop.
//
// (3) It doesn't clear the statuses of a resources that are no longer handled by the Gateway. For example, if
// an HTTPRoute resource no longer has the parentRef to the Gateway resources, the Gateway must update the status
// of the resource to remove the status about the removed parentRef.
//
// (4) If another controllers changes the status of the Gateway/HTTPRoute resource so that the information set by our
// Gateway is removed, our Gateway will not restore the status until the EventLoop invokes the StatusUpdater as a
// result of processing some other new change to a resource(s).
// FIXME(pleshakov): Make updater production ready
// https://github.com/nginxinc/nginx-gateway-fabric/issues/691

// UpdaterImpl needs to be modified to support new resources. Consider making UpdaterImpl extendable, so that it
// goes along the Open-closed principle.
type UpdaterImpl struct {
	lastStatuses lastStatuses
	cfg          UpdaterConfig
	isLeader     bool

	lock sync.Mutex
}

// lastStatuses hold the last saved statuses. Used when leader election is enabled to write the last saved statuses on
// a leader change.
type lastStatuses struct {
	nginxGateway *NginxGatewayStatus
	gatewayAPI   GatewayAPIStatuses
}

// Enable writes the last saved statuses for the Gateway API resources.
// Used in leader election when the Pod starts leading. It's possible that during a leader change,
// some statuses are missed. This will ensure that the latest statuses are written when a new leader takes over.
func (upd *UpdaterImpl) Enable(ctx context.Context) {
	defer upd.lock.Unlock()
	upd.lock.Lock()

	upd.isLeader = true

	upd.cfg.Logger.Info("Writing last statuses")
	upd.updateGatewayAPI(ctx, upd.lastStatuses.gatewayAPI)
	upd.updateNginxGateway(ctx, upd.lastStatuses.nginxGateway)
}

func (upd *UpdaterImpl) Disable() {
	defer upd.lock.Unlock()
	upd.lock.Lock()

	upd.isLeader = false
}

// NewUpdater creates a new Updater.
func NewUpdater(cfg UpdaterConfig) *UpdaterImpl {
	return &UpdaterImpl{
		cfg: cfg,
		// If leader election is enabled then we should not start running as a leader. Instead,
		// we wait for Enable to be invoked by the Leader Elector goroutine.
		isLeader: !cfg.LeaderElectionEnabled,
	}
}

func (upd *UpdaterImpl) Update(ctx context.Context, status Status) {
	// FIXME(pleshakov) Merge the new Conditions in the status with the existing Conditions
	// https://github.com/nginxinc/nginx-gateway-fabric/issues/558

	defer upd.lock.Unlock()
	upd.lock.Lock()

	switch s := status.(type) {
	case *NginxGatewayStatus:
		upd.updateNginxGateway(ctx, s)
	case GatewayAPIStatuses:
		upd.updateGatewayAPI(ctx, s)
	default:
		panic(fmt.Sprintf("unknown status type %T with group name %s", s, status.APIGroup()))
	}
}

func (upd *UpdaterImpl) updateNginxGateway(ctx context.Context, status *NginxGatewayStatus) {
	upd.lastStatuses.nginxGateway = status

	if !upd.isLeader {
		upd.cfg.Logger.Info("Skipping updating Nginx Gateway status because not leader")
		return
	}

	upd.cfg.Logger.Info("Updating Nginx Gateway status")

	if status != nil {
		statusUpdatedCh := make(chan struct{})

		go func() {
			upd.writeStatuses(ctx, status.NsName, &ngfAPI.NginxGateway{}, func(object client.Object) {
				ng := object.(*ngfAPI.NginxGateway)
				ng.Status = ngfAPI.NginxGatewayStatus{
					Conditions: convertConditions(
						status.Conditions,
						status.ObservedGeneration,
						upd.cfg.Clock.Now(),
					),
				}
			})
			statusUpdatedCh <- struct{}{}
		}()
		// Block here as the above goroutine runs. If the context gets canceled, we unblock and the method
		// returns. The goroutine continues but is canceled when the controller exits. If the goroutine
		// finishes and writes to the channel "statusUpdatedCh", we know that it is done updating
		// and this method exits.
		select {
		case <-ctx.Done():
		case <-statusUpdatedCh:
		}
	}
}

func (upd *UpdaterImpl) updateGatewayAPI(ctx context.Context, statuses GatewayAPIStatuses) {
	upd.lastStatuses.gatewayAPI = statuses

	if !upd.isLeader {
		upd.cfg.Logger.Info("Skipping updating Gateway API status because not leader")
		return
	}

	upd.cfg.Logger.Info("Updating Gateway API statuses")

	statusUpdatedCh := make(chan struct{})

	go func() {
		if upd.cfg.UpdateGatewayClassStatus {
			for nsname, gcs := range statuses.GatewayClassStatuses {
				upd.writeStatuses(ctx, nsname, &v1beta1.GatewayClass{}, func(object client.Object) {
					gc := object.(*v1beta1.GatewayClass)
					gc.Status = prepareGatewayClassStatus(gcs, upd.cfg.Clock.Now())
				},
				)
			}
		}

		for nsname, gs := range statuses.GatewayStatuses {
			upd.writeStatuses(ctx, nsname, &v1beta1.Gateway{}, func(object client.Object) {
				gw := object.(*v1beta1.Gateway)
				gw.Status = prepareGatewayStatus(gs, upd.cfg.PodIP, upd.cfg.Clock.Now())
			})
		}

		for nsname, rs := range statuses.HTTPRouteStatuses {
			upd.writeStatuses(ctx, nsname, &v1beta1.HTTPRoute{}, func(object client.Object) {
				hr := object.(*v1beta1.HTTPRoute)
				// statuses.GatewayStatus is never nil when len(statuses.HTTPRouteStatuses) > 0
				hr.Status = prepareHTTPRouteStatus(
					rs,
					upd.cfg.GatewayCtlrName,
					upd.cfg.Clock.Now(),
				)
			})
		}
		statusUpdatedCh <- struct{}{}
	}()
	// Block here as the above goroutine runs. If the context gets canceled, we unblock and the method
	// returns. The goroutine continues but is canceled when the controller exits. If the goroutine
	// finishes and writes to the channel "statusUpdatedCh", we know that it is done updating
	// and this method exits.
	select {
	case <-ctx.Done():
	case <-statusUpdatedCh:
	}

}

func (upd *UpdaterImpl) writeStatuses(
	ctx context.Context,
	nsname types.NamespacedName,
	obj client.Object,
	statusSetter func(client.Object),
) {
	// As a safety net, if the context is canceled we exit the function.
	if ctx.Err() != nil {
		return
	}

	// Using exponential backoff, 200 + 400 + 800 + 1600 ~ 3000ms total wait time.
	attempts := 5
	sleep := time.Millisecond * 200

	// The function handles errors by reporting them in the logs.
	// We need to get the latest version of the resource.
	// Otherwise, the Update status API call can fail.
	// Note: the default client uses a cache for reads, so we're not making an unnecessary API call here.
	// the default is configurable in the Manager options.
	if err := GetObj(ctx, attempts, sleep, obj, upd, nsname); err != nil {
		upd.cfg.Logger.Error(
			err,
			"Failed to get the recent version the resource when updating status",
			"namespace", nsname.Namespace,
			"name", nsname.Name,
			"kind", obj.GetObjectKind().GroupVersionKind().Kind)
		// If we can't get the resource, especially after retrying,
		// log the error and exit the function as we cannot continue.
		return
	}

	statusSetter(obj)

	if err := UpdateObjStatus(ctx, attempts, sleep, obj, upd); err != nil {
		upd.cfg.Logger.Error(
			err,
			"Failed to update status",
			"namespace", nsname.Namespace,
			"name", nsname.Name,
			"kind", obj.GetObjectKind().GroupVersionKind().Kind)
	}
}

func UpdateObjStatus(
	ctx context.Context,
	attempts int,
	sleep time.Duration,
	obj client.Object,
	upd *UpdaterImpl,
) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2
		}
		if err = upd.cfg.Client.Status().Update(ctx, obj); err == nil {
			return nil
		}
	}
	return err
}

func GetObj(
	ctx context.Context,
	attempts int,
	sleep time.Duration,
	obj client.Object,
	upd *UpdaterImpl,
	nsname types.NamespacedName,
) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2
		}
		// apierrors.IsNotFound(err) can happen when the resource is deleted, so no need to retry or return an error.
		if err = upd.cfg.Client.Get(ctx, nsname, obj); err == nil || apierrors.IsNotFound(err) {
			return nil
		}
	}
	return err
}
