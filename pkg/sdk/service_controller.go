package sdk

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type serviceReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   ServiceImpl
}

// RegisterServiceController registers the ServiceController in the manager.
func RegisterServiceController(mgr manager.Manager, impl ServiceImpl) error {
	r := &serviceReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&apiv1.Service{}).
		Complete(r)
}

func (r *serviceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("service", req.NamespacedName)

	log.V(3).Info("Reconciling Service")

	found := true
	var svc apiv1.Service
	err := r.Get(ctx, req.NamespacedName, &svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get Service")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		log.V(3).Info("Removing Service")

		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	log.V(3).Info("Upserting Service")

	r.impl.Upsert(&svc)
	return reconcile.Result{}, nil
}
