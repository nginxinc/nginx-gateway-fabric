package state

import "k8s.io/apimachinery/pkg/types"

// ListenerStatuses holds the statuses of listeners where the key is the name of a listener in the Gateway resource.
type ListenerStatuses map[string]ListenerStatus

// HTTPRouteStatuses holds the statuses of HTTPRoutes where the key is the namespaced name of an HTTPRoute.
type HTTPRouteStatuses map[types.NamespacedName]HTTPRouteStatus

// Statuses holds the status-related information about Gateway API resources.
// It is assumed that only a singe Gateway resource is used.
type Statuses struct {
	GatewayClassStatus *GatewayClassStatus
	ListenerStatuses   ListenerStatuses
	HTTPRouteStatuses  HTTPRouteStatuses
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
		ListenerStatuses:  make(map[string]ListenerStatus),
		HTTPRouteStatuses: make(map[types.NamespacedName]HTTPRouteStatus),
	}

	if graph.GatewayClass != nil {
		statuses.GatewayClassStatus = &GatewayClassStatus{
			Valid:              graph.GatewayClass.Valid,
			ErrorMsg:           graph.GatewayClass.ErrorMsg,
			ObservedGeneration: graph.GatewayClass.Source.Generation,
		}
	}

	gcValidAndExist := graph.GatewayClass != nil && graph.GatewayClass.Valid

	for name, l := range graph.Listeners {
		statuses.ListenerStatuses[name] = ListenerStatus{
			Valid:          l.Valid && gcValidAndExist,
			AttachedRoutes: int32(len(l.Routes)),
		}
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
