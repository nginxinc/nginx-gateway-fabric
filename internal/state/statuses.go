package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
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
	// ListenerStatuses holds the statuses of listeners defined on the Gateway.
	ListenerStatuses ListenerStatuses
	// NsName is the namespaced name of the winning Gateway resource.
	NsName types.NamespacedName
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

// IgnoredGatewayStatuses holds the statuses of the ignored Gateway resources.
type IgnoredGatewayStatuses map[types.NamespacedName]IgnoredGatewayStatus

// IgnoredGatewayStatus holds the status of an ignored Gateway resource.
type IgnoredGatewayStatus struct {
	ObservedGeneration int64
}

// ListenerStatus holds the status-related information about a listener in the Gateway resource.
type ListenerStatus struct {
	// Conditions is the list of conditions for this listener.
	Conditions []conditions.Condition
	// AttachedRoutes is the number of routes attached to the listener.
	AttachedRoutes int32
}

// ParentStatuses holds the statuses of parents where the key is the section name in a parentRef.
type ParentStatuses map[string]ParentStatus

// HTTPRouteStatus holds the status-related information about an HTTPRoute resource.
type HTTPRouteStatus struct {
	// ParentStatuses holds the statuses for parentRefs of the HTTPRoute.
	ParentStatuses ParentStatuses
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
}

// ParentStatus holds status-related information related to how the HTTPRoute binds to a specific parentRef.
type ParentStatus struct {
	// Conditions is the list of conditions that are relevant to the parentRef.
	Conditions []conditions.Condition
}

// GatewayClassStatus holds status-related infortmation about the GatewayClass resource.
type GatewayClassStatus struct {
	// ErrorMsg describe the error when the resource is invalid.
	ErrorMsg string
	// ObservedGeneration is the generation of the resource that was processed.
	ObservedGeneration int64
	// Valid shows if the resource is valid.
	Valid bool
}

// buildStatuses builds statuses from a Graph.
func buildStatuses(graph *graph.Graph) Statuses {
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

		defaultConds := conditions.NewDefaultListenerConditions()

		for name, l := range graph.Gateway.Listeners {
			conds := make([]conditions.Condition, 0, len(l.Conditions)+len(defaultConds)+1) // 1 is for missing GC

			// We add default conds first, so that any additional conditions will override them, which is
			// ensured by DeduplicateConditions.
			conds = append(conds, defaultConds...)
			conds = append(conds, l.Conditions...)

			if !gcValidAndExist {
				// FIXME(pleshakov): Figure out appropriate conditions for the cases when:
				// (1) GatewayClass is invalid.
				// (2) GatewayClass does not exist.
				// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/307
				conds = append(conds, conditions.NewTODO("GatewayClass is invalid or doesn't exist"))
			}

			listenerStatuses[name] = ListenerStatus{
				AttachedRoutes: int32(len(l.Routes)),
				Conditions:     conditions.DeduplicateConditions(conds),
			}
		}

		statuses.GatewayStatus = &GatewayStatus{
			NsName:             client.ObjectKeyFromObject(graph.Gateway.Source),
			ListenerStatuses:   listenerStatuses,
			ObservedGeneration: graph.Gateway.Source.Generation,
		}
	}

	for nsname, gw := range graph.IgnoredGateways {
		statuses.IgnoredGatewayStatuses[nsname] = IgnoredGatewayStatus{ObservedGeneration: gw.Generation}
	}

	for nsname, r := range graph.Routes {
		parentStatuses := make(map[string]ParentStatus)

		baseConds := buildBaseRouteConditions(gcValidAndExist)

		for ref := range r.SectionNameRefs {
			conds := r.GetAllConditionsForSectionName(ref)

			allConds := make([]conditions.Condition, 0, len(conds)+len(baseConds))
			// We add baseConds first, so that any additional conditions will override them, which is
			// ensured by DeduplicateConditions.
			allConds = append(allConds, baseConds...)
			allConds = append(allConds, conds...)

			if ref == "" {
				// FIXME(pleshakov): Gateway API spec does allow empty section names in the status.
				// However, NKG doesn't yet support the empty section names.
				// Once NKG supports them, it will be able to determine which section name the HTTPRoute was bound to.
				// So we won't need this workaround.
				ref = "unbound"
			}

			parentStatuses[ref] = ParentStatus{
				Conditions: conditions.DeduplicateConditions(allConds),
			}
		}

		statuses.HTTPRouteStatuses[nsname] = HTTPRouteStatus{
			ObservedGeneration: r.Source.Generation,
			ParentStatuses:     parentStatuses,
		}
	}

	return statuses
}

func buildBaseRouteConditions(gcValidAndExist bool) []conditions.Condition {
	conds := conditions.NewDefaultRouteConditions()

	// FIXME(pleshakov): Figure out appropriate conditions for the cases when:
	// (1) GatewayClass is invalid.
	// (2) GatewayClass does not exist.
	// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/307
	if !gcValidAndExist {
		conds = append(conds, conditions.NewTODO("GatewayClass is invalid or doesn't exist"))
	}

	return conds
}
