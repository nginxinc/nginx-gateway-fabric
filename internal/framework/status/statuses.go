package status

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
)

// Status is the status of one or more Kubernetes resources that the StatusUpdater will update.
type Status interface {
	// APIGroup returns the GroupName of the resources contained in the status
	APIGroup() string
}

// GatewayAPIStatuses holds the status-related information about Gateway API resources.
type GatewayAPIStatuses struct {
	GatewayClassStatuses GatewayClassStatuses
	GatewayStatuses      GatewayStatuses
	HTTPRouteStatuses    HTTPRouteStatuses
}

func (g GatewayAPIStatuses) APIGroup() string {
	return v1beta1.GroupName
}

// NginxGatewayStatus holds status-related information about the NginxGateway resource.
type NginxGatewayStatus struct {
	// NsName is the NamespacedName of the NginxGateway resource.
	NsName types.NamespacedName
	// Conditions is the list of conditions for this NginxGateway.
	Conditions []conditions.Condition
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

func (n *NginxGatewayStatus) APIGroup() string {
	return ngfAPI.GroupName
}

// NginxProxyStatus holds status-related information about the NginxProxy resource.
type NginxProxyStatus struct {
	// NsName is the NamespacedName of the NginxProxy resource.
	NsName types.NamespacedName
	// Conditions is the list of conditions for this NginxProxy.
	Conditions []conditions.Condition
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

func (n *NginxProxyStatus) APIGroup() string {
	return ngfAPI.GroupName
}

// ListenerStatuses holds the statuses of listeners where the key is the name of a listener in the Gateway resource.
type ListenerStatuses map[string]ListenerStatus

// HTTPRouteStatuses holds the statuses of HTTPRoutes where the key is the namespaced name of an HTTPRoute.
type HTTPRouteStatuses map[types.NamespacedName]HTTPRouteStatus

// GatewayStatuses holds the statuses of Gateways where the key is the namespaced name of a Gateway.
type GatewayStatuses map[types.NamespacedName]GatewayStatus

// GatewayClassStatuses holds the statuses of GatewayClasses where the key is the namespaced name of a GatewayClass.
type GatewayClassStatuses map[types.NamespacedName]GatewayClassStatus

// GatewayStatus holds the status of the winning Gateway resource.
type GatewayStatus struct {
	// ListenerStatuses holds the statuses of listeners defined on the Gateway.
	ListenerStatuses ListenerStatuses
	// Conditions is the list of conditions for this Gateway.
	Conditions []conditions.Condition
	// Addresses holds the list of GatewayStatusAddresses.
	Addresses []v1beta1.GatewayStatusAddress
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
	// Ignored tells whether or not this Gateway is ignored.
	Ignored bool
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
	// Conditions is the list of conditions for this GatewayClass.
	Conditions []conditions.Condition
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}
