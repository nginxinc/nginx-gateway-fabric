package sdk

import (
	"context"

	discoveryV1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type endpointSliceReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   EndpointSliceImpl
}

// RegisterEndpointSliceController registers the EndpointSliceController in the manager.
func RegisterEndpointSliceController(mgr manager.Manager, impl EndpointSliceImpl) error {
	r := &endpointSliceReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&discoveryV1.EndpointSlice{}).
		Complete(r)
}

func (r *endpointSliceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("endpointslice", req.NamespacedName)

	log.V(3).Info("Reconciling Endpoint Slice")

	found := true
	var endpSlice discoveryV1.EndpointSlice
	err := r.Get(ctx, req.NamespacedName, &endpSlice)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get Endpoint Slice")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		log.V(3).Info("Removing Endpoint Slice")

		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	log.V(3).Info("Upserting Endpoint Slice")

	r.impl.Upsert(&endpSlice)
	return reconcile.Result{}, nil
}
