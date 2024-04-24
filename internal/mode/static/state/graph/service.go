package graph

import (
	"k8s.io/apimachinery/pkg/types"
)

func buildReferencedServices(
	routes map[RouteKey]*L7Route,
) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	getServiceNamesFromRoute := func(parentRefs []ParentRef, routeRules []RouteRule) {
		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		attached := false
		for _, ref := range parentRefs {
			if ref.Attachment.Attached {
				attached = true
				break
			}
		}
		if !attached {
			return
		}

		for _, rule := range routeRules {
			for _, ref := range rule.BackendRefs {
				// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
				// we may want to track.
				if ref.SvcNsName != (types.NamespacedName{}) {
					svcNames[ref.SvcNsName] = struct{}{}
				}
			}
		}
	}

	// routes all have populated ParentRefs from when they were created.
	//
	// Get all the service names referenced from all the Routes.
	for _, route := range routes {
		if !route.Valid {
			continue
		}

		getServiceNamesFromRoute(route.ParentRefs, route.Spec.Rules)
	}

	if len(svcNames) == 0 {
		return nil
	}
	return svcNames
}
