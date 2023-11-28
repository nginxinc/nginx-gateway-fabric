package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// GatewayClassPredicate implements a predicate function based on the controllerName of a GatewayClass.
// This predicate will skip events for GatewayClasses that don't reference this controller.
type GatewayClassPredicate struct {
	predicate.Funcs
	ControllerName string
}

// Create implements default CreateEvent filter for validating a GatewayClass controllerName.
func (gcp GatewayClassPredicate) Create(e event.CreateEvent) bool {
	if e.Object == nil {
		return false
	}

	gc, ok := e.Object.(*v1.GatewayClass)
	if !ok {
		return false
	}

	return string(gc.Spec.ControllerName) == gcp.ControllerName
}

// Update implements default UpdateEvent filter for validating a GatewayClass controllerName.
func (gcp GatewayClassPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld != nil {
		gcOld, ok := e.ObjectOld.(*v1.GatewayClass)
		if ok && string(gcOld.Spec.ControllerName) == gcp.ControllerName {
			return true
		}
	}

	if e.ObjectNew != nil {
		gcNew, ok := e.ObjectNew.(*v1.GatewayClass)
		if ok && string(gcNew.Spec.ControllerName) == gcp.ControllerName {
			return true
		}
	}

	return false
}
