package status

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	nkgAPI "github.com/nginxinc/nginx-kubernetes-gateway/apis/v1alpha1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Updater
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . LeaderElector

// LeaderElector reports whether the current Pod is the leader.
type LeaderElector interface {
	IsLeader() bool
}

// Updater updates statuses of the Gateway API resources.
type Updater interface {
	// Update updates the statuses of the resources.
	Update(context.Context, Statuses)
	// WriteLastStatuses writes the last statuses of the resources.
	WriteLastStatuses(ctx context.Context)
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
}

// UpdaterImpl updates statuses of the Gateway API resources.
//
// It has the following limitations:
//
// (1) It doesn't understand the leader election. Only the leader must report the statuses of the resources. Otherwise,
// multiple replicas will step on each other when trying to report statuses for the same resources.
//
// (2) It is not smart. It will update the status of a resource (make an API call) even if it hasn't changed.
//
// (3) It is synchronous, which means the status reporter can slow down the event loop.
// Consider the following cases:
// (a) Sometimes the Gateway will need to update statuses of all resources it handles, which could be ~1000. Making 1000
// status API calls sequentially will take time.
// (b) k8s API can become slow or even timeout. This will increase every update status API call.
// Making UpdaterImpl asynchronous will prevent it from adding variable delays to the event loop.
//
// (4) It doesn't retry on failures. This means there is a chance that some resources will not have up-to-do statuses.
// Statuses are important part of the Gateway API, so we need to ensure that the Gateway always keep the resources
// statuses up-to-date.
//
// (5) It doesn't clear the statuses of a resources that are no longer handled by the Gateway. For example, if
// an HTTPRoute resource no longer has the parentRef to the Gateway resources, the Gateway must update the status
// of the resource to remove the status about the removed parentRef.
//
// (6) If another controllers changes the status of the Gateway/HTTPRoute resource so that the information set by our
// Gateway is removed, our Gateway will not restore the status until the EventLoop invokes the StatusUpdater as a
// result of processing some other new change to a resource(s).
// FIXME(pleshakov): Make updater production ready
// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/691

// UpdaterImpl needs to be modified to support new resources. Consider making UpdaterImpl extendable, so that it
// goes along the Open-closed principle.
type UpdaterImpl struct {
	leaderElector LeaderElector
	lastStatuses  *Statuses
	cfg           UpdaterConfig

	statusLock sync.Mutex
}

// NewUpdater creates a new Updater.
func NewUpdater(cfg UpdaterConfig) *UpdaterImpl {
	return &UpdaterImpl{
		cfg: cfg,
	}
}

// SetLeaderElector sets the LeaderElector of the updater.
func (upd *UpdaterImpl) SetLeaderElector(elector LeaderElector) {
	upd.leaderElector = elector
}

// WriteLastStatuses writes the last saved statuses for the Gateway API resources.
// Used in leader election when the Pod starts leading. It's possible that during a leader change,
// some statuses are missed. This will ensure that the latest statuses are written when a new leader takes over.
func (upd *UpdaterImpl) WriteLastStatuses(ctx context.Context) {
	defer upd.statusLock.Unlock()
	upd.statusLock.Lock()

	if upd.lastStatuses == nil {
		upd.cfg.Logger.Info("No statuses to write")
		return
	}

	upd.cfg.Logger.Info("Writing last statuses")
	upd.update(ctx, *upd.lastStatuses)
}

func (upd *UpdaterImpl) Update(ctx context.Context, statuses Statuses) {
	// FIXME(pleshakov) Merge the new Conditions in the status with the existing Conditions
	// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/558

	defer upd.statusLock.Unlock()
	upd.statusLock.Lock()

	upd.lastStatuses = &statuses

	if !upd.shouldWriteStatus() {
		upd.cfg.Logger.Info("Skipping updating statuses because not leader")
		return
	}

	upd.cfg.Logger.Info("Updating statuses")
	upd.update(ctx, statuses)
}

func (upd *UpdaterImpl) update(ctx context.Context, statuses Statuses) {
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
		select {
		case <-ctx.Done():
			return
		default:
		}

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

	ngStatus := statuses.NginxGatewayStatus
	if ngStatus != nil {
		upd.writeStatuses(ctx, ngStatus.NsName, &nkgAPI.NginxGateway{}, func(object client.Object) {
			ng := object.(*nkgAPI.NginxGateway)
			ng.Status = nkgAPI.NginxGatewayStatus{
				Conditions: convertConditions(
					ngStatus.Conditions,
					ngStatus.ObservedGeneration,
					upd.cfg.Clock.Now(),
				),
			}
		})
	}
}

func (upd *UpdaterImpl) shouldWriteStatus() bool {
	return upd.leaderElector == nil || upd.leaderElector.IsLeader()
}

func (upd *UpdaterImpl) writeStatuses(
	ctx context.Context,
	nsname types.NamespacedName,
	obj client.Object,
	statusSetter func(client.Object),
) {
	// The function handles errors by reporting them in the logs.
	// We need to get the latest version of the resource.
	// Otherwise, the Update status API call can fail.
	// Note: the default client uses a cache for reads, so we're not making an unnecessary API call here.
	// the default is configurable in the Manager options.
	if err := upd.cfg.Client.Get(ctx, nsname, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			upd.cfg.Logger.Error(
				err,
				"Failed to get the recent version the resource when updating status",
				"namespace", nsname.Namespace,
				"name", nsname.Name,
				"kind", obj.GetObjectKind().GroupVersionKind().Kind)
		}
		return
	}

	statusSetter(obj)

	if err := upd.cfg.Client.Status().Update(ctx, obj); err != nil {
		upd.cfg.Logger.Error(
			err,
			"Failed to update status",
			"namespace", nsname.Namespace,
			"name", nsname.Name,
			"kind", obj.GetObjectKind().GroupVersionKind().Kind)
	}
}
