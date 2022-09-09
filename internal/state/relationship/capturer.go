package relationship

import (
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Capturer

// Capturer captures relationships between Kubernetes objects and can be queried for whether a relationship exists
// for a given object.
//
// Currently, it only captures relationships between HTTPRoutes and Services and Services and EndpointSlices,
// but it can be extended to capture additional relationships.
// The relationships between HTTPRoutes -> Services are many to 1,
// so these relationships are tracked using a counter.
// A Service relationship exists if at least one HTTPRoute references it.
// An EndpointSlice relationship exists, if its Service owner is referenced by at least one HTTPRoute.
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
)

// CapturerImpl implements the Capturer interface.
type CapturerImpl struct {
	routesToServices    routeToServicesMap
	serviceRefCount     serviceRefCountMap
	endpointSliceOwners map[types.NamespacedName]types.NamespacedName
}

// NewCapturerImpl creates a new instance of CapturerImpl.
func NewCapturerImpl() *CapturerImpl {
	return &CapturerImpl{
		routesToServices:    make(map[types.NamespacedName]map[types.NamespacedName]struct{}),
		serviceRefCount:     make(map[types.NamespacedName]int),
		endpointSliceOwners: make(map[types.NamespacedName]types.NamespacedName),
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
	}
}

// Remove removes the relationship for the given object from the CapturerImpl.
func (c *CapturerImpl) Remove(resourceType client.Object, nsname types.NamespacedName) {
	switch resourceType.(type) {
	case *v1beta1.HTTPRoute:
		c.deleteForRoute(nsname)
	case *discoveryV1.EndpointSlice:
		delete(c.endpointSliceOwners, nsname)
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

// FIXME(pleshakov): for now, we only support a single backend reference
func getBackendServiceNamesFromRoute(hr *v1beta1.HTTPRoute) map[types.NamespacedName]struct{} {
	svcNames := make(map[types.NamespacedName]struct{})

	for _, rule := range hr.Spec.Rules {
		if len(rule.BackendRefs) == 0 {
			continue
		}
		ref := rule.BackendRefs[0].BackendRef

		if ref.Kind != nil && *ref.Kind != "Service" {
			continue
		}

		ns := hr.Namespace
		if ref.Namespace != nil {
			ns = string(*ref.Namespace)
		}

		svcNames[types.NamespacedName{Namespace: ns, Name: string(ref.Name)}] = struct{}{}
	}

	return svcNames
}
