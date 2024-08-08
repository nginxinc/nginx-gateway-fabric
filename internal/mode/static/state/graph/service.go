package graph

import (
	"k8s.io/apimachinery/pkg/types"
)

func buildReferencedServices(
	l7routes map[RouteKey]*L7Route,
	l4Routes map[L4RouteKey]*L4Route,
) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	attached := func(parentRefs []ParentRef) bool {
		for _, ref := range parentRefs {
			if ref.Attachment.Attached {
				return true
			}
		}

		return false
	}

	// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
	// we may want to track.

	populateServiceNamesForL7Routes := func(routeRules []RouteRule) {
		for _, rule := range routeRules {
			for _, ref := range rule.BackendRefs {
				if ref.SvcNsName != (types.NamespacedName{}) {
					svcNames[ref.SvcNsName] = struct{}{}
				}
			}
		}
	}

	populateServiceNamesForL4Routes := func(route *L4Route) {
		nsname := route.Spec.BackendRef.SvcNsName
		if nsname != (types.NamespacedName{}) {
			svcNames[nsname] = struct{}{}
		}
	}

	// routes all have populated ParentRefs from when they were created.
	//
	// Get all the service names referenced from all the l7 and l4 routes.
	for _, route := range l7routes {
		if !route.Valid {
			continue
		}

		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		if !attached(route.ParentRefs) {
			continue
		}

		populateServiceNamesForL7Routes(route.Spec.Rules)
	}

	for _, route := range l4Routes {
		if !route.Valid {
			continue
		}

		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		if !attached(route.ParentRefs) {
			continue
		}

		populateServiceNamesForL4Routes(route)
	}

	if len(svcNames) == 0 {
		return nil
	}
	return svcNames
}
