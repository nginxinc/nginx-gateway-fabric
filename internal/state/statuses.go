package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

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
	GatewayStatuses        []GatewayStatus
	IgnoredGatewayStatuses IgnoredGatewayStatuses
	HTTPRouteStatuses      HTTPRouteStatuses
}

// GatewayStatus holds the status of the winning Gateway resource.
type GatewayStatus struct {
	Conditions []conditions.Condition
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

// buildStatuses builds statuses from a Graph.
func buildStatuses(graph *graph.Graph) Statuses {
	statuses := Statuses{
		HTTPRouteStatuses:      make(map[types.NamespacedName]HTTPRouteStatus),
		IgnoredGatewayStatuses: make(map[types.NamespacedName]IgnoredGatewayStatus),
	}

	if graph.GatewayClass != nil {
		defaultConds := conditions.NewDefaultGatewayClassConditions()

		conds := make([]conditions.Condition, 0, len(graph.GatewayClass.Conditions)+len(defaultConds))

		// We add default conds first, so that any additional conditions will override them, which is
		// ensured by DeduplicateConditions.
		conds = append(conds, defaultConds...)
		conds = append(conds, graph.GatewayClass.Conditions...)

		statuses.GatewayClassStatus = &GatewayClassStatus{
			Conditions:         conditions.DeduplicateConditions(conds),
			ObservedGeneration: graph.GatewayClass.Source.Generation,
		}
	}

	for _, gw := range graph.Gateways {
		listenerStatuses := make(map[string]ListenerStatus)

		defaultConds := conditions.NewDefaultListenerConditions()

		for name, l := range gw.Listeners {
			conds := make([]conditions.Condition, 0, len(l.Conditions)+len(defaultConds))

			// We add default conds first, so that any additional conditions will override them, which is
			// ensured by DeduplicateConditions.
			conds = append(conds, defaultConds...)
			conds = append(conds, l.Conditions...)

			listenerStatuses[name] = ListenerStatus{
				AttachedRoutes: int32(len(l.Routes)),
				Conditions:     conditions.DeduplicateConditions(conds),
			}
		}

		status := GatewayStatus{
			Conditions:         conditions.NewDefaultGatewayConditions(),
			NsName:             client.ObjectKeyFromObject(gw.Source),
			ListenerStatuses:   listenerStatuses,
			ObservedGeneration: gw.Source.Generation,
		}

		statuses.GatewayStatuses = append(statuses.GatewayStatuses, status)
	}

	for nsname, gw := range graph.IgnoredGateways {
		statuses.IgnoredGatewayStatuses[nsname] = IgnoredGatewayStatus{ObservedGeneration: gw.Generation}
	}

	for nsname, r := range graph.Routes {
		parentStatuses := make([]ParentStatus, 0, len(r.ParentRefs))

		defaultConds := conditions.NewDefaultRouteConditions()

		for _, ref := range r.ParentRefs {
			failedAttachmentCondCount := 0
			if ref.Attachment != nil && !ref.Attachment.Attached {
				failedAttachmentCondCount = 1
			}
			allConds := make([]conditions.Condition, 0, len(r.Conditions)+len(defaultConds)+failedAttachmentCondCount)

			// We add defaultConds first, so that any additional conditions will override them, which is
			// ensured by DeduplicateConditions.
			allConds = append(allConds, defaultConds...)
			allConds = append(allConds, r.Conditions...)
			if failedAttachmentCondCount == 1 {
				allConds = append(allConds, ref.Attachment.FailedCondition)
			}

			routeRef := r.Source.Spec.ParentRefs[ref.Idx]

			parentStatuses = append(parentStatuses, ParentStatus{
				GatewayNsName: ref.Gateway,
				SectionName:   routeRef.SectionName,
				Conditions:    conditions.DeduplicateConditions(allConds),
			})
		}

		statuses.HTTPRouteStatuses[nsname] = HTTPRouteStatus{
			ObservedGeneration: r.Source.Generation,
			ParentStatuses:     parentStatuses,
		}
	}

	return statuses
}
