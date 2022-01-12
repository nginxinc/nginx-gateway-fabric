package configuration

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// httpListener defines an HTTP Listener.
type httpListener struct {
	// hosts include all Hosts that belong to the listener.
	hosts map[string]*Host
	// httpRoutes include all HTTPRoute resources that belong to the listener.
	httpRoutes map[string]*v1alpha2.HTTPRoute
}

// Host is the primary configuration unit of the internal representation.
// It corresponds to an NGINX server block with server_name Value;
// See https://nginx.org/en/docs/http/ngx_http_core_module.html#server
type Host struct {
	// Value is the host value (hostname).
	Value      string
	// PathRouteGroups include all PathRouteGroups that belong to the Host.
	// We use a slice rather than a map to control the order of the routes.
	PathRouteGroups []*PathRouteGroup
}

// PathRouteGroup represents a collection of Routes grouped by a path.
// Among those Routes, there will be routing rules with additional matching criteria. For example, matching of headers.
// The reason we group Routes by Path is how NGINX processes requests: its primary routing rule mechanism is a location block.
// See https://nginx.org/en/docs/http/ngx_http_core_module.html#location
type PathRouteGroup struct {
	// Path is the path (URI).
	Path string
	// Routes include all Routes for that path.
	// Routes are sorted based on the creation timestamp and namespace/name of the Route source to resolve conflicts among
	// multiple same Routes.
	// Sorting is stable so that the Routes retain the order of appearance of the corresponding matches in the corresponding
	// HTTPRoute resources.
	// The first "fired" Route will win in the NGINX configuration.
	Routes []Route
}

// Route represents a Route, which corresponds to a Match in the HTTPRouteRule. If a rule doesn't define any matches,
// that rule is for "/" path.
type Route struct {
	// MatchIdx is the index of the rule in the Rule.Matches or -1 if there are no matches.
	MatchIdx int
	// Rule is the corresponding Routing rule
	Rule     *v1alpha2.HTTPRouteRule
	// Source is the corresponding HTTPRoute resource.
	Source   *v1alpha2.HTTPRoute
}

// Operation defines an operation to be performed for a Host.
type Operation int

const (
	// Delete the config for the Host.
	Delete Operation = iota
	// Upsert the config for the Host.
	Upsert
)

// Represents a change of the Host that needs to be reflected in the NGINX config.
type Change struct {
	// Op is the operation to be performed.
	Op   Operation
	// Host is a reference to the Host associated with the Change.
	Host *Host
}

// Notification represents status information to be reported about an API resource.
type Notification struct {
	// Object is the API resource.
	Object  runtime.Object
	// status information - to be defined
}

// Configuration represents the configuration of the Gateway - a collection of routing rules ready to be transformed
// into NGINX configuration.
type Configuration struct {
	// caches of valid resources
	gatewayClass *v1alpha2.GatewayClass
	gateway      *v1alpha2.Gateway
	httpRoutes   map[string]*v1alpha2.HTTPRoute

	// internal representation of Gateway configuration
	httpListeners map[string]*httpListener
}

func NewConfiguration() *Configuration {
	return &Configuration{
		httpRoutes:    make(map[string]*v1alpha2.HTTPRoute),
		httpListeners: make(map[string]*httpListener),
	}
}

func (c *Configuration) UpsertGatewayClass(gc *v1alpha2.GatewayClass) ([]Change) {
	// validate
	err := validation.ValidateGatewayClass(gc)
	if err != nil {
		c.gatewayClass = nil

		// create changes
		// create notifications (TO-DO)
		return c.updateListeners()
	}

	c.gatewayClass = gc

	// create changes
	// create notifications (TO-DO)
	return c.updateListeners()
}

func (c *Configuration) DeleteGatewayClass() []Change {
	c.gatewayClass = nil
	// create changes
	// create notifications (TO-DO)

	return c.updateListeners()
}

func (c *Configuration) UpsertGateway(gw *v1alpha2.Gateway) []Change {
	err := validation.ValidateGateway(gw)
	if err != nil {
		c.gateway = nil

		// create changes
		// create notifications (TO-DO)
		return c.updateListeners()
	}

	c.gateway = gw

	// create changes
	// create notifications (TO-DO)
	return c.updateListeners()
}

func (c *Configuration) DeleteGateway() []Change {
	c.gateway = nil

	// create changes
	// create notifications (TO-DO)

	return c.updateListeners()
}

func (c *Configuration) UpsertHTTPRoute(httpRoute *v1alpha2.HTTPRoute) []Change {
	key := getResourceKey(&httpRoute.ObjectMeta)

	err := validation.ValidateHTTPRoute(httpRoute)
	if err != nil {
		delete(c.httpRoutes, key)

		// create changes
		// create notifications (TO-DO)
		return c.updateListeners()
	}

	c.httpRoutes[key] = httpRoute

	// create changes
	// create notifications (TO-DO)
	return c.updateListeners()
}

func (c *Configuration) DeleteHTTPRoute(key string) []Change {
	delete(c.httpRoutes, key)

	// create changes
	// create notifications (TO-DO)
	return c.updateListeners()
}

func (c *Configuration) updateListeners() []Change {
	var changes []Change

	// we only have one http listener

	if c.gatewayClass == nil || c.gateway == nil {
		if _, exist := c.httpListeners["http"]; !exist {
			return changes
		}

		_, changes := rebuildHTTPListener(c.httpListeners["http"], c.httpRoutes)

		delete(c.httpListeners, "http")
		return changes
	}

	if _, exist := c.httpListeners["http"]; !exist {
		c.httpListeners["http"] = &httpListener{
			hosts: make(map[string]*Host),
		}
	}

	c.httpListeners["http"], changes = rebuildHTTPListener(c.httpListeners["http"], c.httpRoutes)

	return changes
}

func rebuildHTTPListener(listener *httpListener, httpRoutes map[string]*v1alpha2.HTTPRoute) (*httpListener, []Change) {
	// (1) Build new hosts

	pathRoutesForHost := make(map[string]map[string]*PathRouteGroup)

	// for now, we take in all availble HTTPRoutes
	for _, key := range getSortedKeysForHTTPRoutes(httpRoutes) {
		hr := httpRoutes[key]

		// every hostname x every routing rule
		for _, h := range hr.Spec.Hostnames {
			pathRoutes, exist := pathRoutesForHost[string(h)]
			if !exist {
				pathRoutes = make(map[string]*PathRouteGroup)
				pathRoutesForHost[string(h)] = pathRoutes
			}

			for i := range hr.Spec.Rules {
				rule := &hr.Spec.Rules[i]

				if len(rule.Matches) == 0 {
					pathRoute, exist := pathRoutes["/"]
					if !exist {
						pathRoute = &PathRouteGroup{
							Path: "/",
						}
						pathRoutes["/"] = pathRoute
					}

					pathRoute.Routes = append(pathRoute.Routes, Route{
						Rule:     rule,
						MatchIdx: -1,
						Source:   hr,
					})
				} else {
					for i, m := range rule.Matches {
						path := "/"
						if m.Path != nil && m.Path.Value != nil && *m.Path.Value != "/" {
							path = *m.Path.Value
						}

						pathRoute, exist := pathRoutes[path]
						if !exist {
							pathRoute = &PathRouteGroup{
								Path: path,
							}
							pathRoutes[path] = pathRoute
						}

						pathRoute.Routes = append(pathRoute.Routes, Route{
							Rule:     rule,
							MatchIdx: i,
							Source:   hr,
						})
					}
				}
			}
		}
	}

	//  resolve any route conflicts

	newHosts := make(map[string]*Host)

	for h, pathRoutes := range pathRoutesForHost {
		host := &Host{
			Value: h,
		}

		// this sorting will mess up the order of routes in the HTTPRoutes
		// the order of routes can be important when regexes are used
		// See https://nginx.org/en/docs/http/ngx_http_core_module.html#location to learn how NGINX searches for
		// a location.
		for _, path := range getSortedKeysForPathRoutes(pathRoutes) {
			pathRoute := pathRoutes[path]
			sortRoutes(pathRoute.Routes)

			host.PathRouteGroups = append(host.PathRouteGroups, pathRoute)
		}

		newHosts[h] = host
	}

	// (2) Determine changes in hosts

	var removedHosts, updatedHosts, addedHosts []string

	for _, h := range getSortedKeysForHosts(listener.hosts) {
		_, exists := newHosts[h]
		if !exists {
			removedHosts = append(removedHosts, h)
		}
	}

	for _, h := range getSortedKeysForHosts(newHosts) {
		_, exists := listener.hosts[h]
		if !exists {
			addedHosts = append(addedHosts, h)
		}
	}

	for _, h := range getSortedKeysForHosts(newHosts) {
		oldHost, exists := listener.hosts[h]
		if !exists {
			continue
		}

		if !arePathRoutesEqual(oldHost.PathRouteGroups, newHosts[h].PathRouteGroups) {
			updatedHosts = append(updatedHosts, h)
		}
	}

	// (3) Create changes

	var changes []Change

	for _, h := range removedHosts {
		change := Change{
			Op:   Delete,
			Host: listener.hosts[h],
		}
		changes = append(changes, change)
	}

	for _, h := range updatedHosts {
		change := Change{
			Op:   Upsert,
			Host: newHosts[h],
		}
		changes = append(changes, change)
	}

	for _, h := range addedHosts {
		change := Change{
			Op:   Upsert,
			Host: newHosts[h],
		}
		changes = append(changes, change)
	}

	// (4) Create a new listener

	newListener := &httpListener{
		hosts: newHosts,
		// httpListeners: newHTTPListeners,
		// we can compare httpListeners with listener.httpListeners to determinate which HTTPRoutes no longer
		// handled by the Gateway, so that we can update their statuses.
	}

	return newListener, changes
}

func arePathRoutesEqual(pathRoutes1, pathRoutes2 []*PathRouteGroup) bool {
	if len(pathRoutes1) != len(pathRoutes2) {
		return false
	}

	for i := 0; i < len(pathRoutes1); i++ {
		if pathRoutes1[i].Path != pathRoutes2[i].Path {
			return false
		}

		if len(pathRoutes1[i].Routes) != len(pathRoutes2[i].Routes) {
			return false
		}

		for j := 0; j < len(pathRoutes1[i].Routes); j++ {
			if !compareObjectMetas(&pathRoutes1[i].Routes[j].Source.ObjectMeta, &pathRoutes2[i].Routes[j].Source.ObjectMeta) {
				return false
			}

			// this might not be needed - the comparison above might be enough
			if !reflect.DeepEqual(pathRoutes1[i].Routes[j].Rule, pathRoutes2[i].Routes[j].Rule) {
				return false
			}
		}
	}

	return true
}

func compareObjectMetas(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	// if the spec of a resource is updated, its Generation increases
	// note: annotations are not part of the spec, so their update don't affect the Generation
	return meta1.Namespace == meta2.Namespace &&
		meta1.Name == meta2.Name &&
		meta1.Generation == meta2.Generation
}

func sortRoutes(routes []Route) {
	// stable sort is used so that the order of matches (as defined in each HttpRoute) is preserved
	sort.SliceStable(routes, func(i, j int) bool {
		return lessObjectMeta(&routes[i].Source.ObjectMeta, &routes[j].Source.ObjectMeta)
	})
}

func lessObjectMeta(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		return getResourceKey(meta1) < getResourceKey(meta2)
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}

func getSortedKeysForPathRoutes(pathRoutes map[string]*PathRouteGroup) []string {
	var keys []string

	for k := range pathRoutes {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedKeysForHTTPRoutes(httpRoutes map[string]*v1alpha2.HTTPRoute) []string {
	var keys []string

	for k := range httpRoutes {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedKeysForHosts(hosts map[string]*Host) []string {
	var keys []string

	for k := range hosts {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
