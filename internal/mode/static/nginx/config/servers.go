package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var serversTemplate = gotemplate.Must(gotemplate.New("servers").Parse(serversTemplateText))

const (
	// HeaderMatchSeparator is the separator for constructing header-based match for NJS.
	HeaderMatchSeparator = ":"
	rootPath             = "/"
)

// httpBaseHeaders contains the constant headers set in each HTTP server block.
var httpBaseHeaders = []http.Header{
	{
		Name:  "Host",
		Value: "$gw_api_compliant_host",
	},
	{
		Name:  "X-Forwarded-For",
		Value: "$proxy_add_x_forwarded_for",
	},
	{
		Name:  "Upgrade",
		Value: "$http_upgrade",
	},
	{
		Name:  "Connection",
		Value: "$connection_upgrade",
	},
}

// grpcBaseHeaders contains the constant headers set in each gRPC server block.
var grpcBaseHeaders = []http.Header{
	{
		Name:  "Host",
		Value: "$gw_api_compliant_host",
	},
	{
		Name:  "X-Forwarded-For",
		Value: "$proxy_add_x_forwarded_for",
	},
	{
		Name:  "Authority",
		Value: "$gw_api_compliant_host",
	},
}

func executeServers(conf dataplane.Configuration) []executeResult {
	servers, httpMatchPairs := createServers(conf.HTTPServers, conf.SSLServers)

	serverResult := executeResult{
		dest: httpConfigFile,
		data: helpers.MustExecuteTemplate(serversTemplate, servers),
	}

	// create httpMatchPair conf
	httpMatchConf, err := json.Marshal(httpMatchPairs)
	if err != nil {
		// panic is safe here because we should never fail to marshal the match unless we constructed it incorrectly.
		panic(fmt.Errorf("could not marshal http match pairs: %w", err))
	}

	httpMatchResult := executeResult{
		dest: httpMatchVarsFile,
		data: httpMatchConf,
	}

	additionFileResults := createAdditionFileResults(conf)

	allResults := make([]executeResult, 0, len(additionFileResults)+2)
	allResults = append(allResults, additionFileResults...)
	allResults = append(allResults, serverResult, httpMatchResult)

	return allResults
}

func createAdditionFileResults(conf dataplane.Configuration) []executeResult {
	uniqueAdditions := make(map[string][]byte)

	findUniqueAdditionsForServer := func(server dataplane.VirtualServer) {
		for _, add := range server.Additions {
			uniqueAdditions[createAdditionFileName(add)] = add.Bytes
		}

		for _, pr := range server.PathRules {
			for _, mr := range pr.MatchRules {
				for _, add := range mr.Additions {
					uniqueAdditions[createAdditionFileName(add)] = add.Bytes
				}
			}
		}
	}

	for _, s := range conf.HTTPServers {
		findUniqueAdditionsForServer(s)
	}

	for _, s := range conf.SSLServers {
		findUniqueAdditionsForServer(s)
	}

	results := make([]executeResult, 0, len(uniqueAdditions))

	for filename, contents := range uniqueAdditions {
		results = append(results, executeResult{
			dest: filename,
			data: contents,
		})
	}

	return results
}

func createAdditionFileName(addition dataplane.Addition) string {
	return fmt.Sprintf("%s/%s.conf", includesFolder, addition.Identifier)
}

func createIncludes(additions []dataplane.Addition) []string {
	if len(additions) == 0 {
		return nil
	}

	includes := make([]string, 0, len(additions))

	for _, addition := range additions {
		includes = append(includes, createAdditionFileName(addition))
	}

	return includes
}

func createServers(httpServers, sslServers []dataplane.VirtualServer) ([]http.Server, httpMatchPairs) {
	servers := make([]http.Server, 0, len(httpServers)+len(sslServers))
	finalMatchPairs := make(httpMatchPairs)

	for serverID, s := range httpServers {
		httpServer, matchPairs := createServer(s, serverID)
		servers = append(servers, httpServer)
		maps.Copy(finalMatchPairs, matchPairs)
	}

	for serverID, s := range sslServers {
		sslServer, matchPair := createSSLServer(s, serverID)
		servers = append(servers, sslServer)
		maps.Copy(finalMatchPairs, matchPair)
	}

	return servers, finalMatchPairs
}

func createSSLServer(virtualServer dataplane.VirtualServer, serverID int) (http.Server, httpMatchPairs) {
	if virtualServer.IsDefault {
		return http.Server{
			IsDefaultSSL: true,
			Port:         virtualServer.Port,
		}, nil
	}

	locs, matchPairs, grpc := createLocations(&virtualServer, serverID)

	return http.Server{
		ServerName: virtualServer.Hostname,
		SSL: &http.SSL{
			Certificate:    generatePEMFileName(virtualServer.SSL.KeyPairID),
			CertificateKey: generatePEMFileName(virtualServer.SSL.KeyPairID),
		},
		Locations: locs,
		Port:      virtualServer.Port,
		GRPC:      grpc,
		Includes:  createIncludes(virtualServer.Additions),
	}, matchPairs
}

func createServer(virtualServer dataplane.VirtualServer, serverID int) (http.Server, httpMatchPairs) {
	if virtualServer.IsDefault {
		return http.Server{
			IsDefaultHTTP: true,
			Port:          virtualServer.Port,
		}, nil
	}

	locs, matchPairs, grpc := createLocations(&virtualServer, serverID)

	return http.Server{
		ServerName: virtualServer.Hostname,
		Locations:  locs,
		Port:       virtualServer.Port,
		GRPC:       grpc,
		Includes:   createIncludes(virtualServer.Additions),
	}, matchPairs
}

// rewriteConfig contains the configuration for a location to rewrite paths,
// as specified in a URLRewrite filter.
type rewriteConfig struct {
	// Rewrite rewrites the original URI to the new URI (ex: /coffee -> /beans)
	Rewrite string
}

type httpMatchPairs map[string][]routeMatch

func createLocations(server *dataplane.VirtualServer, serverID int) ([]http.Location, httpMatchPairs, bool) {
	maxLocs, pathsAndTypes := getMaxLocationCountAndPathMap(server.PathRules)
	locs := make([]http.Location, 0, maxLocs)
	matchPairs := make(httpMatchPairs)
	var rootPathExists bool
	var grpc bool

	for pathRuleIdx, rule := range server.PathRules {
		matches := make([]routeMatch, 0, len(rule.MatchRules))

		if rule.Path == rootPath {
			rootPathExists = true
		}

		if rule.GRPC {
			grpc = true
		}

		extLocations := initializeExternalLocations(rule, pathsAndTypes)

		for matchRuleIdx, r := range rule.MatchRules {
			buildLocations := extLocations

			if len(rule.MatchRules) != 1 || !isPathOnlyMatch(r.Match) {
				intLocation, match := initializeInternalLocation(pathRuleIdx, matchRuleIdx, r.Match)
				buildLocations = []http.Location{intLocation}
				matches = append(matches, match)
			}

			includes := createIncludes(r.Additions)

			// buildLocations will either contain the extLocations OR the intLocation.
			// If it contains the intLocation, the extLocations will be added to the final locations after we loop
			// through all the MatchRules.
			// It is safe to modify the buildLocations here by adding includes and filters.
			buildLocations = updateLocationsForIncludes(buildLocations, includes)
			buildLocations = updateLocationsForFilters(r.Filters, buildLocations, r, server.Port, rule.Path, rule.GRPC)
			locs = append(locs, buildLocations...)
		}

		if len(matches) > 0 {
			for i := range extLocations {
				// FIXME(sberman): De-dupe matches and associated locations
				// so we don't need nginx/njs to perform unnecessary matching.
				// https://github.com/nginxinc/nginx-gateway-fabric/issues/662
				var key string
				if server.SSL != nil {
					key = "SSL"
				}
				key += strconv.Itoa(serverID) + "_" + strconv.Itoa(pathRuleIdx)
				extLocations[i].HTTPMatchKey = key
				matchPairs[extLocations[i].HTTPMatchKey] = matches
			}
			locs = append(locs, extLocations...)
		}
	}

	if !rootPathExists {
		locs = append(locs, createDefaultRootLocation())
	}

	return locs, matchPairs, grpc
}

func updateLocationsForIncludes(locations []http.Location, includes []string) []http.Location {
	for i := range locations {
		locations[i].Includes = includes
	}

	return locations
}

// pathAndTypeMap contains a map of paths and any path types defined for that path
// for example, {/foo: {exact: {}, prefix: {}}}.
type pathAndTypeMap map[string]map[dataplane.PathType]struct{}

// To calculate the maximum number of locations, we need to take into account the following:
// 1. Each match rule for a path rule will have one location.
// 2. Each path rule may have an additional location if it contains non-path-only matches.
// 3. Each prefix path rule may have an additional location if it doesn't contain trailing slash.
// 4. There may be an additional location for the default root path.
// We also return a map of all paths and their types.
func getMaxLocationCountAndPathMap(pathRules []dataplane.PathRule) (int, pathAndTypeMap) {
	maxLocs := 1
	pathsAndTypes := make(pathAndTypeMap)
	for _, rule := range pathRules {
		maxLocs += len(rule.MatchRules) + 2
		if pathsAndTypes[rule.Path] == nil {
			pathsAndTypes[rule.Path] = map[dataplane.PathType]struct{}{
				rule.PathType: {},
			}
		} else {
			pathsAndTypes[rule.Path][rule.PathType] = struct{}{}
		}
	}

	return maxLocs, pathsAndTypes
}

func initializeExternalLocations(
	rule dataplane.PathRule,
	pathsAndTypes pathAndTypeMap,
) []http.Location {
	extLocations := make([]http.Location, 0, 2)
	externalLocPath := createPath(rule)
	// If the path type is Prefix and doesn't contain a trailing slash, then we need a second location
	// that handles the Exact prefix case (if it doesn't already exist), and the first location is updated
	// to handle the trailing slash prefix case (if it doesn't already exist)
	if isNonSlashedPrefixPath(rule.PathType, externalLocPath) {
		// if Exact path and/or trailing slash Prefix path already exists, this means some routing rule
		// configures it. The routing rule location has priority over this location, so we don't try to
		// overwrite it and we don't add a duplicate location to NGINX because that will cause an NGINX config error.
		_, exactPathExists := pathsAndTypes[rule.Path][dataplane.PathTypeExact]
		var trailingSlashPrefixPathExists bool
		if pathTypes, exists := pathsAndTypes[rule.Path+"/"]; exists {
			_, trailingSlashPrefixPathExists = pathTypes[dataplane.PathTypePrefix]
		}

		if exactPathExists && trailingSlashPrefixPathExists {
			return []http.Location{}
		}

		if !trailingSlashPrefixPathExists {
			externalLocTrailing := http.Location{
				Path: externalLocPath + "/",
			}
			extLocations = append(extLocations, externalLocTrailing)
		}
		if !exactPathExists {
			externalLocExact := http.Location{
				Path: exactPath(externalLocPath),
			}
			extLocations = append(extLocations, externalLocExact)
		}
	} else {
		externalLoc := http.Location{
			Path: externalLocPath,
		}
		extLocations = []http.Location{externalLoc}
	}

	return extLocations
}

func initializeInternalLocation(
	pathruleIdx,
	matchRuleIdx int,
	match dataplane.Match,
) (http.Location, routeMatch) {
	path := fmt.Sprintf("@rule%d-route%d", pathruleIdx, matchRuleIdx)
	return createMatchLocation(path), createRouteMatch(match, path)
}

// updateLocationsForFilters updates the existing locations with any relevant filters.
func updateLocationsForFilters(
	filters dataplane.HTTPFilters,
	buildLocations []http.Location,
	matchRule dataplane.MatchRule,
	listenerPort int32,
	path string,
	grpc bool,
) []http.Location {
	if filters.InvalidFilter != nil {
		for i := range buildLocations {
			buildLocations[i].Return = &http.Return{Code: http.StatusInternalServerError}
		}
		return buildLocations
	}

	if filters.RequestRedirect != nil {
		ret := createReturnValForRedirectFilter(filters.RequestRedirect, listenerPort)
		for i := range buildLocations {
			buildLocations[i].Return = ret
		}
		return buildLocations
	}

	rewrites := createRewritesValForRewriteFilter(filters.RequestURLRewrite, path)
	proxySetHeaders := generateProxySetHeaders(&matchRule.Filters, grpc)
	responseHeaders := generateResponseHeaders(&matchRule.Filters)
	for i := range buildLocations {
		if rewrites != nil {
			if rewrites.Rewrite != "" {
				buildLocations[i].Rewrites = append(buildLocations[i].Rewrites, rewrites.Rewrite)
			}
		}
		buildLocations[i].ProxySetHeaders = proxySetHeaders
		buildLocations[i].ProxySSLVerify = createProxyTLSFromBackends(matchRule.BackendGroup.Backends)
		proxyPass := createProxyPass(
			matchRule.BackendGroup,
			matchRule.Filters.RequestURLRewrite,
			generateProtocolString(buildLocations[i].ProxySSLVerify, grpc),
			grpc,
		)
		buildLocations[i].ResponseHeaders = responseHeaders
		buildLocations[i].ProxyPass = proxyPass
		buildLocations[i].GRPC = grpc
	}

	return buildLocations
}

func generateProtocolString(ssl *http.ProxySSLVerify, grpc bool) string {
	if !grpc {
		if ssl != nil {
			return "https"
		}
		return "http"
	}
	if ssl != nil {
		return "grpcs"
	}
	return "grpc"
}

func createProxyTLSFromBackends(backends []dataplane.Backend) *http.ProxySSLVerify {
	if len(backends) == 0 {
		return nil
	}
	for _, b := range backends {
		proxyVerify := createProxySSLVerify(b.VerifyTLS)
		if proxyVerify != nil {
			// If any backend has a backend TLS policy defined, then we use that for the proxy SSL verification.
			// We require that all backends in a group have the same backend TLS policy.
			// Verification that all backends in a group have the same backend TLS policy is done in the graph package.
			return proxyVerify
		}
	}
	return nil
}

func createProxySSLVerify(v *dataplane.VerifyTLS) *http.ProxySSLVerify {
	if v == nil {
		return nil
	}
	var trustedCert string
	if v.CertBundleID != "" {
		trustedCert = generateCertBundleFileName(v.CertBundleID)
	} else {
		trustedCert = v.RootCAPath
	}
	return &http.ProxySSLVerify{
		TrustedCertificate: trustedCert,
		Name:               v.Hostname,
	}
}

func createReturnValForRedirectFilter(filter *dataplane.HTTPRequestRedirectFilter, listenerPort int32) *http.Return {
	if filter == nil {
		return nil
	}

	hostname := "$host"
	if filter.Hostname != nil {
		hostname = *filter.Hostname
	}

	code := http.StatusFound
	if filter.StatusCode != nil {
		code = http.StatusCode(*filter.StatusCode)
	}

	port := listenerPort
	if filter.Port != nil {
		port = *filter.Port
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

func createRewritesValForRewriteFilter(filter *dataplane.HTTPURLRewriteFilter, path string) *rewriteConfig {
	if filter == nil {
		return nil
	}

	rewrites := &rewriteConfig{}

	if filter.Path != nil {
		switch filter.Path.Type {
		case dataplane.ReplaceFullPath:
			rewrites.Rewrite = fmt.Sprintf("^ %s break", filter.Path.Replacement)
		case dataplane.ReplacePrefixMatch:
			filterPrefix := filter.Path.Replacement
			if filterPrefix == "" {
				filterPrefix = "/"
			}

			// capture everything after the configured prefix
			regex := fmt.Sprintf("^%s(.*)$", path)
			// replace the configured prefix with the filter prefix and append what was captured
			replacement := fmt.Sprintf("%s$1", filterPrefix)

			// if configured prefix does not end in /, but replacement prefix does end in /,
			// then make sure that we *require* but *don't capture* a trailing slash in the request,
			// otherwise we'll get duplicate slashes in the full replacement
			if strings.HasSuffix(filterPrefix, "/") && !strings.HasSuffix(path, "/") {
				regex = fmt.Sprintf("^%s(?:/(.*))?$", path)
			}

			// if configured prefix ends in / we won't capture it for a request (since it's not in the regex),
			// so append it to the replacement prefix if the replacement prefix doesn't already end in /
			if strings.HasSuffix(path, "/") && !strings.HasSuffix(filterPrefix, "/") {
				replacement = fmt.Sprintf("%s/$1", filterPrefix)
			}

			rewrites.Rewrite = fmt.Sprintf("%s %s break", regex, replacement)
		}
	}

	return rewrites
}

// routeMatch is an internal representation of an HTTPRouteMatch.
// This struct is stored as a key-value pair in /etc/nginx/conf.d/matches.json with a key for the route's path.
// The NJS httpmatches module will look up key specified in the nginx location on the request object
// and compare the request against the Method, Headers, and QueryParams contained in routeMatch.
// If the request satisfies the routeMatch, NGINX will redirect the request to the location RedirectPath.
type routeMatch struct {
	// Method is the HTTPMethod of the HTTPRouteMatch.
	Method string `json:"method,omitempty"`
	// RedirectPath is the path to redirect the request to if the request satisfies the match conditions.
	RedirectPath string `json:"redirectPath,omitempty"`
	// Headers is a list of HTTPHeaders name value pairs with the format "{name}:{value}".
	Headers []string `json:"headers,omitempty"`
	// QueryParams is a list of HTTPQueryParams name value pairs with the format "{name}={value}".
	QueryParams []string `json:"params,omitempty"`
	// Any represents a match with no match conditions.
	Any bool `json:"any,omitempty"`
}

func createRouteMatch(match dataplane.Match, redirectPath string) routeMatch {
	hm := routeMatch{
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
			// duplicate header names are not permitted by the spec
			// only configure the first entry for every header name (case-insensitive)
			lowerName := strings.ToLower(h.Name)
			if _, ok := headerNames[lowerName]; !ok {
				headers = append(headers, createHeaderKeyValString(h))
				headerNames[lowerName] = struct{}{}
			}
		}
		hm.Headers = headers
	}

	if match.QueryParams != nil {
		params := make([]string, 0, len(match.QueryParams))

		for _, p := range match.QueryParams {
			params = append(params, createQueryParamKeyValString(p))
		}
		hm.QueryParams = params
	}

	return hm
}

// The name and values are delimited by "=". A name and value can always be recovered using strings.SplitN(arg,"=", 2).
// Query Parameters are case-sensitive so case is preserved.
func createQueryParamKeyValString(p dataplane.HTTPQueryParamMatch) string {
	return p.Name + "=" + p.Value
}

// The name and values are delimited by ":". A name and value can always be recovered using strings.Split(arg, ":").
// Header names are case-insensitive and header values are case-sensitive.
// Ex. foo:bar == FOO:bar, but foo:bar != foo:BAR,
// We preserve the case of the name here because NGINX allows us to look up the header names in a case-insensitive
// manner.
func createHeaderKeyValString(h dataplane.HTTPHeaderMatch) string {
	return h.Name + HeaderMatchSeparator + h.Value
}

func isPathOnlyMatch(match dataplane.Match) bool {
	return match.Method == nil && len(match.Headers) == 0 && len(match.QueryParams) == 0
}

func createProxyPass(
	backendGroup dataplane.BackendGroup,
	filter *dataplane.HTTPURLRewriteFilter,
	protocol string,
	grpc bool,
) string {
	var requestURI string
	if !grpc {
		if filter == nil || filter.Path == nil {
			requestURI = "$request_uri"
		}
	}

	backendName := backendGroupName(backendGroup)
	if backendGroupNeedsSplit(backendGroup) {
		return protocol + "://$" + convertStringToSafeVariableName(backendName) + requestURI
	}

	return protocol + "://" + backendName + requestURI
}

func createMatchLocation(path string) http.Location {
	return http.Location{
		Path: path,
	}
}

func generateProxySetHeaders(filters *dataplane.HTTPFilters, grpc bool) []http.Header {
	var headers []http.Header
	if !grpc {
		headers = make([]http.Header, len(httpBaseHeaders))
		copy(headers, httpBaseHeaders)
	} else {
		headers = make([]http.Header, len(grpcBaseHeaders))
		copy(headers, grpcBaseHeaders)
	}

	if filters != nil && filters.RequestURLRewrite != nil && filters.RequestURLRewrite.Hostname != nil {
		for i, header := range headers {
			if header.Name == "Host" {
				headers[i].Value = *filters.RequestURLRewrite.Hostname
				break
			}
		}
	}

	if filters == nil || filters.RequestHeaderModifiers == nil {
		return headers
	}

	headerFilter := filters.RequestHeaderModifiers

	headerLen := len(headerFilter.Add) + len(headerFilter.Set) + len(headerFilter.Remove) + len(headers)
	proxySetHeaders := make([]http.Header, 0, headerLen)
	if len(headerFilter.Add) > 0 {
		addHeaders := createHeadersWithVarName(headerFilter.Add)
		proxySetHeaders = append(proxySetHeaders, addHeaders...)
	}
	if len(headerFilter.Set) > 0 {
		setHeaders := createHeaders(headerFilter.Set)
		proxySetHeaders = append(proxySetHeaders, setHeaders...)
	}
	// If the value of a header field is an empty string then this field will not be passed to a proxied server
	for _, h := range headerFilter.Remove {
		proxySetHeaders = append(proxySetHeaders, http.Header{
			Name:  h,
			Value: "",
		})
	}

	return append(proxySetHeaders, headers...)
}

func generateResponseHeaders(filters *dataplane.HTTPFilters) http.ResponseHeaders {
	if filters == nil || filters.ResponseHeaderModifiers == nil {
		return http.ResponseHeaders{}
	}

	headerFilter := filters.ResponseHeaderModifiers
	responseRemoveHeaders := make([]string, len(headerFilter.Remove))

	// Make a deep copy to prevent the slice from being accidentally modified.
	copy(responseRemoveHeaders, headerFilter.Remove)

	return http.ResponseHeaders{
		Add:    createHeaders(headerFilter.Add),
		Set:    createHeaders(headerFilter.Set),
		Remove: responseRemoveHeaders,
	}
}

func createHeadersWithVarName(headers []dataplane.HTTPHeader) []http.Header {
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

func createHeaders(headers []dataplane.HTTPHeader) []http.Header {
	locHeaders := make([]http.Header, 0, len(headers))
	for _, h := range headers {
		locHeaders = append(locHeaders, http.Header{
			Name:  h.Name,
			Value: h.Value,
		})
	}
	return locHeaders
}

func exactPath(path string) string {
	return fmt.Sprintf("= %s", path)
}

// createPath builds the location path depending on the path type.
func createPath(rule dataplane.PathRule) string {
	switch rule.PathType {
	case dataplane.PathTypeExact:
		return exactPath(rule.Path)
	default:
		return rule.Path
	}
}

func createDefaultRootLocation() http.Location {
	return http.Location{
		Path:   "/",
		Return: &http.Return{Code: http.StatusNotFound},
	}
}

// isNonSlashedPrefixPath returns whether or not a path is of type Prefix and does not contain a trailing slash.
func isNonSlashedPrefixPath(pathType dataplane.PathType, path string) bool {
	return pathType == dataplane.PathTypePrefix && !strings.HasSuffix(path, "/")
}
