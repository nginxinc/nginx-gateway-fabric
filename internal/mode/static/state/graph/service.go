package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func buildReferencedServices(
	routes map[types.NamespacedName]*Route,
	services map[types.NamespacedName]*v1.Service,
) map[types.NamespacedName]*v1.Service {
	svcNames := make(map[types.NamespacedName]*v1.Service)

	// routes all have populated ParentRefs from when they were created.
	//
	// Get all the service names referenced from all the HTTPRoutes.
	for _, route := range routes {
		if !route.Valid {
			continue
		}

		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		attached := false
		for _, ref := range route.ParentRefs {
			if ref.Attachment.Attached {
				attached = true
				break
			}
		}
		if !attached {
			continue
		}

		for _, rule := range route.Rules {
			for _, ref := range rule.BackendRefs {
				// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
				// we may want to track.
				if ref.SvcNsName != (types.NamespacedName{}) {
					svcNames[ref.SvcNsName] = services[ref.SvcNsName]
				}
			}
		}
	}

	if len(svcNames) == 0 {
		return nil
	}
	return svcNames
}
