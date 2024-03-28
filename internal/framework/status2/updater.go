package status2

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller"
)

// K8sUpdater updates a resource from the k8s API.
// It allows us to mock the client.Reader.Status.Update method.
type K8sUpdater interface {
	// Update is from client.StatusClient.SubResourceWriter.
	Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error
}

type UpdateRequest struct {
	NsName       types.NamespacedName
	ResourceType client.Object
	Setter       Setter
}

type Setter func(client.Object) bool

type Updater struct {
	client client.Client
	logger logr.Logger
}

func NewUpdater(client client.Client, logger logr.Logger) *Updater {
	return &Updater{
		client: client,
		logger: logger,
	}
}

func (u *Updater) Update(ctx context.Context, reqs ...UpdateRequest) {
	for _, r := range reqs {
		u.writeStatuses(ctx, r.NsName, r.ResourceType, r.Setter)
	}
}

func (u *Updater) writeStatuses(
	ctx context.Context,
	nsname types.NamespacedName,
	obj client.Object,
	statusSetter Setter,
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
		NewRetryUpdateFunc(u.client, u.client.Status(), nsname, obj, u.logger, statusSetter),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		u.logger.Error(
			err,
			"Failed to update status",
			"namespace", nsname.Namespace,
			"name", nsname.Name,
			"kind", obj.GetObjectKind().GroupVersionKind().Kind)
	}
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

type GroupUpdateRequest struct {
	Name    string
	Request []UpdateRequest
}

type CachingGroupUpdater struct {
	updater *Updater
	lock    *sync.Mutex
	groups  map[string]GroupUpdateRequest
	enabled bool
}

func NewCachingGroupUpdater(updater *Updater) *CachingGroupUpdater {
	return &CachingGroupUpdater{
		updater: updater,
		lock:    &sync.Mutex{},
		groups:  make(map[string]GroupUpdateRequest),
	}
}

func (u *CachingGroupUpdater) Update(ctx context.Context, update GroupUpdateRequest) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if len(update.Request) == 0 {
		delete(u.groups, update.Name)
	}

	u.groups[update.Name] = update

	if !u.enabled {
		return
	}

	u.updater.Update(ctx, update.Request...)
}

func (u *CachingGroupUpdater) Enable(ctx context.Context) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.enabled = true

	for _, update := range u.groups {
		u.updater.Update(ctx, update.Request...)
	}
}

func ConditionsEqual(prev, cur []v1.Condition) bool {
	return slices.EqualFunc(prev, cur, func(c1, c2 v1.Condition) bool {
		if c1.ObservedGeneration != c2.ObservedGeneration {
			return false
		}

		if c1.Type != c2.Type {
			return false
		}

		if c1.Status != c2.Status {
			return false
		}

		if c1.Message != c2.Message {
			return false
		}

		return c1.Reason == c2.Reason
	})
}

func ConvertConditions(
	conds []conditions.Condition,
	observedGeneration int64,
	transitionTime v1.Time,
) []v1.Condition {
	apiConds := make([]v1.Condition, len(conds))

	for i := range conds {
		apiConds[i] = v1.Condition{
			Type:               conds[i].Type,
			Status:             conds[i].Status,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             conds[i].Reason,
			Message:            conds[i].Message,
		}
	}

	return apiConds
}
