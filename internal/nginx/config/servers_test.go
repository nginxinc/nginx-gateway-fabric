package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

func TestExecuteServers(t *testing.T) {
	conf := dataplane.Configuration{
		HTTPServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
			},
			{
				Hostname: "example.com",
			},
			{
				Hostname: "cafe.example.com",
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
			},
			{
				Hostname: "example.com",
				SSL: &dataplane.SSL{
					CertificatePath: "cert-path",
				},
			},
			{
				Hostname: "cafe.example.com",
				SSL: &dataplane.SSL{
					CertificatePath: "cert-path",
				},
			},
		},
	}

	expSubStrings := map[string]int{
		"listen 80 default_server;":      1,
		"listen 443 ssl;":                2,
		"listen 443 ssl default_server;": 1,
		"server_name example.com;":       2,
		"server_name cafe.example.com;":  2,
		"ssl_certificate cert-path;":     2,
		"ssl_certificate_key cert-path;": 2,
	}

	servers := string(executeServers(conf))
	for expSubStr, expCount := range expSubStrings {
		if expCount != strings.Count(servers, expSubStr) {
			t.Errorf(
				"executeServers() did not generate servers with substring %q %d times. Servers: %v",
				expSubStr,
				expCount,
				servers,
			)
		}
	}
}

func TestExecuteForDefaultServers(t *testing.T) {
	testcases := []struct {
		msg         string
		conf        dataplane.Configuration
		httpDefault bool
		sslDefault  bool
	}{
		{
			conf:        dataplane.Configuration{},
			httpDefault: false,
			sslDefault:  false,
			msg:         "no default servers",
		},
		{
			conf: dataplane.Configuration{
				HTTPServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
					},
				},
			},
			httpDefault: true,
			sslDefault:  false,
			msg:         "only HTTP default server",
		},
		{
			conf: dataplane.Configuration{
				SSLServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
					},
				},
			},
			httpDefault: false,
			sslDefault:  true,
			msg:         "only HTTPS default server",
		},
		{
			conf: dataplane.Configuration{
				HTTPServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
					},
				},
				SSLServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
					},
				},
			},
			httpDefault: true,
			sslDefault:  true,
			msg:         "both HTTP and HTTPS default servers",
		},
	}

	for _, tc := range testcases {
		cfg := string(executeServers(tc.conf))

		defaultSSLExists := strings.Contains(cfg, "listen 443 ssl default_server")
		defaultHTTPExists := strings.Contains(cfg, "listen 80 default_server")

		if tc.sslDefault && !defaultSSLExists {
			t.Errorf(
				"executeServers() did not generate a config with a default TLS termination server for test: %q",
				tc.msg,
			)
		}

		if !tc.sslDefault && defaultSSLExists {
			t.Errorf("executeServers() generated a config with a default TLS termination server for test: %q", tc.msg)
		}

		if tc.httpDefault && !defaultHTTPExists {
			t.Errorf("executeServers() did not generate a config with a default http server for test: %q", tc.msg)
		}

		if !tc.httpDefault && defaultHTTPExists {
			t.Errorf("executeServers() generated a config with a default http server for test: %q", tc.msg)
		}

		if len(cfg) == 0 {
			t.Errorf("executeServers() generated empty config for test: %q", tc.msg)
		}
	}
}

func TestCreateServers(t *testing.T) {
	const (
		certPath = "/etc/nginx/secrets/cert"
	)

	hr := &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route1",
		},
		Spec: v1beta1.HTTPRouteSpec{
			Hostnames: []v1beta1.Hostname{
				"cafe.example.com",
			},
			Rules: []v1beta1.HTTPRouteRule{
				{
					// matches with path and methods
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodPost),
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodPatch),
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer(
									"/", // should generate an "any" httpmatch since other matches exists for /
								),
								Type: helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
						},
					},
				},
				{
					// A match with all possible fields set
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/test"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodGet),
							Headers: []v1beta1.HTTPHeaderMatch{
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
									Name:  "Version",
									Value: "V1",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
									Name:  "test",
									Value: "foo",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
									Name:  "my-header",
									Value: "my-value",
								},
							},
							QueryParams: []v1beta1.HTTPQueryParamMatch{
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchExact),
									Name:  "GrEat", // query names and values should not be normalized to lowercase
									Value: "EXAMPLE",
								},
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchExact),
									Name:  "test",
									Value: "foo=bar",
								},
							},
						},
					},
				},
				{
					// A match with just path
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-only"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
						},
					},
				},
				{
					// A match with a redirect with implicit port
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/redirect-implicit-port"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
						},
					},
					// redirect is set in the corresponding state.MatchRule
				},
				{
					// A match with a redirect with explicit port
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/redirect-explicit-port"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
						},
					},
					// redirect is set in the corresponding state.MatchRule
				},
				{
					// A match with an invalid filter
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetPointer("/invalid-filter"),
								Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
							},
						},
					},
				},
				{
					// A match using type Exact
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetPointer("/exact"),
								Type:  helpers.GetPointer(v1beta1.PathMatchExact),
							},
						},
					},
				},
				{
					// A match using type Exact with method
					Matches: []v1beta1.HTTPRouteMatch{
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetPointer("/test"),
								Type:  helpers.GetPointer(v1beta1.PathMatchExact),
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodGet),
						},
					},
				},
			},
		},
	}

	hrNsName := types.NamespacedName{Namespace: hr.Namespace, Name: hr.Name}

	fooGroup := dataplane.BackendGroup{
		Source:  hrNsName,
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_foo_80",
				Valid:        true,
				Weight:       1,
			},
		},
	}

	// barGroup has two backends, which should generate a proxy pass with a variable.
	barGroup := dataplane.BackendGroup{
		Source:  hrNsName,
		RuleIdx: 1,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_bar_80",
				Valid:        true,
				Weight:       50,
			},
			{
				UpstreamName: "test_bar2_80",
				Valid:        true,
				Weight:       50,
			},
		},
	}

	// baz group has an invalid backend, which should generate a proxy pass to the invalid ref backend.
	bazGroup := dataplane.BackendGroup{
		Source:  hrNsName,
		RuleIdx: 2,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_baz_80",
				Valid:        false,
				Weight:       1,
			},
		},
	}

	filterGroup1 := dataplane.BackendGroup{Source: hrNsName, RuleIdx: 3}

	filterGroup2 := dataplane.BackendGroup{Source: hrNsName, RuleIdx: 4}

	invalidFilterGroup := dataplane.BackendGroup{Source: hrNsName, RuleIdx: 5}

	cafePathRules := []dataplane.PathRule{
		{
			Path:     "/",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      0,
					BackendGroup: fooGroup,
					Source:       hr,
				},
				{
					MatchIdx:     1,
					RuleIdx:      0,
					BackendGroup: fooGroup,
					Source:       hr,
				},
				{
					MatchIdx:     2,
					RuleIdx:      0,
					BackendGroup: fooGroup,
					Source:       hr,
				},
			},
		},
		{
			Path:     "/test",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      1,
					BackendGroup: barGroup,
					Source:       hr,
				},
			},
		},
		{
			Path:     "/path-only",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      2,
					BackendGroup: bazGroup,
					Source:       hr,
				},
			},
		},
		{
			Path:     "/redirect-implicit-port",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx: 0,
					RuleIdx:  3,
					Source:   hr,
					Filters: dataplane.Filters{
						RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
							Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
						},
					},
					BackendGroup: filterGroup1,
				},
			},
		},
		{
			Path:     "/redirect-explicit-port",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx: 0,
					RuleIdx:  4,
					Source:   hr,
					Filters: dataplane.Filters{
						RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
							Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("bar.example.com")),
							Port:     (*v1beta1.PortNumber)(helpers.GetInt32Pointer(8080)),
						},
					},
					BackendGroup: filterGroup2,
				},
			},
		},
		{
			Path:     "/invalid-filter",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx: 0,
					RuleIdx:  5,
					Source:   hr,
					Filters: dataplane.Filters{
						InvalidFilter: &dataplane.InvalidFilter{},
					},
					BackendGroup: invalidFilterGroup,
				},
			},
		},
		{
			Path:     "/exact",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      6,
					Source:       hr,
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/test",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      7,
					Source:       hr,
					BackendGroup: fooGroup,
				},
			},
		},
	}

	httpServers := []dataplane.VirtualServer{
		{
			IsDefault: true,
		},
		{
			Hostname:  "cafe.example.com",
			PathRules: cafePathRules,
		},
	}

	sslServers := []dataplane.VirtualServer{
		{
			IsDefault: true,
		},
		{
			Hostname:  "cafe.example.com",
			SSL:       &dataplane.SSL{CertificatePath: certPath},
			PathRules: cafePathRules,
		},
	}

	expectedMatchString := func(m []httpMatch) string {
		b, err := json.Marshal(m)
		if err != nil {
			t.Errorf("error marshaling test match: %v", err)
		}
		return string(b)
	}

	slashMatches := []httpMatch{
		{Method: v1beta1.HTTPMethodPost, RedirectPath: "/_prefix_route0"},
		{Method: v1beta1.HTTPMethodPatch, RedirectPath: "/_prefix_route1"},
		{Any: true, RedirectPath: "/_prefix_route2"},
	}
	testMatches := []httpMatch{
		{
			Method:       v1beta1.HTTPMethodGet,
			Headers:      []string{"Version:V1", "test:foo", "my-header:my-value"},
			QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
			RedirectPath: "/test_prefix_route0",
		},
	}
	exactMatches := []httpMatch{
		{
			Method:       v1beta1.HTTPMethodGet,
			RedirectPath: "/test_exact_route0",
		},
	}

	getExpectedLocations := func(isHTTPS bool) []http.Location {
		port := 80
		if isHTTPS {
			port = 443
		}

		return []http.Location{
			{
				Path:      "/_prefix_route0",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:      "/_prefix_route1",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:      "/_prefix_route2",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:         "/",
				HTTPMatchVar: expectedMatchString(slashMatches),
			},
			{
				Path:      "/test_prefix_route0",
				Internal:  true,
				ProxyPass: "http://$test__route1_rule1",
			},
			{
				Path:         "/test",
				HTTPMatchVar: expectedMatchString(testMatches),
			},
			{
				Path:      "/path-only",
				ProxyPass: "http://invalid-backend-ref",
			},
			{
				Path: "/redirect-implicit-port",
				Return: &http.Return{
					Code: 302,
					Body: fmt.Sprintf("$scheme://foo.example.com:%d$request_uri", port),
				},
			},
			{
				Path: "/redirect-explicit-port",
				Return: &http.Return{
					Code: 302,
					Body: "$scheme://bar.example.com:8080$request_uri",
				},
			},
			{
				Path: "/invalid-filter",
				Return: &http.Return{
					Code: http.StatusInternalServerError,
				},
			},
			{
				Path:      "/exact",
				ProxyPass: "http://test_foo_80",
				Exact:     true,
			},
			{
				Path:      "/test_exact_route0",
				ProxyPass: "http://test_foo_80",
				Internal:  true,
			},
			{
				Path:         "/test",
				HTTPMatchVar: expectedMatchString(exactMatches),
				Exact:        true,
			},
		}
	}

	expectedServers := []http.Server{
		{
			IsDefaultHTTP: true,
		},
		{
			ServerName: "cafe.example.com",
			Locations:  getExpectedLocations(false),
		},
		{
			IsDefaultSSL: true,
		},
		{
			ServerName: "cafe.example.com",
			SSL:        &http.SSL{Certificate: certPath, CertificateKey: certPath},
			Locations:  getExpectedLocations(true),
		},
	}

	g := NewGomegaWithT(t)

	result := createServers(httpServers, sslServers)
	g.Expect(helpers.Diff(expectedServers, result)).To(BeEmpty())
}

func TestCreateLocationsRootPath(t *testing.T) {
	g := NewGomegaWithT(t)

	createRoute := func(rootPath bool) *v1beta1.HTTPRoute {
		route := &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "route1",
			},
			Spec: v1beta1.HTTPRouteSpec{
				Hostnames: []v1beta1.Hostname{
					"cafe.example.com",
				},
				Rules: []v1beta1.HTTPRouteRule{
					{
						Matches: []v1beta1.HTTPRouteMatch{
							{
								Path: &v1beta1.HTTPPathMatch{
									Value: helpers.GetStringPointer("/path-1"),
									Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
								},
							},
							{
								Path: &v1beta1.HTTPPathMatch{
									Value: helpers.GetStringPointer("/path-2"),
									Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
								},
							},
						},
					},
				},
			},
		}

		if rootPath {
			route.Spec.Rules[0].Matches = append(route.Spec.Rules[0].Matches, v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Value: helpers.GetStringPointer("/"),
					Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
				},
			})
		}

		return route
	}

	hrWithRootPathRule := createRoute(true)

	hrWithoutRootPathRule := createRoute(false)

	hrNsName := types.NamespacedName{Namespace: "test", Name: "route1"}

	fooGroup := dataplane.BackendGroup{
		Source:  hrNsName,
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_foo_80",
				Valid:        true,
				Weight:       1,
			},
		},
	}

	getPathRules := func(source *v1beta1.HTTPRoute, rootPath bool) []dataplane.PathRule {
		rules := []dataplane.PathRule{
			{
				Path: "/path-1",
				MatchRules: []dataplane.MatchRule{
					{
						Source:       source,
						BackendGroup: fooGroup,
						MatchIdx:     0,
						RuleIdx:      0,
					},
				},
			},
			{
				Path: "/path-2",
				MatchRules: []dataplane.MatchRule{
					{
						Source:       source,
						BackendGroup: fooGroup,
						MatchIdx:     1,
						RuleIdx:      0,
					},
				},
			},
		}

		if rootPath {
			rules = append(rules, dataplane.PathRule{
				Path: "/",
				MatchRules: []dataplane.MatchRule{
					{
						Source:       source,
						BackendGroup: fooGroup,
						MatchIdx:     2,
						RuleIdx:      0,
					},
				},
			})
		}

		return rules
	}

	tests := []struct {
		name         string
		pathRules    []dataplane.PathRule
		expLocations []http.Location
	}{
		{
			name:      "path rules with no root path should generate a default 404 root location",
			pathRules: getPathRules(hrWithoutRootPathRule, false),
			expLocations: []http.Location{
				{
					Path:      "/path-1",
					ProxyPass: "http://test_foo_80",
				},
				{
					Path:      "/path-2",
					ProxyPass: "http://test_foo_80",
				},
				{
					Path: "/",
					Return: &http.Return{
						Code: http.StatusNotFound,
					},
				},
			},
		},
		{
			name:      "path rules with a root path should not generate a default 404 root path",
			pathRules: getPathRules(hrWithRootPathRule, true),
			expLocations: []http.Location{
				{
					Path:      "/path-1",
					ProxyPass: "http://test_foo_80",
				},
				{
					Path:      "/path-2",
					ProxyPass: "http://test_foo_80",
				},
				{
					Path:      "/",
					ProxyPass: "http://test_foo_80",
				},
			},
		},
		{
			name:      "nil path rules should generate a default 404 root path",
			pathRules: nil,
			expLocations: []http.Location{
				{
					Path: "/",
					Return: &http.Return{
						Code: http.StatusNotFound,
					},
				},
			},
		},
	}

	for _, test := range tests {
		locs := createLocations(test.pathRules, 80)
		g.Expect(locs).To(Equal(test.expLocations), fmt.Sprintf("test case: %s", test.name))
	}
}

func TestCreateReturnValForRedirectFilter(t *testing.T) {
	const listenerPort = 123

	tests := []struct {
		filter   *v1beta1.HTTPRequestRedirectFilter
		expected *http.Return
		msg      string
	}{
		{
			filter:   nil,
			expected: nil,
			msg:      "filter is nil",
		},
		{
			filter: &v1beta1.HTTPRequestRedirectFilter{},
			expected: &http.Return{
				Code: http.StatusFound,
				Body: "$scheme://$host:123$request_uri",
			},
			msg: "all fields are empty",
		},
		{
			filter: &v1beta1.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetStringPointer("https"),
				Hostname:   (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
				Port:       (*v1beta1.PortNumber)(helpers.GetInt32Pointer(2022)),
				StatusCode: helpers.GetIntPointer(101),
			},
			expected: &http.Return{
				Code: 101,
				Body: "https://foo.example.com:2022$request_uri",
			},
			msg: "all fields are set",
		},
	}

	for _, test := range tests {
		result := createReturnValForRedirectFilter(test.filter, listenerPort)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("createReturnValForRedirectFilter() mismatch %q (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestCreateHTTPMatch(t *testing.T) {
	testPath := "/internal_loc"

	testPathMatch := v1beta1.HTTPPathMatch{Value: helpers.GetStringPointer("/")}
	testMethodMatch := helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodPut)
	testHeaderMatches := []v1beta1.HTTPHeaderMatch{
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
			Name:  "header-1",
			Value: "val-1",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
			Name:  "header-2",
			Value: "val-2",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch headers.
			Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchRegularExpression),
			Name:  "ignore-this-header",
			Value: "val",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
			Name:  "header-3",
			Value: "val-3",
		},
	}

	testDuplicateHeaders := make([]v1beta1.HTTPHeaderMatch, 0, 5)
	duplicateHeaderMatch := v1beta1.HTTPHeaderMatch{
		Type:  helpers.GetHeaderMatchTypePointer(v1beta1.HeaderMatchExact),
		Name:  "HEADER-2", // header names are case-insensitive
		Value: "val-2",
	}
	testDuplicateHeaders = append(testDuplicateHeaders, testHeaderMatches...)
	testDuplicateHeaders = append(testDuplicateHeaders, duplicateHeaderMatch)

	testQueryParamMatches := []v1beta1.HTTPQueryParamMatch{
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchExact),
			Name:  "arg1",
			Value: "val1",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchExact),
			Name:  "arg2",
			Value: "val2=another-val",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch args
			Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchRegularExpression),
			Name:  "ignore-this-arg",
			Value: "val",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1beta1.QueryParamMatchExact),
			Name:  "arg3",
			Value: "==val3",
		},
	}

	expectedHeaders := []string{"header-1:val-1", "header-2:val-2", "header-3:val-3"}
	expectedArgs := []string{"arg1=val1", "arg2=val2=another-val", "arg3===val3"}

	tests := []struct {
		match    v1beta1.HTTPRouteMatch
		msg      string
		expected httpMatch
	}{
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &testPathMatch,
			},
			expected: httpMatch{
				Any:          true,
				RedirectPath: testPath,
			},
			msg: "path only match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path:   &testPathMatch, // A path match with a method should not set the Any field to true
				Method: testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				RedirectPath: testPath,
			},
			msg: "method only match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				RedirectPath: testPath,
				Headers:      expectedHeaders,
			},
			msg: "headers only match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "query params only match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Method:      testMethodMatch,
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method and query params match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Method:  testMethodMatch,
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "method and headers match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
				Headers:     testHeaderMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "query params and headers match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Headers:     testHeaderMatches,
				QueryParams: testQueryParamMatches,
				Method:      testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method, headers, and query params match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Headers: testDuplicateHeaders,
			},
			expected: httpMatch{
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "duplicate header names",
		},
	}
	for _, tc := range tests {
		result := createHTTPMatch(tc.match, testPath)
		if diff := helpers.Diff(result, tc.expected); diff != "" {
			t.Errorf("createHTTPMatch() returned incorrect httpMatch for test case: %q, diff: %+v", tc.msg, diff)
		}
	}
}

func TestCreateQueryParamKeyValString(t *testing.T) {
	expected := "key=value"

	result := createQueryParamKeyValString(
		v1beta1.HTTPQueryParamMatch{
			Name:  "key",
			Value: "value",
		},
	)
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}

	expected = "KeY=vaLUe=="

	result = createQueryParamKeyValString(
		v1beta1.HTTPQueryParamMatch{
			Name:  "KeY",
			Value: "vaLUe==",
		},
	)
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestCreateHeaderKeyValString(t *testing.T) {
	expected := "kEy:vALUe"

	result := createHeaderKeyValString(
		v1beta1.HTTPHeaderMatch{
			Name:  "kEy",
			Value: "vALUe",
		},
	)

	if result != expected {
		t.Errorf("createHeaderKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestIsPathOnlyMatch(t *testing.T) {
	tests := []struct {
		match    v1beta1.HTTPRouteMatch
		msg      string
		expected bool
	}{
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
			},
			expected: true,
			msg:      "path only match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodGet),
			},
			expected: false,
			msg:      "method defined in match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Name:  "header",
						Value: "val",
					},
				},
			},
			expected: false,
			msg:      "headers defined in match",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Name:  "arg",
						Value: "val",
					},
				},
			},
			expected: false,
			msg:      "query params defined in match",
		},
	}

	for _, tc := range tests {
		result := isPathOnlyMatch(tc.match)

		if result != tc.expected {
			t.Errorf("isPathOnlyMatch() returned %t but expected %t for test case %q", result, tc.expected, tc.msg)
		}
	}
}

func TestCreateProxyPass(t *testing.T) {
	expected := "http://10.0.0.1:80"

	result := createProxyPass("10.0.0.1:80")
	if result != expected {
		t.Errorf("createProxyPass() returned %s but expected %s", result, expected)
	}
}

func TestCreateProxyPassForVar(t *testing.T) {
	expected := "http://$my_variable"

	result := createProxyPassForVar("my-variable")
	if result != expected {
		t.Errorf("createProxyPassForVar() returned %s but expected %s", result, expected)
	}
}

func TestCreateMatchLocation(t *testing.T) {
	expected := http.Location{
		Path:     "/path",
		Internal: true,
	}

	result := createMatchLocation("/path")
	if result != expected {
		t.Errorf("createMatchLocation() returned %v but expected %v", result, expected)
	}
}

func TestCreatePathForMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		expected string
		pathType dataplane.PathType
		panic    bool
	}{
		{
			expected: "/path_prefix_route1",
			pathType: dataplane.PathTypePrefix,
		},
		{
			expected: "/path_exact_route1",
			pathType: dataplane.PathTypeExact,
		},
	}

	for _, tc := range tests {
		result := createPathForMatch("/path", tc.pathType, 1)
		g.Expect(result).To(Equal(tc.expected))
	}
}
