package graph

import (
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func buildReferencedServicesNames(
	clusterHTTPRoutes map[types.NamespacedName]*gatewayv1.HTTPRoute,
) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	// Get all the service names referenced from all the HTTPRoutes
	for _, hr := range clusterHTTPRoutes {
		tempSvcNames := getBackendServiceNamesFromRoute(hr)
		for k, v := range tempSvcNames {
			svcNames[k] = v
		}
	}

	if len(svcNames) == 0 {
		return nil
	}
	return svcNames
}

func getBackendServiceNamesFromRoute(hr *gatewayv1.HTTPRoute) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	for _, rule := range hr.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if ref.Kind != nil && *ref.Kind != "Service" {
				continue
			}

			ns := hr.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}

			svcNames[types.NamespacedName{Namespace: ns, Name: string(ref.Name)}] = struct{}{}
		}
	}

	return svcNames
}
