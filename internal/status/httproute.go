package status

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
)

// prepareHTTPRouteStatus prepares the status for an HTTPRoute resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple attached/not attached status per each parentRef.
// Extend support to cover more cases.
func prepareHTTPRouteStatus(
	status newstate.HTTPRouteStatus,
	gwNsName types.NamespacedName,
	gatewayCtlrName string,
	clock Clock,
) v1alpha2.HTTPRouteStatus {
	parents := make([]v1alpha2.RouteParentStatus, 0, len(status.ParentStatuses))

	// FIXME(pleshakov) Maintain the order from the HTTPRoute resource
	names := make([]string, 0, len(status.ParentStatuses))
	for name := range status.ParentStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		ps := status.ParentStatuses[name]

		var status, reason string
		if ps.Attached {
			status = "True"
			reason = "Attached"
		} else {
			status = "False"
			reason = "Not attached"
		}

		sectionName := name

		p := v1alpha2.RouteParentStatus{
			ParentRef: v1alpha2.ParentRef{
				Namespace:   (*v1alpha2.Namespace)(&gwNsName.Namespace),
				Name:        v1alpha2.ObjectName(gwNsName.Name),
				SectionName: (*v1alpha2.SectionName)(&sectionName),
			},
			ControllerName: v1alpha2.GatewayController(gatewayCtlrName),
			Conditions: []metav1.Condition{
				{
					Type:   string(v1alpha2.ConditionRouteAccepted),
					Status: metav1.ConditionStatus(status),
					// FIXME(pleshakov) Set the observed generation to the last processed generation of the HTTPRoute resource.
					ObservedGeneration: 123,
					LastTransitionTime: metav1.NewTime(clock.Now()),
					Reason:             reason,
					Message:            "", // FIXME(pleshakov): Figure out a good message
				},
			},
		}
		parents = append(parents, p)
	}

	return v1alpha2.HTTPRouteStatus{
		RouteStatus: v1alpha2.RouteStatus{
			Parents: parents,
		},
	}
}
