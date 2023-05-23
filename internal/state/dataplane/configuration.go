package dataplane

import (
	"context"
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
)

type PathType string

const (
	wildcardHostname          = "~^"
	PathTypePrefix   PathType = "prefix"
	PathTypeExact    PathType = "exact"
)

// Configuration is an intermediate representation of dataplane configuration.
type Configuration struct {
	// HTTPServers holds all HTTPServers.
	// We assume that all servers are HTTP and listen on port 80.
	HTTPServers []VirtualServer
	// SSLServers holds all SSLServers.
	// We assume that all SSL servers listen on port 443.
	SSLServers []VirtualServer
	// Upstreams holds all unique Upstreams.
	Upstreams []Upstream
	// BackendGroups holds all unique BackendGroups.
	BackendGroups []BackendGroup
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
	// PathType is simplified path type. For example, prefix or exact.
	PathType PathType
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
	Source *v1beta1.HTTPRoute
	// BackendGroup is the group of Backends that the rule routes to.
	BackendGroup BackendGroup
	// MatchIdx is the index of the rule in the Rule.Matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
}

// BackendGroup represents a group of Backends for a routing rule in an HTTPRoute.
type BackendGroup struct {
	// Source is the NamespacedName of the HTTPRoute the group belongs to.
	Source types.NamespacedName
	// Backends is a list of Backends in the Group.
	Backends []Backend
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
}

// Name returns the name of the backend group.
// This name must be unique across all HTTPRoutes and all rules within the same HTTPRoute.
// The RuleIdx is used to make the name unique across all rules within the same HTTPRoute.
// The RuleIdx may change for a given rule if an update is made to the HTTPRoute, but it will always match the index
// of the rule in the stored HTTPRoute.
func (bg *BackendGroup) Name() string {
	return fmt.Sprintf("%s__%s_rule%d", bg.Source.Namespace, bg.Source.Name, bg.RuleIdx)
}

// Backend represents a Backend for a routing rule.
type Backend struct {
	// UpstreamName is the name of the upstream for this backend.
	UpstreamName string
	// Weight is the weight of the BackendRef.
	// The possible values of weight are 0-1,000,000.
	// If weight is 0, no traffic should be forwarded for this entry.
	Weight int32
	// Valid indicates whether the Backend is valid.
	Valid bool
}

// GetMatch returns the HTTPRouteMatch of the Route .
func (r *MatchRule) GetMatch() v1beta1.HTTPRouteMatch {
	return r.Source.Spec.Rules[r.RuleIdx].Matches[r.MatchIdx]
}

// BuildConfiguration builds the Configuration from the Graph.
func BuildConfiguration(ctx context.Context, g *graph.Graph, resolver resolver.ServiceResolver) Configuration {
	if g.GatewayClass == nil || !g.GatewayClass.Valid {
		return Configuration{}
	}

	if g.Gateway == nil {
		return Configuration{}
	}

	upstreams := buildUpstreams(ctx, g.Gateway.Listeners, resolver)
	httpServers, sslServers := buildServers(g.Gateway.Listeners)
	backendGroups := buildBackendGroups(append(httpServers, sslServers...))

	config := Configuration{
		HTTPServers:   httpServers,
		SSLServers:    sslServers,
		Upstreams:     upstreams,
		BackendGroups: backendGroups,
	}

	return config
}

func buildBackendGroups(servers []VirtualServer) []BackendGroup {
	type key struct {
		nsname  types.NamespacedName
		ruleIdx int
	}

	// There can be duplicate backend groups if a route is attached to multiple listeners.
	// We use a map to deduplicate them.
	uniqueGroups := make(map[key]BackendGroup)

	for _, s := range servers {
		for _, pr := range s.PathRules {
			for _, mr := range pr.MatchRules {
				group := mr.BackendGroup

				key := key{
					nsname:  group.Source,
					ruleIdx: group.RuleIdx,
				}

				uniqueGroups[key] = group
			}
		}
	}

	numGroups := len(uniqueGroups)
	if len(uniqueGroups) == 0 {
		return nil
	}

	groups := make([]BackendGroup, 0, numGroups)
	for _, group := range uniqueGroups {
		groups = append(groups, group)
	}

	return groups
}

func newBackendGroup(refs []graph.BackendRef, sourceNsName types.NamespacedName, ruleIdx int) BackendGroup {
	var backends []Backend

	if len(refs) > 0 {
		backends = make([]Backend, 0, len(refs))
	}

	for _, ref := range refs {
		backends = append(backends, Backend{
			UpstreamName: ref.ServicePortReference(),
			Weight:       ref.Weight,
			Valid:        ref.Valid,
		})
	}

	return BackendGroup{
		Backends: backends,
		Source:   sourceNsName,
		RuleIdx:  ruleIdx,
	}
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

type pathAndType struct {
	path     string
	pathType v1beta1.PathMatchType
}

type hostPathRules struct {
	rulesPerHost     map[string]map[pathAndType]PathRule
	listenersForHost map[string]*graph.Listener
	httpsListeners   []*graph.Listener
	listenersExist   bool
}

func newHostPathRules() *hostPathRules {
	return &hostPathRules{
		rulesPerHost:     make(map[string]map[pathAndType]PathRule),
		listenersForHost: make(map[string]*graph.Listener),
		httpsListeners:   make([]*graph.Listener, 0),
	}
}

func (hpr *hostPathRules) upsertListener(l *graph.Listener) {
	hpr.listenersExist = true

	if l.Source.Protocol == v1beta1.HTTPSProtocolType {
		hpr.httpsListeners = append(hpr.httpsListeners, l)
	}

	for routeNsName, r := range l.Routes {
		var hostnames []string
		for _, p := range r.ParentRefs {
			if val, exist := p.Attachment.AcceptedHostnames[string(l.Source.Name)]; exist {
				hostnames = val
			}
		}

		for _, h := range hostnames {
			if prevListener, exists := hpr.listenersForHost[h]; exists {
				// override the previous listener if the new one has a more specific hostname
				if listenerHostnameMoreSpecific(l.Source.Hostname, prevListener.Source.Hostname) {
					hpr.listenersForHost[h] = l
				}
			} else {
				hpr.listenersForHost[h] = l
			}

			if _, exist := hpr.rulesPerHost[h]; !exist {
				hpr.rulesPerHost[h] = make(map[pathAndType]PathRule)
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

					key := pathAndType{
						path:     path,
						pathType: *m.Path.Type,
					}

					rule, exist := hpr.rulesPerHost[h][key]
					if !exist {
						rule.Path = path
						rule.PathType = convertPathType(*m.Path.Type)
					}

					rule.MatchRules = append(rule.MatchRules, MatchRule{
						MatchIdx:     j,
						RuleIdx:      i,
						Source:       r.Source,
						BackendGroup: newBackendGroup(r.Rules[i].BackendRefs, routeNsName, i),
						Filters:      filters,
					})

					hpr.rulesPerHost[h][key] = rule
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
			if s.PathRules[i].Path != s.PathRules[j].Path {
				return s.PathRules[i].Path < s.PathRules[j].Path
			}

			return s.PathRules[i].PathType < s.PathRules[j].PathType
		})

		servers = append(servers, s)
	}

	for _, l := range hpr.httpsListeners {
		hostname := getListenerHostname(l.Source.Hostname)
		// Generate a 404 ssl server block for listeners with no routes or listeners with wildcard (match-all) routes.
		// This server overrides the default ssl server.
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

func buildUpstreams(
	ctx context.Context,
	listeners map[string]*graph.Listener,
	resolver resolver.ServiceResolver,
) []Upstream {
	// There can be duplicate upstreams if multiple routes reference the same upstream.
	// We use a map to deduplicate them.
	uniqueUpstreams := make(map[string]Upstream)

	for _, l := range listeners {

		if !l.Valid {
			continue
		}

		for _, route := range l.Routes {
			for _, rule := range route.Rules {
				if !rule.ValidMatches || !rule.ValidFilters {
					// don't generate upstreams for rules that have invalid matches or filters
					continue
				}
				for _, br := range rule.BackendRefs {
					if br.Valid {
						upstreamName := br.ServicePortReference()
						_, exist := uniqueUpstreams[upstreamName]

						if exist {
							continue
						}

						var errMsg string

						eps, err := resolver.Resolve(ctx, br.Svc, br.Port)
						if err != nil {
							errMsg = err.Error()
						}

						uniqueUpstreams[upstreamName] = Upstream{
							Name:      upstreamName,
							Endpoints: eps,
							ErrorMsg:  errMsg,
						}
					}
				}
			}
		}
	}

	if len(uniqueUpstreams) == 0 {
		return nil
	}

	upstreams := make([]Upstream, 0, len(uniqueUpstreams))

	for _, up := range uniqueUpstreams {
		upstreams = append(upstreams, up)
	}
	return upstreams
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

func convertPathType(pathType v1beta1.PathMatchType) PathType {
	switch pathType {
	case v1beta1.PathMatchPathPrefix:
		return PathTypePrefix
	case v1beta1.PathMatchExact:
		return PathTypeExact
	default:
		panic(fmt.Sprintf("unsupported path type: %s", pathType))
	}
}

// listenerHostnameMoreSpecific returns true if host1 is more specific than host2 (using length).
//
// Since the only caller of this function specifies listener hostnames that are both
// bound to the same route hostname, this function assumes that host1 and host2 match, either
// exactly or as a substring.
//
// For example:
// - foo.example.com and "" (host1 wins)
// Non-example:
// - foo.example.com and bar.example.com (should not be given to this function)
//
// As we add regex support, we should put in the proper
// validation and error handling for this function to ensure that the hostnames are actually matching,
// to avoid the unintended inputs above for the invalid case.
func listenerHostnameMoreSpecific(host1, host2 *v1beta1.Hostname) bool {
	var host1Str, host2Str string
	if host1 != nil {
		host1Str = string(*host1)
	}

	if host2 != nil {
		host2Str = string(*host2)
	}

	return len(host1Str) >= len(host2Str)
}
