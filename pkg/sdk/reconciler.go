package sdk

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler[T ObjectConstraint] struct {
	client.Client
	impl         Implementation[T]
	resourceKind string
}

var _ reconcile.Reconciler = &reconciler[client.Object]{}

func RegisterController[T ObjectConstraint](mgr manager.Manager, impl Implementation[T]) error {
	return RegisterControllerWithEventFilter(mgr, impl, nil)
}

// To handle new controllers from PR https://github.com/nginxinc/nginx-kubernetes-gateway/pull/221
func RegisterControllerWithEventFilter[T ObjectConstraint](
	mgr manager.Manager,
	impl Implementation[T],
	filter predicate.Predicate,
) error {
	var obj T

	r := &reconciler[T]{
		Client:       mgr.GetClient(),
		impl:         impl,
		resourceKind: reflect.TypeOf(obj).Elem().Name(),
	}

	obj = reflect.New(reflect.TypeOf(obj).Elem()).Interface().(T)

	builder := ctlr.NewControllerManagedBy(mgr).For(obj)

	if filter != nil {
		builder = builder.WithEventFilter(filter)
	}

	return builder.Complete(r)
}

func (r *reconciler[T]) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	log.Info(fmt.Sprintf("Reconciling %s", r.resourceKind))

	found := true

	var obj T
	obj = reflect.New(reflect.TypeOf(obj).Elem()).Interface().(T)

	err := r.Get(ctx, req.NamespacedName, obj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Failed to get %s", r.resourceKind))
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	r.impl.Upsert(obj)
	return reconcile.Result{}, nil
}
