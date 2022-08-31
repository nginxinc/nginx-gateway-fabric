package sdk

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type serviceReconciler struct {
	client.Client
	scheme *runtime.Scheme
	impl   ServiceImpl
}

// ServicePortsChangedPredicate implements an update predicate function based on the Ports of a Service.
// This predicate will skip update events that have no change in the Service Ports and TargetPorts.
type ServicePortsChangedPredicate struct {
	predicate.Funcs
}

// ports contains the ports that the Gateway cares about.
type ports struct {
	servicePort int32
	targetPort  intstr.IntOrString
}

// Update implements default UpdateEvent filter for validating Service port changes.
func (ServicePortsChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		return false
	}
	if e.ObjectNew == nil {
		return false
	}

	oldSvc, ok := e.ObjectOld.(*apiv1.Service)
	if !ok {
		return false
	}

	newSvc, ok := e.ObjectNew.(*apiv1.Service)
	if !ok {
		return false
	}

	oldPorts := oldSvc.Spec.Ports
	newPorts := newSvc.Spec.Ports

	if len(oldPorts) != len(newPorts) {
		return true
	}

	oldPortSet := make(map[ports]struct{})
	newPortSet := make(map[ports]struct{})

	for i := 0; i < len(oldSvc.Spec.Ports); i++ {
		oldPortSet[ports{servicePort: oldPorts[i].Port, targetPort: oldPorts[i].TargetPort}] = struct{}{}
		newPortSet[ports{servicePort: newPorts[i].Port, targetPort: newPorts[i].TargetPort}] = struct{}{}
	}

	for pd := range oldPortSet {
		if _, exists := newPortSet[pd]; exists {
			delete(newPortSet, pd)
		} else {
			return true
		}
	}

	return len(newPortSet) > 0
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
		WithEventFilter(ServicePortsChangedPredicate{}).
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
