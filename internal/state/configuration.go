package state

import (
	"sort"

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
// An HTTPRoute is guaranteed to have at least one rule with one match.
// If no rule or match is specified by the user, the default rule {{path:{ type: "PathPrefix", value: "/"}}} is set by the schema.
type MatchRule struct {
	// MatchIdx is the index of the rule in the Rule.Matches.
	MatchIdx int
	// RuleIdx is the index of the corresponding rule in the HTTPRoute.
	RuleIdx int
	// Source is the corresponding HTTPRoute resource.
	Source *v1alpha2.HTTPRoute
}

// GetMatch returns the HTTPRouteMatch of the Route .
func (r *MatchRule) GetMatch() v1alpha2.HTTPRouteMatch {
	return r.Source.Spec.Rules[r.RuleIdx].Matches[r.MatchIdx]
}

// buildConfiguration builds the Configuration from the graph.
func buildConfiguration(graph *graph) Configuration {
	if graph.GatewayClass == nil || !graph.GatewayClass.Valid {
		return Configuration{}
	}

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
			sortMatchRules(r.MatchRules)

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

func getPath(path *v1alpha2.HTTPPathMatch) string {
	if path == nil || path.Value == nil || *path.Value == "" {
		return "/"
	}
	return *path.Value
}
