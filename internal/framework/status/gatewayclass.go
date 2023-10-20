package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// prepareGatewayClassStatus prepares the status for the GatewayClass resource.
func prepareGatewayClassStatus(status GatewayClassStatus, transitionTime metav1.Time) v1.GatewayClassStatus {
	return v1.GatewayClassStatus{
		Conditions: convertConditions(status.Conditions, status.ObservedGeneration, transitionTime),
	}
}
