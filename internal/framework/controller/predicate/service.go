package predicate

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ServicePortsChangedPredicate implements an update predicate function based on the Ports of a Service.
// This predicate will skip update events that have no change in the Service Ports and TargetPorts.
type ServicePortsChangedPredicate struct {
	predicate.Funcs
}

// ports contains the ports that the Gateway cares about.
type ports struct {
	targetPort  intstr.IntOrString
	servicePort int32
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

	for i := range len(oldSvc.Spec.Ports) {
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

// GatewayServicePredicate implements predicate functions for this Pod's Service.
type GatewayServicePredicate struct {
	predicate.Funcs
	NSName types.NamespacedName
}

// Update implements the default UpdateEvent filter for the Gateway Service.
func (gsp GatewayServicePredicate) Update(e event.UpdateEvent) bool {
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

	if client.ObjectKeyFromObject(newSvc) != gsp.NSName {
		return false
	}

	if oldSvc.Spec.Type != newSvc.Spec.Type {
		return true
	}

	if newSvc.Spec.Type == apiv1.ServiceTypeLoadBalancer {
		oldIngress := oldSvc.Status.LoadBalancer.Ingress
		newIngress := newSvc.Status.LoadBalancer.Ingress

		if len(oldIngress) != len(newIngress) {
			return true
		}

		for i, ingress := range oldIngress {
			if ingress.IP != newIngress[i].IP || ingress.Hostname != newIngress[i].Hostname {
				return true
			}
		}
	}

	return false
}
