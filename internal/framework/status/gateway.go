package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// prepareGatewayStatus prepares the status for a Gateway resource.
func prepareGatewayStatus(
	gatewayStatus GatewayStatus,
	transitionTime metav1.Time,
) v1.GatewayStatus {
	listenerStatuses := make([]v1.ListenerStatus, 0, len(gatewayStatus.ListenerStatuses))

	for _, s := range gatewayStatus.ListenerStatuses {
		listenerStatuses = append(listenerStatuses, v1.ListenerStatus{
			Name:           s.Name,
			SupportedKinds: s.SupportedKinds,
			AttachedRoutes: s.AttachedRoutes,
			Conditions:     convertConditions(s.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
		})
	}

	return v1.GatewayStatus{
		Listeners:  listenerStatuses,
		Addresses:  gatewayStatus.Addresses,
		Conditions: convertConditions(gatewayStatus.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
	}
}
