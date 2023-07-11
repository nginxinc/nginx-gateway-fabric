package status

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
)

// ListenerStatuses holds the statuses of listeners where the key is the name of a listener in the Gateway resource.
type ListenerStatuses map[string]ListenerStatus

// HTTPRouteStatuses holds the statuses of HTTPRoutes where the key is the namespaced name of an HTTPRoute.
type HTTPRouteStatuses map[types.NamespacedName]HTTPRouteStatus

// GatewayStatuses holds the statuses of Gateways where the key is the namespaced name of a Gateway.
type GatewayStatuses map[types.NamespacedName]GatewayStatus

// GatewayClassStatuses holds the statuses of GatewayClasses where the key is the namespaced name of a GatewayClass.
type GatewayClassStatuses map[types.NamespacedName]GatewayClassStatus

// Statuses holds the status-related information about Gateway API resources.
type Statuses struct {
	GatewayClassStatuses GatewayClassStatuses
	GatewayStatuses      GatewayStatuses
	HTTPRouteStatuses    HTTPRouteStatuses
}

// GatewayStatus holds the status of the winning Gateway resource.
type GatewayStatus struct {
	// ListenerStatuses holds the statuses of listeners defined on the Gateway.
	ListenerStatuses ListenerStatuses
	// Conditions is the list of conditions for this Gateway.
	Conditions []conditions.Condition
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

// ListenerStatus holds the status-related information about a listener in the Gateway resource.
type ListenerStatus struct {
	// Conditions is the list of conditions for this listener.
	Conditions []conditions.Condition
	// SupportedKinds is the list of SupportedKinds for this listener.
	SupportedKinds []v1beta1.RouteGroupKind
	// AttachedRoutes is the number of routes attached to the listener.
	AttachedRoutes int32
}

// HTTPRouteStatus holds the status-related information about an HTTPRoute resource.
type HTTPRouteStatus struct {
	// ParentStatuses holds the statuses for parentRefs of the HTTPRoute.
	ParentStatuses []ParentStatus
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

// ParentStatus holds status-related information related to how the HTTPRoute binds to a specific parentRef.
type ParentStatus struct {
	// GatewayNsName is the Namespaced name of the Gateway, which the parentRef references.
	GatewayNsName types.NamespacedName
	// SectionName is the SectionName of the parentRef.
	SectionName *v1beta1.SectionName
	// Conditions is the list of conditions that are relevant to the parentRef.
	Conditions []conditions.Condition
}

// GatewayClassStatus holds status-related information about the GatewayClass resource.
type GatewayClassStatus struct {
	Conditions         []conditions.Condition
	ObservedGeneration int64
}
