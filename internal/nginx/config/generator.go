package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
const nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
type Generator interface {
	// Generate generates NGINX configuration from internal representation.
	Generate(configuration state.Configuration) ([]byte, Warnings)
}

// GeneratorImpl is an implementation of Generator
type GeneratorImpl struct {
	executor     *templateExecutor
	serviceStore state.ServiceStore
}

// NewGeneratorImpl creates a new GeneratorImpl.
func NewGeneratorImpl(serviceStore state.ServiceStore) *GeneratorImpl {
	return &GeneratorImpl{
		executor:     newTemplateExecutor(),
		serviceStore: serviceStore,
	}
}

func (g *GeneratorImpl) Generate(conf state.Configuration) ([]byte, Warnings) {
	http, warnings := generate(conf, g.serviceStore)

	return g.executor.ExecuteForHTTP(http), warnings
}

func generate(conf state.Configuration, serviceStore state.ServiceStore) (http, Warnings) {
	warnings := newWarnings()

	confServers := append(conf.HTTPServers, conf.SSLServers...)

	httpCfg := http{
		// capacity is all the conf servers + default ssl & http servers
		Servers: make([]server, 0, len(confServers)+2),
	}

	if len(conf.HTTPServers) > 0 {
		defaultHTTPServer := generateDefaultHTTPServer()

		httpCfg.Servers = append(httpCfg.Servers, defaultHTTPServer)
	}

	if len(conf.SSLServers) > 0 {
		defaultSSLServer := generateDefaultSSLServer()

		httpCfg.Servers = append(httpCfg.Servers, defaultSSLServer)
	}

	for _, s := range confServers {
		httpCfg.Servers = append(httpCfg.Servers, generateServer(s))
	}

	upstreams, warns := generateUpstreams(confServers, serviceStore)
	warnings.Add(warns)

	httpCfg.Upstreams = upstreams

	return httpCfg, warnings
}

func generateDefaultSSLServer() server {
	return server{IsDefaultSSL: true}
}

func generateDefaultHTTPServer() server {
	return server{IsDefaultHTTP: true}
}

func generateServer(virtualServer state.VirtualServer) server {
	locs := make([]location, 0, len(virtualServer.PathRules)) // FIXME(pleshakov): expand with rule.Routes

	for _, rule := range virtualServer.PathRules {

		matches := make([]httpMatch, 0, len(rule.MatchRules))

		for ruleIdx, r := range rule.MatchRules {

			upstreamName := generateUpstreamName(r)

			m := r.GetMatch()

			// handle case where the only route is a path-only match
			// generate a standard location block without http_matches.
			if len(rule.MatchRules) == 1 && isPathOnlyMatch(m) {
				locs = append(locs, location{
					Path:      rule.Path,
					ProxyPass: generateProxyPass(upstreamName),
				})
			} else {
				path := createPathForMatch(rule.Path, ruleIdx)
				locs = append(locs, generateMatchLocation(path, upstreamName))
				matches = append(matches, createHTTPMatch(m, path))
			}
		}

		if len(matches) > 0 {
			b, err := json.Marshal(matches)

			if err != nil {
				// panic is safe here because we should never fail to marshal the match unless we constructed it incorrectly.
				panic(fmt.Errorf("could not marshal http match: %w", err))
			}

			pathLoc := location{
				Path:         rule.Path,
				HTTPMatchVar: string(b),
			}

			locs = append(locs, pathLoc)
		}
	}
	s := server{
		ServerName: virtualServer.Hostname,
		Locations:  locs,
	}
	if virtualServer.SSL != nil {
		s.SSL = &ssl{
			Certificate:    virtualServer.SSL.CertificatePath,
			CertificateKey: virtualServer.SSL.CertificatePath,
		}
	}

	return s
}

func generateUpstreamName(rule state.MatchRule) string {
	return fmt.Sprintf("%s_%s_rule%d", rule.Source.Namespace, rule.Source.Name, rule.RuleIdx)
}

func generateUpstreams(virtualServers []state.VirtualServer, serviceStore state.ServiceStore) ([]upstream, Warnings) {
	// FIXME(kate-osborn): This logic is required to prevent duplicate upstreams.
	// We should decouple upstreams from virtual servers to avoid having to do this.
	upstreamNameMap := make(map[string]struct{})

	// populate map so we can use it to calculate capacity
	for _, vs := range virtualServers {
		for _, rule := range vs.PathRules {
			for _, r := range rule.MatchRules {
				upstreamNameMap[generateUpstreamName(r)] = struct{}{}
			}
		}
	}

	warnings := newWarnings()
	upstreams := make([]upstream, 0, len(upstreamNameMap))

	for _, vs := range virtualServers {
		for _, rule := range vs.PathRules {
			for _, r := range rule.MatchRules {
				upstreamName := generateUpstreamName(r)

				if _, exists := upstreamNameMap[upstreamName]; exists {
					up, err := generateUpstream(
						r.Source.Spec.Rules[r.RuleIdx].BackendRefs,
						r.Source.Namespace,
						serviceStore,
						upstreamName,
					)
					// delete from map so we don't duplicate upstreams
					delete(upstreamNameMap, upstreamName)

					if err != nil {
						warnings.AddWarning(r.Source, err.Error())
					}

					upstreams = append(upstreams, up)
				}
			}
		}
	}

	return upstreams, warnings
}

func generateUpstream(backendRefs []v1alpha2.HTTPBackendRef, parentNS string, serviceStore state.ServiceStore, upstreamName string) (upstream, error) {
	nginx502Upstream := upstream{
		Name: upstreamName,
		Servers: []upstreamServer{
			{
				Address: nginx502Server,
			},
		},
	}

	if len(backendRefs) == 0 {
		return nginx502Upstream, errors.New("empty backend refs")
	}
	// FIXME(pleshakov): for now, we only support a single backend reference
	ref := backendRefs[0].BackendRef

	if ref.Kind != nil && *ref.Kind != "Service" {
		return nginx502Upstream, errors.New("ref must be of kind service")
	}

	if ref.Port == nil {
		return nginx502Upstream, errors.New("ref must contain port")
	}

	ns := parentNS
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	svcNsname := types.NamespacedName{
		Namespace: ns,
		Name:      string(ref.Name),
	}

	endpoints, err := serviceStore.Resolve(svcNsname, int32(*ref.Port))

	if err != nil {
		return nginx502Upstream, err
	}

	if len(endpoints) == 0 {
		return nginx502Upstream, errors.New("no endpoints found for backend ref")
	}

	upstreamServers := make([]upstreamServer, len(endpoints))
	for idx, ep := range endpoints {
		upstreamServers[idx] = upstreamServer{
			Address: fmt.Sprintf("%s:%d", ep.Address, ep.Port),
		}
	}

	upstream := upstream{
		Name:    upstreamName,
		Servers: upstreamServers,
	}

	return upstream, nil
}

func generateProxyPass(address string) string {
	return "http://" + address
}

func generateMatchLocation(path, address string) location {
	return location{
		Path:      path,
		ProxyPass: generateProxyPass(address),
		Internal:  true,
	}
}

func createPathForMatch(path string, routeIdx int) string {
	return fmt.Sprintf("%s_route%d", path, routeIdx)
}

// httpMatch is an internal representation of an HTTPRouteMatch.
// This struct is marshaled into a string and stored as a variable in the nginx location block for the route's path.
// The NJS httpmatches module will lookup this variable on the request object and compare the request against the Method, Headers, and QueryParams contained in httpMatch.
// If the request satisfies the httpMatch, the request will be internally redirected to the location RedirectPath by NGINX.
type httpMatch struct {
	// Any represents a match with no match conditions.
	Any bool `json:"any,omitempty"`
	// Method is the HTTPMethod of the HTTPRouteMatch.
	Method v1alpha2.HTTPMethod `json:"method,omitempty"`
	// Headers is a list of HTTPHeaders name value pairs with the format "{name}:{value}".
	Headers []string `json:"headers,omitempty"`
	// QueryParams is a list of HTTPQueryParams name value pairs with the format "{name}={value}".
	QueryParams []string `json:"params,omitempty"`
	// RedirectPath is the path to redirect the request to if the request satisfies the match conditions.
	RedirectPath string `json:"redirectPath,omitempty"`
}

func createHTTPMatch(match v1alpha2.HTTPRouteMatch, redirectPath string) httpMatch {
	hm := httpMatch{
		RedirectPath: redirectPath,
	}

	if isPathOnlyMatch(match) {
		hm.Any = true
		return hm
	}

	if match.Method != nil {
		hm.Method = *match.Method
	}

	if match.Headers != nil {
		headers := make([]string, 0, len(match.Headers))
		headerNames := make(map[string]struct{})

		// FIXME(kate-osborn): For now we only support type "Exact".
		for _, h := range match.Headers {
			if *h.Type == v1alpha2.HeaderMatchExact {
				// duplicate header names are not permitted by the spec
				// only configure the first entry for every header name (case-insensitive)
				lowerName := strings.ToLower(string(h.Name))
				if _, ok := headerNames[lowerName]; !ok {
					headers = append(headers, createHeaderKeyValString(h))
					headerNames[lowerName] = struct{}{}
				}
			}
		}
		hm.Headers = headers
	}

	if match.QueryParams != nil {
		params := make([]string, 0, len(match.QueryParams))

		// FIXME(kate-osborn): For now we only support type "Exact".
		for _, p := range match.QueryParams {
			if *p.Type == v1alpha2.QueryParamMatchExact {
				params = append(params, createQueryParamKeyValString(p))
			}
		}
		hm.QueryParams = params
	}

	return hm
}

// The name and values are delimited by "=". A name and value can always be recovered using strings.SplitN(arg,"=", 2).
// Query Parameters are case-sensitive so case is preserved.
func createQueryParamKeyValString(p v1alpha2.HTTPQueryParamMatch) string {
	return p.Name + "=" + p.Value
}

// The name and values are delimited by ":". A name and value can always be recovered using strings.Split(arg, ":").
// Header names are case-insensitive while header values are case-sensitive (e.g. foo:bar == FOO:bar, but foo:bar != foo:BAR).
// We preserve the case of the name here because NGINX allows us to lookup the header names in a case-insensitive manner.
func createHeaderKeyValString(h v1alpha2.HTTPHeaderMatch) string {
	return string(h.Name) + ":" + h.Value
}

func isPathOnlyMatch(match v1alpha2.HTTPRouteMatch) bool {
	return match.Method == nil && match.Headers == nil && match.QueryParams == nil
}
