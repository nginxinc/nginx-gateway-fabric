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

type reconciler struct {
	client.Client
	impl         Implementation
	objectType   client.Object
	resourceKind string
}

var _ reconcile.Reconciler = &reconciler{}

func RegisterController(mgr manager.Manager, impl Implementation, objectType client.Object) error {
	return RegisterControllerWithEventFilter(mgr, impl, objectType, nil)
}

// To handle new controllers from PR https://github.com/nginxinc/nginx-kubernetes-gateway/pull/221
func RegisterControllerWithEventFilter(
	mgr manager.Manager,
	impl Implementation,
	objectType client.Object,
	filter predicate.Predicate,
) error {

	r := &reconciler{
		Client:       mgr.GetClient(),
		impl:         impl,
		objectType:   objectType,
		resourceKind: reflect.TypeOf(objectType).Elem().Name(),
	}

	builder := ctlr.NewControllerManagedBy(mgr).For(objectType)

	if filter != nil {
		builder = builder.WithEventFilter(filter)
	}

	return builder.Complete(r)
}

func newObject(objectType client.Object) client.Object {
	// without Elem(), t will be a pointer to the type. For example, *v1beta1.Gateway, not v1beta1.Gateway
	t := reflect.TypeOf(objectType).Elem()

	return reflect.New(t).Interface().(client.Object)
}

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	log.Info(fmt.Sprintf("Reconciling %s", r.resourceKind))

	found := true
	obj := newObject(r.objectType)
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
