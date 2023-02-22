package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

// Route represents an HTTPRoute.
type Route struct {
	// Source is the source resource of the Route.
	// FIXME(pleshakov)
	// For now, we assume that the source is only HTTPRoute.
	// Later we can support more types - TLSRoute, TCPRoute and UDPRoute.
	Source *v1beta1.HTTPRoute

	// ValidSectionNameRefs includes the sectionNames from the parentRefs of the HTTPRoute that are valid -- i.e.
	// the Gateway resource has a corresponding valid listener.
	ValidSectionNameRefs map[string]struct{}
	// InvalidSectionNameRefs includes the sectionNames from the parentRefs of the HTTPRoute that are invalid.
	// The Condition describes why the sectionName is invalid.
	InvalidSectionNameRefs map[string]conditions.Condition
	// BackendGroups includes the backend groups of the HTTPRoute.
	// There's one BackendGroup per rule in the HTTPRoute.
	// The BackendGroups are stored in order of the rules.
	// Ex: Source.Spec.Rules[0] -> BackendGroups[0].
	BackendGroups []BackendGroup
}

// bindHTTPRouteToListeners tries to bind an HTTPRoute to listener.
// There are three possibilities:
// (1) HTTPRoute will be ignored.
// (2) HTTPRoute will be processed but not bound.
// (3) HTTPRoute will be processed and bound to a listener.
func bindHTTPRouteToListeners(
	ghr *v1beta1.HTTPRoute,
	gw *v1beta1.Gateway,
	ignoredGws map[types.NamespacedName]*v1beta1.Gateway,
	listeners map[string]*Listener,
) (ignored bool, r *Route) {
	if len(ghr.Spec.ParentRefs) == 0 {
		// ignore HTTPRoute without refs
		return true, nil
	}

	r = &Route{
		Source:                 ghr,
		ValidSectionNameRefs:   make(map[string]struct{}),
		InvalidSectionNameRefs: make(map[string]conditions.Condition),
	}

	// FIXME (pleshakov) Handle the case when parent refs are duplicated

	processed := false

	for _, p := range ghr.Spec.ParentRefs {
		// FIXME(pleshakov) Support empty section name
		if p.SectionName == nil || *p.SectionName == "" {
			continue
		}

		// if the namespace is missing, assume the namespace of the HTTPRoute
		ns := ghr.Namespace
		if p.Namespace != nil {
			ns = string(*p.Namespace)
		}

		name := string(*p.SectionName)

		// Below we will figure out what Gateway resource the parentRef references and act accordingly. There are 3 cases.

		// Case 1: the parentRef references the winning Gateway.

		if gw != nil && gw.Namespace == ns && gw.Name == string(p.Name) {

			// Find a listener

			// FIXME(pleshakov)
			// For now, let's do simple matching.
			// However, we need to also support wildcard matching.
			// More over, we need to handle cases when a Route host matches multiple HTTP listeners on the same port when
			// sectionName is empty and only choose one listener.
			// For example:
			// - Route with host foo.example.com;
			// - listener 1 for port 80 with hostname foo.example.com
			// - listener 2 for port 80 with hostname *.example.com;
			// In this case, the Route host foo.example.com should choose listener 1, as it is a more specific match.

			processed = true

			l, exists := listeners[name]
			if !exists {
				// FIXME(pleshakov): Add a proper condition once it is available in the Gateway API.
				// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/306
				r.InvalidSectionNameRefs[name] = conditions.NewTODO("listener is not found")
				continue
			}

			if !l.Valid {
				r.InvalidSectionNameRefs[name] = conditions.NewRouteInvalidListener()
				continue
			}

			accepted := findAcceptedHostnames(l.Source.Hostname, ghr.Spec.Hostnames)

			if len(accepted) > 0 {
				for _, h := range accepted {
					l.AcceptedHostnames[h] = struct{}{}
				}
				r.ValidSectionNameRefs[name] = struct{}{}
				l.Routes[client.ObjectKeyFromObject(ghr)] = r
			} else {
				r.InvalidSectionNameRefs[name] = conditions.NewRouteNoMatchingListenerHostname()
			}

			continue
		}

		// Case 2: the parentRef references an ignored Gateway resource.

		key := types.NamespacedName{Namespace: ns, Name: string(p.Name)}

		if _, exist := ignoredGws[key]; exist {
			// FIXME(pleshakov): Add a proper condition.
			// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/306
			r.InvalidSectionNameRefs[name] = conditions.NewTODO("Gateway is ignored")

			processed = true
			continue
		}

		// Case 3: the parentRef references some unrelated to this NGINX Gateway Gateway or other resource.

		// Do nothing
	}

	if !processed {
		return true, nil
	}

	return false, r
}

func findAcceptedHostnames(listenerHostname *v1beta1.Hostname, routeHostnames []v1beta1.Hostname) []string {
	hostname := getHostname(listenerHostname)

	match := func(h v1beta1.Hostname) bool {
		if hostname == "" {
			return true
		}
		return string(h) == hostname
	}

	var result []string

	for _, h := range routeHostnames {
		if match(h) {
			result = append(result, string(h))
		}
	}

	return result
}

func getHostname(h *v1beta1.Hostname) string {
	if h == nil {
		return ""
	}
	return string(*h)
}
