package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
const nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Generator

// Generator generates NGINX configuration.
type Generator interface {
	// Generate generates NGINX configuration from internal representation.
	Generate(configuration newstate.Configuration) ([]byte, Warnings)
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

func (g *GeneratorImpl) Generate(conf newstate.Configuration) ([]byte, Warnings) {
	warnings := newWarnings()

	servers := httpServers{
		Servers: make([]server, 0, len(conf.HTTPServers)),
	}

	for _, s := range conf.HTTPServers {
		cfg, warns := generate(s, g.serviceStore)

		servers.Servers = append(servers.Servers, cfg)
		warnings.Add(warns)
	}

	return g.executor.ExecuteForHTTPServers(servers), warnings
}

func generate(httpServer newstate.HTTPServer, serviceStore state.ServiceStore) (server, Warnings) {
	warnings := newWarnings()

	locs := make([]location, 0, len(httpServer.PathRules)) // FIXME(pleshakov): expand with rules.Routes

	for _, rules := range httpServer.PathRules {
		// number of routes in a group is always at least 1
		// otherwise, it is a bug in the state.Configuration code, so it is OK to panic here
		r := rules.MatchRules[0] // FIXME(pleshakov): for now, we only handle the first route in case there are multiple routes
		address, err := getBackendAddress(r.Source.Spec.Rules[r.RuleIdx].BackendRefs, r.Source.Namespace, serviceStore)
		if err != nil {
			warnings.AddWarning(r.Source, err.Error())
		}

		match, exists := r.GetMatch()
		if exists && matchLocationNeeded(match) {
			// FIXME(kate-osborn): route index is hardcoded to 0 for now.
			// Once we support multiple routes we will need to change this to the index of the current route.
			path := createPathForMatch(rules.Path, 0)
			// generate location block for this rule and match
			mLoc := generateMatchLocation(path, address)
			// generate the http_matches variable value
			b, err := json.Marshal(createHTTPMatch(match, path))
			if err != nil {
				// panic is safe here because we should never fail to marshal the match unless we constructed it incorrectly.
				panic(fmt.Errorf("could not marshal http match: %w", err))
			}

			loc := location{
				Path:         rules.Path,
				HTTPMatchVar: string(b),
			}
			locs = append(locs, loc, mLoc)
		} else {
			loc := location{
				Path:      rules.Path,
				ProxyPass: generateProxyPass(address),
			}
			locs = append(locs, loc)
		}
	}

	return server{
		ServerName: httpServer.Hostname,
		Locations:  locs,
	}, warnings
}

func generateProxyPass(address string) string {
	if address == "" {
		return "http://" + nginx502Server
	}
	return "http://" + address
}

func getBackendAddress(
	refs []v1alpha2.HTTPBackendRef,
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

	if match.Method != nil {
		hm.Method = *match.Method
	}

	if match.Headers != nil {
		headers := make([]string, 0, len(match.Headers))

		// FIXME(kate-osborn): For now we only support type "Exact".
		for _, h := range match.Headers {
			if *h.Type == v1alpha2.HeaderMatchExact {
				headers = append(headers, createHeaderKeyValString(h))
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

// A match location is needed if the match specifies at least one of the following: Method, Headers, or QueryParams.
func matchLocationNeeded(match v1alpha2.HTTPRouteMatch) bool {
	return match.Method != nil || match.Headers != nil || match.QueryParams != nil
}
