package dataplane

import (
	"context"
	"fmt"
	"sort"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
)

const wildcardHostname = "~^"

// Configuration is an intermediate representation of dataplane configuration.
type Configuration struct {
	Key   string
	Ports graph.GatewayPorts
	// HTTPServers holds all HTTPServers.
	// FIXME(pleshakov) We assume that all servers are HTTP and listen on port 80.
	HTTPServers []VirtualServer
	// SSLServers holds all SSLServers.
	// FIXME(kate-osborn) We assume that all SSL servers listen on port 443.
	SSLServers []VirtualServer
	// Upstreams holds all unique Upstreams.
	Upstreams []Upstream
	// BackendGroups holds all unique BackendGroups.
	// FIXME(pleshakov): Ensure Configuration doesn't include types from the graph package.
	BackendGroups []graph.BackendGroup
}

// VirtualServer is a virtual server.
type VirtualServer struct {
	// SSL holds the SSL configuration options for the server.
	SSL *SSL
	// Hostname is the hostname of the server.
	Hostname string
	// PathRules is a collection of routing rules.
	PathRules []PathRule
	// IsDefault indicates whether the server is the default server.
	IsDefault bool
}

type Upstream struct {
	// Name is the name of the Upstream. Will be unique for each service/port combination.
	Name string
	// ErrorMsg contains the error message if the Upstream is invalid.
	ErrorMsg string
	// Endpoints are the endpoints of the Upstream.
	Endpoints []resolver.Endpoint
}

type SSL struct {
	// CertificatePath is the path to the certificate file.
	CertificatePath string
}

// PathRule represents routing rules that share a common path.
type PathRule struct {
	// Path is a path. For example, '/hello'.
	Path string
	// MatchRules holds routing rules.
	MatchRules []MatchRule
}

// InvalidFilter is a special filter for handling the case when configured filters are invalid.
type InvalidFilter struct{}

// Filters hold the filters for a MatchRule.
type Filters struct {
	InvalidFilter   *InvalidFilter
	RequestRedirect *v1beta1.HTTPRequestRedirectFilter
}

// MatchRule represents a routing rule. It corresponds directly to a Match in the HTTPRoute resource.
// An HTTPRoute is guaranteed to have at least one rule with one match.
// If no rule or match is specified by the user, the default rule {{path:{ type: "PathPrefix", value: "/"}}}
// is set by the schema.
type MatchRule struct {
	// Filters holds the filters for the MatchRule.
	Filters Filters
	// Source is the corresponding HTTPRoute resource.
	// FIXME(pleshakov): Consider referencing only the parts needed for the config generation rather than
	// the entire resource.
	Source *v1beta1.HTTPRoute
	// BackendGroup is the group of Backends that the rule routes to.
	BackendGroup graph.BackendGroup
	// MatchIdx is the index of the rule in the Rule.Matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
}

// GetMatch returns the HTTPRouteMatch of the Route .
func (r *MatchRule) GetMatch() v1beta1.HTTPRouteMatch {
	return r.Source.Spec.Rules[r.RuleIdx].Matches[r.MatchIdx]
}

// BuildConfiguration builds the Configuration from the Graph.
// FIXME(pleshakov) For now we only handle paths with prefix matches. Handle exact and regex matches
func BuildConfiguration(ctx context.Context, g *graph.Graph, resolver resolver.ServiceResolver) []Configuration {
	if g.GatewayClass == nil || !g.GatewayClass.Valid {
		return nil
	}

	if len(g.Gateways) == 0 {
		return nil
	}

	confs := make([]Configuration, 0, len(g.Gateways))

	// for each Gateway, build its servers

	for _, gw := range g.Gateways {
		upstreamsMap := buildUpstreamsMap(ctx, gw.Listeners, resolver)
		httpServers, sslServers := buildServers(gw.Listeners)
		backendGroups := buildBackendGroups(gw.Listeners)

		config := Configuration{
			Key:           fmt.Sprintf("gateway__%s__%s", gw.Source.Namespace, gw.Source.Name),
			Ports:         gw.Ports,
			HTTPServers:   httpServers,
			SSLServers:    sslServers,
			Upstreams:     upstreamsMapToSlice(upstreamsMap),
			BackendGroups: backendGroups,
		}

		confs = append(confs, config)
	}

	return confs
}

func upstreamsMapToSlice(upstreamsMap map[string]Upstream) []Upstream {
	if len(upstreamsMap) == 0 {
		return nil
	}

	upstreams := make([]Upstream, 0, len(upstreamsMap))
	for _, upstream := range upstreamsMap {
		upstreams = append(upstreams, upstream)
	}

	return upstreams
}

func buildBackendGroups(listeners map[string]*graph.Listener) []graph.BackendGroup {
	// There can be duplicate backend groups if a route is attached to multiple listeners.
	// We use a map to deduplicate them.
	uniqueGroups := make(map[string]graph.BackendGroup)

	for _, l := range listeners {

		if !l.Valid {
			continue
		}

		for _, r := range l.Routes {
			for _, group := range r.GetAllBackendGroups() {
				if _, ok := uniqueGroups[group.GroupName()]; !ok {
					uniqueGroups[group.GroupName()] = group
				}
			}
		}
	}

	numGroups := len(uniqueGroups)
	if len(uniqueGroups) == 0 {
		return nil
	}

	groups := make([]graph.BackendGroup, 0, numGroups)
	for _, group := range uniqueGroups {
		groups = append(groups, group)
	}

	return groups
}

func buildServers(listeners map[string]*graph.Listener) (http, ssl []VirtualServer) {
	rulesForProtocol := map[v1beta1.ProtocolType]*hostPathRules{
		v1beta1.HTTPProtocolType:  newHostPathRules(),
		v1beta1.HTTPSProtocolType: newHostPathRules(),
	}

	for _, l := range listeners {
		if l.Valid {
			rules := rulesForProtocol[l.Source.Protocol]
			rules.upsertListener(l)
		}
	}

	httpRules := rulesForProtocol[v1beta1.HTTPProtocolType]
	sslRules := rulesForProtocol[v1beta1.HTTPSProtocolType]

	return httpRules.buildServers(), sslRules.buildServers()
}

type hostPathRules struct {
	rulesPerHost     map[string]map[string]PathRule
	listenersForHost map[string]*graph.Listener
	httpsListeners   []*graph.Listener
	listenersExist   bool
}

func newHostPathRules() *hostPathRules {
	return &hostPathRules{
		rulesPerHost:     make(map[string]map[string]PathRule),
		listenersForHost: make(map[string]*graph.Listener),
		httpsListeners:   make([]*graph.Listener, 0),
	}
}

func (hpr *hostPathRules) upsertListener(l *graph.Listener) {
	hpr.listenersExist = true

	if l.Source.Protocol == v1beta1.HTTPSProtocolType {
		hpr.httpsListeners = append(hpr.httpsListeners, l)
	}

	for _, r := range l.Routes {
		var hostnames []string

		for _, h := range r.Source.Spec.Hostnames {
			if _, exist := l.AcceptedHostnames[string(h)]; exist {
				hostnames = append(hostnames, string(h))
			}
		}

		for _, h := range hostnames {
			hpr.listenersForHost[h] = l

			if _, exist := hpr.rulesPerHost[h]; !exist {
				hpr.rulesPerHost[h] = make(map[string]PathRule)
			}
		}

		for i, rule := range r.Source.Spec.Rules {
			if !r.Rules[i].ValidMatches {
				continue
			}

			var filters Filters
			if r.Rules[i].ValidFilters {
				filters = createFilters(rule.Filters)
			} else {
				filters = Filters{
					InvalidFilter: &InvalidFilter{},
				}
			}

			for _, h := range hostnames {
				for j, m := range rule.Matches {
					path := getPath(m.Path)

					rule, exist := hpr.rulesPerHost[h][path]
					if !exist {
						rule.Path = path
					}

					rule.MatchRules = append(rule.MatchRules, MatchRule{
						MatchIdx:     j,
						RuleIdx:      i,
						Source:       r.Source,
						BackendGroup: r.Rules[i].BackendGroup,
						Filters:      filters,
					})

					hpr.rulesPerHost[h][path] = rule
				}
			}
		}
	}
}

func (hpr *hostPathRules) buildServers() []VirtualServer {
	servers := make([]VirtualServer, 0, len(hpr.rulesPerHost)+len(hpr.httpsListeners))

	for h, rules := range hpr.rulesPerHost {
		s := VirtualServer{
			Hostname:  h,
			PathRules: make([]PathRule, 0, len(rules)),
		}

		l, ok := hpr.listenersForHost[h]
		if !ok {
			panic(fmt.Sprintf("no listener found for hostname: %s", h))
		}

		if l.SecretPath != "" {
			s.SSL = &SSL{CertificatePath: l.SecretPath}
		}

		for _, r := range rules {
			sortMatchRules(r.MatchRules)

			s.PathRules = append(s.PathRules, r)
		}

		// We sort the path rules so the order is preserved after reconfiguration.
		sort.Slice(s.PathRules, func(i, j int) bool {
			return s.PathRules[i].Path < s.PathRules[j].Path
		})

		servers = append(servers, s)
	}

	for _, l := range hpr.httpsListeners {
		hostname := getListenerHostname(l.Source.Hostname)
		// generate a 404 ssl server block for listeners with no routes or listeners with wildcard (match-all) routes
		// FIXME(kate-osborn): when we support regex hostnames (e.g. *.example.com)
		// we will have to modify this check to catch regex hostnames.
		if len(l.Routes) == 0 || hostname == wildcardHostname {
			s := VirtualServer{
				Hostname: hostname,
			}

			if l.SecretPath != "" {
				s.SSL = &SSL{CertificatePath: l.SecretPath}
			}

			servers = append(servers, s)
		}
	}

	// if any listeners exist, we need to generate a default server block.
	if hpr.listenersExist {
		servers = append(servers, VirtualServer{IsDefault: true})
	}

	// We sort the servers so the order is preserved after reconfiguration.
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})

	return servers
}

func buildUpstreamsMap(
	ctx context.Context,
	listeners map[string]*graph.Listener,
	resolver resolver.ServiceResolver,
) map[string]Upstream {
	// There can be duplicate upstreams if multiple routes reference the same upstream.
	// We use a map to deduplicate them.
	uniqueUpstreams := make(map[string]Upstream)

	for _, l := range listeners {

		if !l.Valid {
			continue
		}

		for _, route := range l.Routes {
			for _, group := range route.GetAllBackendGroups() {
				for _, backend := range group.Backends {
					if name := backend.Name; name != "" {
						_, exist := uniqueUpstreams[name]

						if exist {
							continue
						}

						var errMsg string

						eps, err := resolver.Resolve(ctx, backend.Svc, backend.Port)
						if err != nil {
							errMsg = err.Error()
						}

						uniqueUpstreams[name] = Upstream{
							Name:      name,
							Endpoints: eps,
							ErrorMsg:  errMsg,
						}
					}
				}
			}
		}
	}

	return uniqueUpstreams
}

func getListenerHostname(h *v1beta1.Hostname) string {
	if h == nil || *h == "" {
		return wildcardHostname
	}

	return string(*h)
}

func getPath(path *v1beta1.HTTPPathMatch) string {
	if path == nil || path.Value == nil || *path.Value == "" {
		return "/"
	}
	return *path.Value
}

func createFilters(filters []v1beta1.HTTPRouteFilter) Filters {
	var result Filters

	for _, f := range filters {
		switch f.Type {
		case v1beta1.HTTPRouteFilterRequestRedirect:
			result.RequestRedirect = f.RequestRedirect
			// using the first filter
			return result
		}
	}

	return result
}
