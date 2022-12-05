package state

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver/resolverfakes"
)

func TestBuildConfiguration(t *testing.T) {
	createRoute := func(name string, hostname string, listenerName string, paths ...string) *v1beta1.HTTPRoute {
		rules := make([]v1beta1.HTTPRouteRule, 0, len(paths))
		for _, p := range paths {
			rules = append(rules, v1beta1.HTTPRouteRule{
				Matches: []v1beta1.HTTPRouteMatch{
					{
						Path: &v1beta1.HTTPPathMatch{
							Value: helpers.GetStringPointer(p),
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

	addFilters := func(hr *v1beta1.HTTPRoute, filters []v1beta1.HTTPRouteFilter) *v1beta1.HTTPRoute {
		for i := range hr.Spec.Rules {
			hr.Spec.Rules[i].Filters = filters
		}
		return hr
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

	createBackendGroup := func(nsname types.NamespacedName, idx int) BackendGroup {
		return BackendGroup{
			Source:  nsname,
			RuleIdx: idx,
			Backends: []BackendRef{
				{
					Name:   fooUpstreamName,
					Svc:    fooSvc,
					Port:   80,
					Valid:  true,
					Weight: 1,
				},
			},
		}
	}

	createInternalRoute := func(source *v1beta1.HTTPRoute, validSectionName string, groups ...BackendGroup) *route {
		r := &route{
			Source:                 source,
			InvalidSectionNameRefs: make(map[string]conditions.RouteCondition),
			ValidSectionNameRefs:   map[string]struct{}{validSectionName: {}},
			BackendGroups:          groups,
		}
		return r
	}

	createTestResources := func(name, hostname, listenerName string, paths ...string) (
		*v1beta1.HTTPRoute, []BackendGroup, *route,
	) {
		hr := createRoute(name, hostname, listenerName, paths...)
		groups := make([]BackendGroup, 0, len(paths))
		for idx := range paths {
			groups = append(groups, createBackendGroup(types.NamespacedName{Namespace: "test", Name: name}, idx))
		}

		route := createInternalRoute(hr, listenerName, groups...)
		return hr, groups, route
	}

	hr1, hr1Groups, routeHR1 := createTestResources("hr-1", "foo.example.com", "listener-80-1", "/")
	hr2, hr2Groups, routeHR2 := createTestResources("hr-2", "bar.example.com", "listener-80-1", "/")
	hr3, hr3Groups, routeHR3 := createTestResources("hr-3", "foo.example.com", "listener-80-1", "/", "/third")
	hr4, hr4Groups, routeHR4 := createTestResources("hr-4", "foo.example.com", "listener-80-1", "/fourth", "/")

	httpsHR1, httpsHR1Groups, httpsRouteHR1 := createTestResources(
		"https-hr-1",
		"foo.example.com",
		"listener-443-1",
		"/",
	)

	httpsHR2, httpsHR2Groups, httpsRouteHR2 := createTestResources(
		"https-hr-2",
		"bar.example.com",
		"listener-443-1",
		"/",
	)

	httpsHR3, httpsHR3Groups, httpsRouteHR3 := createTestResources(
		"https-hr-3",
		"foo.example.com",
		"listener-443-1",
		"/", "/third",
	)

	httpsHR4, httpsHR4Groups, httpsRouteHR4 := createTestResources(
		"https-hr-4",
		"foo.example.com",
		"listener-443-1",
		"/fourth", "/",
	)

	httpsHR5, httpsHR5Groups, httpsRouteHR5 := createTestResources(
		"https-hr-5",
		"example.com",
		"listener-443-with-hostname",
		"/",
	)

	redirect := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
		},
	}

	hr5 := addFilters(
		createRoute("hr-5", "foo.example.com", "listener-80-1", "/"),
		[]v1beta1.HTTPRouteFilter{redirect},
	)

	hr5BackendGroup := BackendGroup{
		Source:  types.NamespacedName{Namespace: hr5.Namespace, Name: hr5.Name},
		RuleIdx: 0,
	}

	routeHR5 := &route{
		Source:                 hr5,
		InvalidSectionNameRefs: make(map[string]conditions.RouteCondition),
		ValidSectionNameRefs:   map[string]struct{}{"listener-80-1": {}},
		BackendGroups:          []BackendGroup{hr5BackendGroup},
	}

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
		graph    *graph
		expWarns Warnings
		msg      string
		expConf  Configuration
	}{
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source:    &v1beta1.Gateway{},
					Listeners: map[string]*listener{},
				},
				Routes: map[types.NamespacedName]*route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
			},
			msg: "no listeners and routes",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source:            listener80,
							Valid:             true,
							Routes:            map[types.NamespacedName]*route{},
							AcceptedHostnames: map[string]struct{}{},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
			},
			msg: "http listener with no routes",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-443-1": {
							Source:            listener443, // nil hostname
							Valid:             true,
							Routes:            map[types.NamespacedName]*route{},
							AcceptedHostnames: map[string]struct{}{},
							SecretPath:        secretPath,
						},
						"listener-443-with-hostname": {
							Source:            listener443WithHostname, // non-nil hostname
							Valid:             true,
							Routes:            map[types.NamespacedName]*route{},
							AcceptedHostnames: map[string]struct{}{},
							SecretPath:        secretPath,
						},
					},
				},
				Routes: map[types.NamespacedName]*route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
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
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"invalid-listener": {
							Source: invalidListener,
							Valid:  false,
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
								{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
								"bar.example.com": {},
							},
							SecretPath: "",
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
			},
			expWarns: Warnings{
				httpsHR1: []string{"cannot configure routes for listener invalid-listener; listener is invalid"},
				httpsHR2: []string{"cannot configure routes for listener invalid-listener; listener is invalid"},
			},
			msg: "invalid listener",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*route{
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
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
					{Namespace: "test", Name: "hr-2"}: routeHR2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: hr2Groups[0],
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
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: hr1Groups[0],
										Source:       hr1,
									},
								},
							},
						},
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{hr1Groups[0], hr2Groups[0]},
			},
			msg: "one http listener with two routes for different hostnames",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-443-1": {
							Source:     listener443,
							Valid:      true,
							SecretPath: secretPath,
							Routes: map[types.NamespacedName]*route{
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
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
					{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: httpsHR2Groups[0],
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
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: httpsHR5Groups[0],
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
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: httpsHR1Groups[0],
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
				BackendGroups: []BackendGroup{httpsHR1Groups[0], httpsHR2Groups[0], httpsHR5Groups[0]},
			},
			msg: "two https listeners each with routes for different hostnames",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*route{
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
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
								{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-3"}:       routeHR3,
					{Namespace: "test", Name: "hr-4"}:       routeHR4,
					{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
					{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: hr3Groups[0],
										Source:       hr3,
									},
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: hr4Groups[1],
										Source:       hr4,
									},
								},
							},
							{
								Path: "/fourth",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: hr4Groups[0],
										Source:       hr4,
									},
								},
							},
							{
								Path: "/third",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: hr3Groups[1],
										Source:       hr3,
									},
								},
							},
						},
					},
				},
				SSLServers: []VirtualServer{
					{
						Hostname: "foo.example.com",
						SSL: &SSL{
							CertificatePath: secretPath,
						},
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: httpsHR3Groups[0],
										Source:       httpsHR3,
									},
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: httpsHR4Groups[1],
										Source:       httpsHR4,
									},
								},
							},
							{
								Path: "/fourth",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										BackendGroup: httpsHR4Groups[0],
										Source:       httpsHR4,
									},
								},
							},
							{
								Path: "/third",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      1,
										BackendGroup: httpsHR3Groups[1],
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
					hr3Groups[0],
					hr3Groups[1],
					hr4Groups[0],
					hr4Groups[1],
					httpsHR3Groups[0],
					httpsHR3Groups[1],
					httpsHR4Groups[0],
					httpsHR4Groups[1],
				},
			},
			msg: "one http and one https listener with two routes with the same hostname with and without collisions",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source:   &v1beta1.GatewayClass{},
					Valid:    false,
					ErrorMsg: "error",
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "hr-1"}: routeHR1,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
				},
			},
			expConf: Configuration{},
			msg:     "invalid gatewayclass",
		},
		{
			graph: &graph{
				GatewayClass: nil,
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "hr-1"}: routeHR1,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-1"}: routeHR1,
				},
			},
			expConf: Configuration{},
			msg:     "missing gatewayclass",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: nil,
				Routes:  map[types.NamespacedName]*route{},
			},
			expConf: Configuration{},
			msg:     "missing gateway",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{},
					Valid:  true,
				},
				Gateway: &gateway{
					Source: &v1beta1.Gateway{},
					Listeners: map[string]*listener{
						"listener-80-1": {
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*route{
								{Namespace: "test", Name: "hr-5"}: routeHR5,
							},
							AcceptedHostnames: map[string]struct{}{
								"foo.example.com": {},
							},
						},
					},
				},
				Routes: map[types.NamespacedName]*route{
					{Namespace: "test", Name: "hr-5"}: routeHR5,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path: "/",
								MatchRules: []MatchRule{
									{
										MatchIdx:     0,
										RuleIdx:      0,
										Source:       hr5,
										BackendGroup: hr5BackendGroup,
										Filters: Filters{
											RequestRedirect: redirect.RequestRedirect,
										},
									},
								},
							},
						},
					},
				},
				SSLServers:    []VirtualServer{},
				BackendGroups: []BackendGroup{hr5BackendGroup},
			},
			msg: "one http listener with one route with filters",
		},
	}

	for _, test := range tests {
		result, warns := buildConfiguration(context.TODO(), test.graph, fakeResolver)

		sort.Slice(result.BackendGroups, func(i, j int) bool {
			return result.BackendGroups[i].GroupName() < result.BackendGroups[j].GroupName()
		})

		sort.Slice(result.Upstreams, func(i, j int) bool {
			return result.Upstreams[i].Name < result.Upstreams[j].Name
		})

		if diff := cmp.Diff(test.expConf, result); diff != "" {
			t.Errorf("buildConfiguration() %q mismatch for configuration (-want +got):\n%s", test.msg, diff)
		}

		if diff := cmp.Diff(test.expWarns, warns); diff != "" {
			t.Errorf("buildConfiguration() %q mismatch for warnings (-want +got):\n%s", test.msg, diff)
		}
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

	createBackendGroup := func(serviceNames ...string) BackendGroup {
		var backends []BackendRef
		for _, name := range serviceNames {
			backends = append(backends, BackendRef{
				Name: name,
				Svc:  &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name}},
			})
		}
		return BackendGroup{
			Backends: backends,
		}
	}

	hr1Group0 := createBackendGroup("foo", "bar")

	hr1Group1 := createBackendGroup("baz", "", "") // empty service names should be ignored

	hr2Group0 := createBackendGroup("foo", "baz") // shouldn't duplicate foo and baz upstream

	hr2Group1 := createBackendGroup("nil-endpoints")

	hr3Group0 := createBackendGroup("baz") // shouldn't duplicate baz upstream

	hr4Group0 := createBackendGroup("empty-endpoints", "")

	hr4Group1 := createBackendGroup("baz2")

	invalidGroup := createBackendGroup("invalid")

	routes := map[types.NamespacedName]*route{
		{Name: "hr1", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr1Group0, hr1Group1},
		},
		{Name: "hr2", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr2Group0, hr2Group1},
		},
		{Name: "hr3", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr3Group0},
		},
	}

	routes2 := map[types.NamespacedName]*route{
		{Name: "hr4", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr4Group0, hr4Group1},
		},
	}

	invalidRoutes := map[types.NamespacedName]*route{
		{Name: "invalid", Namespace: "test"}: {
			BackendGroups: []BackendGroup{invalidGroup},
		},
	}

	listeners := map[string]*listener{
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

	expUpstreams := map[string]Upstream{
		"bar": {
			Name:      "bar",
			Endpoints: barEndpoints,
		},
		"baz": {
			Name:      "baz",
			Endpoints: bazEndpoints,
		},
		"baz2": {
			Name:      "baz2",
			Endpoints: baz2Endpoints,
		},
		"empty-endpoints": {
			Name:      "empty-endpoints",
			Endpoints: []resolver.Endpoint{},
			ErrorMsg:  emptyEndpointsErrMsg,
		},
		"foo": {
			Name:      "foo",
			Endpoints: fooEndpoints,
		},
		"nil-endpoints": {
			Name:      "nil-endpoints",
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

	upstreams := buildUpstreamsMap(context.TODO(), listeners, fakeResolver)

	if diff := cmp.Diff(expUpstreams, upstreams); diff != "" {
		t.Errorf("buildUpstreamsMap() mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildBackendGroups(t *testing.T) {
	createBackendGroup := func(name string, ruleIdx int, backendNames ...string) BackendGroup {
		backends := make([]BackendRef, len(backendNames))
		for i, name := range backendNames {
			backends[i] = BackendRef{Name: name}
		}

		return BackendGroup{
			Source:   types.NamespacedName{Namespace: "test", Name: name},
			RuleIdx:  ruleIdx,
			Backends: backends,
		}
	}

	hr1Rule0 := createBackendGroup("hr1", 0, "foo", "bar")

	hr1Rule1 := createBackendGroup("hr1", 1, "foo")

	hr2Rule0 := createBackendGroup("hr2", 0, "foo", "bar")

	hr2Rule1 := createBackendGroup("hr2", 1, "foo")

	hr3Rule0 := createBackendGroup("hr3", 0, "foo", "bar")

	hr3Rule1 := createBackendGroup("hr3", 1, "foo")

	hrInvalid := createBackendGroup("hr-invalid", 0, "invalid")

	invalidRoutes := map[types.NamespacedName]*route{
		{Name: "invalid", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hrInvalid},
		},
	}

	routes := map[types.NamespacedName]*route{
		{Name: "hr1", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr1Rule0, hr1Rule1},
		},
		{Name: "hr2", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr2Rule0, hr2Rule1},
		},
	}

	routes2 := map[types.NamespacedName]*route{
		// this backend group is a dupe and should be ignored.
		{Name: "hr1", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr1Rule0, hr1Rule1},
		},
		{Name: "hr3", Namespace: "test"}: {
			BackendGroups: []BackendGroup{hr3Rule0, hr3Rule1},
		},
	}

	listeners := map[string]*listener{
		"invalid-listener": {
			Valid:  false,
			Routes: invalidRoutes, // routes on invalid listener should be ignored.
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

	expGroups := []BackendGroup{
		hr1Rule0,
		hr1Rule1,
		hr2Rule0,
		hr2Rule1,
		hr3Rule0,
		hr3Rule1,
	}

	result := buildBackendGroups(listeners)

	sort.Slice(result, func(i, j int) bool {
		return result[i].GroupName() < result[j].GroupName()
	})

	if diff := helpers.Diff(expGroups, result); diff != "" {
		t.Errorf("buildBackendGroups() mismatch: %+v", diff)
	}
}

func TestBuildWarnings(t *testing.T) {
	createBackendRefs := func(names ...string) []BackendRef {
		backends := make([]BackendRef, len(names))
		for idx, name := range names {
			backends[idx] = BackendRef{Name: name}
		}

		return backends
	}

	createBackendGroup := func(sourceName string, backends []BackendRef, errMsgs ...string) BackendGroup {
		return BackendGroup{
			Source:   types.NamespacedName{Namespace: "test", Name: sourceName},
			Backends: backends,
			Errors:   errMsgs,
		}
	}

	hr1BackendGroup0 := createBackendGroup(
		"hr1",
		createBackendRefs("foo"),
		"error1-1", "error1-2", "error1-3",
	)

	hr1BackendGroup1 := createBackendGroup(
		"hr1",
		createBackendRefs("bar"),
	)

	hr2BackendGroup0 := createBackendGroup(
		"hr2",
		createBackendRefs("foo", "bar"),
	)

	hr2BackendGroup1 := createBackendGroup(
		"hr2",
		createBackendRefs("resolve-error"),
		"error2",
	)

	hr3BackendGroup0 := createBackendGroup(
		"hr3",
		createBackendRefs(""), // empty backend name should be skipped
		"error3",
	)

	hr3BackendGroup1 := createBackendGroup(
		"hr3",
		createBackendRefs("dne"),
	)

	hrInvalidGroup := createBackendGroup(
		"hr-invalid",
		createBackendRefs("invalid"),
		"invalid",
	)

	hr1 := &v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr1", Namespace: "test"}}
	hr2 := &v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr2", Namespace: "test"}}
	hr3 := &v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr3", Namespace: "test"}}
	hrInvalid := &v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "hr-invalid", Namespace: "test"}}

	invalidRoutes := map[types.NamespacedName]*route{
		{Name: "invalid", Namespace: "test"}: {
			Source:        hrInvalid,
			BackendGroups: []BackendGroup{hrInvalidGroup},
		},
	}

	routes := map[types.NamespacedName]*route{
		{Name: "hr1", Namespace: "test"}: {
			Source:        hr1,
			BackendGroups: []BackendGroup{hr1BackendGroup0, hr1BackendGroup1},
		},
		{Name: "hr2", Namespace: "test"}: {
			Source:        hr2,
			BackendGroups: []BackendGroup{hr2BackendGroup0, hr2BackendGroup1},
		},
	}

	routes2 := map[types.NamespacedName]*route{
		{Name: "hr3", Namespace: "test"}: {
			Source:        hr3,
			BackendGroups: []BackendGroup{hr3BackendGroup0, hr3BackendGroup1},
		},
	}

	upstreamMap := map[string]Upstream{
		"foo":           {},
		"bar":           {},
		"resolve-error": {ErrorMsg: "resolve error"},
	}

	graph := &graph{
		Gateway: &gateway{
			Listeners: map[string]*listener{
				"invalid-listener": {
					Source: v1beta1.Listener{
						Name: "invalid",
					},
					Valid:  false,
					Routes: invalidRoutes,
				},
				"listener": {
					Source: v1beta1.Listener{
						Name: "valid",
					},
					Valid:  true,
					Routes: routes,
				},
				"listener2": {
					Source: v1beta1.Listener{
						Name: "valid2",
					},
					Valid:  true,
					Routes: routes2,
				},
			},
		},
	}

	expWarns := Warnings{
		hr1: []string{
			"invalid backend ref: error1-1",
			"invalid backend ref: error1-2",
			"invalid backend ref: error1-3",
		},
		hr2: []string{
			"invalid backend ref: error2",
			"cannot resolve backend ref: resolve error",
		},
		hr3: []string{
			"invalid backend ref: error3",
			"cannot resolve backend ref; internal error: upstream dne not found in map",
		},
		hrInvalid: []string{"cannot configure routes for listener invalid; listener is invalid"},
	}

	warns := buildWarnings(graph, upstreamMap)
	if diff := cmp.Diff(expWarns, warns); diff != "" {
		t.Errorf("buildWarnings() mismatch (-want +got):\n%s", diff)
	}
}

func TestUpstreamsMapToSlice(t *testing.T) {
	fooUpstream := Upstream{
		Name: "foo",
		Endpoints: []resolver.Endpoint{
			{Address: "10.0.0.0", Port: 80},
			{Address: "10.0.0.0", Port: 81},
		},
	}

	barUpstream := Upstream{
		Name:      "bar",
		ErrorMsg:  "error",
		Endpoints: nil,
	}

	bazUpstream := Upstream{
		Name: "baz",
		Endpoints: []resolver.Endpoint{
			{Address: "11.0.0.0", Port: 80},
		},
	}

	upstreamMap := map[string]Upstream{
		"foo": fooUpstream,
		"bar": barUpstream,
		"baz": bazUpstream,
	}

	expUpstreams := []Upstream{
		barUpstream,
		bazUpstream,
		fooUpstream,
	}

	upstreams := upstreamsMapToSlice(upstreamMap)

	sort.Slice(upstreams, func(i, j int) bool {
		return upstreams[i].Name < upstreams[j].Name
	})

	if diff := cmp.Diff(expUpstreams, upstreams); diff != "" {
		t.Errorf("upstreamMapToSlice() mismatch (-want +got):\n%s", diff)
	}
}
