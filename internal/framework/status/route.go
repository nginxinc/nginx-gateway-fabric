package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// prepareHTTPRouteStatus prepares the status for an HTTPRoute resource.
func prepareHTTPRouteStatus(
	oldStatus v1.HTTPRouteStatus,
	status RouteStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1.HTTPRouteStatus {
	return v1.HTTPRouteStatus{
		RouteStatus: v1.RouteStatus{
			Parents: buildRouteParentStatuses(oldStatus.Parents, status, gatewayCtlrName, transitionTime),
		},
	}
}

// prepareGRPCRouteStatus prepares the status for a GRPCRoute resource.
func prepareGRPCRouteStatus(
	oldStatus v1alpha2.GRPCRouteStatus,
	status RouteStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1alpha2.GRPCRouteStatus {
	return v1alpha2.GRPCRouteStatus{
		RouteStatus: v1.RouteStatus{
			Parents: buildRouteParentStatuses(oldStatus.Parents, status, gatewayCtlrName, transitionTime),
		},
	}
}

func buildRouteParentStatuses(
	oldParents []v1.RouteParentStatus,
	status RouteStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) []v1.RouteParentStatus {
	// maxParents is the max number of parent statuses which is the sum of all new parent statuses and all old parent
	// statuses.
	maxParents := len(status.ParentStatuses) + len(oldParents)
	parents := make([]v1.RouteParentStatus, 0, maxParents)

	// keep all the parent statuses that belong to other controllers
	for _, os := range oldParents {
		if string(os.ControllerName) != gatewayCtlrName {
			parents = append(parents, os)
		}
	}

	for _, ps := range status.ParentStatuses {
		// reassign the iteration variable inside the loop to fix implicit memory aliasing
		ps := ps
		p := v1.RouteParentStatus{
			ParentRef: v1.ParentReference{
				Namespace:   (*v1.Namespace)(&ps.GatewayNsName.Namespace),
				Name:        v1.ObjectName(ps.GatewayNsName.Name),
				SectionName: ps.SectionName,
			},
			ControllerName: v1.GatewayController(gatewayCtlrName),
			Conditions:     convertConditions(ps.Conditions, status.ObservedGeneration, transitionTime),
		}
		parents = append(parents, p)
	}
	return parents
}
