package config

import (
	"encoding/json"
	"fmt"
	"strings"
	gotemplate "text/template"

	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

var serversTemplate = gotemplate.Must(gotemplate.New("servers").Parse(serversTemplateText))

const (
	// HeaderMatchSeparator is the separator for constructing header-based match for NJS.
	HeaderMatchSeparator = ":"
	rootPath             = "/"
)

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
		return http.Server{
			IsDefaultSSL: true,
			Port:         virtualServer.Port,
		}
	}

	return http.Server{
		ServerName: virtualServer.Hostname,
		SSL: &http.SSL{
			Certificate:    virtualServer.SSL.CertificatePath,
			CertificateKey: virtualServer.SSL.CertificatePath,
		},
		Locations: createLocations(virtualServer.PathRules, virtualServer.Port),
		Port:      virtualServer.Port,
	}
}

func createServer(virtualServer dataplane.VirtualServer) http.Server {
	if virtualServer.IsDefault {
		return http.Server{
			IsDefaultHTTP: true,
			Port:          virtualServer.Port,
		}
	}

	return http.Server{
		ServerName: virtualServer.Hostname,
		Locations:  createLocations(virtualServer.PathRules, virtualServer.Port),
		Port:       virtualServer.Port,
	}
}

func createLocations(pathRules []dataplane.PathRule, listenerPort int32) []http.Location {
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
					Path:  rule.Path,
					Exact: rule.PathType == dataplane.PathTypeExact,
				}
			} else {
				path := createPathForMatch(rule.Path, rule.PathType, matchRuleIdx)
				loc = createMatchLocation(path)
				matches = append(matches, createHTTPMatch(m, path))
			}

			if r.Filters.InvalidFilter != nil {
				loc.Return = &http.Return{Code: http.StatusInternalServerError}
				locs = append(locs, loc)
				continue
			}

			// There could be a case when the filter has the type set but not the corresponding field.
			// For example, type is v1beta1.HTTPRouteFilterRequestRedirect, but RequestRedirect field is nil.
			// The imported Webhook validation webhook catches that.

			// FIXME(pleshakov): Ensure dataplane.Configuration -related types don't include v1beta1 types, so that
			// we don't need to make any assumptions like above here. After fixing this, ensure that there is a test
			// for checking the imported Webhook validation catches the case above.
			// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/660

			// RequestRedirect and proxying are mutually exclusive.
			if r.Filters.RequestRedirect != nil {
				loc.Return = createReturnValForRedirectFilter(r.Filters.RequestRedirect, listenerPort)
				locs = append(locs, loc)
				continue
			}

			backendName := backendGroupName(r.BackendGroup)
			loc.ProxySetHeaders = generateProxySetHeaders(r.Filters.RequestHeaderModifiers)

			if backendGroupNeedsSplit(r.BackendGroup) {
				loc.ProxyPass = createProxyPassForVar(backendName)
			} else {
				loc.ProxyPass = createProxyPass(backendName)
			}

			locs = append(locs, loc)
		}

		if len(matches) > 0 {
			// FIXME(sberman): De-dupe matches and associated locations
			// so we don't need nginx/njs to perform unnecessary matching.
			// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/662
			b, err := json.Marshal(matches)
			if err != nil {
				// panic is safe here because we should never fail to marshal the match unless we constructed it incorrectly.
				panic(fmt.Errorf("could not marshal http match: %w", err))
			}

			pathLoc := http.Location{
				Path:         rule.Path,
				Exact:        rule.PathType == dataplane.PathTypeExact,
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

func createReturnValForRedirectFilter(filter *v1beta1.HTTPRequestRedirectFilter, listenerPort int32) *http.Return {
	if filter == nil {
		return nil
	}

	hostname := "$host"
	if filter.Hostname != nil {
		hostname = string(*filter.Hostname)
	}

	code := http.StatusFound
	if filter.StatusCode != nil {
		code = http.StatusCode(*filter.StatusCode)
	}

	port := listenerPort
	if filter.Port != nil {
		port = int32(*filter.Port)
	}

	hostnamePort := fmt.Sprintf("%s:%d", hostname, port)

	scheme := "$scheme"
	if filter.Scheme != nil {
		scheme = *filter.Scheme
		// Don't specify the port in the return url if the scheme is
		// well known and the port is already set to the correct well known port
		if (port == 80 && scheme == "http") || (port == 443 && scheme == "https") {
			hostnamePort = hostname
		}
		if filter.Port == nil {
			// Don't specify the port in the return url if the scheme is
			// well known and the port is not specified by the user
			if scheme == "http" || scheme == "https" {
				hostnamePort = hostname
			}
		}
	}

	return &http.Return{
		Code: code,
		Body: fmt.Sprintf("%s://%s$request_uri", scheme, hostnamePort),
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
	return string(p.Name) + "=" + p.Value
}

// The name and values are delimited by ":". A name and value can always be recovered using strings.Split(arg, ":").
// Header names are case-insensitive and header values are case-sensitive.
// Ex. foo:bar == FOO:bar, but foo:bar != foo:BAR,
// We preserve the case of the name here because NGINX allows us to look up the header names in a case-insensitive
// manner.
func createHeaderKeyValString(h v1beta1.HTTPHeaderMatch) string {
	return string(h.Name) + HeaderMatchSeparator + h.Value
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

func generateProxySetHeaders(filters *dataplane.HTTPHeaderFilter) []http.Header {
	if filters == nil {
		return nil
	}
	proxySetHeaders := make([]http.Header, 0, len(filters.Add)+len(filters.Set)+len(filters.Remove))
	if len(filters.Add) > 0 {
		addHeaders := convertAddHeaders(filters.Add)
		proxySetHeaders = append(proxySetHeaders, addHeaders...)
	}
	if len(filters.Set) > 0 {
		setHeaders := convertSetHeaders(filters.Set)
		proxySetHeaders = append(proxySetHeaders, setHeaders...)
	}
	// If the value of a header field is an empty string then this field will not be passed to a proxied server
	for _, h := range filters.Remove {
		proxySetHeaders = append(proxySetHeaders, http.Header{
			Name:  h,
			Value: "",
		})
	}
	return proxySetHeaders
}

func convertAddHeaders(headers []dataplane.HTTPHeader) []http.Header {
	locHeaders := make([]http.Header, 0, len(headers))
	for _, h := range headers {
		mapVarName := "${" + generateAddHeaderMapVariableName(h.Name) + "}"
		locHeaders = append(locHeaders, http.Header{
			Name:  h.Name,
			Value: mapVarName + h.Value,
		})
	}
	return locHeaders
}

func convertSetHeaders(headers []dataplane.HTTPHeader) []http.Header {
	locHeaders := make([]http.Header, 0, len(headers))
	for _, h := range headers {
		locHeaders = append(locHeaders, http.Header{
			Name:  h.Name,
			Value: h.Value,
		})
	}
	return locHeaders
}

func createPathForMatch(path string, pathType dataplane.PathType, routeIdx int) string {
	return fmt.Sprintf("%s_%s_route%d", path, pathType, routeIdx)
}

func createDefaultRootLocation() http.Location {
	return http.Location{
		Path:   "/",
		Return: &http.Return{Code: http.StatusNotFound},
	}
}
