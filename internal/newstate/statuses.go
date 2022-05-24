package newstate

import "k8s.io/apimachinery/pkg/types"

// Statuses holds the status-related information about Gateway API resources.
// It is assumed that only a singe Gateway resource is used.
type Statuses struct {
	// the key is the name of a listener in the Gateway resource.
	ListenerStatuses  map[string]ListenerStatus
	HTTPRouteStatuses map[types.NamespacedName]HTTPRouteStatus
}

// ListenerStatus holds the status-related information about a listener in the Gateway resource.
type ListenerStatus struct {
	// Valid shows if the listener is valid.
	Valid bool
	// AttachedRoutes is the number of routes attached to the listener.
	AttachedRoutes int32
}

// HTTPRouteStatus holds the status-related information about an HTTPRoute.
type HTTPRouteStatus struct {
	// the key is the section name in a parentRef.
	ParentStatuses map[string]ParentStatus
}

// ParentStatus holds status-related information related to how the HTTPRoute binds to a specific parentRef.
type ParentStatus struct {
	// Attached is true if the route attaches to the parent (listener).
	Attached bool
}

// buildStatuses builds statuses from a graph.
func buildStatuses(graph *graph) Statuses {
	statuses := Statuses{
		ListenerStatuses:  make(map[string]ListenerStatus),
		HTTPRouteStatuses: make(map[types.NamespacedName]HTTPRouteStatus),
	}

	for name, l := range graph.Listeners {
		statuses.ListenerStatuses[name] = ListenerStatus{
			Valid:          l.Valid,
			AttachedRoutes: int32(len(l.Routes)),
		}
	}

	for nsname, r := range graph.Routes {
		parentStatuses := make(map[string]ParentStatus)

		for ref := range r.ValidSectionNameRefs {
			parentStatuses[ref] = ParentStatus{
				Attached: true,
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
