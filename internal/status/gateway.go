package status

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

const (
	// GetawayReasonGatewayConflict indicates there are multiple Gateway resources for NGINX Gateway to choose from,
	// and NGINX Gateway ignored the resource in question and picked another Gateway as the winner.
	// NGINX Gateway will use this reason with GatewayConditionReady (false).
	GetawayReasonGatewayConflict v1alpha2.GatewayConditionReason = "GatewayConflict"

	// GatewayMessageGatewayConflict is message that describes GetawayReasonGatewayConflict.
	GatewayMessageGatewayConflict = "The resource is ignored due to a conflicting Gateway resource"
)

// prepareGatewayStatus prepares the status for a Gateway resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple valid/invalid status per each listener.
// Extend support to cover more cases.
func prepareGatewayStatus(gatewayStatus state.GatewayStatus, transitionTime metav1.Time) v1alpha2.GatewayStatus {
	listenerStatuses := make([]v1alpha2.ListenerStatus, 0, len(gatewayStatus.ListenerStatuses))

	// FIXME(pleshakov) Maintain the order from the Gateway resource
	names := make([]string, 0, len(gatewayStatus.ListenerStatuses))
	for name := range gatewayStatus.ListenerStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := gatewayStatus.ListenerStatuses[name]

		var (
			status metav1.ConditionStatus
			reason v1alpha2.ListenerConditionReason
		)

		if s.Valid {
			status = metav1.ConditionTrue
			reason = v1alpha2.ListenerReasonReady
		} else {
			status = metav1.ConditionFalse
			reason = v1alpha2.ListenerReasonInvalid
		}

		cond := metav1.Condition{
			Type:   string(v1alpha2.ListenerConditionReady),
			Status: status,
			// FIXME(pleshakov) Set the observed generation to the last processed generation of the Gateway resource.
			ObservedGeneration: 123,
			LastTransitionTime: transitionTime,
			Reason:             string(reason),
			Message:            "", // FIXME(pleshakov) Come up with a good message
		}

		listenerStatuses = append(listenerStatuses, v1alpha2.ListenerStatus{
			Name: v1alpha2.SectionName(name),
			SupportedKinds: []v1alpha2.RouteGroupKind{
				{
					Kind: "HTTPRoute", // FIXME(pleshakov) Set it based on the listener
				},
			},
			AttachedRoutes: s.AttachedRoutes,
			Conditions:     []metav1.Condition{cond},
		})
	}

	return v1alpha2.GatewayStatus{
		Listeners:  listenerStatuses,
		Conditions: nil, // FIXME(pleshakov) Create conditions for the Gateway resource.
	}
}

// prepareIgnoredGatewayStatus prepares the status for an ignored Gateway resource.
// TODO: is it reasonable to not set the listener statuses?
func prepareIgnoredGatewayStatus(status state.IgnoredGatewayStatus, transitionTime metav1.Time) v1alpha2.GatewayStatus {
	return v1alpha2.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(v1alpha2.GatewayConditionReady),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: status.ObservedGeneration,
				LastTransitionTime: transitionTime,
				Reason:             string(GetawayReasonGatewayConflict),
				Message:            GatewayMessageGatewayConflict,
			},
		},
	}
}
