package state

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// httpListener defines an HTTP Listener.
type httpListener struct {
	// hosts include all Hosts that belong to the listener.
	hosts hosts
	// httpRoutes include all HTTPRoute resources that belong to the listener.
	httpRoutes httpRoutes
}

type hosts map[string]*Host

func (hs hosts) Keys() []string {
	keys := make([]string, 0, len(hs))

	for k := range hs {
		keys = append(keys, k)
	}

	return keys
}

type httpRoutes map[string]*v1alpha2.HTTPRoute

func (hrs httpRoutes) Keys() []string {
	keys := make([]string, 0, len(hrs))

	for k := range hrs {
		keys = append(keys, k)
	}

	return keys
}

// Host is the primary configuration unit of the internal representation.
// It corresponds to an NGINX server block with server_name Value;
// See https://nginx.org/en/docs/http/ngx_http_core_module.html#server
type Host struct {
	// Value is the host value (hostname).
	Value string
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
	// Routes are sorted based on the creation timestamp and namespace/name of the Route source (HTTPRoute). This way
	// the ordering resolves the conflicts among any conflicting rules.
	// Sorting is stable so that the Routes retain the order of appearance of the corresponding matches in the corresponding
	// HTTPRoute resources.
	// The first "fired" Route will win in the NGINX configuration.
	Routes []Route
}

type pathRoutesForHosts map[string]*PathRouteGroup

func (prs pathRoutesForHosts) Keys() []string {
	keys := make([]string, 0, len(prs))

	for k := range prs {
		keys = append(keys, k)
	}

	return keys
}

// Route represents a Route, which corresponds to a Match in the HTTPRouteRule. If a rule doesn't define any matches,
// it is assumed that the rule is for "/" path.
type Route struct {
	// MatchIdx is the index of the rule in the Rule.Matches or -1 if there are no matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
	// Source is the corresponding HTTPRoute resource.
	Source *v1alpha2.HTTPRoute
}

// Operation defines an operation to be performed for a Host.
type Operation int

const (
	// Delete the config for the Host.
	Delete Operation = iota
	// Upsert the config for the Host.
	Upsert
)

// Change represents a change of the Host that needs to be reflected in the NGINX config.
type Change struct {
	// Op is the operation to be performed.
	Op Operation
	// Host is a reference to the Host associated with the Change.
	Host *Host
}

// StatusUpdate represents an update to the status of a resource.
type StatusUpdate struct {
	// Object is the resource.
	Object runtime.Object
	// Status is the status field of the resource
	// The Status include only the new conditions. This means that the status reporter component will need to merge
	// the new conditions with the existing conditions of the resource.
	Status interface{}
}

// Configuration represents the configuration of the Gateway - a collection of routing rules ready to be transformed
// into NGINX configuration.
type Configuration struct {
	// caches of valid resources
	httpRoutes httpRoutes

	// internal representation of Gateway configuration
	httpListeners map[string]*httpListener

	gatewayCtlrName string
	clock           Clock
}

// NewConfiguration creates a Configuration.
func NewConfiguration(gatewayCtlrName string, clock Clock) *Configuration {
	c := &Configuration{
		httpRoutes:      make(httpRoutes),
		httpListeners:   make(map[string]*httpListener),
		gatewayCtlrName: gatewayCtlrName,
		clock:           clock,
	}

	// Until we process the GatewayClass and Gateway resources, we assume the "http" listener always exists.
	c.httpListeners["http"] = &httpListener{
		hosts: make(hosts),
	}

	return c
}

// UpsertHTTPRoute upserts an HTTPRoute into the Configuration.
func (c *Configuration) UpsertHTTPRoute(httpRoute *v1alpha2.HTTPRoute) ([]Change, []StatusUpdate) {
	key := getResourceKey(&httpRoute.ObjectMeta)

	oldHR, exist := c.httpRoutes[key]
	if exist && compareObjectMetas(&oldHR.ObjectMeta, &httpRoute.ObjectMeta) {
		// nothing to do - the resource hasn't changed
		return nil, nil
	}

	c.httpRoutes[key] = httpRoute

	return c.updateListeners()
}

// DeleteHTTPRoute deletes an HTTPRoute from the Configuration.
func (c *Configuration) DeleteHTTPRoute(nsname types.NamespacedName) ([]Change, []StatusUpdate) {
	delete(c.httpRoutes, nsname.String())

	return c.updateListeners()
}

func (c *Configuration) updateListeners() ([]Change, []StatusUpdate) {
	var changes []Change

	// for now, we support only one listener

	c.httpListeners["http"], changes = rebuildHTTPListener(c.httpListeners["http"], c.httpRoutes)

	var statusUpdates []StatusUpdate

	listener := c.httpListeners["http"]

	// TO-DO: optimize it so that we only update the status of the affected (changed) httpRoutes
	for _, key := range getSortedKeys(listener.httpRoutes) {
		update := StatusUpdate{
			Object: listener.httpRoutes[key],
			Status: &v1alpha2.HTTPRouteStatus{
				RouteStatus: v1alpha2.RouteStatus{
					Parents: []v1alpha2.RouteParentStatus{
						{
							ControllerName: v1alpha2.GatewayController(c.gatewayCtlrName),
							Conditions: []metav1.Condition{
								{
									Type:               string(v1alpha2.ConditionRouteAccepted),
									Status:             "True",
									ObservedGeneration: listener.httpRoutes[key].Generation,
									LastTransitionTime: metav1.NewTime(c.clock.Now()),
									Reason:             string(v1alpha2.ConditionRouteAccepted),
									Message:            "", // TO-DO: figure out a good message
								},
							},
						},
					},
				},
			},
		}
		statusUpdates = append(statusUpdates, update)
	}

	// TO-DO: remove the accepted condition for the excluded (no longer handled) Routes

	return changes, statusUpdates
}

func rebuildHTTPListener(listener *httpListener, httpRoutes httpRoutes) (*httpListener, []Change) {
	pathRoutesForHosts := buildPathRoutesForHosts(httpRoutes)

	newHosts, newHTTPRoutes := buildHostsAndDetermineHTTPRoutes(pathRoutesForHosts)

	removedHosts, updatedHosts, addedHosts := determineChangesInHosts(listener, newHosts)

	changes := createChanges(removedHosts, updatedHosts, addedHosts, listener.hosts, newHosts)

	newListener := &httpListener{
		hosts:      newHosts,
		httpRoutes: newHTTPRoutes,
	}

	return newListener, changes
}

func createChanges(removedHosts []string, updatedHosts []string, addedHosts []string, oldHosts hosts, newHosts hosts) []Change {
	var changes []Change

	for _, h := range removedHosts {
		change := Change{
			Op:   Delete,
			Host: oldHosts[h],
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

	return changes
}

func determineChangesInHosts(listener *httpListener, newHosts hosts) (removedHosts []string, updatedHosts []string, addedHosts []string) {
	for _, h := range getSortedKeys(listener.hosts) {
		_, exists := newHosts[h]
		if !exists {
			removedHosts = append(removedHosts, h)
		}
	}

	for _, h := range getSortedKeys(newHosts) {
		_, exists := listener.hosts[h]
		if !exists {
			addedHosts = append(addedHosts, h)
		}
	}

	for _, h := range getSortedKeys(newHosts) {
		oldHost, exists := listener.hosts[h]
		if !exists {
			continue
		}

		if !arePathRoutesEqual(oldHost.PathRouteGroups, newHosts[h].PathRouteGroups) {
			updatedHosts = append(updatedHosts, h)
		}
	}
	return removedHosts, updatedHosts, addedHosts
}

func buildHostsAndDetermineHTTPRoutes(pathRoutesForHosts map[string]pathRoutesForHosts) (hosts, httpRoutes) {
	hosts := make(hosts)
	routes := make(httpRoutes)

	for h, pathRoutes := range pathRoutesForHosts {
		host := &Host{
			Value: h,
		}

		// This sorting (getSortedKeysForPathRoutes) will mess up the original order of rules in the HTTPRoutes.
		// The order of routes can be important when regexes are used
		// See https://nginx.org/en/docs/http/ngx_http_core_module.html#location to learn how NGINX searches for
		// a location.
		// This comment is to be aware of that. However, it is not yet clear whether it is a problem.
		for _, path := range getSortedKeys(pathRoutes) {
			pathRoute := pathRoutes[path]

			sortRoutes(pathRoute.Routes)

			host.PathRouteGroups = append(host.PathRouteGroups, pathRoute)

			for _, r := range pathRoute.Routes {
				key := getResourceKey(&r.Source.ObjectMeta)
				routes[key] = r.Source
			}
		}

		hosts[h] = host
	}
	return hosts, routes
}

func buildPathRoutesForHosts(httpRoutes httpRoutes) map[string]pathRoutesForHosts {
	pathRoutesForHosts := make(map[string]pathRoutesForHosts)

	// for now, we take in all available HTTPRoutes
	for _, key := range getSortedKeys(httpRoutes) {
		hr := httpRoutes[key]

		// every hostname x every routing rule
		for _, h := range hr.Spec.Hostnames {
			pathRoutes, exist := pathRoutesForHosts[string(h)]
			if !exist {
				pathRoutes = make(map[string]*PathRouteGroup)
				pathRoutesForHosts[string(h)] = pathRoutes
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
						MatchIdx: -1,
						RuleIdx:  i,
						Source:   hr,
					})
				} else {
					for j, m := range rule.Matches {
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
							MatchIdx: j,
							RuleIdx:  i,
							Source:   hr,
						})
					}
				}
			}
		}
	}

	return pathRoutesForHosts
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

			// DeepEqual might not be needed - the comparison above might be enough
			idx1 := pathRoutes1[i].Routes[j].RuleIdx
			rule1 := pathRoutes1[i].Routes[j].Source.Spec.Rules[idx1]

			idx2 := pathRoutes2[i].Routes[j].RuleIdx
			rule2 := pathRoutes2[i].Routes[j].Source.Spec.Rules[idx2]

			if !reflect.DeepEqual(rule1, rule2) {
				return false
			}
		}
	}

	return true
}

func compareObjectMetas(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	// Two resources are different if:
	// (1) They have different namespaces or names.
	// (2) They have the same namespace and name (resources are the same resource) but their specs are different.
	// If their specs are different, their Generations are different too. So we only test their Generations.
	// note: annotations are not part of the spec, so their update doesn't affect the Generation.
	return meta1.Namespace == meta2.Namespace &&
		meta1.Name == meta2.Name &&
		meta1.Generation == meta2.Generation
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
