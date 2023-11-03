package relationship

import (
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/index"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Capturer

// Capturer captures relationships between Kubernetes objects and can be queried for whether a relationship exists
// for a given object.
//
// The relationships between HTTPRoutes -> Services are many to 1,
// so these relationships are tracked using a counter.
// A Service relationship exists if at least one HTTPRoute references it.
// An EndpointSlice relationship exists if its Service owner is referenced by at least one HTTPRoute.
//
// A Namespace relationship exists if it has labels that match a Gateway listener's label selector.
type Capturer interface {
	Capture(obj client.Object)
	Remove(resourceType client.Object, nsname types.NamespacedName)
	Exists(resourceType client.Object, nsname types.NamespacedName) bool
}

type (
	// routeToServicesMap maps HTTPRoute names to the set of Services it references.
	routeToServicesMap map[types.NamespacedName]map[types.NamespacedName]struct{}
	// serviceRefCountMap maps Service names to the number of HTTPRoutes that reference it.
	serviceRefCountMap map[types.NamespacedName]int
	// gatewayLabelSelectorsMap maps Gateways to the label selectors that their listeners use for allowed routes
	gatewayLabelSelectorsMap map[types.NamespacedName][]labels.Selector
	// namespaceCfg holds information about a namespace
	// - labels that it contains
	// - gateways that reference it (if labels match)
	namespaceCfg struct {
		labelMap map[string]string
		gateways map[types.NamespacedName]struct{}
	}
	// namespaces is a collection of namespaces in the system
	namespaces map[types.NamespacedName]namespaceCfg
)

func (n namespaceCfg) match() bool {
	return len(n.gateways) > 0
}

// CapturerImpl implements the Capturer interface.
type CapturerImpl struct {
	routesToServices      routeToServicesMap
	serviceRefCount       serviceRefCountMap
	gatewayLabelSelectors gatewayLabelSelectorsMap
	namespaces            namespaces
	paramsRef             *v1beta1.ParametersReference
	endpointSliceOwners   map[types.NamespacedName]types.NamespacedName
	gcName                string
}

// NewCapturerImpl creates a new instance of CapturerImpl.
func NewCapturerImpl(gcName string) *CapturerImpl {
	return &CapturerImpl{
		routesToServices:      make(routeToServicesMap),
		serviceRefCount:       make(serviceRefCountMap),
		gatewayLabelSelectors: make(gatewayLabelSelectorsMap),
		namespaces:            make(namespaces),
		endpointSliceOwners:   make(map[types.NamespacedName]types.NamespacedName),
		gcName:                gcName,
	}
}

// Capture captures relationships for the given object.
func (c *CapturerImpl) Capture(obj client.Object) {
	switch o := obj.(type) {
	case *v1beta1.HTTPRoute:
		c.upsertForRoute(o)
	case *discoveryV1.EndpointSlice:
		svcName := index.GetServiceNameFromEndpointSlice(o)
		if svcName != "" {
			c.endpointSliceOwners[client.ObjectKeyFromObject(o)] = types.NamespacedName{
				Namespace: o.Namespace,
				Name:      svcName,
			}
		}
	case *v1beta1.Gateway:
		c.upsertForGateway(o)
	case *v1.Namespace:
		nsLabels := o.GetLabels()
		gateways := c.matchingGateways(nsLabels)
		nsCfg := namespaceCfg{
			labelMap: nsLabels,
			gateways: gateways,
		}
		c.namespaces[client.ObjectKeyFromObject(o)] = nsCfg
	case *v1beta1.GatewayClass:
		if o.Spec.ParametersRef != nil && o.Name == c.gcName {
			c.paramsRef = o.Spec.ParametersRef
		}
	}
}

// Remove removes the relationship for the given object from the CapturerImpl.
func (c *CapturerImpl) Remove(resourceType client.Object, nsname types.NamespacedName) {
	switch resourceType.(type) {
	case *v1beta1.HTTPRoute:
		c.deleteForRoute(nsname)
	case *discoveryV1.EndpointSlice:
		delete(c.endpointSliceOwners, nsname)
	case *v1beta1.Gateway:
		c.removeGatewayLabelSelector(nsname)
	case *v1.Namespace:
		delete(c.namespaces, nsname)
	case *v1beta1.GatewayClass:
		if nsname.Name == c.gcName {
			c.paramsRef = nil
		}
	}
}

// Exists returns true if the given object has a relationship with another object.
func (c *CapturerImpl) Exists(resourceType client.Object, nsname types.NamespacedName) bool {
	switch resourceType.(type) {
	case *v1.Service:
		return c.serviceRefCount[nsname] > 0
	case *discoveryV1.EndpointSlice:
		svcOwner, exists := c.endpointSliceOwners[nsname]
		return exists && c.serviceRefCount[svcOwner] > 0
	case *v1.Namespace:
		cfg, exists := c.namespaces[nsname]
		return exists && cfg.match()
	case *ngfAPI.NginxProxy:
		if c.paramsRef != nil {
			return c.paramsRef.Namespace != nil &&
				c.paramsRef.Group == ngfAPI.GroupName &&
				c.paramsRef.Kind == v1beta1.Kind("NginxProxy") &&
				c.paramsRef.Name == nsname.Name &&
				string(*c.paramsRef.Namespace) == nsname.Namespace
		}
	}

	return false
}

// GetRefCountForService is used for unit testing purposes. It is not exposed through the Capturer interface.
func (c *CapturerImpl) GetRefCountForService(svcName types.NamespacedName) int {
	return c.serviceRefCount[svcName]
}

func (c *CapturerImpl) upsertForRoute(route *v1beta1.HTTPRoute) {
	oldServices := c.routesToServices[client.ObjectKeyFromObject(route)]
	newServices := getBackendServiceNamesFromRoute(route)

	for svc := range oldServices {
		if _, exist := newServices[svc]; !exist {
			c.decrementRefCount(svc)
		}
	}

	for svc := range newServices {
		if _, exist := oldServices[svc]; !exist {
			c.serviceRefCount[svc]++
		}
	}

	c.routesToServices[client.ObjectKeyFromObject(route)] = newServices
}

func (c *CapturerImpl) deleteForRoute(routeName types.NamespacedName) {
	services := c.routesToServices[routeName]

	for svc := range services {
		c.decrementRefCount(svc)
	}

	delete(c.routesToServices, routeName)
}

func (c *CapturerImpl) decrementRefCount(svcName types.NamespacedName) {
	if count, exist := c.serviceRefCount[svcName]; exist {
		if count == 1 {
			delete(c.serviceRefCount, svcName)

			return
		}

		c.serviceRefCount[svcName]--
	}
}

func getBackendServiceNamesFromRoute(hr *v1beta1.HTTPRoute) map[types.NamespacedName]struct{} {
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

func (c *CapturerImpl) upsertForGateway(gw *v1beta1.Gateway) {
	var selectors []labels.Selector
	for _, listener := range gw.Spec.Listeners {
		if selector := graph.GetAllowedRouteLabelSelector(listener); selector != nil {
			convertedSelector, err := metav1.LabelSelectorAsSelector(selector)
			if err == nil {
				selectors = append(selectors, convertedSelector)
			}
		}
	}

	gatewayName := client.ObjectKeyFromObject(gw)
	if len(selectors) > 0 {
		c.gatewayLabelSelectors[gatewayName] = selectors
		for ns, cfg := range c.namespaces {
			var gatewayMatches bool
			for _, selector := range selectors {
				if selector.Matches(labels.Set(cfg.labelMap)) {
					gatewayMatches = true
					cfg.gateways[gatewayName] = struct{}{}
					break
				}
			}
			if !gatewayMatches {
				// if gateway was previously referenced by this namespace, clean it up
				delete(cfg.gateways, gatewayName)
			}
			c.namespaces[ns] = cfg
		}
	} else if _, exists := c.gatewayLabelSelectors[gatewayName]; exists {
		// label selectors existed previously for this gateway, so clean up any references to them
		c.removeGatewayLabelSelector(gatewayName)
	}
}

// matchingGateways looks through all existing label selectors defined by listeners in a gateway,
// and if any matches are found, returns a map of those gateways
func (c *CapturerImpl) matchingGateways(labelMap map[string]string) map[types.NamespacedName]struct{} {
	gateways := make(map[types.NamespacedName]struct{})
	for gw, selectors := range c.gatewayLabelSelectors {
		for _, selector := range selectors {
			if selector.Matches(labels.Set(labelMap)) {
				gateways[gw] = struct{}{}
				break
			}
		}
	}

	return gateways
}

func (c *CapturerImpl) removeGatewayLabelSelector(gatewayName types.NamespacedName) {
	delete(c.gatewayLabelSelectors, gatewayName)
	for ns, cfg := range c.namespaces {
		delete(cfg.gateways, gatewayName)
		c.namespaces[ns] = cfg
	}
}
