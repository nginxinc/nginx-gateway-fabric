package status

import (
	"os"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

const (
	// GetawayReasonGatewayConflict indicates there are multiple Gateway resources for NGINX Gateway to choose from,
	// and NGINX Gateway ignored the resource in question and picked another Gateway as the winner.
	// NGINX Gateway will use this reason with GatewayConditionReady (false).
	GetawayReasonGatewayConflict v1beta1.GatewayConditionReason = "GatewayConflict"

	// GatewayMessageGatewayConflict is message that describes GetawayReasonGatewayConflict.
	GatewayMessageGatewayConflict = "The resource is ignored due to a conflicting Gateway resource"
)

// prepareGatewayStatus prepares the status for a Gateway resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple valid/invalid status per each listener.
// Extend support to cover more cases.
func prepareGatewayStatus(gatewayStatus state.GatewayStatus, transitionTime metav1.Time) v1beta1.GatewayStatus {
	listenerStatuses := make([]v1beta1.ListenerStatus, 0, len(gatewayStatus.ListenerStatuses))

	// FIXME(pleshakov) Maintain the order from the Gateway resource
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
					Kind: "HTTPRoute", // FIXME(pleshakov) Set it based on the listener
				},
			},
			AttachedRoutes: s.AttachedRoutes,
			Conditions:     convertConditions(s.Conditions, gatewayStatus.ObservedGeneration, transitionTime),
		})
	}

	ipAddrType := v1beta1.IPAddressType
	podIP := v1beta1.GatewayAddress{
		Type:  &ipAddrType,
		Value: os.Getenv("POD_IP"),
	}

	return v1beta1.GatewayStatus{
		Listeners:  listenerStatuses,
		Addresses:  []v1beta1.GatewayAddress{podIP},
		Conditions: nil, // FIXME(pleshakov) Create conditions for the Gateway resource.
	}
}

// prepareIgnoredGatewayStatus prepares the status for an ignored Gateway resource.
// TODO: is it reasonable to not set the listener statuses?
// FIXME(pleshakov): Simplify the code, so that we don't need a separate prepareIgnoredGatewayStatus
// and state.IgnoredGatewayStatus.
func prepareIgnoredGatewayStatus(status state.IgnoredGatewayStatus, transitionTime metav1.Time) v1beta1.GatewayStatus {
	return v1beta1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(v1beta1.GatewayConditionReady),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: status.ObservedGeneration,
				LastTransitionTime: transitionTime,
				Reason:             string(GetawayReasonGatewayConflict),
				Message:            GatewayMessageGatewayConflict,
			},
		},
	}
}
