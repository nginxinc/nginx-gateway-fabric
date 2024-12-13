package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// A ReferencedService represents a Kubernetes Service that is referenced by a Route and that belongs to the
// winning Gateway. It does not contain the v1.Service object, because Services are resolved when building
// the dataplane.Configuration.
type ReferencedService struct {
	// Policies is a list of NGF Policies that target this Service.
	Policies []*Policy
}

func buildReferencedServices(
	l7routes map[RouteKey]*L7Route,
	l4Routes map[L4RouteKey]*L4Route,
	gw *Gateway,
) map[types.NamespacedName]*ReferencedService {
	if gw == nil {
		return nil
	}

	referencedServices := make(map[types.NamespacedName]*ReferencedService)

	belongsToWinningGw := func(refs []ParentRef) bool {
		for _, ref := range refs {
			if ref.Gateway == client.ObjectKeyFromObject(gw.Source) {
				return true
			}
		}

		return false
	}

	// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
	// we may want to track.
	addServicesForL7Routes := func(routeRules []RouteRule) {
		for _, rule := range routeRules {
			for _, ref := range rule.BackendRefs {
				if ref.SvcNsName != (types.NamespacedName{}) {
					referencedServices[ref.SvcNsName] = &ReferencedService{
						Policies: nil,
					}
				}
			}
		}
	}

	addServicesForL4Routes := func(route *L4Route) {
		nsname := route.Spec.BackendRef.SvcNsName
		if nsname != (types.NamespacedName{}) {
			referencedServices[nsname] = &ReferencedService{
				Policies: nil,
			}
		}
	}

	// routes all have populated ParentRefs from when they were created.
	//
	// Get all the service names referenced from all the l7 and l4 routes.
	for _, route := range l7routes {
		if !route.Valid {
			continue
		}

		if !belongsToWinningGw(route.ParentRefs) {
			continue
		}

		addServicesForL7Routes(route.Spec.Rules)
	}

	for _, route := range l4Routes {
		if !route.Valid {
			continue
		}

		if !belongsToWinningGw(route.ParentRefs) {
			continue
		}

		addServicesForL4Routes(route)
	}

	if len(referencedServices) == 0 {
		return nil
	}

	return referencedServices
}
