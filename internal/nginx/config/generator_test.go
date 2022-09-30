package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/resolver"
)

func TestGenerate(t *testing.T) {
	// Note: this test only verifies that Generate() returns a byte array with http upstreams and server blocks.
	// It does not test the correctness of those blocks. That functionality is covered by other tests in this file.
	generator := NewGeneratorImpl()

	conf := state.Configuration{
		HTTPServers: []state.VirtualServer{
			{
				Hostname: "example.com",
			},
		},
		SSLServers: []state.VirtualServer{
			{
				Hostname: "example.com",
			},
		},
		Upstreams: []state.Upstream{
			{
				Name:      "up",
				Endpoints: nil,
			},
		},
	}

	cfg := string(generator.Generate(conf))

	if !strings.Contains(cfg, "listen 80") {
		t.Errorf("Generate() did not generate a config with an HTTP server; config: %s", cfg)
	}

	if !strings.Contains(cfg, "listen 443") {
		t.Errorf("Generate() did not generate a config with an SSL server; config: %s", cfg)
	}

	if !strings.Contains(cfg, "upstream") {
		t.Errorf("Generate() did not generate a config with an upstream block; config: %s", cfg)
	}
}

func TestGenerateForDefaultServers(t *testing.T) {
	generator := NewGeneratorImpl()

	testcases := []struct {
		conf        state.Configuration
		httpDefault bool
		sslDefault  bool
		msg         string
	}{
		{
			conf:        state.Configuration{},
			httpDefault: false,
			sslDefault:  false,
			msg:         "no servers",
		},
		{
			conf: state.Configuration{
				HTTPServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: true,
			sslDefault:  false,
			msg:         "only HTTP servers",
		},
		{
			conf: state.Configuration{
				SSLServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: false,
			sslDefault:  true,
			msg:         "only HTTPS servers",
		},
		{
			conf: state.Configuration{
				HTTPServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
				SSLServers: []state.VirtualServer{
					{
						Hostname: "example.com",
					},
				},
			},
			httpDefault: true,
			sslDefault:  true,
			msg:         "both HTTP and HTTPS servers",
		},
	}

	for _, tc := range testcases {
		cfg := generator.Generate(tc.conf)

		defaultSSLExists := strings.Contains(string(cfg), "listen 443 ssl default_server")
		defaultHTTPExists := strings.Contains(string(cfg), "listen 80 default_server")

		if tc.sslDefault && !defaultSSLExists {
			t.Errorf("Generate() did not generate a config with a default TLS termination server for test: %q", tc.msg)
		}

		if !tc.sslDefault && defaultSSLExists {
			t.Errorf("Generate() generated a config with a default TLS termination server for test: %q", tc.msg)
		}

		if tc.httpDefault && !defaultHTTPExists {
			t.Errorf("Generate() did not generate a config with a default http server for test: %q", tc.msg)
		}

		if !tc.httpDefault && defaultHTTPExists {
			t.Errorf("Generate() generated a config with a default http server for test: %q", tc.msg)
		}

		if len(cfg) == 0 {
			t.Errorf("Generate() generated empty config for test: %q", tc.msg)
		}
	}
}

func TestGenerateHTTPServers(t *testing.T) {
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
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodPost),
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
							},
							Method: helpers.GetHTTPMethodPointer(v1beta1.HTTPMethodPatch),
						},
						{
							Path: &v1beta1.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"), // should generate an "any" httpmatch since other matches exists for /
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
							},
						},
					},
					// redirect is set in the corresponding state.MatchRule
				},
			},
		},
	}

	cafePathRules := []state.PathRule{
		{
			Path: "/",
			MatchRules: []state.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      0,
					UpstreamName: "test_foo_80",
					Source:       hr,
				},
				{
					MatchIdx:     1,
					RuleIdx:      0,
					UpstreamName: "test_foo_80",
					Source:       hr,
				},
				{
					MatchIdx:     2,
					RuleIdx:      0,
					UpstreamName: "test_foo_80",
					Source:       hr,
				},
			},
		},
		{
			Path: "/test",
			MatchRules: []state.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      1,
					UpstreamName: "test_bar_80",
					Source:       hr,
				},
			},
		},
		{
			Path: "/path-only",
			MatchRules: []state.MatchRule{
				{
					MatchIdx:     0,
					RuleIdx:      2,
					UpstreamName: "test_baz_80",
					Source:       hr,
				},
			},
		},
		{
			Path: "/redirect-implicit-port",
			MatchRules: []state.MatchRule{
				{
					MatchIdx: 0,
					RuleIdx:  3,
					Source:   hr,
					Filters: state.Filters{
						RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
							Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("foo.example.com")),
						},
					},
				},
			},
		},
		{
			Path: "/redirect-explicit-port",
			MatchRules: []state.MatchRule{
				{
					MatchIdx: 0,
					RuleIdx:  4,
					Source:   hr,
					Filters: state.Filters{
						RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
							Hostname: (*v1beta1.PreciseHostname)(helpers.GetStringPointer("bar.example.com")),
							Port:     (*v1beta1.PortNumber)(helpers.GetInt32Pointer(8080)),
						},
					},
				},
			},
		},
	}

	httpServers := []state.VirtualServer{
		{
			Hostname:  "cafe.example.com",
			PathRules: cafePathRules,
		},
	}

	sslServers := []state.VirtualServer{
		{
			Hostname:  "cafe.example.com",
			SSL:       &state.SSL{CertificatePath: certPath},
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
		{Method: v1beta1.HTTPMethodPost, RedirectPath: "/_route0"},
		{Method: v1beta1.HTTPMethodPatch, RedirectPath: "/_route1"},
		{Any: true, RedirectPath: "/_route2"},
	}
	testMatches := []httpMatch{
		{
			Method:       v1beta1.HTTPMethodGet,
			Headers:      []string{"Version:V1", "test:foo", "my-header:my-value"},
			QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
			RedirectPath: "/test_route0",
		},
	}

	getExpectedLocations := func(isHTTPS bool) []location {
		port := 80
		if isHTTPS {
			port = 443
		}

		return []location{
			{
				Path:      "/_route0",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:      "/_route1",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:      "/_route2",
				Internal:  true,
				ProxyPass: "http://test_foo_80",
			},
			{
				Path:         "/",
				HTTPMatchVar: expectedMatchString(slashMatches),
			},
			{
				Path:      "/test_route0",
				Internal:  true,
				ProxyPass: "http://test_bar_80",
			},
			{
				Path:         "/test",
				HTTPMatchVar: expectedMatchString(testMatches),
			},
			{
				Path:      "/path-only",
				ProxyPass: "http://test_baz_80",
			},
			{
				Path: "/redirect-implicit-port",
				Return: &returnVal{
					Code: 302,
					URL:  fmt.Sprintf("$scheme://foo.example.com:%d$request_uri", port),
				},
			},
			{
				Path: "/redirect-explicit-port",
				Return: &returnVal{
					Code: 302,
					URL:  "$scheme://bar.example.com:8080$request_uri",
				},
			},
		}
	}

	expectedServers := []server{
		{
			IsDefaultHTTP: true,
		},
		{
			IsDefaultSSL: true,
		},
		{
			ServerName: "cafe.example.com",
			Locations:  getExpectedLocations(false),
		},
		{
			ServerName: "cafe.example.com",
			SSL:        &ssl{Certificate: certPath, CertificateKey: certPath},
			Locations:  getExpectedLocations(true),
		},
	}

	conf := state.Configuration{
		HTTPServers: httpServers,
		SSLServers:  sslServers,
	}

	result := generateHTTPServers(conf)

	if diff := cmp.Diff(expectedServers, result.Servers); diff != "" {
		t.Errorf("generate() mismatch on servers (-want +got):\n%s", diff)
	}
}

func TestGenerateProxyPass(t *testing.T) {
	expected := "http://10.0.0.1:80"

	result := generateProxyPass("10.0.0.1:80")
	if result != expected {
		t.Errorf("generateProxyPass() returned %s but expected %s", result, expected)
	}
}

func TestGenerateHTTPUpstreams(t *testing.T) {
	stateUpstreams := []state.Upstream{
		{
			Name: "up1",
			Endpoints: []resolver.Endpoint{
				{
					Address: "10.0.0.0",
					Port:    80,
				},
				{
					Address: "10.0.0.1",
					Port:    80,
				},
				{
					Address: "10.0.0.2",
					Port:    80,
				},
			},
		},
		{
			Name: "up2",
			Endpoints: []resolver.Endpoint{
				{
					Address: "11.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name:      "up3",
			Endpoints: []resolver.Endpoint{},
		},
	}

	expUpstreams := httpUpstreams{
		Upstreams: []upstream{
			{
				Name: "up1",
				Servers: []upstreamServer{
					{
						Address: "10.0.0.0:80",
					},
					{
						Address: "10.0.0.1:80",
					},
					{
						Address: "10.0.0.2:80",
					},
				},
			},
			{
				Name: "up2",
				Servers: []upstreamServer{
					{
						Address: "11.0.0.0:80",
					},
				},
			},
			{
				Name: "up3",
				Servers: []upstreamServer{
					{
						Address: nginx502Server,
					},
				},
			},
			{
				Name: state.InvalidBackendRef,
				Servers: []upstreamServer{
					{
						Address: nginx502Server,
					},
				},
			},
		},
	}

	result := generateHTTPUpstreams(stateUpstreams)
	if diff := cmp.Diff(expUpstreams, result); diff != "" {
		t.Errorf("generateHTTPUpstreams() mismatch (-want +got):\n%s", diff)
	}
}

func TestGenerateUpstream(t *testing.T) {
	tests := []struct {
		stateUpstream    state.Upstream
		expectedUpstream upstream
		msg              string
	}{
		{
			stateUpstream: state.Upstream{
				Name:      "nil-endpoints",
				Endpoints: nil,
			},
			expectedUpstream: upstream{Name: "nil-endpoints", Servers: []upstreamServer{{Address: nginx502Server}}},
			msg:              "nil endpoints",
		},
		{
			stateUpstream: state.Upstream{
				Name:      "no-endpoints",
				Endpoints: []resolver.Endpoint{},
			},
			expectedUpstream: upstream{Name: "no-endpoints", Servers: []upstreamServer{{Address: nginx502Server}}},
			msg:              "no endpoints",
		},
		{
			stateUpstream: state.Upstream{
				Name: "multiple-endpoints",
				Endpoints: []resolver.Endpoint{
					{
						Address: "10.0.0.1",
						Port:    80,
					},
					{
						Address: "10.0.0.2",
						Port:    80,
					},
					{
						Address: "10.0.0.3",
						Port:    80,
					},
				},
			},
			expectedUpstream: upstream{Name: "multiple-endpoints", Servers: []upstreamServer{{Address: "10.0.0.1:80"}, {Address: "10.0.0.2:80"}, {Address: "10.0.0.3:80"}}},
			msg:              "multiple endpoints",
		},
	}

	for _, test := range tests {
		result := generateUpstream(test.stateUpstream)
		if diff := cmp.Diff(test.expectedUpstream, result); diff != "" {
			t.Errorf("generateUpstream() %q mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestGenerateReturnValForRedirectFilter(t *testing.T) {
	const listenerPort = 123

	tests := []struct {
		filter   *v1beta1.HTTPRequestRedirectFilter
		expected *returnVal
		msg      string
	}{
		{
			filter:   nil,
			expected: nil,
			msg:      "filter is nil",
		},
		{
			filter: &v1beta1.HTTPRequestRedirectFilter{},
			expected: &returnVal{
				Code: statusFound,
				URL:  "$scheme://$host:123$request_uri",
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
			expected: &returnVal{
				Code: 101,
				URL:  "https://foo.example.com:2022$request_uri",
			},
			msg: "all fields are set",
		},
	}

	for _, test := range tests {
		result := generateReturnValForRedirectFilter(test.filter, listenerPort)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("generateReturnValForRedirectFilter() mismatch '%s' (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestGenerateMatchLocation(t *testing.T) {
	expected := location{
		Path:     "/path",
		Internal: true,
	}

	result := generateMatchLocation("/path")
	if result != expected {
		t.Errorf("generateMatchLocation() returned %v but expected %v", result, expected)
	}
}

func TestCreatePathForMatch(t *testing.T) {
	expected := "/path_route1"

	result := createPathForMatch("/path", 1)
	if result != expected {
		t.Errorf("createPathForMatch() returned %q but expected %q", result, expected)
	}
}

func TestCreateArgKeyValString(t *testing.T) {
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

func TestMatchLocationNeeded(t *testing.T) {
	tests := []struct {
		match    v1beta1.HTTPRouteMatch
		expected bool
		msg      string
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
		expected httpMatch
		msg      string
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
