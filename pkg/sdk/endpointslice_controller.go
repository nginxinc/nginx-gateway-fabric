package sdk

import (
	"context"
	"fmt"
	"time"

	discoveryV1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// KubernetesServiceNameIndexField is the name of the Index Field used to index EndpointSlices by their service
	// owners.
	KubernetesServiceNameIndexField = "k8sServiceName"
	// KubernetesServiceNameLabel is the label used to identify the Kubernetes service name on an EndpointSlice.
	KubernetesServiceNameLabel = "kubernetes.io/service-name"
	// addIndexFieldTimeout is the timeout used for adding an Index Field to the EndpointSlice cache.
	addIndexFieldTimeout = 2 * time.Minute
)

type endpointSliceReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   EndpointSliceImpl
}

// ServiceNameIndexFunc is a client.IndexerFunc that parses a Kubernetes object and returns the value of the
// Kubernetes service-name label.
// Used to index EndpointSlices by their service owners.
func ServiceNameIndexFunc(obj client.Object) []string {
	slice, ok := obj.(*discoveryV1.EndpointSlice)
	if !ok {
		panic(fmt.Sprintf("expected an EndpointSlice; got %T", obj))
	}

	name := GetServiceNameFromEndpointSlice(slice)
	if name == "" {
		return nil
	}

	return []string{name}
}

// GetServiceNameFromEndpointSlice returns the value of the Kubernetes service-name label from an EndpointSlice.
func GetServiceNameFromEndpointSlice(slice *discoveryV1.EndpointSlice) string {
	if slice.Labels == nil {
		return ""
	}

	return slice.Labels[KubernetesServiceNameLabel]
}

// RegisterEndpointSliceController registers the EndpointSliceController in the manager.
func RegisterEndpointSliceController(ctx context.Context, mgr manager.Manager, impl EndpointSliceImpl) error {
	r := &endpointSliceReconciler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		impl:   impl,
	}

	ctx, cancel := context.WithTimeout(ctx, addIndexFieldTimeout)
	defer cancel()

	err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&discoveryV1.EndpointSlice{},
		KubernetesServiceNameIndexField,
		ServiceNameIndexFunc,
	)
	if err != nil {
		return fmt.Errorf("failed to add service name index for EndpointSlices: %w", err)
	}

	return ctlr.NewControllerManagedBy(mgr).
		For(&discoveryV1.EndpointSlice{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *endpointSliceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("endpointslice", req.NamespacedName)

	log.V(3).Info("Reconciling EndpointSlice")

	found := true
	var endpSlice discoveryV1.EndpointSlice
	err := r.Get(ctx, req.NamespacedName, &endpSlice)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get EndpointSlice")
			return reconcile.Result{}, err
		}
		found = false
	}

	if !found {
		log.V(3).Info("Removing EndpointSlice")

		r.impl.Remove(req.NamespacedName)
		return reconcile.Result{}, nil
	}

	log.V(3).Info("Upserting EndpointSlice")

	r.impl.Upsert(&endpSlice)
	return reconcile.Result{}, nil
}
