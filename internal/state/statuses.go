package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListenerStatuses holds the statuses of listeners where the key is the name of a listener in the Gateway resource.
type ListenerStatuses map[string]ListenerStatus

// HTTPRouteStatuses holds the statuses of HTTPRoutes where the key is the namespaced name of an HTTPRoute.
type HTTPRouteStatuses map[types.NamespacedName]HTTPRouteStatus

// Statuses holds the status-related information about Gateway API resources.
type Statuses struct {
	GatewayClassStatus     *GatewayClassStatus
	GatewayStatus          *GatewayStatus
	IgnoredGatewayStatuses IgnoredGatewayStatuses
	HTTPRouteStatuses      HTTPRouteStatuses
}

// GatewayStatus holds the status of the winning Gateway resource.
type GatewayStatus struct {
	NsName           types.NamespacedName
	ListenerStatuses ListenerStatuses
}

// IgnoredGatewayStatuses holds the statuses of the ignored Gateway resources.
type IgnoredGatewayStatuses map[types.NamespacedName]IgnoredGatewayStatus

// IgnoredGatewayStatus holds the status of an ignored Gateway resource.
type IgnoredGatewayStatus struct {
	ObservedGeneration int64
}

// ListenerStatus holds the status-related information about a listener in the Gateway resource.
type ListenerStatus struct {
	// Valid shows if the listener is valid.
	Valid bool
	// AttachedRoutes is the number of routes attached to the listener.
	AttachedRoutes int32
}

// ParentStatuses holds the statuses of parents where the key is the section name in a parentRef.
type ParentStatuses map[string]ParentStatus

type HTTPRouteStatus struct {
	ParentStatuses ParentStatuses
}

// ParentStatus holds status-related information related to how the HTTPRoute binds to a specific parentRef.
type ParentStatus struct {
	// Attached is true if the route attaches to the parent (listener).
	Attached bool
}

// GatewayClassStatus holds status-related infortmation about the GatewayClass resource.
type GatewayClassStatus struct {
	// Valid shows if the resource is valid.
	Valid bool
	// ErrorMsg describe the error when the resource is invalid.
	ErrorMsg string
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

// buildStatuses builds statuses from a graph.
func buildStatuses(graph *graph) Statuses {
	statuses := Statuses{
		HTTPRouteStatuses:      make(map[types.NamespacedName]HTTPRouteStatus),
		IgnoredGatewayStatuses: make(map[types.NamespacedName]IgnoredGatewayStatus),
	}

	if graph.GatewayClass != nil {
		statuses.GatewayClassStatus = &GatewayClassStatus{
			Valid:              graph.GatewayClass.Valid,
			ErrorMsg:           graph.GatewayClass.ErrorMsg,
			ObservedGeneration: graph.GatewayClass.Source.Generation,
		}
	}

	gcValidAndExist := graph.GatewayClass != nil && graph.GatewayClass.Valid

	if graph.Gateway != nil {
		listenerStatuses := make(map[string]ListenerStatus)

		for name, l := range graph.Gateway.Listeners {
			listenerStatuses[name] = ListenerStatus{
				Valid:          l.Valid && gcValidAndExist,
				AttachedRoutes: int32(len(l.Routes)),
			}
		}

		statuses.GatewayStatus = &GatewayStatus{
			NsName:           client.ObjectKeyFromObject(graph.Gateway.Source),
			ListenerStatuses: listenerStatuses,
		}
	}

	for nsname, gw := range graph.IgnoredGateways {
		statuses.IgnoredGatewayStatuses[nsname] = IgnoredGatewayStatus{ObservedGeneration: gw.Generation}
	}

	for nsname, r := range graph.Routes {
		parentStatuses := make(map[string]ParentStatus)

		for ref := range r.ValidSectionNameRefs {
			parentStatuses[ref] = ParentStatus{
				Attached: gcValidAndExist, // Attached only when GatewayClass is valid and exists
			}
		}
		for ref := range r.InvalidSectionNameRefs {
			parentStatuses[ref] = ParentStatus{
				Attached: false,
			}
		}

		statuses.HTTPRouteStatuses[nsname] = HTTPRouteStatus{
			ParentStatuses: parentStatuses,
		}
	}

	return statuses
}
