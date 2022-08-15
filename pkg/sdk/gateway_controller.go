package sdk

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

type gatewayReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   GatewayImpl
}

func RegisterGatewayController(mgr manager.Manager, impl GatewayImpl) error {
	r := &gatewayReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&v1beta1.Gateway{}).
		Complete(r)
}

func (r *gatewayReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("gateway", req.Name)
	log.V(3).Info("Reconciling Gateway")

	found := true
	var gw v1beta1.Gateway
	err := r.Get(ctx, req.NamespacedName, &gw)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get Gateway")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	r.impl.Upsert(&gw)
	return reconcile.Result{}, nil
}
