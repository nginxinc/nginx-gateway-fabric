package sdk

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
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
		For(&v1alpha2.Gateway{}).
		Complete(r)
}

func (r *gatewayReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logr.FromContext(ctx).WithValues("gateway", req.Name)
	log.V(3).Info("Reconciling Gateway")

	var found bool = true
	var gw v1alpha2.Gateway
	err := r.Get(ctx, req.NamespacedName, &gw)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get Gateway")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		r.impl.Remove(req.Name)
		return reconcile.Result{}, nil
	}

	r.impl.Upsert(&gw)
	return reconcile.Result{}, nil
}
