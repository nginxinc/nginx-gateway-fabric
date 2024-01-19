package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteServers(t *testing.T) {
	conf := dataplane.Configuration{
		HTTPServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
				Port:      8080,
			},
			{
				Hostname: "example.com",
				Port:     8080,
			},
			{
				Hostname: "cafe.example.com",
				Port:     8080,
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
				Port:      8443,
			},
			{
				Hostname: "example.com",
				SSL: &dataplane.SSL{
					KeyPairID: "test-keypair",
				},
				Port: 8443,
			},
			{
				Hostname: "cafe.example.com",
				SSL: &dataplane.SSL{
					KeyPairID: "test-keypair",
				},
				Port: 8443,
			},
		},
	}

	expSubStrings := map[string]int{
		"listen 8080 default_server;":                              1,
		"listen 8080;":                                             2,
		"listen 8443 ssl;":                                         2,
		"listen 8443 ssl default_server;":                          1,
		"server_name example.com;":                                 2,
		"server_name cafe.example.com;":                            2,
		"ssl_certificate /etc/nginx/secrets/test-keypair.pem;":     2,
		"ssl_certificate_key /etc/nginx/secrets/test-keypair.pem;": 2,
	}
	g := NewWithT(t)
	servers := string(executeServers(conf))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(strings.Count(servers, expSubStr)).To(Equal(expCount))
	}
}

func TestExecuteForDefaultServers(t *testing.T) {
	testcases := []struct {
		msg       string
		httpPorts []int
		sslPorts  []int
		conf      dataplane.Configuration
	}{
		{
			conf: dataplane.Configuration{},
			msg:  "no default servers",
		},
		{
			conf: dataplane.Configuration{
				HTTPServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
				},
			},
			httpPorts: []int{80},
			msg:       "only HTTP default server",
		},
		{
			conf: dataplane.Configuration{
				SSLServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
				},
			},
			sslPorts: []int{443},
			msg:      "only HTTPS default server",
		},
		{
			conf: dataplane.Configuration{
				HTTPServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
						Port:      80,
					},
					{
						IsDefault: true,
						Port:      8080,
					},
				},
				SSLServers: []dataplane.VirtualServer{
					{
						IsDefault: true,
						Port:      443,
					},
					{
						IsDefault: true,
						Port:      8443,
					},
				},
			},
			httpPorts: []int{80, 8080},
			sslPorts:  []int{443, 8443},
			msg:       "multiple HTTP and HTTPS default servers",
		},
	}

	sslDefaultFmt := "listen %d ssl default_server"
	httpDefaultFmt := "listen %d default_server"

	for _, tc := range testcases {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			cfg := string(executeServers(tc.conf))

			for _, expPort := range tc.httpPorts {
				g.Expect(cfg).To(ContainSubstring(fmt.Sprintf(httpDefaultFmt, expPort)))
			}

			for _, expPort := range tc.sslPorts {
				g.Expect(cfg).To(ContainSubstring(fmt.Sprintf(sslDefaultFmt, expPort)))
			}
		})
	}
}

func TestCreateServers(t *testing.T) {
	const (
		sslKeyPairID = "test-keypair"
	)

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
					Match: dataplane.Match{
						Method: helpers.GetPointer("POST"),
					},
					BackendGroup: fooGroup,
				},
				{
					Match: dataplane.Match{
						Method: helpers.GetPointer("PATCH"),
					},
					BackendGroup: fooGroup,
				},
				{
					// should generate an "any" httpmatch since other matches exists for /
					Match:        dataplane.Match{},
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/test",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					// A match with all possible fields set
					Match: dataplane.Match{
						Method: helpers.GetPointer("GET"),
						Headers: []dataplane.HTTPHeaderMatch{
							{
								Name:  "Version",
								Value: "V1",
							},
							{
								Name:  "test",
								Value: "foo",
							},
							{
								Name:  "my-header",
								Value: "my-value",
							},
						},
						QueryParams: []dataplane.HTTPQueryParamMatch{
							{
								// query names and values should not be normalized to lowercase
								Name:  "GrEat",
								Value: "EXAMPLE",
							},
							{
								Name:  "test",
								Value: "foo=bar",
							},
						},
					},
					BackendGroup: barGroup,
				},
			},
		},
		{
			Path:     "/path-only",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: bazGroup,
				},
			},
		},
		{
			Path:     "/redirect-implicit-port",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{},
					Filters: dataplane.HTTPFilters{
						RequestRedirect: &dataplane.HTTPRequestRedirectFilter{
							Hostname: helpers.GetPointer("foo.example.com"),
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
					Match: dataplane.Match{},
					Filters: dataplane.HTTPFilters{
						RequestRedirect: &dataplane.HTTPRequestRedirectFilter{
							Hostname: helpers.GetPointer("bar.example.com"),
							Port:     helpers.GetPointer[int32](8080),
						},
					},
					BackendGroup: filterGroup2,
				},
			},
		},
		{
			Path:     "/redirect-with-headers",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{
						Headers: []dataplane.HTTPHeaderMatch{
							{
								Name:  "redirect",
								Value: "this",
							},
						},
					},
					Filters: dataplane.HTTPFilters{
						RequestRedirect: &dataplane.HTTPRequestRedirectFilter{
							Hostname: helpers.GetPointer("foo.example.com"),
							Port:     helpers.GetPointer[int32](8080),
						},
					},
					BackendGroup: filterGroup1,
				},
			},
		},
		{
			Path:     "/rewrite",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{},
					Filters: dataplane.HTTPFilters{
						RequestURLRewrite: &dataplane.HTTPURLRewriteFilter{
							Hostname: helpers.GetPointer("new.example.com"),
							Path: &dataplane.HTTPPathModifier{
								Type:        dataplane.ReplaceFullPath,
								Replacement: "/replacement",
							},
						},
					},
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/rewrite-with-headers",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{
						Headers: []dataplane.HTTPHeaderMatch{
							{
								Name:  "rewrite",
								Value: "this",
							},
						},
					},
					Filters: dataplane.HTTPFilters{
						RequestURLRewrite: &dataplane.HTTPURLRewriteFilter{
							Hostname: helpers.GetPointer("new.example.com"),
							Path: &dataplane.HTTPPathModifier{
								Type:        dataplane.ReplacePrefixMatch,
								Replacement: "/prefix-replacement",
							},
						},
					},
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/invalid-filter",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{},
					Filters: dataplane.HTTPFilters{
						InvalidFilter: &dataplane.InvalidHTTPFilter{},
					},
					BackendGroup: invalidFilterGroup,
				},
			},
		},
		{
			Path:     "/invalid-filter-with-headers",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{
						Headers: []dataplane.HTTPHeaderMatch{
							{
								Name:  "filter",
								Value: "this",
							},
						},
					},
					Filters: dataplane.HTTPFilters{
						InvalidFilter: &dataplane.InvalidHTTPFilter{},
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
					Match:        dataplane.Match{},
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/test",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{
						Method: helpers.GetPointer("GET"),
					},
					BackendGroup: fooGroup,
				},
			},
		},
		{
			Path:     "/proxy-set-headers",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: fooGroup,
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
			},
		},
	}

	httpServers := []dataplane.VirtualServer{
		{
			IsDefault: true,
			Port:      8080,
		},
		{
			Hostname:  "cafe.example.com",
			PathRules: cafePathRules,
			Port:      8080,
		},
	}

	sslServers := []dataplane.VirtualServer{
		{
			IsDefault: true,
			Port:      8443,
		},
		{
			Hostname:  "cafe.example.com",
			SSL:       &dataplane.SSL{KeyPairID: sslKeyPairID},
			PathRules: cafePathRules,
			Port:      8443,
		},
	}

	expectedMatchString := func(m []httpMatch) string {
		g := NewWithT(t)
		b, err := json.Marshal(m)
		g.Expect(err).ToNot(HaveOccurred())
		return string(b)
	}

	slashMatches := []httpMatch{
		{Method: "POST", RedirectPath: "@rule0-route0"},
		{Method: "PATCH", RedirectPath: "@rule0-route1"},
		{Any: true, RedirectPath: "@rule0-route2"},
	}
	testMatches := []httpMatch{
		{
			Method:       "GET",
			Headers:      []string{"Version:V1", "test:foo", "my-header:my-value"},
			QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
			RedirectPath: "@rule1-route0",
		},
	}
	exactMatches := []httpMatch{
		{
			Method:       "GET",
			RedirectPath: "@rule11-route0",
		},
	}
	redirectHeaderMatches := []httpMatch{
		{
			Headers:      []string{"redirect:this"},
			RedirectPath: "@rule5-route0",
		},
	}
	rewriteHeaderMatches := []httpMatch{
		{
			Headers:      []string{"rewrite:this"},
			RedirectPath: "@rule7-route0",
		},
	}
	rewriteProxySetHeaders := []http.Header{
		{
			Name:  "Host",
			Value: "new.example.com",
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
	invalidFilterHeaderMatches := []httpMatch{
		{
			Headers:      []string{"filter:this"},
			RedirectPath: "@rule9-route0",
		},
	}

	getExpectedLocations := func(isHTTPS bool) []http.Location {
		port := 8080
		if isHTTPS {
			port = 8443
		}

		return []http.Location{
			{
				Path:            "@rule0-route0",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:            "@rule0-route1",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:            "@rule0-route2",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:         "/",
				HTTPMatchVar: expectedMatchString(slashMatches),
			},
			{
				Path:            "@rule1-route0",
				ProxyPass:       "http://$test__route1_rule1$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:         "/test/",
				HTTPMatchVar: expectedMatchString(testMatches),
			},
			{
				Path:            "/path-only/",
				ProxyPass:       "http://invalid-backend-ref$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:            "= /path-only",
				ProxyPass:       "http://invalid-backend-ref$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path: "/redirect-implicit-port/",
				Return: &http.Return{
					Code: 302,
					Body: fmt.Sprintf("$scheme://foo.example.com:%d$request_uri", port),
				},
			},
			{
				Path: "= /redirect-implicit-port",
				Return: &http.Return{
					Code: 302,
					Body: fmt.Sprintf("$scheme://foo.example.com:%d$request_uri", port),
				},
			},
			{
				Path: "/redirect-explicit-port/",
				Return: &http.Return{
					Code: 302,
					Body: "$scheme://bar.example.com:8080$request_uri",
				},
			},
			{
				Path: "= /redirect-explicit-port",
				Return: &http.Return{
					Code: 302,
					Body: "$scheme://bar.example.com:8080$request_uri",
				},
			},
			{
				Path: "@rule5-route0",
				Return: &http.Return{
					Body: "$scheme://foo.example.com:8080$request_uri",
					Code: 302,
				},
			},
			{
				Path:         "/redirect-with-headers/",
				HTTPMatchVar: expectedMatchString(redirectHeaderMatches),
			},
			{
				Path:         "= /redirect-with-headers",
				HTTPMatchVar: expectedMatchString(redirectHeaderMatches),
			},
			{
				Path:            "/rewrite/",
				Rewrites:        []string{"^ /replacement break"},
				ProxyPass:       "http://test_foo_80",
				ProxySetHeaders: rewriteProxySetHeaders,
			},
			{
				Path:            "= /rewrite",
				Rewrites:        []string{"^ /replacement break"},
				ProxyPass:       "http://test_foo_80",
				ProxySetHeaders: rewriteProxySetHeaders,
			},
			{
				Path:            "@rule7-route0",
				Rewrites:        []string{"^/rewrite-with-headers(.*)$ /prefix-replacement$1 break"},
				ProxyPass:       "http://test_foo_80",
				ProxySetHeaders: rewriteProxySetHeaders,
			},
			{
				Path:         "/rewrite-with-headers/",
				HTTPMatchVar: expectedMatchString(rewriteHeaderMatches),
			},
			{
				Path:         "= /rewrite-with-headers",
				HTTPMatchVar: expectedMatchString(rewriteHeaderMatches),
			},
			{
				Path: "/invalid-filter/",
				Return: &http.Return{
					Code: http.StatusInternalServerError,
				},
			},
			{
				Path: "= /invalid-filter",
				Return: &http.Return{
					Code: http.StatusInternalServerError,
				},
			},
			{
				Path: "@rule9-route0",
				Return: &http.Return{
					Code: http.StatusInternalServerError,
				},
			},
			{
				Path:         "/invalid-filter-with-headers/",
				HTTPMatchVar: expectedMatchString(invalidFilterHeaderMatches),
			},
			{
				Path:         "= /invalid-filter-with-headers",
				HTTPMatchVar: expectedMatchString(invalidFilterHeaderMatches),
			},
			{
				Path:            "= /exact",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:            "@rule11-route0",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: baseHeaders,
			},
			{
				Path:         "= /test",
				HTTPMatchVar: expectedMatchString(exactMatches),
			},
			{
				Path:      "/proxy-set-headers/",
				ProxyPass: "http://test_foo_80$request_uri",
				ProxySetHeaders: []http.Header{
					{
						Name:  "my-header",
						Value: "${my_header_header_var}some-value-123",
					},
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
				},
			},
			{
				Path:      "= /proxy-set-headers",
				ProxyPass: "http://test_foo_80$request_uri",
				ProxySetHeaders: []http.Header{
					{
						Name:  "my-header",
						Value: "${my_header_header_var}some-value-123",
					},
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
				},
			},
		}
	}

	expectedPEMPath := fmt.Sprintf("/etc/nginx/secrets/%s.pem", sslKeyPairID)

	expectedServers := []http.Server{
		{
			IsDefaultHTTP: true,
			Port:          8080,
		},
		{
			ServerName: "cafe.example.com",
			Locations:  getExpectedLocations(false),
			Port:       8080,
		},
		{
			IsDefaultSSL: true,
			Port:         8443,
		},
		{
			ServerName: "cafe.example.com",
			SSL: &http.SSL{
				Certificate:    expectedPEMPath,
				CertificateKey: expectedPEMPath,
			},
			Locations: getExpectedLocations(true),
			Port:      8443,
		},
	}

	g := NewWithT(t)

	result := createServers(httpServers, sslServers)
	g.Expect(helpers.Diff(expectedServers, result)).To(BeEmpty())
}

func TestCreateServersConflicts(t *testing.T) {
	fooGroup := dataplane.BackendGroup{
		Source:  types.NamespacedName{Namespace: "test", Name: "route"},
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_foo_80",
				Valid:        true,
				Weight:       1,
			},
		},
	}
	barGroup := dataplane.BackendGroup{
		Source:  types.NamespacedName{Namespace: "test", Name: "route"},
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_bar_80",
				Valid:        true,
				Weight:       1,
			},
		},
	}
	bazGroup := dataplane.BackendGroup{
		Source:  types.NamespacedName{Namespace: "test", Name: "route"},
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_baz_80",
				Valid:        true,
				Weight:       1,
			},
		},
	}

	tests := []struct {
		name    string
		rules   []dataplane.PathRule
		expLocs []http.Location
	}{
		{
			name: "/coffee prefix, /coffee exact",
			rules: []dataplane.PathRule{
				{
					Path:     "/coffee",
					PathType: dataplane.PathTypePrefix,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: fooGroup,
						},
					},
				},
				{
					Path:     "/coffee",
					PathType: dataplane.PathTypeExact,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: barGroup,
						},
					},
				},
			},
			expLocs: []http.Location{
				{
					Path:            "/coffee/",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "= /coffee",
					ProxyPass:       "http://test_bar_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				createDefaultRootLocation(),
			},
		},
		{
			name: "/coffee prefix, /coffee/ prefix",
			rules: []dataplane.PathRule{
				{
					Path:     "/coffee",
					PathType: dataplane.PathTypePrefix,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: fooGroup,
						},
					},
				},
				{
					Path:     "/coffee/",
					PathType: dataplane.PathTypePrefix,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: barGroup,
						},
					},
				},
			},
			expLocs: []http.Location{
				{
					Path:            "= /coffee",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "/coffee/",
					ProxyPass:       "http://test_bar_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				createDefaultRootLocation(),
			},
		},
		{
			name: "/coffee prefix, /coffee/ prefix, /coffee exact",
			rules: []dataplane.PathRule{
				{
					Path:     "/coffee",
					PathType: dataplane.PathTypePrefix,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: fooGroup,
						},
					},
				},
				{
					Path:     "/coffee/",
					PathType: dataplane.PathTypePrefix,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: barGroup,
						},
					},
				},
				{
					Path:     "/coffee",
					PathType: dataplane.PathTypeExact,
					MatchRules: []dataplane.MatchRule{
						{
							Match:        dataplane.Match{},
							BackendGroup: bazGroup,
						},
					},
				},
			},
			expLocs: []http.Location{
				{
					Path:            "/coffee/",
					ProxyPass:       "http://test_bar_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "= /coffee",
					ProxyPass:       "http://test_baz_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				createDefaultRootLocation(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpServers := []dataplane.VirtualServer{
				{
					IsDefault: true,
					Port:      8080,
				},
				{
					Hostname:  "cafe.example.com",
					PathRules: test.rules,
					Port:      8080,
				},
			}
			expectedServers := []http.Server{
				{
					IsDefaultHTTP: true,
					Port:          8080,
				},
				{
					ServerName: "cafe.example.com",
					Locations:  test.expLocs,
					Port:       8080,
				},
			}

			g := NewWithT(t)

			result := createServers(httpServers, []dataplane.VirtualServer{})
			g.Expect(helpers.Diff(expectedServers, result)).To(BeEmpty())
		})
	}
}

func TestCreateLocationsRootPath(t *testing.T) {
	g := NewWithT(t)

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

	getPathRules := func(rootPath bool) []dataplane.PathRule {
		rules := []dataplane.PathRule{
			{
				Path: "/path-1",
				MatchRules: []dataplane.MatchRule{
					{
						Match:        dataplane.Match{},
						BackendGroup: fooGroup,
					},
				},
			},
			{
				Path: "/path-2",
				MatchRules: []dataplane.MatchRule{
					{
						Match:        dataplane.Match{},
						BackendGroup: fooGroup,
					},
				},
			},
		}

		if rootPath {
			rules = append(rules, dataplane.PathRule{
				Path: "/",
				MatchRules: []dataplane.MatchRule{
					{
						Match:        dataplane.Match{},
						BackendGroup: fooGroup,
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
			pathRules: getPathRules(false /* rootPath */),
			expLocations: []http.Location{
				{
					Path:            "/path-1",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "/path-2",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
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
			pathRules: getPathRules(true /* rootPath */),
			expLocations: []http.Location{
				{
					Path:            "/path-1",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "/path-2",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
				},
				{
					Path:            "/",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: baseHeaders,
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
		t.Run(test.name, func(t *testing.T) {
			locs := createLocations(test.pathRules, 80)
			g.Expect(locs).To(Equal(test.expLocations))
		})
	}
}

func TestCreateReturnValForRedirectFilter(t *testing.T) {
	const listenerPortCustom = 123
	const listenerPortHTTP = 80
	const listenerPortHTTPS = 443

	tests := []struct {
		filter       *dataplane.HTTPRequestRedirectFilter
		expected     *http.Return
		msg          string
		listenerPort int32
		path         string
	}{
		{
			filter:       nil,
			expected:     nil,
			listenerPort: listenerPortCustom,
			msg:          "filter is nil",
			path:         "",
		},
		{
			filter:       &dataplane.HTTPRequestRedirectFilter{},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: http.StatusFound,
				Body: "$scheme://$host:123$request_uri",
			},
			msg:  "all fields are empty",
			path: "/path",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](2022),
				StatusCode: helpers.GetPointer(301),
				Path: helpers.GetPointer(dataplane.HTTPPathModifier{
					Replacement: "/xyz",
					Type:        dataplane.ReplaceFullPath,
				}),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com:2022/xyz$request_uri",
			},
			msg:  "all fields are set",
			path: "/foo/bar",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com$request_uri",
			},
			msg:  "listenerPort is custom, scheme is set, no port",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Hostname:   helpers.GetPointer("foo.example.com"),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortHTTPS,
			expected: &http.Return{
				Code: 301,
				Body: "$scheme://foo.example.com:443$request_uri",
			},
			msg:  "no scheme, listenerPort https, no port is set",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortHTTPS,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com$request_uri",
			},
			msg:  "scheme is https, listenerPort https, no port is set",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("http"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortHTTP,
			expected: &http.Return{
				Code: 301,
				Body: "http://foo.example.com$request_uri",
			},
			msg:  "scheme is http, listenerPort http, no port is set",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("http"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](80),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "http://foo.example.com$request_uri",
			},
			msg:  "scheme is http, port http",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](443),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com$request_uri",
			},
			msg:  "scheme is https, port https",
			path: "",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](2022),
				StatusCode: helpers.GetPointer(301),
				Path: helpers.GetPointer(dataplane.HTTPPathModifier{
					Replacement: "/xyz",
					Type:        dataplane.ReplaceFullPath,
				}),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com:2022/xyz$request_uri",
			},
			msg:  "ReplaceFullPath set",
			path: "/foo/bar",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](2022),
				StatusCode: helpers.GetPointer(301),
				Path: helpers.GetPointer(dataplane.HTTPPathModifier{
					Replacement: "/xyz",
					Type:        dataplane.ReplacePrefixMatch,
				}),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com:2022^/foo/bar(.*)$ /xyz$1$request_uri",
			},
			msg:  "replacePrefixMatch set",
			path: "/foo/bar",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](2022),
				StatusCode: helpers.GetPointer(301),
				Path: helpers.GetPointer(dataplane.HTTPPathModifier{
					Replacement: "/xyz/",
					Type:        dataplane.ReplacePrefixMatch,
				}),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com:2022^foo/bar(?:/(.*))?$ /xyz/$1$request_uri",
			},
			msg:  "replacePrefixMatch without path / suffix",
			path: "foo/bar",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := createReturnValForRedirectFilter(test.filter, test.listenerPort, test.path)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestCreateRewritesValForRewriteFilter(t *testing.T) {
	tests := []struct {
		filter   *dataplane.HTTPURLRewriteFilter
		expected *rewriteConfig
		msg      string
		path     string
	}{
		{
			filter:   nil,
			expected: nil,
			msg:      "filter is nil",
		},
		{
			filter:   &dataplane.HTTPURLRewriteFilter{},
			expected: &rewriteConfig{},
			msg:      "all fields are empty",
		},
		{
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplaceFullPath,
					Replacement: "/full-path",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^ /full-path break",
			},
			msg: "full path",
		},
		{
			path: "/original",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "/prefix-path",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original(.*)$ /prefix-path$1 break",
			},
			msg: "prefix path no trailing slashes",
		},
		{
			path: "/original",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original(?:/(.*))?$ /$1 break",
			},
			msg: "prefix path empty string",
		},
		{
			path: "/original",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "/",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original(?:/(.*))?$ /$1 break",
			},
			msg: "prefix path /",
		},
		{
			path: "/original",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "/trailing/",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original(?:/(.*))?$ /trailing/$1 break",
			},
			msg: "prefix path replacement with trailing /",
		},
		{
			path: "/original/",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "/prefix-path",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original/(.*)$ /prefix-path/$1 break",
			},
			msg: "prefix path original with trailing /",
		},
		{
			path: "/original/",
			filter: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{
					Type:        dataplane.ReplacePrefixMatch,
					Replacement: "/trailing/",
				},
			},
			expected: &rewriteConfig{
				Rewrite: "^/original/(.*)$ /trailing/$1 break",
			},
			msg: "prefix path both with trailing slashes",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := createRewritesValForRewriteFilter(test.filter, test.path)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestCreateHTTPMatch(t *testing.T) {
	testPath := "/internal_loc"

	testMethodMatch := helpers.GetPointer("PUT")
	testHeaderMatches := []dataplane.HTTPHeaderMatch{
		{
			Name:  "header-1",
			Value: "val-1",
		},
		{
			Name:  "header-2",
			Value: "val-2",
		},
		{
			Name:  "header-3",
			Value: "val-3",
		},
	}

	testDuplicateHeaders := make([]dataplane.HTTPHeaderMatch, 0, 4)
	duplicateHeaderMatch := dataplane.HTTPHeaderMatch{
		Name:  "HEADER-2", // header names are case-insensitive
		Value: "val-2",
	}
	testDuplicateHeaders = append(testDuplicateHeaders, testHeaderMatches...)
	testDuplicateHeaders = append(testDuplicateHeaders, duplicateHeaderMatch)

	testQueryParamMatches := []dataplane.HTTPQueryParamMatch{
		{
			Name:  "arg1",
			Value: "val1",
		},
		{
			Name:  "arg2",
			Value: "val2=another-val",
		},
		{
			Name:  "arg3",
			Value: "==val3",
		},
	}

	expectedHeaders := []string{"header-1:val-1", "header-2:val-2", "header-3:val-3"}
	expectedArgs := []string{"arg1=val1", "arg2=val2=another-val", "arg3===val3"}

	tests := []struct {
		match    dataplane.Match
		msg      string
		expected httpMatch
	}{
		{
			match: dataplane.Match{},
			expected: httpMatch{
				Any:          true,
				RedirectPath: testPath,
			},
			msg: "path only match",
		},
		{
			match: dataplane.Match{
				Method: testMethodMatch, // A path match with a method should not set the Any field to true
			},
			expected: httpMatch{
				Method:       "PUT",
				RedirectPath: testPath,
			},
			msg: "method only match",
		},
		{
			match: dataplane.Match{
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				RedirectPath: testPath,
				Headers:      expectedHeaders,
			},
			msg: "headers only match",
		},
		{
			match: dataplane.Match{
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "query params only match",
		},
		{
			match: dataplane.Match{
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
			match: dataplane.Match{
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
			match: dataplane.Match{
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
			match: dataplane.Match{
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
			match: dataplane.Match{
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
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := createHTTPMatch(tc.match, testPath)
			g.Expect(helpers.Diff(result, tc.expected)).To(BeEmpty())
		})
	}
}

func TestCreateQueryParamKeyValString(t *testing.T) {
	g := NewWithT(t)

	expected := "key=value"

	result := createQueryParamKeyValString(
		dataplane.HTTPQueryParamMatch{
			Name:  "key",
			Value: "value",
		},
	)

	g.Expect(result).To(Equal(expected))

	expected = "KeY=vaLUe=="

	result = createQueryParamKeyValString(
		dataplane.HTTPQueryParamMatch{
			Name:  "KeY",
			Value: "vaLUe==",
		},
	)

	g.Expect(result).To(Equal(expected))
}

func TestCreateHeaderKeyValString(t *testing.T) {
	g := NewWithT(t)

	expected := "kEy:vALUe"

	result := createHeaderKeyValString(
		dataplane.HTTPHeaderMatch{
			Name:  "kEy",
			Value: "vALUe",
		},
	)

	g.Expect(result).To(Equal(expected))
}

func TestIsPathOnlyMatch(t *testing.T) {
	tests := []struct {
		msg      string
		match    dataplane.Match
		expected bool
	}{
		{
			match:    dataplane.Match{},
			expected: true,
			msg:      "path only match",
		},
		{
			match: dataplane.Match{
				Method: helpers.GetPointer("GET"),
			},
			expected: false,
			msg:      "method defined in match",
		},
		{
			match: dataplane.Match{
				Headers: []dataplane.HTTPHeaderMatch{
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
			match: dataplane.Match{
				QueryParams: []dataplane.HTTPQueryParamMatch{
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
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := isPathOnlyMatch(tc.match)
			g.Expect(result).To(Equal(tc.expected))
		})
	}
}

func TestCreateProxyPass(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		rewrite  *dataplane.HTTPURLRewriteFilter
		expected string
		grp      dataplane.BackendGroup
	}{
		{
			expected: "http://10.0.0.1:80$request_uri",
			grp: dataplane.BackendGroup{
				Backends: []dataplane.Backend{
					{
						UpstreamName: "10.0.0.1:80",
						Valid:        true,
						Weight:       1,
					},
				},
			},
		},
		{
			expected: "http://$ns1__bg_rule0$request_uri",
			grp: dataplane.BackendGroup{
				Source: types.NamespacedName{Namespace: "ns1", Name: "bg"},
				Backends: []dataplane.Backend{
					{
						UpstreamName: "my-variable",
						Valid:        true,
						Weight:       1,
					},
					{
						UpstreamName: "my-variable2",
						Valid:        true,
						Weight:       1,
					},
				},
			},
		},
		{
			expected: "http://10.0.0.1:80",
			rewrite: &dataplane.HTTPURLRewriteFilter{
				Path: &dataplane.HTTPPathModifier{},
			},
			grp: dataplane.BackendGroup{
				Backends: []dataplane.Backend{
					{
						UpstreamName: "10.0.0.1:80",
						Valid:        true,
						Weight:       1,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		result := createProxyPass(tc.grp, tc.rewrite)
		g.Expect(result).To(Equal(tc.expected))
	}
}

func TestCreateMatchLocation(t *testing.T) {
	g := NewWithT(t)

	expected := http.Location{
		Path: "/path",
	}

	result := createMatchLocation("/path")
	g.Expect(result).To(Equal(expected))
}

func TestGenerateProxySetHeaders(t *testing.T) {
	tests := []struct {
		filters         *dataplane.HTTPFilters
		msg             string
		expectedHeaders []http.Header
	}{
		{
			msg: "header filter",
			filters: &dataplane.HTTPFilters{
				RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
					Add: []dataplane.HTTPHeader{
						{
							Name:  "Authorization",
							Value: "my-auth",
						},
					},
					Set: []dataplane.HTTPHeader{
						{
							Name:  "Accept-Encoding",
							Value: "gzip",
						},
					},
					Remove: []string{"my-header"},
				},
			},
			expectedHeaders: []http.Header{
				{
					Name:  "Authorization",
					Value: "${authorization_header_var}my-auth",
				},
				{
					Name:  "Accept-Encoding",
					Value: "gzip",
				},
				{
					Name:  "my-header",
					Value: "",
				},
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
			},
		},
		{
			msg: "with url rewrite hostname",
			filters: &dataplane.HTTPFilters{
				RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
					Add: []dataplane.HTTPHeader{
						{
							Name:  "Authorization",
							Value: "my-auth",
						},
					},
				},
				RequestURLRewrite: &dataplane.HTTPURLRewriteFilter{
					Hostname: helpers.GetPointer("rewrite-hostname"),
				},
			},
			expectedHeaders: []http.Header{
				{
					Name:  "Authorization",
					Value: "${authorization_header_var}my-auth",
				},
				{
					Name:  "Host",
					Value: "rewrite-hostname",
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
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			headers := generateProxySetHeaders(tc.filters)
			g.Expect(headers).To(Equal(tc.expectedHeaders))
		})
	}
}
