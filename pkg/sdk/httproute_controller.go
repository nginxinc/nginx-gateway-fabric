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

type httpRouteReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   HTTPRouteImpl
}

// RegisterHTTPRouteController registers the HTTPRouteController in the manager.
func RegisterHTTPRouteController(mgr manager.Manager, impl HTTPRouteImpl) error {
	r := &httpRouteReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&v1beta1.HTTPRoute{}).
		Complete(r)
}

func (r *httpRouteReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("httpRoute", req.NamespacedName)

	log.V(3).Info("Reconciling HTTPRoute")

	found := true
	var hr v1beta1.HTTPRoute
	err := r.Get(ctx, req.NamespacedName, &hr)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get HTTPRoute")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		log.V(3).Info("Removing HTTPRoute")

		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	log.V(3).Info("Upserting HTTPRoute")

	r.impl.Upsert(&hr)
	return reconcile.Result{}, nil
}
