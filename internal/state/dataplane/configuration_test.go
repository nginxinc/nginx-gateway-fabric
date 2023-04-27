package dataplane

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver/resolverfakes"
)

func TestBuildConfiguration(t *testing.T) {
	const (
		invalidMatchesPath = "/not-valid-matches"
		invalidFiltersPath = "/not-valid-filters"
	)

	createRoute := func(name, hostname, listenerName string, paths ...pathAndType) *v1beta1.HTTPRoute {
		rules := make([]v1beta1.HTTPRouteRule, 0, len(paths))
		for _, p := range paths {
			rules = append(rules, v1beta1.HTTPRouteRule{
				Matches: []v1beta1.HTTPRouteMatch{
					{
						Path: &v1beta1.HTTPPathMatch{
							Value: helpers.GetStringPointer(p.path),
							Type:  helpers.GetPointer(p.pathType),
						},
					},
				},
			})
		}
		return &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: []v1beta1.ParentReference{
						{
							Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
							Name:        "gateway",
							SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer(listenerName)),
						},
					},
				},
				Hostnames: []v1beta1.Hostname{
					v1beta1.Hostname(hostname),
				},
				Rules: rules,
			},
		}
	}

	addFilters := func(hr *v1beta1.HTTPRoute, filters []v1beta1.HTTPRouteFilter) {
		for i := range hr.Spec.Rules {
			hr.Spec.Rules[i].Filters = filters
		}
	}

	fooUpstreamName := "test_foo_80"

	fooEndpoints := []resolver.Endpoint{
		{
			Address: "10.0.0.0",
			Port:    8080,
		},
	}

	fooUpstream := Upstream{
		Name:      fooUpstreamName,
		Endpoints: fooEndpoints,
	}

	fooSvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test"}}

	fakeResolver := &resolverfakes.FakeServiceResolver{}
	fakeResolver.ResolveReturns(fooEndpoints, nil)

	validBackendRef := graph.BackendRef{
		Svc:    fooSvc,
		Port:   80,
		Valid:  true,
		Weight: 1,
	}

	expValidBackend := Backend{
		UpstreamName: fooUpstreamName,
		Weight:       1,
		Valid:        true,
	}

	createBackendRefs := func(validRule bool) []graph.BackendRef {
		if !validRule {
			return nil
		}

		return []graph.BackendRef{validBackendRef}
	}

	createRules := func(hr *v1beta1.HTTPRoute, paths []pathAndType) []graph.Rule {
		rules := make([]graph.Rule, len(hr.Spec.Rules))

		for i := range paths {
			validMatches := paths[i].path != invalidMatchesPath
			validFilters := paths[i].path != invalidFiltersPath
			validRule := validMatches && validFilters

			rules[i] = graph.Rule{
				ValidMatches: validMatches,
				ValidFilters: validFilters,
				BackendRefs:  createBackendRefs(validRule),
			}
		}

		return rules
	}

	createInternalRoute := func(
		source *v1beta1.HTTPRoute,
		paths []pathAndType,
	) *graph.Route {
		r := &graph.Route{
			Source: source,
			Rules:  createRules(source, paths),
		}
		return r
	}

	createExpBackendGroupsForRoute := func(route *graph.Route) []BackendGroup {
		groups := make([]BackendGroup, 0)

		for idx, r := range route.Rules {
			var backends []Backend
			if r.ValidFilters && r.ValidMatches {
				backends = []Backend{expValidBackend}
			}

			groups = append(groups, BackendGroup{
				Backends: backends,
				Source:   client.ObjectKeyFromObject(route.Source),
				RuleIdx:  idx,
			})
		}

		return groups
	}

	createTestResources := func(name, hostname, listenerName string, paths ...pathAndType) (
		*v1beta1.HTTPRoute, []BackendGroup, *graph.Route,
	) {
		hr := createRoute(name, hostname, listenerName, paths...)
		route := createInternalRoute(hr, paths)
		groups := createExpBackendGroupsForRoute(route)
		return hr, groups, route
	}

	prefix := v1beta1.PathMatchPathPrefix

	hr1, expHR1Groups, routeHR1 := createTestResources(
		"hr-1",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/", pathType: prefix},
	)
	hr2, expHR2Groups, routeHR2 := createTestResources(
		"hr-2",
		"bar.example.com",
		"listener-80-1",
		pathAndType{path: "/", pathType: prefix},
	)
	hr3, expHR3Groups, routeHR3 := createTestResources(
		"hr-3",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/", pathType: prefix},
		pathAndType{path: "/third", pathType: prefix},
	)
	hr4, expHR4Groups, routeHR4 := createTestResources(
		"hr-4",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/fourth", pathType: prefix},
		pathAndType{path: "/", pathType: prefix},
	)
	hr5, expHR5Groups, routeHR5 := createTestResources(
		"hr-5",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/", pathType: prefix}, pathAndType{path: invalidFiltersPath, pathType: prefix},
	)

	redirect := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
		},
	}
	addFilters(hr5, []v1beta1.HTTPRouteFilter{redirect})

	hr6, expHR6Groups, routeHR6 := createTestResources(
		"hr-6",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: invalidMatchesPath, pathType: prefix},
	)

	hr7, hr7Groups, routeHR7 := createTestResources(
		"hr-7",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: "/valid", pathType: v1beta1.PathMatchExact},
	)

	httpsHR1, expHTTPSHR1Groups, httpsRouteHR1 := createTestResources(
		"https-hr-1",
		"foo.example.com",
		"listener-443-1",
		pathAndType{path: "/", pathType: prefix},
	)

	httpsHR2, expHTTPSHR2Groups, httpsRouteHR2 := createTestResources(
		"https-hr-2",
		"bar.example.com",
		"listener-443-1",
		pathAndType{path: "/", pathType: prefix},
	)

	httpsHR3, expHTTPSHR3Groups, httpsRouteHR3 := createTestResources(
		"https-hr-3",
		"foo.example.com",
		"listener-443-1",
		pathAndType{path: "/", pathType: prefix}, pathAndType{path: "/third", pathType: prefix},
	)

	httpsHR4, expHTTPSHR4Groups, httpsRouteHR4 := createTestResources(
		"https-hr-4",
		"foo.example.com",
		"listener-443-1",
		pathAndType{path: "/fourth", pathType: prefix}, pathAndType{path: "/", pathType: prefix},
	)

	httpsHR5, expHTTPSHR5Groups, httpsRouteHR5 := createTestResources(
		"https-hr-5",
		"example.com",
		"listener-443-with-hostname",
		pathAndType{path: "/", pathType: prefix},
	)

	httpsHR6, expHTTPSHR6Groups, httpsRouteHR6 := createTestResources(
		"https-hr-6",
		"foo.example.com",
		"listener-443-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: invalidMatchesPath, pathType: prefix},
	)

	listener80 := v1beta1.Listener{
		Name:     "listener-80-1",
		Hostname: nil,
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}

	listener443 := v1beta1.Listener{
		Name:     "listener-443-1",
		Hostname: nil,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
					Name:      "secret",
					Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
				},
			},
		},
	}
	hostname := v1beta1.Hostname("example.com")

	listener443WithHostname := v1beta1.Listener{
		Name:     "listener-443-with-hostname",
		Hostname: &hostname,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
					Name:      "secret",
					Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
				},
			},
		},
	}

	invalidListener := v1beta1.Listener{
		Name:     "invalid-listener",
		Hostname: nil,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS:      nil, // missing TLS config
	}

	// nolint:gosec
	secretPath := "/etc/nginx/secrets/secret"

	tests := []struct {
		graph   *graph.Graph
		msg     string
		expConf Configuration
	}{
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source:    &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{},
				},
				Routes: map[types.NamespacedName]*graph.Route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
			},
			msg: "no listeners and routes",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source:            listener80,
							Valid:             true,
							Routes:            map[types.NamespacedName]*graph.Route{},
							AcceptedHostnames: map[string]struct{}{},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
				},
				SSLServers: []VirtualServer{},
			},
			msg: "http listener with no routes",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-443-1": {
							Source:            listener443, // nil hostname
							Valid:             true,
							Routes:            map[types.NamespacedName]*graph.Route{},
							AcceptedHostnames: map[string]struct{}{},
							SecretPath:        secretPath,
						},
						"listener-443-with-hostname": {
							Source:            listener443WithHostname, // non-nil hostname
							Valid:             true,
							Routes:            map[types.NamespacedName]*graph.Route{},
							AcceptedHostnames: map[string]struct{}{},
							SecretPath:        secretPath,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: string(hostname),
						SSL:      &SSL{CertificatePath: secretPath},
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{CertificatePath: secretPath},
					},
				},
			},
			msg: "https listeners with no routes",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"invalid-listener": {
							Source: invalidListener,
							Valid:  false,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
			},
			msg: "invalid listener",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-1"}: routeHR1,
								{Namespace: "test", Name: "hr-2"}: routeHR2,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
								"bar.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
					{Namespace: "test", Name: "hr-2"}: routeHR2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHR2Groups[0],
										Source:       hr2,
									},
								},
							},
						},
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHR1Groups[0],
										Source:       hr1,
									},
								},
							},
						},
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHR1Groups[0], expHR2Groups[0]},
			},
			msg: "one http listener with two routes for different hostnames",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-443-1": {
							Source:     listener443,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
								{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
								"bar.example.com": {},
							},
						},
						"listener-443-with-hostname": {
							Source:     listener443WithHostname,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
					{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR2Groups[0],
										Source:       httpsHR2,
									},
								},
							},
						},
						SSL: &SSL{
							CertificatePath: secretPath,
						},
					},
					{
						Hostname: "example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       httpsHR5,
									},
								},
							},
						},
						SSL: &SSL{
							CertificatePath: secretPath,
						},
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR1Groups[0],
										Source:       httpsHR1,
									},
								},
							},
						},
						SSL: &SSL{
							CertificatePath: secretPath,
						},
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{CertificatePath: secretPath},
					},
				},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHTTPSHR1Groups[0], expHTTPSHR2Groups[0], expHTTPSHR5Groups[0]},
			},
			msg: "two https listeners each with routes for different hostnames",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-3"}: routeHR3,
								{Namespace: "test", Name: "hr-4"}: routeHR4,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
						"listener-443-1": {
							Source:     listener443,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
								{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-3"}:       routeHR3,
					{Namespace: "test", Name: "hr-4"}:       routeHR4,
					{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
					{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHR3Groups[0],
										Source:       hr3,
									},
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: expHR4Groups[1],
										Source:       hr4,
									},
								},
							},
							{
								Path:     "/fourth",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHR4Groups[0],
										Source:       hr4,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: expHR3Groups[1],
										Source:       hr3,
									},
								},
							},
						},
					},
				},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						SSL: &SSL{
							CertificatePath: secretPath,
						},
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR3Groups[0],
										Source:       httpsHR3,
									},
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: expHTTPSHR4Groups[1],
										Source:       httpsHR4,
									},
								},
							},
							{
								Path:     "/fourth",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR4Groups[0],
										Source:       httpsHR4,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: expHTTPSHR3Groups[1],
										Source:       httpsHR3,
									},
								},
							},
						},
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{CertificatePath: secretPath},
					},
				},
				Upstreams: []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{
					expHR3Groups[0],
					expHR3Groups[1],
					expHR4Groups[0],
					expHR4Groups[1],
					expHTTPSHR3Groups[0],
					expHTTPSHR3Groups[1],
					expHTTPSHR4Groups[0],
					expHTTPSHR4Groups[1],
				},
			},
			msg: "one http and one https listener with two routes with the same hostname with and without collisions",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  false,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-1"}: routeHR1,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
				},
			},
			expConf: Configuration{},
			msg:     "invalid gatewayclass",
		},
		{
			graph: &graph.Graph{
				GatewayClass: nil,
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-1"}: routeHR1,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
				},
			},
			expConf: Configuration{},
			msg:     "missing gatewayclass",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: nil,
				Routes:  map[types.NamespacedName]*graph.Route{},
			},
			expConf: Configuration{},
			msg:     "missing gateway",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-5"}: routeHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-5"}: routeHR5,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										Source:       hr5,
										BackendGroup: expHR5Groups[0],
										Filters: Filters{
											RequestRedirect: redirect.RequestRedirect,
										},
									},
								},
							},
							{
								Path:     invalidFiltersPath,
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										Source:       hr5,
										BackendGroup: expHR5Groups[1],
										Filters: Filters{
											InvalidFilter: &InvalidFilter{},
										},
									},
								},
							},
						},
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHR5Groups[0], expHR5Groups[1]},
			},
			msg: "one http listener with one route with filters",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-6"}: routeHR6,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
						"listener-443-1": {
							Source:     listener443,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-6"}: httpsRouteHR6,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-6"}:       routeHR6,
					{Namespace: "test", Name: "https-hr-6"}: httpsRouteHR6,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHR6Groups[0],
										Source:       hr6,
									},
								},
							},
						},
					},
				},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						SSL: &SSL{
							CertificatePath: secretPath,
						},
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR6Groups[0],
										Source:       httpsHR6,
									},
								},
							},
						},
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{CertificatePath: secretPath},
					},
				},
				Upstreams: []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{
					expHR6Groups[0],
					expHTTPSHR6Groups[0],
				},
			},
			msg: "one http and one https listener with routes with valid and invalid rules",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-7"}: routeHR7,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-7"}: routeHR7,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypeExact,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: hr7Groups[1],
										Source:       hr7,
									},
								},
							},
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: hr7Groups[0],
										Source:       hr7,
									},
								},
							},
						},
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{hr7Groups[0], hr7Groups[1]},
			},
			msg: "duplicate paths with different types",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &graph.Gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*graph.Listener{
						"listener-443-with-hostname": {
							Source:     listener443WithHostname,
							Valid:      true,
							SecretPath: "secret-path-https-listener-2",
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"example.com": {},
							},
						},
						"listener-443-1": {
							Source:     listener443,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
					},
					{
						Hostname: "example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									// duplicate match rules since two listeners both match this route's hostname
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       httpsHR5,
									},
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       httpsHR5,
									},
								},
							},
						},
						SSL: &SSL{
							CertificatePath: "secret-path-https-listener-2",
						},
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{CertificatePath: secretPath},
					},
				},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHTTPSHR5Groups[0]},
			},
			msg: "two https listeners with different hostnames but same route; chooses listener with more specific hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result := BuildConfiguration(context.TODO(), test.graph, fakeResolver)

			sort.Slice(result.BackendGroups, func(i, j int) bool {
				return result.BackendGroups[i].Name() < result.BackendGroups[j].Name()
			})

			sort.Slice(result.Upstreams, func(i, j int) bool {
				return result.Upstreams[i].Name < result.Upstreams[j].Name
			})

			if diff := cmp.Diff(test.expConf, result); diff != "" {
				t.Errorf("BuildConfiguration() %q mismatch for configuration (-want +got):\n%s", test.msg, diff)
			}
		})
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		path     *v1beta1.HTTPPathMatch
		expected string
		msg      string
	}{
		{
			path:     &v1beta1.HTTPPathMatch{Value: helpers.GetStringPointer("/abc")},
			expected: "/abc",
			msg:      "normal case",
		},
		{
			path:     nil,
			expected: "/",
			msg:      "nil path",
		},
		{
			path:     &v1beta1.HTTPPathMatch{Value: nil},
			expected: "/",
			msg:      "nil value",
		},
		{
			path:     &v1beta1.HTTPPathMatch{Value: helpers.GetStringPointer("")},
			expected: "/",
			msg:      "empty value",
		},
	}

	for _, test := range tests {
		result := getPath(test.path)
		if result != test.expected {
			t.Errorf("getPath() returned %q but expected %q for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestCreateFilters(t *testing.T) {
	redirect1 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
		},
	}
	redirect2 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("bar.example.com")),
		},
	}

	tests := []struct {
		expected Filters
		msg      string
		filters  []v1beta1.HTTPRouteFilter
	}{
		{
			filters:  []v1beta1.HTTPRouteFilter{},
			expected: Filters{},
			msg:      "no filters",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
			},
			expected: Filters{
				RequestRedirect: redirect1.RequestRedirect,
			},
			msg: "one filter",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
				redirect2,
			},
			expected: Filters{
				RequestRedirect: redirect1.RequestRedirect,
			},
			msg: "two filters, first wins",
		},
	}

	for _, test := range tests {
		result := createFilters(test.filters)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("createFilters() %q mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestMatchRuleGetMatch(t *testing.T) {
	hr := &v1beta1.HTTPRoute{
		Spec: v1beta1.HTTPRouteSpec{
			Rules: []v1beta1.HTTPRouteRule{
				{
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-1"),
							},
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-2"),
							},
						},
					},
				},
				{
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-3"),
							},
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-4"),
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name    string
		expPath string
		rule    MatchRule
	}{
		{
			name:    "first match in first rule",
			expPath: "/path-1",
			rule:    MatchRule{MatchIdx: 0, RuleIdx: 0, Source: hr},
		},
		{
			name:    "second match in first rule",
			expPath: "/path-2",
			rule:    MatchRule{MatchIdx: 1, RuleIdx: 0, Source: hr},
		},
		{
			name:    "second match in second rule",
			expPath: "/path-4",
			rule:    MatchRule{MatchIdx: 1, RuleIdx: 1, Source: hr},
		},
	}

	for _, tc := range tests {
		actual := tc.rule.GetMatch()
		if *actual.Path.Value != tc.expPath {
			t.Errorf(
				"MatchRule.GetMatch() returned incorrect match with path: %s, expected path: %s for test case: %q",
				*actual.Path.Value,
				tc.expPath,
				tc.name,
			)
		}
	}
}

func TestGetListenerHostname(t *testing.T) {
	var emptyHostname v1beta1.Hostname
	var hostname v1beta1.Hostname = "example.com"

	tests := []struct {
		hostname *v1beta1.Hostname
		expected string
		msg      string
	}{
		{
			hostname: nil,
			expected: wildcardHostname,
			msg:      "nil hostname",
		},
		{
			hostname: &emptyHostname,
			expected: wildcardHostname,
			msg:      "empty hostname",
		},
		{
			hostname: &hostname,
			expected: string(hostname),
			msg:      "normal hostname",
		},
	}

	for _, test := range tests {
		result := getListenerHostname(test.hostname)
		if result != test.expected {
			t.Errorf(
				"getListenerHostname() returned %q but expected %q for the case of %q",
				result,
				test.expected,
				test.msg,
			)
		}
	}
}

func refsToValidRules(refs ...[]graph.BackendRef) []graph.Rule {
	rules := make([]graph.Rule, 0, len(refs))

	for _, ref := range refs {
		rules = append(rules, graph.Rule{
			ValidMatches: true,
			ValidFilters: true,
			BackendRefs:  ref,
		})
	}

	return rules
}

func TestBuildUpstreams(t *testing.T) {
	fooEndpoints := []resolver.Endpoint{
		{
			Address: "10.0.0.0",
			Port:    8080,
		},
		{
			Address: "10.0.0.1",
			Port:    8080,
		},
		{
			Address: "10.0.0.2",
			Port:    8080,
		},
	}

	barEndpoints := []resolver.Endpoint{
		{
			Address: "11.0.0.0",
			Port:    80,
		},
		{
			Address: "11.0.0.1",
			Port:    80,
		},
		{
			Address: "11.0.0.2",
			Port:    80,
		},
		{
			Address: "11.0.0.3",
			Port:    80,
		},
	}

	bazEndpoints := []resolver.Endpoint{
		{
			Address: "12.0.0.0",
			Port:    80,
		},
	}

	baz2Endpoints := []resolver.Endpoint{
		{
			Address: "13.0.0.0",
			Port:    80,
		},
	}

	createBackendRefs := func(serviceNames ...string) []graph.BackendRef {
		var backends []graph.BackendRef
		for _, name := range serviceNames {
			backends = append(backends, graph.BackendRef{
				Svc:   &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name}},
				Port:  80,
				Valid: name != "",
			})
		}
		return backends
	}

	hr1Refs0 := createBackendRefs("foo", "bar")

	hr1Refs1 := createBackendRefs("baz", "", "") // empty service names should be ignored

	hr2Refs0 := createBackendRefs("foo", "baz") // shouldn't duplicate foo and baz upstream

	hr2Refs1 := createBackendRefs("nil-endpoints")

	hr3Refs0 := createBackendRefs("baz") // shouldn't duplicate baz upstream

	hr4Refs0 := createBackendRefs("empty-endpoints", "")

	hr4Refs1 := createBackendRefs("baz2")

	invalidRefs := createBackendRefs("invalid")

	routes := map[types.NamespacedName]*graph.Route{
		{Name: "hr1", Namespace: "test"}: {
			Rules: refsToValidRules(hr1Refs0, hr1Refs1),
		},
		{Name: "hr2", Namespace: "test"}: {
			Rules: refsToValidRules(hr2Refs0, hr2Refs1),
		},
		{Name: "hr3", Namespace: "test"}: {
			Rules: refsToValidRules(hr3Refs0),
		},
	}

	routes2 := map[types.NamespacedName]*graph.Route{
		{Name: "hr4", Namespace: "test"}: {
			Rules: refsToValidRules(hr4Refs0, hr4Refs1),
		},
	}

	invalidRoutes := map[types.NamespacedName]*graph.Route{
		{Name: "invalid", Namespace: "test"}: {
			Rules: refsToValidRules(invalidRefs),
		},
	}

	listeners := map[string]*graph.Listener{
		"invalid-listener": {
			Valid:  false,
			Routes: invalidRoutes, // shouldn't be included since listener is invalid
		},
		"listener-1": {
			Valid:  true,
			Routes: routes,
		},
		"listener-2": {
			Valid:  true,
			Routes: routes2,
		},
	}

	emptyEndpointsErrMsg := "empty endpoints error"
	nilEndpointsErrMsg := "nil endpoints error"

	expUpstreams := []Upstream{
		{
			Name:      "test_bar_80",
			Endpoints: barEndpoints,
		},
		{
			Name:      "test_baz2_80",
			Endpoints: baz2Endpoints,
		},
		{
			Name:      "test_baz_80",
			Endpoints: bazEndpoints,
		},
		{
			Name:      "test_empty-endpoints_80",
			Endpoints: []resolver.Endpoint{},
			ErrorMsg:  emptyEndpointsErrMsg,
		},
		{
			Name:      "test_foo_80",
			Endpoints: fooEndpoints,
		},
		{
			Name:      "test_nil-endpoints_80",
			Endpoints: nil,
			ErrorMsg:  nilEndpointsErrMsg,
		},
	}

	fakeResolver := &resolverfakes.FakeServiceResolver{}
	fakeResolver.ResolveCalls(func(ctx context.Context, svc *v1.Service, port int32) ([]resolver.Endpoint, error) {
		switch svc.Name {
		case "bar":
			return barEndpoints, nil
		case "baz":
			return bazEndpoints, nil
		case "baz2":
			return baz2Endpoints, nil
		case "empty-endpoints":
			return []resolver.Endpoint{}, errors.New(emptyEndpointsErrMsg)
		case "foo":
			return fooEndpoints, nil
		case "nil-endpoints":
			return nil, errors.New(nilEndpointsErrMsg)
		default:
			return nil, fmt.Errorf("unexpected service %s", svc.Name)
		}
	})

	g := NewGomegaWithT(t)

	upstreams := buildUpstreams(context.TODO(), listeners, fakeResolver)
	g.Expect(upstreams).To(ConsistOf(expUpstreams))
}

func TestBuildBackendGroups(t *testing.T) {
	createBackendGroup := func(name string, ruleIdx int, backendNames ...string) BackendGroup {
		backends := make([]Backend, len(backendNames))
		for i, name := range backendNames {
			backends[i] = Backend{UpstreamName: name}
		}

		return BackendGroup{
			Source:   types.NamespacedName{Namespace: "test", Name: name},
			RuleIdx:  ruleIdx,
			Backends: backends,
		}
	}

	hr1Group0 := createBackendGroup("hr1", 0, "foo", "bar")

	hr1Group1 := createBackendGroup("hr1", 1, "foo")

	hr2Group0 := createBackendGroup("hr2", 0, "foo", "bar")

	hr2Group1 := createBackendGroup("hr2", 1, "foo")

	hr3Group0 := createBackendGroup("hr3", 0, "foo", "bar")

	hr3Group1 := createBackendGroup("hr3", 1, "foo")

	// groups with no backends should still be included
	hrNoBackends := createBackendGroup("no-backends", 0)

	createServer := func(groups ...BackendGroup) VirtualServer {
		matchRules := make([]MatchRule, 0, len(groups))
		for _, g := range groups {
			matchRules = append(matchRules, MatchRule{BackendGroup: g})
		}

		server := VirtualServer{
			PathRules: []PathRule{
				{
					MatchRules: matchRules,
				},
			},
		}

		return server
	}
	servers := []VirtualServer{
		createServer(hr1Group0, hr1Group1),
		createServer(hr2Group0, hr2Group1),
		createServer(hr3Group0, hr3Group1),
		createServer(hr1Group0, hr1Group1), // next three are duplicates
		createServer(hr2Group0, hr2Group1),
		createServer(hr3Group0, hr3Group1),
		createServer(hrNoBackends),
	}

	expGroups := []BackendGroup{
		hr1Group0,
		hr1Group1,
		hr2Group0,
		hr2Group1,
		hr3Group0,
		hr3Group1,
		hrNoBackends,
	}

	g := NewGomegaWithT(t)

	result := buildBackendGroups(servers)

	g.Expect(result).To(ConsistOf(expGroups))
}

func TestConvertPathType(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		expected PathType
		pathType v1beta1.PathMatchType
		panic    bool
	}{
		{
			expected: PathTypePrefix,
			pathType: v1beta1.PathMatchPathPrefix,
		},
		{
			expected: PathTypeExact,
			pathType: v1beta1.PathMatchExact,
		},
		{
			pathType: v1beta1.PathMatchRegularExpression,
			panic:    true,
		},
	}

	for _, tc := range tests {
		if tc.panic {
			g.Expect(func() { convertPathType(tc.pathType) }).To(Panic())
		} else {
			result := convertPathType(tc.pathType)
			g.Expect(result).To(Equal(tc.expected))
		}
	}
}

func TestHostnameMoreSpecific(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		host1     *v1beta1.Hostname
		host2     *v1beta1.Hostname
		msg       string
		host1Wins bool
	}{
		{
			host1:     nil,
			host2:     helpers.GetPointer(v1beta1.Hostname("")),
			host1Wins: true,
			msg:       "host1 nil; host2 empty",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("")),
			host2:     nil,
			host1Wins: true,
			msg:       "host1 empty; host2 nil",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("")),
			host2:     helpers.GetPointer(v1beta1.Hostname("")),
			host1Wins: true,
			msg:       "both hosts empty",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("example.com")),
			host2:     helpers.GetPointer(v1beta1.Hostname("")),
			host1Wins: true,
			msg:       "host1 has value; host2 empty",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("example.com")),
			host2:     helpers.GetPointer(v1beta1.Hostname("foo.example.com")),
			host1Wins: false,
			msg:       "host2 longer than host1",
		},
	}

	for _, tc := range tests {
		g.Expect(listenerHostnameMoreSpecific(tc.host1, tc.host2)).To(Equal(tc.host1Wins), tc.msg)
	}
}
