package status

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// prepareHTTPRouteStatus prepares the status for an HTTPRoute resource.
// FIXME(pleshakov): Be compliant with in the Gateway API.
// Currently, we only support simple attached/not attached status per each parentRef.
// Extend support to cover more cases.
func prepareHTTPRouteStatus(
	status state.HTTPRouteStatus,
	gwNsName types.NamespacedName,
	gatewayCtlrName string,
	transitionTime metav1.Time,
) v1beta1.HTTPRouteStatus {
	parents := make([]v1beta1.RouteParentStatus, 0, len(status.ParentStatuses))

	// FIXME(pleshakov) Maintain the order from the HTTPRoute resource
	names := make([]string, 0, len(status.ParentStatuses))
	for name := range status.ParentStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		ps := status.ParentStatuses[name]

		sectionName := name

		p := v1beta1.RouteParentStatus{
			ParentRef: v1beta1.ParentReference{
				Namespace:   (*v1beta1.Namespace)(&gwNsName.Namespace),
				Name:        v1beta1.ObjectName(gwNsName.Name),
				SectionName: (*v1beta1.SectionName)(&sectionName),
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
