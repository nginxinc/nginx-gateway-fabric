package sdk

import (
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type gatewayClassReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   GatewayClassImpl
}

func RegisterGatewayClassController(mgr manager.Manager, impl GatewayClassImpl) error {
	r := &gatewayClassReconciler{
		impl:   impl,
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.GatewayClass{}).
		Complete(r)
}

func (r *gatewayClassReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("gatewayclass", req.Name)
	log.V(3).Info("Reconciling GatewayClass")

	var gc v1alpha2.GatewayClass
	found := true

	err := r.Get(ctx, req.NamespacedName, &gc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get GatewayClass")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	r.impl.Upsert(&gc)
	return reconcile.Result{}, nil
}
