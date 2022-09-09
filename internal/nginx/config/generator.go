package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

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
	warnings := newWarnings()

	confServers := append(conf.HTTPServers, conf.SSLServers...)

	servers := httpServers{
		// capacity is all the conf servers + default ssl & http servers
		Servers: make([]server, 0, len(confServers)+2),
	}

	if len(conf.HTTPServers) > 0 {
		defaultHTTPServer := generateDefaultHTTPServer()

		servers.Servers = append(servers.Servers, defaultHTTPServer)
	}

	if len(conf.SSLServers) > 0 {
		defaultSSLServer := generateDefaultSSLServer()

		servers.Servers = append(servers.Servers, defaultSSLServer)
	}

	for _, s := range confServers {
		cfg, warns := generate(s, g.serviceStore)

		servers.Servers = append(servers.Servers, cfg)
		warnings.Add(warns)
	}

	return g.executor.ExecuteForHTTPServers(servers), warnings
}

func generateDefaultSSLServer() server {
	return server{IsDefaultSSL: true}
}

func generateDefaultHTTPServer() server {
	return server{IsDefaultHTTP: true}
}

func generate(virtualServer state.VirtualServer, serviceStore state.ServiceStore) (server, Warnings) {
	warnings := newWarnings()

	s := server{ServerName: virtualServer.Hostname}

	listenerPort := 80

	if virtualServer.SSL != nil {
		s.SSL = &ssl{
			Certificate:    virtualServer.SSL.CertificatePath,
			CertificateKey: virtualServer.SSL.CertificatePath,
		}

		listenerPort = 443
	}

	if len(virtualServer.PathRules) == 0 {
		// generate default "/" 404 location
		s.Locations = []location{{Path: "/", Return: &returnVal{Code: statusNotFound}}}
		return s, warnings
	}

	locs := make([]location, 0, len(virtualServer.PathRules)) // FIXME(pleshakov): expand with rule.Routes
	for _, rule := range virtualServer.PathRules {
		matches := make([]httpMatch, 0, len(rule.MatchRules))

		for ruleIdx, r := range rule.MatchRules {
			m := r.GetMatch()

			var loc location

			// handle case where the only route is a path-only match
			// generate a standard location block without http_matches.
			if len(rule.MatchRules) == 1 && isPathOnlyMatch(m) {
				loc = location{
					Path: rule.Path,
				}
			} else {
				path := createPathForMatch(rule.Path, ruleIdx)
				loc = generateMatchLocation(path)
				matches = append(matches, createHTTPMatch(m, path))
			}

			// FIXME(pleshakov): There could be a case when the filter has the type set but not the corresponding field.
			// For example, type is v1beta1.HTTPRouteFilterRequestRedirect, but RequestRedirect field is nil.
			// The validation webhook catches that.
			// If it doesn't work as expected, such situation is silently handled below in findFirstFilters.
			// Consider reporting an error. But that should be done in a separate validation layer.

			// RequestRedirect and proxying are mutually exclusive.
			if r.Filters.RequestRedirect != nil {
				loc.Return = generateReturnValForRedirectFilter(r.Filters.RequestRedirect, listenerPort)
			} else {
				address, err := getBackendAddress(r.Source.Spec.Rules[r.RuleIdx].BackendRefs, r.Source.Namespace, serviceStore)
				if err != nil {
					warnings.AddWarning(r.Source, err.Error())
				}

				loc.ProxyPass = generateProxyPass(address)
			}

			locs = append(locs, loc)
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

	s.Locations = locs

	return s, warnings
}

func generateProxyPass(address string) string {
	if address == "" {
		return "http://" + nginx502Server
	}
	return "http://" + address
}

func generateReturnValForRedirectFilter(filter *v1beta1.HTTPRequestRedirectFilter, listenerPort int) *returnVal {
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
	code := statusFound
	if filter.StatusCode != nil {
		code = statusCode(*filter.StatusCode)
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

	return &returnVal{
		Code: code,
		URL:  fmt.Sprintf("%s://%s:%d$request_uri", scheme, hostname, port),
	}
}

func getBackendAddress(
	refs []v1beta1.HTTPBackendRef,
	parentNS string,
	serviceStore state.ServiceStore,
) (string, error) {
	if len(refs) == 0 {
		return "", errors.New("empty backend refs")
	}

	// FIXME(pleshakov): for now, we only support a single backend reference
	ref := refs[0].BackendRef

	if ref.Kind != nil && *ref.Kind != "Service" {
		return "", fmt.Errorf("unsupported kind %s", *ref.Kind)
	}

	ns := parentNS
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	address, err := serviceStore.Resolve(types.NamespacedName{Namespace: ns, Name: string(ref.Name)})
	if err != nil {
		return "", fmt.Errorf("service %s/%s cannot be resolved: %w", ns, ref.Name, err)
	}

	if ref.Port == nil {
		return "", errors.New("port is nil")
	}

	return fmt.Sprintf("%s:%d", address, *ref.Port), nil
}

func generateMatchLocation(path string) location {
	return location{
		Path:     path,
		Internal: true,
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
	Method v1beta1.HTTPMethod `json:"method,omitempty"`
	// Headers is a list of HTTPHeaders name value pairs with the format "{name}:{value}".
	Headers []string `json:"headers,omitempty"`
	// QueryParams is a list of HTTPQueryParams name value pairs with the format "{name}={value}".
	QueryParams []string `json:"params,omitempty"`
	// RedirectPath is the path to redirect the request to if the request satisfies the match conditions.
	RedirectPath string `json:"redirectPath,omitempty"`
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
// Header names are case-insensitive while header values are case-sensitive (e.g. foo:bar == FOO:bar, but foo:bar != foo:BAR).
// We preserve the case of the name here because NGINX allows us to lookup the header names in a case-insensitive manner.
func createHeaderKeyValString(h v1beta1.HTTPHeaderMatch) string {
	return string(h.Name) + ":" + h.Value
}

func isPathOnlyMatch(match v1beta1.HTTPRouteMatch) bool {
	return match.Method == nil && match.Headers == nil && match.QueryParams == nil
}
