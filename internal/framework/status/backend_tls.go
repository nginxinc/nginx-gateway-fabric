package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// prepareBackendTLSPolicyStatus prepares the status for a BackendTLSPolicy resource.
func prepareBackendTLSPolicyStatus(
	oldStatus v1alpha2.PolicyStatus,
	status BackendTLSPolicyStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1alpha2.PolicyStatus {
	// maxAncestors is the max number of ancestor statuses which is the sum of all new ancestor statuses and all old
	// ancestor statuses.
	maxAncestors := len(status.AncestorStatuses) + len(oldStatus.Ancestors)
	ancestors := make([]v1alpha2.PolicyAncestorStatus, 0, maxAncestors)

	// keep all the ancestor statuses that belong to other controllers
	for _, os := range oldStatus.Ancestors {
		if string(os.ControllerName) != gatewayCtlrName {
			ancestors = append(ancestors, os)
		}
	}

	for _, as := range status.AncestorStatuses {
		// reassign the iteration variable inside the loop to fix implicit memory aliasing
		as := as
		a := v1alpha2.PolicyAncestorStatus{
			AncestorRef: v1.ParentReference{
				Namespace: (*v1.Namespace)(&as.GatewayNsName.Namespace),
				Name:      v1alpha2.ObjectName(as.GatewayNsName.Name),
			},
			ControllerName: v1alpha2.GatewayController(gatewayCtlrName),
			Conditions:     convertConditions(as.Conditions, status.ObservedGeneration, transitionTime),
		}
		ancestors = append(ancestors, a)
	}

	return v1alpha2.PolicyStatus{
		Ancestors: ancestors,
	}
}
