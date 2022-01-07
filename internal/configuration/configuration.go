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

type Configuration struct {
	// caches of valid resources
	gatewayClass *v1alpha2.GatewayClass
	gateway      *v1alpha2.Gateway
	httpRoutes   map[string]*v1alpha2.HTTPRoute

	// internal representation of Gateway configuration
	httpListeners map[string]*httpListener
}

type httpListener struct {
	hosts map[string]*Host
	// httpRoutes map[string]*v1alpha2.HTTPRoute - to quickly answer if a route is part of the listener
}

// corresponds to an NGINX server block with server_name Value;
type Host struct {
	Value  string
	Routes []*PathRoute
}

// NGINX uses locations as a primary routing mechanism
// Because of that, we will represent routing rules as two-layer structure
// First layer - PathRoutes, each for its own path
// Second layer - Routes for a particular path

type PathRoute struct {
	Path string
	// sorted based on the creation timestamp and namespace/name of the source
	// sorting must be stable to preserve the order in HTTPRoute
	Routes []Route
}

type Route struct {
	MatchIdx int // id of the rule in Rule.Matches or -1 if there are no matches
	Rule     *v1alpha2.HTTPRouteRule
	Source   *v1alpha2.HTTPRoute // where the Route comes from
}

type Operation int

const (
	Delete Operation = iota
	Upsert
)

type Change struct {
	Op   Operation
	Host *Host
}

type Notification struct {
	Object  runtime.Object
	Reason  string
	Message string
}

func NewConfiguration() *Configuration {
	return &Configuration{
		httpRoutes:    make(map[string]*v1alpha2.HTTPRoute),
		httpListeners: make(map[string]*httpListener),
	}
}

func (c *Configuration) UpsertGatewayClass(gc *v1alpha2.GatewayClass) []Change {
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

	pathRoutesForHost := make(map[string]map[string]*PathRoute)

	// for now, we take in all availble HTTPRoutes
	for _, key := range getSortedKeysForHTTPRoutes(httpRoutes) {
		hr := httpRoutes[key]

		// every hostname x every routing rule
		for _, h := range hr.Spec.Hostnames {
			pathRoutes, exist := pathRoutesForHost[string(h)]
			if !exist {
				pathRoutes = make(map[string]*PathRoute)
				pathRoutesForHost[string(h)] = pathRoutes
			}

			for i := range hr.Spec.Rules {
				rule := &hr.Spec.Rules[i]

				if len(rule.Matches) == 0 {
					pathRoute, exist := pathRoutes["/"]
					if !exist {
						pathRoute = &PathRoute{
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
							pathRoute = &PathRoute{
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

			host.Routes = append(host.Routes, pathRoute)
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

		if !arePathRoutesEqual(oldHost.Routes, newHosts[h].Routes) {
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

func arePathRoutesEqual(pathRoutes1, pathRoutes2 []*PathRoute) bool {
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

func getSortedKeysForPathRoutes(pathRoutes map[string]*PathRoute) []string {
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
