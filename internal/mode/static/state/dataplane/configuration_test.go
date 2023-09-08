package dataplane

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/resolver"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/resolver/resolverfakes"
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
							Value: helpers.GetPointer(p.path),
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
							Namespace:   (*v1beta1.Namespace)(helpers.GetPointer("test")),
							Name:        "gateway",
							SectionName: (*v1beta1.SectionName)(helpers.GetPointer(listenerName)),
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

	fooSvc := &apiv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test"}}

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
		listenerName string,
		paths []pathAndType,
	) *graph.Route {
		hostnames := make([]string, 0, len(source.Spec.Hostnames))
		for _, h := range source.Spec.Hostnames {
			hostnames = append(hostnames, string(h))
		}
		r := &graph.Route{
			Source: source,
			Rules:  createRules(source, paths),
			ParentRefs: []graph.ParentRef{
				{
					Attachment: &graph.ParentRefAttachmentStatus{
						AcceptedHostnames: map[string][]string{
							listenerName: hostnames,
						},
					},
				},
			},
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
		route := createInternalRoute(hr, listenerName, paths)
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
			Hostname: (*v1beta1.PreciseHostname)(helpers.GetPointer("foo.example.com")),
		},
	}
	addFilters(hr5, []v1beta1.HTTPRouteFilter{redirect})
	expRedirect := HTTPRequestRedirectFilter{
		Hostname: helpers.GetPointer("foo.example.com"),
	}

	hr6, expHR6Groups, routeHR6 := createTestResources(
		"hr-6",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: invalidMatchesPath, pathType: prefix},
	)

	hr7, expHR7Groups, routeHR7 := createTestResources(
		"hr-7",
		"foo.example.com",
		"listener-80-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: "/valid", pathType: v1beta1.PathMatchExact},
	)

	hr8, expHR8Groups, routeHR8 := createTestResources(
		"hr-8",
		"foo.example.com", // same as hr3
		"listener-8080",
		pathAndType{path: "/", pathType: prefix},
		pathAndType{path: "/third", pathType: prefix},
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
	// add extra attachment for this route for duplicate listener test
	httpsRouteHR5.ParentRefs[0].Attachment.AcceptedHostnames["listener-443-1"] = []string{"example.com"}

	httpsHR6, expHTTPSHR6Groups, httpsRouteHR6 := createTestResources(
		"https-hr-6",
		"foo.example.com",
		"listener-443-1",
		pathAndType{path: "/valid", pathType: prefix}, pathAndType{path: invalidMatchesPath, pathType: prefix},
	)

	httpsHR7, expHTTPSHR7Groups, httpsRouteHR7 := createTestResources(
		"https-hr-7",
		"foo.example.com", // same as httpsHR3
		"listener-8443",
		pathAndType{path: "/", pathType: prefix}, pathAndType{path: "/third", pathType: prefix},
	)

	secret1NsName := types.NamespacedName{Namespace: "test", Name: "secret-1"}
	secret1 := &graph.Secret{
		Source: &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret1NsName.Name,
				Namespace: secret1NsName.Namespace,
			},
			Data: map[string][]byte{
				apiv1.TLSCertKey:       []byte("cert-1"),
				apiv1.TLSPrivateKeyKey: []byte("privateKey-1"),
			},
		},
	}

	secret2NsName := types.NamespacedName{Namespace: "test", Name: "secret-2"}
	secret2 := &graph.Secret{
		Source: &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret2NsName.Name,
				Namespace: secret2NsName.Namespace,
			},
			Data: map[string][]byte{
				apiv1.TLSCertKey:       []byte("cert-2"),
				apiv1.TLSPrivateKeyKey: []byte("privateKey-2"),
			},
		},
	}

	listener80 := v1beta1.Listener{
		Name:     "listener-80-1",
		Hostname: nil,
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}

	listener8080 := v1beta1.Listener{
		Name:     "listener-8080",
		Hostname: nil,
		Port:     8080,
		Protocol: v1beta1.HTTPProtocolType,
	}

	listener443 := v1beta1.Listener{
		Name:     "listener-443-1",
		Hostname: nil,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
					Namespace: helpers.GetPointer(v1beta1.Namespace(secret1NsName.Namespace)),
					Name:      v1beta1.ObjectName(secret1NsName.Name),
				},
			},
		},
	}

	listener8443 := v1beta1.Listener{
		Name:     "listener-8443",
		Hostname: nil,
		Port:     8443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
					Namespace: helpers.GetPointer(v1beta1.Namespace(secret2NsName.Namespace)),
					Name:      v1beta1.ObjectName(secret2NsName.Name),
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
			Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
					Namespace: helpers.GetPointer(v1beta1.Namespace(secret2NsName.Namespace)),
					Name:      v1beta1.ObjectName(secret2NsName.Name),
				},
			},
		},
	}

	invalidListener := v1beta1.Listener{
		Name:     "invalid-listener",
		Hostname: nil,
		Port:     443,
		Protocol: v1beta1.HTTPSProtocolType,
		TLS: &v1beta1.GatewayTLSConfig{
			// Mode is missing, that's why invalid
			CertificateRefs: []v1beta1.SecretObjectReference{
				{
					Kind:      helpers.GetPointer[v1beta1.Kind]("Secret"),
					Namespace: helpers.GetPointer(v1beta1.Namespace(secret1NsName.Namespace)),
					Name:      v1beta1.ObjectName(secret1NsName.Name),
				},
			},
		},
	}

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
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{},
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
							Source: listener80,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{},
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
				},
				SSLServers:  []VirtualServer{},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{},
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
							Source:         listener443, // nil hostname
							Valid:          true,
							Routes:         map[types.NamespacedName]*graph.Route{},
							ResolvedSecret: &secret1NsName,
						},
						"listener-443-with-hostname": {
							Source:         listener443WithHostname, // non-nil hostname
							Valid:          true,
							Routes:         map[types.NamespacedName]*graph.Route{},
							ResolvedSecret: &secret2NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
					secret2NsName: secret2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						Hostname: string(hostname),
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-2"},
						Port:     443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
					},
				},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
					"ssl_keypair_test_secret-2": {
						Cert: []byte("cert-2"),
						Key:  []byte("privateKey-2"),
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
							Source:         invalidListener,
							Valid:          false,
							ResolvedSecret: &secret1NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers:  []VirtualServer{},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{},
			},
			msg: "invalid https listener with resolved secret",
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
						Port:      80,
					},
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR2Groups[0],
										Source:       &hr2.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR1Groups[0],
										Source:       &hr1.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHR1Groups[0], expHR2Groups[0]},
				SSLKeyPairs:   map[SSLKeyPairID]SSLKeyPair{},
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
							Source: listener443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
								{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
							},
							ResolvedSecret: &secret1NsName,
						},
						"listener-443-with-hostname": {
							Source: listener443WithHostname,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							ResolvedSecret: &secret2NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-1"}: httpsRouteHR1,
					{Namespace: "test", Name: "https-hr-2"}: httpsRouteHR2,
					{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
					secret2NsName: secret2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						Hostname: "bar.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR2Groups[0],
										Source:       &httpsHR2.ObjectMeta,
									},
								},
							},
						},
						SSL:  &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port: 443,
					},
					{
						Hostname: "example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       &httpsHR5.ObjectMeta,
									},
								},
							},
						},
						SSL:  &SSL{KeyPairID: "ssl_keypair_test_secret-2"},
						Port: 443,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR1Groups[0],
										Source:       &httpsHR1.ObjectMeta,
									},
								},
							},
						},
						SSL:  &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port: 443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
					},
				},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHTTPSHR1Groups[0], expHTTPSHR2Groups[0], expHTTPSHR5Groups[0]},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
					"ssl_keypair_test_secret-2": {
						Cert: []byte("cert-2"),
						Key:  []byte("privateKey-2"),
					},
				},
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
						},
						"listener-443-1": {
							Source: listener443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
								{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
							},
							ResolvedSecret: &secret1NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-3"}:       routeHR3,
					{Namespace: "test", Name: "hr-4"}:       routeHR4,
					{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
					{Namespace: "test", Name: "https-hr-4"}: httpsRouteHR4,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR3Groups[0],
										Source:       &hr3.ObjectMeta,
									},
									{
										BackendGroup: expHR4Groups[1],
										Source:       &hr4.ObjectMeta,
									},
								},
							},
							{
								Path:     "/fourth",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR4Groups[0],
										Source:       &hr4.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR3Groups[1],
										Source:       &hr3.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
				},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						Hostname: "foo.example.com",
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR3Groups[0],
										Source:       &httpsHR3.ObjectMeta,
									},
									{
										BackendGroup: expHTTPSHR4Groups[1],
										Source:       &httpsHR4.ObjectMeta,
									},
								},
							},
							{
								Path:     "/fourth",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR4Groups[0],
										Source:       &httpsHR4.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR3Groups[1],
										Source:       &httpsHR3.ObjectMeta,
									},
								},
							},
						},
						Port: 443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
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
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
				},
			},
			msg: "one http and one https listener with two routes with the same hostname with and without collisions",
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
							},
						},
						"listener-8080": {
							Source: listener8080,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "hr-8"}: routeHR8,
							},
						},
						"listener-443-1": {
							Source: listener443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
							},
							ResolvedSecret: &secret1NsName,
						},
						"listener-8443": {
							Source: listener8443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-7"}: httpsRouteHR7,
							},
							ResolvedSecret: &secret1NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-3"}:       routeHR3,
					{Namespace: "test", Name: "hr-8"}:       routeHR8,
					{Namespace: "test", Name: "https-hr-3"}: httpsRouteHR3,
					{Namespace: "test", Name: "https-hr-7"}: httpsRouteHR7,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR3Groups[0],
										Source:       &hr3.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR3Groups[1],
										Source:       &hr3.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
					{
						IsDefault: true,
						Port:      8080,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR8Groups[0],
										Source:       &hr8.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR8Groups[1],
										Source:       &hr8.ObjectMeta,
									},
								},
							},
						},
						Port: 8080,
					},
				},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						Hostname: "foo.example.com",
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR3Groups[0],
										Source:       &httpsHR3.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR3Groups[1],
										Source:       &httpsHR3.ObjectMeta,
									},
								},
							},
						},
						Port: 443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
					},
					{
						IsDefault: true,
						Port:      8443,
					},
					{
						Hostname: "foo.example.com",
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR7Groups[0],
										Source:       &httpsHR7.ObjectMeta,
									},
								},
							},
							{
								Path:     "/third",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR7Groups[1],
										Source:       &httpsHR7.ObjectMeta,
									},
								},
							},
						},
						Port: 8443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     8443,
					},
				},
				Upstreams: []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{
					expHR3Groups[0],
					expHR3Groups[1],
					expHR8Groups[0],
					expHR8Groups[1],
					expHTTPSHR3Groups[0],
					expHTTPSHR3Groups[1],
					expHTTPSHR7Groups[0],
					expHTTPSHR7Groups[1],
				},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
				},
			},

			msg: "multiple http and https listener; different ports",
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
						Port:      80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										Source:       &hr5.ObjectMeta,
										BackendGroup: expHR5Groups[0],
										Filters: HTTPFilters{
											RequestRedirect: &expRedirect,
										},
									},
								},
							},
							{
								Path:     invalidFiltersPath,
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										Source:       &hr5.ObjectMeta,
										BackendGroup: expHR5Groups[1],
										Filters: HTTPFilters{
											InvalidFilter: &InvalidHTTPFilter{},
										},
									},
								},
							},
						},
						Port: 80,
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHR5Groups[0], expHR5Groups[1]},
				SSLKeyPairs:   map[SSLKeyPairID]SSLKeyPair{},
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
						},
						"listener-443-1": {
							Source: listener443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-6"}: httpsRouteHR6,
							},
							ResolvedSecret: &secret1NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "hr-6"}:       routeHR6,
					{Namespace: "test", Name: "https-hr-6"}: httpsRouteHR6,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR6Groups[0],
										Source:       &hr6.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
				},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						Hostname: "foo.example.com",
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHTTPSHR6Groups[0],
										Source:       &httpsHR6.ObjectMeta,
									},
								},
							},
						},
						Port: 443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
					},
				},
				Upstreams: []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{
					expHR6Groups[0],
					expHTTPSHR6Groups[0],
				},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
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
						Port:      80,
					},
					{
						Hostname: "foo.example.com",
						PathRules: []PathRule{
							{
								Path:     "/valid",
								PathType: PathTypeExact,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR7Groups[1],
										Source:       &hr7.ObjectMeta,
									},
								},
							},
							{
								Path:     "/valid",
								PathType: PathTypePrefix,
								MatchRules: []MatchRule{
									{
										BackendGroup: expHR7Groups[0],
										Source:       &hr7.ObjectMeta,
									},
								},
							},
						},
						Port: 80,
					},
				},
				SSLServers:    []VirtualServer{},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHR7Groups[0], expHR7Groups[1]},
				SSLKeyPairs:   map[SSLKeyPairID]SSLKeyPair{},
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
							Source: listener443WithHostname,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							ResolvedSecret: &secret2NsName,
						},
						"listener-443-1": {
							Source: listener443,
							Valid:  true,
							Routes: map[types.NamespacedName]*graph.Route{
								{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
							},
							ResolvedSecret: &secret1NsName,
						},
					},
				},
				Routes: map[types.NamespacedName]*graph.Route{
					{Namespace: "test", Name: "https-hr-5"}: httpsRouteHR5,
				},
				ReferencedSecrets: map[types.NamespacedName]*graph.Secret{
					secret1NsName: secret1,
					secret2NsName: secret2,
				},
			},
			expConf: Configuration{
				HTTPServers: []VirtualServer{},
				SSLServers: []VirtualServer{
					{
						IsDefault: true,
						Port:      443,
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
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       &httpsHR5.ObjectMeta,
									},
									{
										BackendGroup: expHTTPSHR5Groups[0],
										Source:       &httpsHR5.ObjectMeta,
									},
								},
							},
						},
						SSL:  &SSL{KeyPairID: "ssl_keypair_test_secret-2"},
						Port: 443,
					},
					{
						Hostname: wildcardHostname,
						SSL:      &SSL{KeyPairID: "ssl_keypair_test_secret-1"},
						Port:     443,
					},
				},
				Upstreams:     []Upstream{fooUpstream},
				BackendGroups: []BackendGroup{expHTTPSHR5Groups[0]},
				SSLKeyPairs: map[SSLKeyPairID]SSLKeyPair{
					"ssl_keypair_test_secret-1": {
						Cert: []byte("cert-1"),
						Key:  []byte("privateKey-1"),
					},
					"ssl_keypair_test_secret-2": {
						Cert: []byte("cert-2"),
						Key:  []byte("privateKey-2"),
					},
				},
			},
			msg: "two https listeners with different hostnames but same route; chooses listener with more specific hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := BuildConfiguration(context.TODO(), test.graph, fakeResolver, 1)

			g.Expect(result.BackendGroups).To(ConsistOf(test.expConf.BackendGroups))
			g.Expect(result.Upstreams).To(ConsistOf(test.expConf.Upstreams))
			g.Expect(result.HTTPServers).To(ConsistOf(test.expConf.HTTPServers))
			g.Expect(result.SSLServers).To(ConsistOf(test.expConf.SSLServers))
			g.Expect(result.SSLKeyPairs).To(Equal(test.expConf.SSLKeyPairs))
			g.Expect(result.Version).To(Equal(1))
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
			path:     &v1beta1.HTTPPathMatch{Value: helpers.GetPointer("/abc")},
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
			path:     &v1beta1.HTTPPathMatch{Value: helpers.GetPointer("")},
			expected: "/",
			msg:      "empty value",
		},
	}

	for _, test := range tests {
		g := NewGomegaWithT(t)
		result := getPath(test.path)
		g.Expect(result).To(Equal(test.expected))
	}
}

func TestCreateFilters(t *testing.T) {
	redirect1 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: helpers.GetPointer[v1beta1.PreciseHostname]("foo.example.com"),
		},
	}
	redirect2 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: helpers.GetPointer[v1beta1.PreciseHostname]("bar.example.com"),
		},
	}
	requestHeaderModifiers1 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
		RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
			Set: []v1beta1.HTTPHeader{
				{
					Name:  "MyBespokeHeader",
					Value: "my-value",
				},
			},
		},
	}
	requestHeaderModifiers2 := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
		RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
			Add: []v1beta1.HTTPHeader{
				{
					Name:  "Content-Accepted",
					Value: "gzip",
				},
			},
		},
	}

	expectedRedirect1 := HTTPRequestRedirectFilter{
		Hostname: helpers.GetPointer("foo.example.com"),
	}
	expectedHeaderModifier1 := HTTPHeaderFilter{
		Set: []HTTPHeader{
			{
				Name:  "MyBespokeHeader",
				Value: "my-value",
			},
		},
	}

	tests := []struct {
		expected HTTPFilters
		msg      string
		filters  []v1beta1.HTTPRouteFilter
	}{
		{
			filters:  []v1beta1.HTTPRouteFilter{},
			expected: HTTPFilters{},
			msg:      "no filters",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
			},
			expected: HTTPFilters{
				RequestRedirect: &expectedRedirect1,
			},
			msg: "one filter",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
				redirect2,
			},
			expected: HTTPFilters{
				RequestRedirect: &expectedRedirect1,
			},
			msg: "two filters, first wins",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
				redirect2,
				requestHeaderModifiers1,
			},
			expected: HTTPFilters{
				RequestRedirect:        &expectedRedirect1,
				RequestHeaderModifiers: &expectedHeaderModifier1,
			},
			msg: "two redirect filters, one request header modifier, first redirect wins",
		},
		{
			filters: []v1beta1.HTTPRouteFilter{
				redirect1,
				redirect2,
				requestHeaderModifiers1,
				requestHeaderModifiers2,
			},
			expected: HTTPFilters{
				RequestRedirect:        &expectedRedirect1,
				RequestHeaderModifiers: &expectedHeaderModifier1,
			},
			msg: "two redirect filters, two request header modifier, first value for each wins",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := createHTTPFilters(test.filters)

			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
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
		g := NewGomegaWithT(t)
		result := getListenerHostname(test.hostname)
		g.Expect(result).To(Equal(test.expected))
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
				Svc:   &apiv1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name}},
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
	fakeResolver.ResolveCalls(func(ctx context.Context, svc *apiv1.Service, port int32) ([]resolver.Endpoint, error) {
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

func TestHostnameMoreSpecific(t *testing.T) {
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
			host1:     helpers.GetPointer(v1beta1.Hostname("")),
			host2:     helpers.GetPointer(v1beta1.Hostname("example.com")),
			host1Wins: false,
			msg:       "host2 has value; host1 empty",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("foo.example.com")),
			host2:     helpers.GetPointer(v1beta1.Hostname("*.example.com")),
			host1Wins: true,
			msg:       "host1 more specific than host2",
		},
		{
			host1:     helpers.GetPointer(v1beta1.Hostname("*.example.com")),
			host2:     helpers.GetPointer(v1beta1.Hostname("foo.example.com")),
			host1Wins: false,
			msg:       "host2 more specific than host1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)

			g.Expect(listenerHostnameMoreSpecific(tc.host1, tc.host2)).To(Equal(tc.host1Wins))
		})
	}
}
