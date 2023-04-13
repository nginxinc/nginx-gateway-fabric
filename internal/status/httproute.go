package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// prepareHTTPRouteStatus prepares the status for an HTTPRoute resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple attached/not attached status per each parentRef.
// Extend support to cover more cases.
func prepareHTTPRouteStatus(
	status state.HTTPRouteStatus,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1beta1.HTTPRouteStatus {
	parents := make([]v1beta1.RouteParentStatus, 0, len(status.ParentStatuses))

	for _, ps := range status.ParentStatuses {
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
