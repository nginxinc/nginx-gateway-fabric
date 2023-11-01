package predicate

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

// NginxProxyPredicate implements a predicate function for an NginxProxy resource.
// This predicate will skip events for NginxProxies that aren't linked to the GatewayClass
// for this controller.
type NginxProxyPredicate struct {
	predicate.Funcs
	Client         client.Client
	ControllerName string
}

// Create implements default CreateEvent filter for validating an NginxProxy resource.
func (npp NginxProxyPredicate) Create(e event.CreateEvent) bool {
	if e.Object == nil {
		return false
	}

	np, ok := e.Object.(*ngfAPI.NginxProxy)
	if !ok {
		return false
	}

	return npp.validateNginxProxy(np)
}

// Update implements default UpdateEvent filter for validating an NginxProxy.
func (npp NginxProxyPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectNew == nil {
		return false
	}
	npNew, ok := e.ObjectNew.(*ngfAPI.NginxProxy)
	if !ok {
		return false
	}

	return npp.validateNginxProxy(npNew)
}

// Delete implements default DeleteEvent filter for validating an NginxProxy resource.
func (npp NginxProxyPredicate) Delete(e event.DeleteEvent) bool {
	if e.Object == nil {
		return false
	}

	np, ok := e.Object.(*ngfAPI.NginxProxy)
	if !ok {
		return false
	}

	return npp.validateNginxProxy(np)
}

// validateNginxProxy ensures that the NginxProxy resource being processed is the same one
// that is referenced in the GatewayClass for this controller.
func (npp NginxProxyPredicate) validateNginxProxy(np *ngfAPI.NginxProxy) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var gcList v1beta1.GatewayClassList
	if err := npp.Client.List(ctx, &gcList); err != nil {
		return false
	}

	for _, gc := range gcList.Items {
		if string(gc.Spec.ControllerName) == npp.ControllerName && gc.Spec.ParametersRef != nil {
			if gc.Spec.ParametersRef.Group == ngfAPI.GroupName &&
				gc.Spec.ParametersRef.Kind == v1beta1.Kind("NginxProxy") &&
				gc.Spec.ParametersRef.Name == np.Name &&
				gc.Spec.ParametersRef.Namespace != nil &&
				*gc.Spec.ParametersRef.Namespace == v1beta1.Namespace(np.Namespace) {
				return true
			}
		}
	}

	return false
}
