package status

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// prepareGatewayStatus prepares the status for a Gateway resource.
func prepareGatewayStatus(
	gatewayStatus state.GatewayStatus,
	podIP string,
	transitionTime metav1.Time,
) v1beta1.GatewayStatus {
	listenerStatuses := make([]v1beta1.ListenerStatus, 0, len(gatewayStatus.ListenerStatuses))

	// FIXME(pleshakov) Maintain the order from the Gateway resource
	// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/689
	names := make([]string, 0, len(gatewayStatus.ListenerStatuses))
	for name := range gatewayStatus.ListenerStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := gatewayStatus.ListenerStatuses[name]

		listenerStatuses = append(listenerStatuses, v1beta1.ListenerStatus{
			Name: v1beta1.SectionName(name),
			SupportedKinds: []v1beta1.RouteGroupKind{
				{
					Kind: "HTTPRoute", // FIXME(pleshakov) Set it based on the listener https://github.com/nginxinc/nginx-kubernetes-gateway/issues/690
				},
			},
			AttachedRoutes: s.AttachedRoutes,
			Conditions:     convertConditions(s.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
		})
	}

	ipAddrType := v1beta1.IPAddressType
	gwPodIP := v1beta1.GatewayAddress{
		Type:  &ipAddrType,
		Value: podIP,
	}

	return v1beta1.GatewayStatus{
		Listeners:  listenerStatuses,
		Addresses:  []v1beta1.GatewayAddress{gwPodIP},
		Conditions: convertConditions(gatewayStatus.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
	}
}
