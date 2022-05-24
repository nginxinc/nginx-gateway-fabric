package state

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// Configuration is an internal representation of Gateway configuration.
// We can think of Configuration as an intermediate state between the Gateway API resources and the data plane (NGINX)
// configuration.
type Configuration struct {
	// HTTPServers holds all HTTPServers.
	// FIXME(pleshakov) We assume that all servers are HTTP and listen on port 80.
	HTTPServers []HTTPServer
}

// HTTPServer is a virtual server.
type HTTPServer struct {
	// Hostname is the hostname of the server.
	Hostname string
	// PathRules is a collection of routing rules.
	PathRules []PathRule
}

// PathRule represents routing rules that share a common path.
type PathRule struct {
	// Path is a path. For example, '/hello'.
	Path string
	// MatchRules holds routing rules.
	MatchRules []MatchRule
}

// MatchRule represents a routing rule. It corresponds directly to a Match in the HTTPRoute resource.
type MatchRule struct {
	// MatchIdx is the index of the rule in the Rule.Matches or -1 if there are no matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
	// Source is the corresponding HTTPRoute resource.
	Source *v1alpha2.HTTPRoute
}

// GetMatch returns the HTTPRouteMatch of the Route and true if it exists.
// If there is no Match defined on the Route, GetMatch returns an empty HTTPRouteMatch and false.
func (r *MatchRule) GetMatch() (v1alpha2.HTTPRouteMatch, bool) {
	if r.MatchIdx == -1 {
		return v1alpha2.HTTPRouteMatch{}, false
	}
	return r.Source.Spec.Rules[r.RuleIdx].Matches[r.MatchIdx], true
}

// buildConfiguration builds the Configuration from the graph.
func buildConfiguration(graph *graph) Configuration {
	// FIXME(pleshakov) For now we only handle paths with prefix matches. Handle exact and regex matches
	pathRulesForHosts := make(map[string]map[string]PathRule)

	for _, l := range graph.Listeners {
		for _, r := range l.Routes {
			var hostnames []string

			for _, h := range r.Source.Spec.Hostnames {
				if _, exist := l.AcceptedHostnames[string(h)]; exist {
					hostnames = append(hostnames, string(h))
				}
			}

			for _, h := range hostnames {
				if _, exist := pathRulesForHosts[h]; !exist {
					pathRulesForHosts[h] = make(map[string]PathRule)
				}
			}

			for i, rule := range r.Source.Spec.Rules {
				for _, h := range hostnames {
					for j, m := range rule.Matches {
						path := getPath(m.Path)

						rule, exist := pathRulesForHosts[h][path]
						if !exist {
							rule.Path = path
						}

						rule.MatchRules = append(rule.MatchRules, MatchRule{
							MatchIdx: j,
							RuleIdx:  i,
							Source:   r.Source,
						})

						pathRulesForHosts[h][path] = rule
					}
				}
			}
		}
	}

	httpServers := make([]HTTPServer, 0, len(pathRulesForHosts))

	for h, rules := range pathRulesForHosts {
		s := HTTPServer{
			Hostname:  h,
			PathRules: make([]PathRule, 0, len(rules)),
		}

		for _, r := range rules {
			// sort matches in every PathRule based on the Source timestamp and its namespace/name
			// for conflict resolution of conflicting rules
			// stable sort so that the order of matches within one HTTPRoute is preserved
			sort.SliceStable(r.MatchRules, func(i, j int) bool {
				return lessObjectMeta(&r.MatchRules[i].Source.ObjectMeta, &r.MatchRules[j].Source.ObjectMeta)
			})

			s.PathRules = append(s.PathRules, r)
		}

		// sort rules for predictable order
		sort.Slice(s.PathRules, func(i, j int) bool {
			return s.PathRules[i].Path < s.PathRules[j].Path
		})

		httpServers = append(httpServers, s)
	}

	// sort servers for predictable order
	sort.Slice(httpServers, func(i, j int) bool {
		return httpServers[i].Hostname < httpServers[j].Hostname
	})

	return Configuration{
		HTTPServers: httpServers,
	}
}

func lessObjectMeta(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		if meta1.Namespace == meta2.Namespace {
			return meta1.Name < meta2.Name
		}
		return meta1.Namespace < meta2.Namespace
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}

func getPath(path *v1alpha2.HTTPPathMatch) string {
	if path == nil || path.Value == nil || *path.Value == "" {
		return "/"
	}
	return *path.Value
}
