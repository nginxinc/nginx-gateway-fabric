package graph

import (
	"k8s.io/apimachinery/pkg/types"
)

// A ReferencedService represents a Kubernetes Service that is referenced by a Route.
// It does not contain the v1.Service object, because Services are resolved when building the dataplane.Configuration.
type ReferencedService struct {
	// ParentGateways is a list of unique attached parent Gateways for the Routes that reference this Service.
	ParentGateways []types.NamespacedName
	// Policies is a list of NGF Policies that target this Service.
	Policies []*Policy
}

func buildReferencedServices(
	l7routes map[RouteKey]*L7Route,
	l4Routes map[L4RouteKey]*L4Route,
) map[types.NamespacedName]*ReferencedService {
	referencedServices := make(map[types.NamespacedName]*ReferencedService)

	attached := func(parentRefs []ParentRef) bool {
		for _, ref := range parentRefs {
			if ref.Attachment != nil && ref.Attachment.Attached {
				return true
			}
		}

		return false
	}

	// Processes both valid and invalid BackendRefs as invalid ones still have referenced services
	// we may want to track.

	addServicesForL7Routes := func(routeRules []RouteRule, parentGateways []types.NamespacedName) {
		for _, rule := range routeRules {
			for _, ref := range rule.BackendRefs {
				if ref.SvcNsName != (types.NamespacedName{}) {
					referencedServices[ref.SvcNsName] = &ReferencedService{
						ParentGateways: parentGateways,
						Policies:       nil,
					}
				}
			}
		}
	}

	addServicesForL4Routes := func(route *L4Route, parentGateways []types.NamespacedName) {
		nsname := route.Spec.BackendRef.SvcNsName
		if nsname != (types.NamespacedName{}) {
			referencedServices[nsname] = &ReferencedService{
				ParentGateways: parentGateways,
				Policies:       nil,
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

		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		if !attached(route.ParentRefs) {
			continue
		}

		addServicesForL7Routes(route.Spec.Rules, getUniqueAttachedParentGateways(route.ParentRefs))
	}

	for _, route := range l4Routes {
		if !route.Valid {
			continue
		}

		// If none of the ParentRefs are attached to the Gateway, we want to skip the route.
		if !attached(route.ParentRefs) {
			continue
		}

		addServicesForL4Routes(route, getUniqueAttachedParentGateways(route.ParentRefs))
	}

	if len(referencedServices) == 0 {
		return nil
	}

	return referencedServices
}

func getUniqueAttachedParentGateways(parentRefs []ParentRef) []types.NamespacedName {
	gatewayMap := map[types.NamespacedName]struct{}{}
	for _, ref := range parentRefs {
		if ref.Attachment == nil || !ref.Attachment.Attached {
			continue
		}
		gatewayMap[ref.Gateway] = struct{}{}
	}

	uniqueGateways := make([]types.NamespacedName, 0, len(gatewayMap))
	for gateway := range gatewayMap {
		uniqueGateways = append(uniqueGateways, gateway)
	}

	return uniqueGateways
}
