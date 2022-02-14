package status

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Updater

// Updater updates statuses of the Gateway API resources.
type Updater interface {
	// ProcessStatusUpdates updates the statuses according to the provided updates.
	ProcessStatusUpdates(context.Context, []state.StatusUpdate) error
}

// updaterImpl reports statuses for the Gateway API resources.
//
// It has the following limitations:
//
// (1) It doesn't understand the leader election. Only the leader must report the statuses of the resources. Otherwise,
// multiple replicas will step on each other when trying to report statuses for the same resources.
// TO-DO: address limitation (1)
//
// (2) It is synchronous, which means the status reporter can slow down the event loop.
// Consider the following cases:
// (a) Sometimes the Gateway will need to update statuses of all resources it handles, which could be ~1000. Making 1000
// status API calls sequentially will take time.
// (b) k8s API can become slow or even timeout. This will increase every update status API call.
// Making updaterImpl asynchronous will prevent it from adding variable delays to the event loop.
// TO-DO: address limitation (2)
//
// (3) It doesn't retry on failures. This means there is a chance that some resources will not have up-to-do statuses.
// Statuses are important part of the Gateway API, so we need to ensure that the Gateway always keep the resources
// statuses up-to-date.
// TO-DO: address limitation (3)
type updaterImpl struct {
	client client.Client
	logger logr.Logger
}

// NewUpdater creates a new Updater.
func NewUpdater(client client.Client, logger logr.Logger) Updater {
	return &updaterImpl{
		client: client,
		logger: logger.WithName("statusUpdater"),
	}
}

// ProcessStatusUpdates updates the statuses according to the provided updates.
func (upd *updaterImpl) ProcessStatusUpdates(ctx context.Context, updates []state.StatusUpdate) error {
	for _, u := range updates {

		switch s := u.Status.(type) {
		case *v1alpha2.HTTPRouteStatus:
			upd.logger.Info("Processing a status update for HTTPRoute",
				"namespace", u.NamespacedName.Namespace,
				"name", u.NamespacedName.Name)

			var hr v1alpha2.HTTPRoute

			upd.update(ctx, u.NamespacedName, &hr, func(object client.Object) {
				route := object.(*v1alpha2.HTTPRoute)
				// TO-DO: merge the conditions in the status with the conditions in the route.Status properly, because
				// right now, we are replacing the conditions.
				route.Status = *s
			})
		default:
			return fmt.Errorf("unknown status type %T", u.Status)
		}
	}

	return nil
}

func (upd *updaterImpl) update(ctx context.Context, nsname types.NamespacedName, obj client.Object, statusSetter func(client.Object)) {
	// The function handles errors by reporting them in the logs.
	// TO-DO: figure out appropriate log level for these errors. Perhaps 3?

	// We need to get the latest version of the resource.
	// Otherwise, the Update status API call can fail.
	// Note: the default client uses a cache for reads, so we're not making an unnecessary API call here.
	// the default is configurable in the Manager options.
	err := upd.client.Get(ctx, nsname, obj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			upd.logger.Error(err, "Failed to get the recent version the resource when updating status")
		}
		return
	}

	statusSetter(obj)

	err = upd.client.Status().Update(ctx, obj)
	if err != nil {
		upd.logger.Error(err, "Failed to update status")
	}
}
