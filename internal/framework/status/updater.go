package status

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Updater

// Updater updates statuses of the Gateway API resources.
// Updater can be disabled. In this case, it will stop updating the statuses of resources, while
// always saving the statuses of the last Update call. This is used to support multiple replicas of
// control plane being able to run simultaneously where only the leader will update statuses.
type Updater interface {
	// Update updates the statuses of the resources.
	Update(context.Context, Status)
	// UpdateAddresses updates the Gateway Addresses when the Gateway Service changes.
	UpdateAddresses(context.Context, []v1beta1.GatewayStatusAddress)
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
// (1) It is synchronous, which means the status reporter can slow down the event loop.
// Consider the following cases:
// (a) Sometimes the Gateway will need to update statuses of all resources it handles, which could be ~1000. Making 1000
// status API calls sequentially will take time.
// (b) k8s API can become slow or even timeout. This will increase every update status API call.
// Making UpdaterImpl asynchronous will prevent it from adding variable delays to the event loop.
//
// (2) It doesn't clear the statuses of a resources that are no longer handled by the Gateway. For example, if
// an HTTPRoute resource no longer has the parentRef to the Gateway resources, the Gateway must update the status
// of the resource to remove the status about the removed parentRef.
//
// (3) If another controllers changes the status of the Gateway/HTTPRoute resource so that the information set by our
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
	nginxProxy   *NginxProxyStatus
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
	upd.updateNginxProxy(ctx, upd.lastStatuses.nginxProxy)
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
	case *NginxProxyStatus:
		upd.updateNginxProxy(ctx, s)
	case GatewayAPIStatuses:
		upd.updateGatewayAPI(ctx, s)
	default:
		panic(fmt.Sprintf("unknown status type %T with group name %s", s, status.APIGroup()))
	}
}

func (upd *UpdaterImpl) updateNginxGateway(ctx context.Context, status *NginxGatewayStatus) {
	upd.lastStatuses.nginxGateway = status

	if !upd.isLeader {
		upd.cfg.Logger.Info("Skipping updating NginxGateway status because not leader")
		return
	}

	if status != nil {
		upd.cfg.Logger.Info("Updating NginxGateway status")

		upd.writeStatuses(
			ctx,
			status.NsName,
			&ngfAPI.NginxGateway{},
			newNginxGatewayStatusSetter(upd.cfg.Clock, *status),
		)
	}
}

func (upd *UpdaterImpl) updateNginxProxy(ctx context.Context, status *NginxProxyStatus) {
	upd.lastStatuses.nginxProxy = status

	if !upd.isLeader {
		upd.cfg.Logger.Info("Skipping updating NginxProxy status because not leader")
		return
	}

	if status != nil {
		upd.cfg.Logger.Info("Updating NginxProxy status")

		upd.writeStatuses(
			ctx,
			status.NsName,
			&ngfAPI.NginxProxy{},
			newNginxProxyStatusSetter(upd.cfg.Clock, *status),
		)
	}
}

func (upd *UpdaterImpl) updateGatewayAPI(ctx context.Context, statuses GatewayAPIStatuses) {
	upd.lastStatuses.gatewayAPI = statuses

	if !upd.isLeader {
		upd.cfg.Logger.Info("Skipping updating Gateway API status because not leader")
		return
	}

	upd.cfg.Logger.Info("Updating Gateway API statuses")

	if upd.cfg.UpdateGatewayClassStatus {
		for nsname, gcs := range statuses.GatewayClassStatuses {
			select {
			case <-ctx.Done():
				return
			default:
			}

			upd.writeStatuses(ctx, nsname, &v1beta1.GatewayClass{}, newGatewayClassStatusSetter(upd.cfg.Clock, gcs))
		}
	}

	for nsname, gs := range statuses.GatewayStatuses {
		select {
		case <-ctx.Done():
			return
		default:
		}

		upd.writeStatuses(ctx, nsname, &v1beta1.Gateway{}, newGatewayStatusSetter(upd.cfg.Clock, gs))
	}

	for nsname, rs := range statuses.HTTPRouteStatuses {
		select {
		case <-ctx.Done():
			return
		default:
		}

		upd.writeStatuses(
			ctx,
			nsname,
			&v1beta1.HTTPRoute{},
			newHTTPRouteStatusSetter(upd.cfg.GatewayCtlrName, upd.cfg.Clock, rs),
		)
	}
}

func (upd *UpdaterImpl) writeStatuses(
	ctx context.Context,
	nsname types.NamespacedName,
	obj client.Object,
	statusSetter setter,
) {
	err := wait.ExponentialBackoffWithContext(
		ctx,
		wait.Backoff{
			Duration: time.Millisecond * 200,
			Factor:   2,
			Jitter:   0.5,
			Steps:    4,
			Cap:      time.Millisecond * 3000,
		},
		// Function returns true if the condition is satisfied, or an error if the loop should be aborted.
		NewRetryUpdateFunc(upd.cfg.Client, upd.cfg.Client.Status(), nsname, obj, upd.cfg.Logger, statusSetter),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		upd.cfg.Logger.Error(
			err,
			"Failed to update status",
			"namespace", nsname.Namespace,
			"name", nsname.Name,
			"kind", obj.GetObjectKind().GroupVersionKind().Kind)
	}
}

// UpdateAddresses is called when the Gateway Status needs its addresses updated.
func (upd *UpdaterImpl) UpdateAddresses(ctx context.Context, addresses []v1beta1.GatewayStatusAddress) {
	defer upd.lock.Unlock()
	upd.lock.Lock()

	for name, status := range upd.lastStatuses.gatewayAPI.GatewayStatuses {
		if status.Ignored {
			continue
		}
		status.Addresses = addresses
		upd.lastStatuses.gatewayAPI.GatewayStatuses[name] = status
	}

	upd.updateGatewayAPI(ctx, upd.lastStatuses.gatewayAPI)
}

// NewRetryUpdateFunc returns a function which will be used in wait.ExponentialBackoffWithContext.
// The function will attempt to Update a kubernetes resource and will be retried in
// wait.ExponentialBackoffWithContext if an error occurs. Exported for testing purposes.
//
// wait.ExponentialBackoffWithContext will retry if this function returns nil as its error,
// which is what we want if we encounter an error from the functions we call. However,
// the linter will complain if we return nil if an error was found.
//
//nolint:nilerr
func NewRetryUpdateFunc(
	getter controller.Getter,
	updater K8sUpdater,
	nsname types.NamespacedName,
	obj client.Object,
	logger logr.Logger,
	statusSetter func(client.Object) bool,
) func(ctx context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		// The function handles errors by reporting them in the logs.
		// We need to get the latest version of the resource.
		// Otherwise, the Update status API call can fail.
		// Note: the default client uses a cache for reads, so we're not making an unnecessary API call here.
		// the default is configurable in the Manager options.
		if err := getter.Get(ctx, nsname, obj); err != nil {
			// apierrors.IsNotFound(err) can happen when the resource is deleted,
			// so no need to retry or return an error.
			if apierrors.IsNotFound(err) {
				return true, nil
			}

			logger.V(1).Info(
				"Encountered error when getting resource to update status",
				"error", err,
				"namespace", nsname.Namespace,
				"name", nsname.Name,
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
			)

			return false, nil
		}

		if !statusSetter(obj) {
			logger.V(1).Info(
				"Skipping status update because there's no change",
				"namespace", nsname.Namespace,
				"name", nsname.Name,
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
			)

			return true, nil
		}

		if err := updater.Update(ctx, obj); err != nil {
			logger.V(1).Info(
				"Encountered error updating status",
				"error", err,
				"namespace", nsname.Namespace,
				"name", nsname.Name,
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
			)

			return false, nil
		}

		return true, nil
	}
}
