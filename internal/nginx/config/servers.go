package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

var serversTemplate = template.Must(template.New("servers").Parse(serversTemplateText))

const rootPath = "/"

func executeServers(conf dataplane.Configuration) []byte {
	servers := createServers(conf.HTTPServers, conf.SSLServers)

	return execute(serversTemplate, servers)
}

func createServers(httpServers, sslServers []dataplane.VirtualServer) []http.Server {
	servers := make([]http.Server, 0, len(httpServers)+len(sslServers))

	for _, s := range httpServers {
		servers = append(servers, createServer(s))
	}

	for _, s := range sslServers {
		servers = append(servers, createSSLServer(s))
	}

	return servers
}

func createSSLServer(virtualServer dataplane.VirtualServer) http.Server {
	if virtualServer.IsDefault {
		return createDefaultSSLServer()
	}

	return http.Server{
		ServerName: virtualServer.Hostname,
		SSL: &http.SSL{
			Certificate:    virtualServer.SSL.CertificatePath,
			CertificateKey: virtualServer.SSL.CertificatePath,
		},
		Locations: createLocations(virtualServer.PathRules, 443),
	}
}

func createServer(virtualServer dataplane.VirtualServer) http.Server {
	if virtualServer.IsDefault {
		return createDefaultHTTPServer()
	}

	return http.Server{
		ServerName: virtualServer.Hostname,
		Locations:  createLocations(virtualServer.PathRules, 80),
	}
}

func createLocations(pathRules []dataplane.PathRule, listenerPort int) []http.Location {
	lenPathRules := len(pathRules)

	if lenPathRules == 0 {
		return []http.Location{createDefaultRootLocation()}
	}

	// To calculate the maximum number of locations, we need to take into account the following:
	// 1. Each match rule for a path rule will have one location.
	// 2. Each path rule may have an additional location if it contains non-path-only matches.
	// 3. There may be an additional location for the default root path.
	maxLocs := 1
	for _, rules := range pathRules {
		maxLocs += len(rules.MatchRules) + 1
	}

	locs := make([]http.Location, 0, maxLocs)

	rootPathExists := false

	for _, rule := range pathRules {
		matches := make([]httpMatch, 0, len(rule.MatchRules))

		if rule.Path == rootPath {
			rootPathExists = true
		}

		for matchRuleIdx, r := range rule.MatchRules {
			m := r.GetMatch()

			var loc http.Location

			// handle case where the only route is a path-only match
			// generate a standard location block without http_matches.
			if len(rule.MatchRules) == 1 && isPathOnlyMatch(m) {
				loc = http.Location{
					Path: rule.Path,
				}
			} else {
				path := createPathForMatch(rule.Path, matchRuleIdx)
				loc = createMatchLocation(path)
				matches = append(matches, createHTTPMatch(m, path))
			}

			// FIXME(pleshakov): There could be a case when the filter has the type set but not the corresponding field.
			// For example, type is v1beta1.HTTPRouteFilterRequestRedirect, but RequestRedirect field is nil.
			// The validation webhook catches that.
			// If it doesn't work as expected, such situation is silently handled below in findFirstFilters.
			// Consider reporting an error. But that should be done in a separate validation layer.

			// RequestRedirect and proxying are mutually exclusive.
			if r.Filters.RequestRedirect != nil {
				loc.Return = createReturnValForRedirectFilter(r.Filters.RequestRedirect, listenerPort)

				locs = append(locs, loc)
				continue
			}

			backendName := backendGroupName(r.BackendGroup)

			if backendGroupNeedsSplit(r.BackendGroup) {
				loc.ProxyPass = createProxyPassForVar(backendName)
			} else {
				loc.ProxyPass = createProxyPass(backendName)
			}

			locs = append(locs, loc)
		}

		if len(matches) > 0 {
			b, err := json.Marshal(matches)
			if err != nil {
				// panic is safe here because we should never fail to marshal the match unless we constructed it incorrectly.
				panic(fmt.Errorf("could not marshal http match: %w", err))
			}

			pathLoc := http.Location{
				Path:         rule.Path,
				HTTPMatchVar: string(b),
			}

			locs = append(locs, pathLoc)
		}
	}

	if !rootPathExists {
		locs = append(locs, createDefaultRootLocation())
	}

	return locs
}

func createDefaultSSLServer() http.Server {
	return http.Server{IsDefaultSSL: true}
}

func createDefaultHTTPServer() http.Server {
	return http.Server{IsDefaultHTTP: true}
}

func createReturnValForRedirectFilter(filter *v1beta1.HTTPRequestRedirectFilter, listenerPort int) *http.Return {
	if filter == nil {
		return nil
	}

	hostname := "$host"
	if filter.Hostname != nil {
		hostname = string(*filter.Hostname)
	}

	// FIXME(pleshakov): Unknown values here must result in the implementation setting the Attached Condition for
	// the Route to  `status: False`, with a Reason of `UnsupportedValue`. In that case, all routes of the Route will be
	// ignored. NGINX will return 500. This should be implemented in the validation layer.
	code := http.StatusFound
	if filter.StatusCode != nil {
		code = http.StatusCode(*filter.StatusCode)
	}

	port := listenerPort
	if filter.Port != nil {
		port = int(*filter.Port)
	}

	// FIXME(pleshakov): Same as the FIXME about StatusCode above.
	scheme := "$scheme"
	if filter.Scheme != nil {
		scheme = *filter.Scheme
	}

	return &http.Return{
		Code: code,
		URL:  fmt.Sprintf("%s://%s:%d$request_uri", scheme, hostname, port),
	}
}

// httpMatch is an internal representation of an HTTPRouteMatch.
// This struct is marshaled into a string and stored as a variable in the nginx location block for the route's path.
// The NJS httpmatches module will look up this variable on the request object and compare the request against the
// Method, Headers, and QueryParams contained in httpMatch.
// If the request satisfies the httpMatch, NGINX will redirect the request to the location RedirectPath.
type httpMatch struct {
	// Method is the HTTPMethod of the HTTPRouteMatch.
	Method v1beta1.HTTPMethod `json:"method,omitempty"`
	// RedirectPath is the path to redirect the request to if the request satisfies the match conditions.
	RedirectPath string `json:"redirectPath,omitempty"`
	// Headers is a list of HTTPHeaders name value pairs with the format "{name}:{value}".
	Headers []string `json:"headers,omitempty"`
	// QueryParams is a list of HTTPQueryParams name value pairs with the format "{name}={value}".
	QueryParams []string `json:"params,omitempty"`
	// Any represents a match with no match conditions.
	Any bool `json:"any,omitempty"`
}

func createHTTPMatch(match v1beta1.HTTPRouteMatch, redirectPath string) httpMatch {
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
			if *h.Type == v1beta1.HeaderMatchExact {
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
			if *p.Type == v1beta1.QueryParamMatchExact {
				params = append(params, createQueryParamKeyValString(p))
			}
		}
		hm.QueryParams = params
	}

	return hm
}

// The name and values are delimited by "=". A name and value can always be recovered using strings.SplitN(arg,"=", 2).
// Query Parameters are case-sensitive so case is preserved.
func createQueryParamKeyValString(p v1beta1.HTTPQueryParamMatch) string {
	return p.Name + "=" + p.Value
}

// The name and values are delimited by ":". A name and value can always be recovered using strings.Split(arg, ":").
// Header names are case-insensitive and header values are case-sensitive.
// Ex. foo:bar == FOO:bar, but foo:bar != foo:BAR,
// We preserve the case of the name here because NGINX allows us to look up the header names in a case-insensitive
// manner.
func createHeaderKeyValString(h v1beta1.HTTPHeaderMatch) string {
	return string(h.Name) + ":" + h.Value
}

func isPathOnlyMatch(match v1beta1.HTTPRouteMatch) bool {
	return match.Method == nil && match.Headers == nil && match.QueryParams == nil
}

func createProxyPass(address string) string {
	return "http://" + address
}

func createProxyPassForVar(variable string) string {
	return "http://$" + convertStringToSafeVariableName(variable)
}

func createMatchLocation(path string) http.Location {
	return http.Location{
		Path:     path,
		Internal: true,
	}
}

func createPathForMatch(path string, routeIdx int) string {
	return fmt.Sprintf("%s_route%d", path, routeIdx)
}

func createDefaultRootLocation() http.Location {
	return http.Location{
		Path:   "/",
		Return: &http.Return{Code: http.StatusNotFound},
	}
}
