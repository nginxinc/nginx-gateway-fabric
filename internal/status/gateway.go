package status

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
)

// prepareGatewayStatus prepares the status for a Gateway resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple valid/invalid status per each listener.
// Extend support to cover more cases.
func prepareGatewayStatus(statuses newstate.ListenerStatuses, clock Clock) v1alpha2.GatewayStatus {
	listenerStatuses := make([]v1alpha2.ListenerStatus, 0, len(statuses))

	// FIXME(pleshakov) Maintain the order from the Gateway resource
	names := make([]string, 0, len(statuses))
	for name := range statuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := statuses[name]

		var status, reason string

		if s.Valid {
			status = "True"
			reason = string(v1alpha2.ListenerReasonReady)
		} else {
			status = "False"
			reason = string(v1alpha2.ListenerReasonInvalid)
		}

		cond := metav1.Condition{
			Type:   string(v1alpha2.ListenerConditionReady),
			Status: metav1.ConditionStatus(status),
			// FIXME(pleshakov) Set the observed generation to the last processed generation of the Gateway resource.
			ObservedGeneration: 123,
			LastTransitionTime: metav1.NewTime(clock.Now()),
			Reason:             reason,
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
