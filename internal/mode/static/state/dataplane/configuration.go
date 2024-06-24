package dataplane

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/observability"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

const (
	wildcardHostname    = "~^"
	alpineSSLRootCAPath = "/etc/ssl/cert.pem"
)

// BuildConfiguration builds the Configuration from the Graph.
func BuildConfiguration(
	ctx context.Context,
	g *graph.Graph,
	resolver resolver.ServiceResolver,
	generator policies.ConfigGenerator,
	configVersion int,
) Configuration {
	if g.GatewayClass == nil || !g.GatewayClass.Valid {
		return Configuration{Version: configVersion}
	}

	if g.Gateway == nil {
		return Configuration{Version: configVersion}
	}

	upstreams := buildUpstreams(ctx, g.Gateway.Listeners, resolver)
	httpServers, sslServers := buildServers(g, generator)
	backendGroups := buildBackendGroups(append(httpServers, sslServers...))
	keyPairs := buildSSLKeyPairs(g.ReferencedSecrets, g.Gateway.Listeners)
	certBundles := buildCertBundles(g.ReferencedCaCertConfigMaps, backendGroups)
	telemetry := buildTelemetry(g)
	baseHTTPConfig := buildBaseHTTPConfig(g)
	var tlsServers []Layer4VirtualServer

	config := Configuration{
		HTTPServers:           httpServers,
		SSLServers:            sslServers,
		TLSPassthroughServers: tlsServers,
		Upstreams:             upstreams,
		BackendGroups:         backendGroups,
		SSLKeyPairs:           keyPairs,
		Version:               configVersion,
		CertBundles:           certBundles,
		Telemetry:             telemetry,
		BaseHTTPConfig:        baseHTTPConfig,
	}

	return config
}

// buildSSLKeyPairs builds the SSLKeyPairs from the Secrets. It will only include Secrets that are referenced by
// valid listeners, so that we don't include unused Secrets in the configuration of the data plane.
func buildSSLKeyPairs(
	secrets map[types.NamespacedName]*graph.Secret,
	listeners []*graph.Listener,
) map[SSLKeyPairID]SSLKeyPair {
	keyPairs := make(map[SSLKeyPairID]SSLKeyPair)

	for _, l := range listeners {
		if l.Valid && l.ResolvedSecret != nil {
			id := generateSSLKeyPairID(*l.ResolvedSecret)
			secret := secrets[*l.ResolvedSecret]
			// The Data map keys are guaranteed to exist by the graph package.
			// the Source field is guaranteed to be non-nil by the graph package.
			keyPairs[id] = SSLKeyPair{
				Cert: secret.Source.Data[apiv1.TLSCertKey],
				Key:  secret.Source.Data[apiv1.TLSPrivateKeyKey],
			}
		}
	}

	return keyPairs
}

func buildCertBundles(
	caCertConfigMaps map[types.NamespacedName]*graph.CaCertConfigMap,
	backendGroups []BackendGroup,
) map[CertBundleID]CertBundle {
	bundles := make(map[CertBundleID]CertBundle)
	refByBG := make(map[CertBundleID]struct{})

	// We only need to build the cert bundles if there are valid backend groups that reference them.
	if len(backendGroups) == 0 {
		return bundles
	}
	for _, bg := range backendGroups {
		if bg.Backends == nil {
			continue
		}
		for _, b := range bg.Backends {
			if !b.Valid || b.VerifyTLS == nil {
				continue
			}
			refByBG[b.VerifyTLS.CertBundleID] = struct{}{}
		}
	}

	for cmName, cm := range caCertConfigMaps {
		id := generateCertBundleID(cmName)
		if _, exists := refByBG[id]; exists {
			if cm.CACert != nil || len(cm.CACert) > 0 {
				// the cert could be base64 encoded or plaintext
				data := make([]byte, base64.StdEncoding.DecodedLen(len(cm.CACert)))
				_, err := base64.StdEncoding.Decode(data, cm.CACert)
				if err != nil {
					data = cm.CACert
				}
				bundles[id] = CertBundle(data)
			}
		}
	}

	return bundles
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
			VerifyTLS:    convertBackendTLS(ref.BackendTLSPolicy),
		})
	}

	return BackendGroup{
		Backends: backends,
		Source:   sourceNsName,
		RuleIdx:  ruleIdx,
	}
}

func convertBackendTLS(btp *graph.BackendTLSPolicy) *VerifyTLS {
	if btp == nil || !btp.Valid {
		return nil
	}
	verify := &VerifyTLS{}
	if btp.CaCertRef.Name != "" {
		verify.CertBundleID = generateCertBundleID(btp.CaCertRef)
	} else {
		verify.RootCAPath = alpineSSLRootCAPath
	}
	verify.Hostname = string(btp.Source.Spec.Validation.Hostname)
	return verify
}

func buildServers(g *graph.Graph, generator policies.ConfigGenerator) (http, ssl []VirtualServer) {
	rulesForProtocol := map[v1.ProtocolType]portPathRules{
		v1.HTTPProtocolType:  make(portPathRules),
		v1.HTTPSProtocolType: make(portPathRules),
	}

	for _, l := range g.Gateway.Listeners {
		if l.Valid {
			rules := rulesForProtocol[l.Source.Protocol][l.Source.Port]
			if rules == nil {
				rules = newHostPathRules(generator)
				rulesForProtocol[l.Source.Protocol][l.Source.Port] = rules
			}

			rules.upsertListener(l, g.GlobalSettings)
		}
	}

	httpRules := rulesForProtocol[v1.HTTPProtocolType]
	sslRules := rulesForProtocol[v1.HTTPSProtocolType]

	httpServers, sslServers := httpRules.buildServers(), sslRules.buildServers()

	additions := buildAdditions(g.Gateway.Policies, g.GlobalSettings, generator)

	for i := range httpServers {
		httpServers[i].Additions = additions
	}

	for i := range sslServers {
		sslServers[i].Additions = additions
	}

	return httpServers, sslServers
}

// portPathRules keeps track of hostPathRules per port.
type portPathRules map[v1.PortNumber]*hostPathRules

func (p portPathRules) buildServers() []VirtualServer {
	serverCount := 0
	for _, rules := range p {
		serverCount += rules.maxServerCount()
	}

	servers := make([]VirtualServer, 0, serverCount)

	for _, rules := range p {
		servers = append(servers, rules.buildServers()...)
	}

	return servers
}

type pathAndType struct {
	path     string
	pathType v1.PathMatchType
}

type hostPathRules struct {
	generator        policies.ConfigGenerator
	rulesPerHost     map[string]map[pathAndType]PathRule
	listenersForHost map[string]*graph.Listener
	httpsListeners   []*graph.Listener
	port             int32
	listenersExist   bool
}

func newHostPathRules(generator policies.ConfigGenerator) *hostPathRules {
	return &hostPathRules{
		rulesPerHost:     make(map[string]map[pathAndType]PathRule),
		listenersForHost: make(map[string]*graph.Listener),
		httpsListeners:   make([]*graph.Listener, 0),
		generator:        generator,
	}
}

func (hpr *hostPathRules) upsertListener(l *graph.Listener, globalSettings *policies.GlobalSettings) {
	hpr.listenersExist = true
	hpr.port = int32(l.Source.Port)

	if l.Source.Protocol == v1.HTTPSProtocolType {
		hpr.httpsListeners = append(hpr.httpsListeners, l)
	}

	for _, r := range l.Routes {
		if !r.Valid {
			continue
		}

		hpr.upsertRoute(r, l, globalSettings)
	}
}

func (hpr *hostPathRules) upsertRoute(
	route *graph.L7Route,
	listener *graph.Listener,
	globalSettings *policies.GlobalSettings,
) {
	var hostnames []string
	GRPC := route.RouteType == graph.RouteTypeGRPC

	var objectSrc *metav1.ObjectMeta

	if GRPC {
		objectSrc = &helpers.MustCastObject[*v1.GRPCRoute](route.Source).ObjectMeta
	} else {
		objectSrc = &helpers.MustCastObject[*v1.HTTPRoute](route.Source).ObjectMeta
	}

	for _, p := range route.ParentRefs {
		if val, exist := p.Attachment.AcceptedHostnames[string(listener.Source.Name)]; exist {
			hostnames = val
		}
	}

	for _, h := range hostnames {
		if prevListener, exists := hpr.listenersForHost[h]; exists {
			// override the previous listener if the new one has a more specific hostname
			if listenerHostnameMoreSpecific(listener.Source.Hostname, prevListener.Source.Hostname) {
				hpr.listenersForHost[h] = listener
			}
		} else {
			hpr.listenersForHost[h] = listener
		}

		if _, exist := hpr.rulesPerHost[h]; !exist {
			hpr.rulesPerHost[h] = make(map[pathAndType]PathRule)
		}
	}

	for i, rule := range route.Spec.Rules {
		if !rule.ValidMatches {
			continue
		}

		var filters HTTPFilters
		if rule.ValidFilters {
			filters = createHTTPFilters(rule.Filters)
		} else {
			filters = HTTPFilters{
				InvalidFilter: &InvalidHTTPFilter{},
			}
		}

		additions := buildAdditions(route.Policies, globalSettings, hpr.generator)

		for _, h := range hostnames {
			for _, m := range rule.Matches {
				path := getPath(m.Path)

				key := pathAndType{
					path:     path,
					pathType: *m.Path.Type,
				}

				hostRule, exist := hpr.rulesPerHost[h][key]
				if !exist {
					hostRule.Path = path
					hostRule.PathType = convertPathType(*m.Path.Type)
				}

				routeNsName := client.ObjectKeyFromObject(route.Source)

				hostRule.GRPC = GRPC

				hostRule.MatchRules = append(hostRule.MatchRules, MatchRule{
					Source:       objectSrc,
					BackendGroup: newBackendGroup(rule.BackendRefs, routeNsName, i),
					Filters:      filters,
					Match:        convertMatch(m),
					Additions:    additions,
				})

				hpr.rulesPerHost[h][key] = hostRule
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
			Port:      hpr.port,
		}

		l, ok := hpr.listenersForHost[h]
		if !ok {
			panic(fmt.Sprintf("no listener found for hostname: %s", h))
		}

		if l.ResolvedSecret != nil {
			s.SSL = &SSL{
				KeyPairID: generateSSLKeyPairID(*l.ResolvedSecret),
			}
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
				Port:     hpr.port,
			}

			if l.ResolvedSecret != nil {
				s.SSL = &SSL{
					KeyPairID: generateSSLKeyPairID(*l.ResolvedSecret),
				}
			}

			servers = append(servers, s)
		}
	}

	// if any listeners exist, we need to generate a default server block.
	if hpr.listenersExist {
		servers = append(servers, VirtualServer{
			IsDefault: true,
			Port:      hpr.port,
		})
	}

	// We sort the servers so the order is preserved after reconfiguration.
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})

	return servers
}

// maxServerCount returns the maximum number of VirtualServers that can be built from the host path rules.
func (hpr *hostPathRules) maxServerCount() int {
	// to calculate max # of servers we add up:
	// - # of hostnames
	// - # of https listeners - this is to account for https wildcard default servers
	// - default server - for every hostPathRules we generate 1 default server
	return len(hpr.rulesPerHost) + len(hpr.httpsListeners) + 1
}

func buildUpstreams(
	ctx context.Context,
	listeners []*graph.Listener,
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
			if !route.Valid {
				continue
			}

			for _, rule := range route.Spec.Rules {
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

						eps, err := resolver.Resolve(ctx, br.SvcNsName, br.ServicePort)
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

func getListenerHostname(h *v1.Hostname) string {
	if h == nil || *h == "" {
		return wildcardHostname
	}

	return string(*h)
}

func getPath(path *v1.HTTPPathMatch) string {
	if path == nil || path.Value == nil || *path.Value == "" {
		return "/"
	}
	return *path.Value
}

func createHTTPFilters(filters []v1.HTTPRouteFilter) HTTPFilters {
	var result HTTPFilters

	for _, f := range filters {
		switch f.Type {
		case v1.HTTPRouteFilterRequestRedirect:
			if result.RequestRedirect == nil {
				// using the first filter
				result.RequestRedirect = convertHTTPRequestRedirectFilter(f.RequestRedirect)
			}
		case v1.HTTPRouteFilterURLRewrite:
			if result.RequestURLRewrite == nil {
				// using the first filter
				result.RequestURLRewrite = convertHTTPURLRewriteFilter(f.URLRewrite)
			}
		case v1.HTTPRouteFilterRequestHeaderModifier:
			if result.RequestHeaderModifiers == nil {
				// using the first filter
				result.RequestHeaderModifiers = convertHTTPHeaderFilter(f.RequestHeaderModifier)
			}
		case v1.HTTPRouteFilterResponseHeaderModifier:
			if result.ResponseHeaderModifiers == nil {
				// using the first filter
				result.ResponseHeaderModifiers = convertHTTPHeaderFilter(f.ResponseHeaderModifier)
			}
		}
	}
	return result
}

// listenerHostnameMoreSpecific returns true if host1 is more specific than host2.
func listenerHostnameMoreSpecific(host1, host2 *v1.Hostname) bool {
	var host1Str, host2Str string
	if host1 != nil {
		host1Str = string(*host1)
	}

	if host2 != nil {
		host2Str = string(*host2)
	}

	return graph.GetMoreSpecificHostname(host1Str, host2Str) == host1Str
}

// generateSSLKeyPairID generates an ID for the SSL key pair based on the Secret namespaced name.
// It is guaranteed to be unique per unique namespaced name.
// The ID is safe to use as a file name.
func generateSSLKeyPairID(secret types.NamespacedName) SSLKeyPairID {
	return SSLKeyPairID(fmt.Sprintf("ssl_keypair_%s_%s", secret.Namespace, secret.Name))
}

// generateCertBundleID generates an ID for the certificate bundle based on the ConfigMap namespaced name.
// It is guaranteed to be unique per unique namespaced name.
// The ID is safe to use as a file name.
func generateCertBundleID(configMap types.NamespacedName) CertBundleID {
	return CertBundleID(fmt.Sprintf("cert_bundle_%s_%s", configMap.Namespace, configMap.Name))
}

// buildTelemetry generates the Otel configuration.
func buildTelemetry(g *graph.Graph) Telemetry {
	if g.NginxProxy == nil || !g.NginxProxy.Valid ||
		g.NginxProxy.Source.Spec.Telemetry == nil ||
		g.NginxProxy.Source.Spec.Telemetry.Exporter == nil {
		return Telemetry{}
	}

	serviceName := fmt.Sprintf("ngf:%s:%s", g.Gateway.Source.Namespace, g.Gateway.Source.Name)
	telemetry := g.NginxProxy.Source.Spec.Telemetry
	if telemetry.ServiceName != nil {
		serviceName = serviceName + ":" + *telemetry.ServiceName
	}

	tel := Telemetry{
		Endpoint:    telemetry.Exporter.Endpoint,
		ServiceName: serviceName,
	}

	if telemetry.Exporter.BatchCount != nil {
		tel.BatchCount = *telemetry.Exporter.BatchCount
	}
	if telemetry.Exporter.BatchSize != nil {
		tel.BatchSize = *telemetry.Exporter.BatchSize
	}
	if telemetry.Exporter.Interval != nil {
		tel.Interval = string(*telemetry.Exporter.Interval)
	}

	// FIXME(sberman): https://github.com/nginxinc/nginx-gateway-fabric/issues/2038
	// Find a generic way to include relevant policy info at the http context so we don't need policy-specific
	// logic in this function
	ratioMap := make(map[string]int32)
	for _, pol := range g.NGFPolicies {
		if obsPol, ok := pol.Source.(*ngfAPI.ObservabilityPolicy); ok {
			if obsPol.Spec.Tracing != nil && obsPol.Spec.Tracing.Ratio != nil && *obsPol.Spec.Tracing.Ratio > 0 {
				ratioName := observability.CreateRatioVarName(obsPol)
				ratioMap[ratioName] = *obsPol.Spec.Tracing.Ratio
			}
		}
	}

	tel.Ratios = make([]Ratio, 0, len(ratioMap))
	for name, ratio := range ratioMap {
		tel.Ratios = append(tel.Ratios, Ratio{Name: name, Value: ratio})
	}

	return tel
}

// buildBaseHTTPConfig generates the base http context config that should be applied to all servers.
func buildBaseHTTPConfig(g *graph.Graph) BaseHTTPConfig {
	baseConfig := BaseHTTPConfig{
		// HTTP2 should be enabled by default
		HTTP2: true,
	}
	if g.NginxProxy == nil || !g.NginxProxy.Valid {
		return baseConfig
	}

	if g.NginxProxy.Source.Spec.DisableHTTP2 {
		baseConfig.HTTP2 = false
	}

	return baseConfig
}

func buildAdditions(
	policies []*graph.Policy,
	globalSettings *policies.GlobalSettings,
	generator policies.ConfigGenerator,
) []Addition {
	if len(policies) == 0 {
		return nil
	}

	additions := make([]Addition, 0, len(policies))

	for _, policy := range policies {
		if !policy.Valid {
			continue
		}

		additions = append(additions, Addition{
			Bytes: generator.Generate(policy.Source, globalSettings),
			Identifier: fmt.Sprintf(
				"%s_%s_%s",
				policy.Source.GetObjectKind().GroupVersionKind().Kind,
				policy.Source.GetNamespace(),
				policy.Source.GetName(),
			),
		})
	}

	return additions
}
