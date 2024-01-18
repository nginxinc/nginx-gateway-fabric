package graph

import (
	"k8s.io/apimachinery/pkg/types"
)

// routes all have populated ParentRefs from when they were created
func buildReferencedServices(
	routes map[types.NamespacedName]*Route,
) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	// Get all the service names referenced from all the HTTPRoutes.
	for _, route := range routes {
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

		if !route.Valid {
			continue
		}

		for idx := range route.Rules {
			// Do I still need to do these checks? As the check is made in backend_ref, basically
			// if the route rules are not valid, it will not populate route.Rules[idx].BackendRefs, thus
			// if we just do the for loop over the route.Rules[idx].BackendRefs = backendRefs, we should be fine?
			if !route.Rules[idx].ValidMatches {
				continue
			}
			if !route.Rules[idx].ValidFilters {
				continue
			}

			for _, ref := range route.Rules[idx].BackendRefs {
				// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
				// we may want to track.
				if ref.SvcNsName != (types.NamespacedName{}) {
					svcNames[ref.SvcNsName] = struct{}{}
				}
			}
		}
	}

	if len(svcNames) == 0 {
		return nil
	}
	return svcNames
}
