package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// prepareHTTPRouteStatus prepares the status for an HTTPRoute resource.
func prepareHTTPRouteStatus(
	oldStatus v1beta1.HTTPRouteStatus,
	status HTTPRouteStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1beta1.HTTPRouteStatus {
	// maxParents is the max number of parent statuses which is the sum of all new parent statuses and all old parent
	// statuses.
	maxParents := len(status.ParentStatuses) + len(oldStatus.Parents)
	parents := make([]v1beta1.RouteParentStatus, 0, maxParents)

	// keep all the parent statuses that belong to other controllers
	for _, os := range oldStatus.Parents {
		if string(os.ControllerName) != gatewayCtlrName {
			parents = append(parents, os)
		}
	}

	for _, ps := range status.ParentStatuses {
		// reassign the iteration variable inside the loop to fix implicit memory aliasing
		ps := ps
		p := v1beta1.RouteParentStatus{
			ParentRef: v1beta1.ParentReference{
				Namespace:   (*v1beta1.Namespace)(&ps.GatewayNsName.Namespace),
				Name:        v1beta1.ObjectName(ps.GatewayNsName.Name),
				SectionName: ps.SectionName,
			},
			ControllerName: v1beta1.GatewayController(gatewayCtlrName),
			Conditions:     convertConditions(ps.Conditions, status.ObservedGeneration, transitionTime),
		}
		parents = append(parents, p)
	}

	return v1beta1.HTTPRouteStatus{
		RouteStatus: v1beta1.RouteStatus{
			Parents: parents,
		},
	}
}
