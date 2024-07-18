package config

import (
	"fmt"
	"maps"
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
				Additions: []dataplane.Addition{
					{
						Bytes:      []byte("addition-1"),
						Identifier: "addition-1",
					},
				},
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
				PathRules: []dataplane.PathRule{
					{
						Path:     "/",
						PathType: dataplane.PathTypePrefix,
						MatchRules: []dataplane.MatchRule{
							{
								Match: dataplane.Match{},
								BackendGroup: dataplane.BackendGroup{
									Source:  types.NamespacedName{Namespace: "test", Name: "route1"},
									RuleIdx: 0,
									Backends: []dataplane.Backend{
										{
											UpstreamName: "test_foo_443",
											Valid:        true,
											Weight:       1,
											VerifyTLS: &dataplane.VerifyTLS{
												CertBundleID: "test-foo",
												Hostname:     "test-foo.example.com",
											},
										},
									},
								},
							},
						},
					},
				},
				Additions: []dataplane.Addition{
					{
						Bytes:      []byte("addition-1"),
						Identifier: "addition-1", // duplicate
					},
					{
						Bytes:      []byte("addition-2"),
						Identifier: "addition-2",
					},
				},
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
		"proxy_ssl_server_name on;":                                1,
	}

	type assertion func(g *WithT, data string)

	expectedResults := map[string]assertion{
		httpConfigFile: func(g *WithT, data string) {
			for expSubStr, expCount := range expSubStrings {
				g.Expect(strings.Count(data, expSubStr)).To(Equal(expCount))
			}
		},
		httpMatchVarsFile: func(g *WithT, data string) {
			g.Expect(data).To(Equal("{}"))
		},
		includesFolder + "/addition-1.conf": func(g *WithT, data string) {
			g.Expect(data).To(Equal("addition-1"))
		},
		includesFolder + "/addition-2.conf": func(g *WithT, data string) {
			g.Expect(data).To(Equal("addition-2"))
		},
	}
	g := NewWithT(t)

	results := executeServers(conf)
	g.Expect(results).To(HaveLen(len(expectedResults)))

	for _, res := range results {
		g.Expect(expectedResults).To(HaveKey(res.dest), "executeServers returned unexpected result destination")

		assertData := expectedResults[res.dest]
		assertData(g, string(res.data))
	}
}

func TestExecuteServersForIPFamily(t *testing.T) {
	httpServers := []dataplane.VirtualServer{
		{
			IsDefault: true,
			Port:      8080,
		},
		{
			Hostname: "example.com",
			Port:     8080,
		},
	}
	sslServers := []dataplane.VirtualServer{
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
	}
	tests := []struct {
		msg                string
		expectedHTTPConfig map[string]int
		config             dataplane.Configuration
	}{
		{
			msg: "http and ssl servers with IPv4 IP family",
			config: dataplane.Configuration{
				HTTPServers: httpServers,
				SSLServers:  sslServers,
				BaseHTTPConfig: dataplane.BaseHTTPConfig{
					IPFamily: dataplane.IPv4,
				},
			},
			expectedHTTPConfig: map[string]int{
				"listen 8080 default_server;":                              1,
				"listen 8080;":                                             1,
				"listen 8443 ssl default_server;":                          1,
				"listen 8443 ssl;":                                         1,
				"server_name example.com;":                                 2,
				"ssl_certificate /etc/nginx/secrets/test-keypair.pem;":     1,
				"ssl_certificate_key /etc/nginx/secrets/test-keypair.pem;": 1,
				"ssl_reject_handshake on;":                                 1,
			},
		},
		{
			msg: "http and ssl servers with IPv6 IP family",
			config: dataplane.Configuration{
				HTTPServers: httpServers,
				SSLServers:  sslServers,
				BaseHTTPConfig: dataplane.BaseHTTPConfig{
					IPFamily: dataplane.IPv6,
				},
			},
			expectedHTTPConfig: map[string]int{
				"listen [::]:8080 default_server;":                         1,
				"listen [::]:8080;":                                        1,
				"listen [::]:8443 ssl default_server;":                     1,
				"listen [::]:8443 ssl;":                                    1,
				"server_name example.com;":                                 2,
				"ssl_certificate /etc/nginx/secrets/test-keypair.pem;":     1,
				"ssl_certificate_key /etc/nginx/secrets/test-keypair.pem;": 1,
				"ssl_reject_handshake on;":                                 1,
			},
		},
		{
			msg: "http and ssl servers with Dual IP family",
			config: dataplane.Configuration{
				HTTPServers: httpServers,
				SSLServers:  sslServers,
				BaseHTTPConfig: dataplane.BaseHTTPConfig{
					IPFamily: dataplane.Dual,
				},
			},
			expectedHTTPConfig: map[string]int{
				"listen 8080 default_server;":                              1,
				"listen 8080;":                                             1,
				"listen 8443 ssl default_server;":                          1,
				"listen 8443 ssl;":                                         1,
				"server_name example.com;":                                 2,
				"ssl_certificate /etc/nginx/secrets/test-keypair.pem;":     1,
				"ssl_certificate_key /etc/nginx/secrets/test-keypair.pem;": 1,
				"ssl_reject_handshake on;":                                 1,
				"listen [::]:8080 default_server;":                         1,
				"listen [::]:8080;":                                        1,
				"listen [::]:8443 ssl default_server;":                     1,
				"listen [::]:8443 ssl;":                                    1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)
			results := executeServers(test.config)
			g.Expect(results).To(HaveLen(2))
			serverConf := string(results[0].data)
			httpMatchConf := string(results[1].data)
			g.Expect(httpMatchConf).To(Equal("{}"))

			for expSubStr, expCount := range test.expectedHTTPConfig {
				g.Expect(strings.Count(serverConf, expSubStr)).To(Equal(expCount))
			}
		})
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

			serverResults := executeServers(tc.conf)
			g.Expect(serverResults).To(HaveLen(2))
			serverConf := string(serverResults[0].data)
			httpMatchConf := string(serverResults[1].data)
			g.Expect(httpMatchConf).To(Equal("{}"))

			for _, expPort := range tc.httpPorts {
				g.Expect(serverConf).To(ContainSubstring(fmt.Sprintf(httpDefaultFmt, expPort)))
			}

			for _, expPort := range tc.sslPorts {
				g.Expect(serverConf).To(ContainSubstring(fmt.Sprintf(sslDefaultFmt, expPort)))
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

	btpGroup := dataplane.BackendGroup{
		Source:  hrNsName,
		RuleIdx: 3,
		Backends: []dataplane.Backend{
			{
				UpstreamName: "test_btp_80",
				Valid:        true,
				Weight:       1,
				VerifyTLS: &dataplane.VerifyTLS{
					CertBundleID: "test-btp",
					Hostname:     "test-btp.example.com",
				},
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
			Path:     "/backend-tls-policy",
			PathType: dataplane.PathTypePrefix,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: btpGroup,
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
						ResponseHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-header-response",
									Value: "some-value-response-123",
								},
							},
						},
					},
				},
			},
		},
		{
			Path:     "/grpc/method",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: fooGroup,
				},
			},
			GRPC: true,
		},
		{
			Path:     "/grpc-with-backend-tls-policy/method",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: btpGroup,
				},
			},
			GRPC: true,
		},
		{
			Path:     "/addition-path-only-match",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					Match:        dataplane.Match{},
					BackendGroup: fooGroup,
					Additions: []dataplane.Addition{
						{
							Bytes:      []byte("path-only-match-addition"),
							Identifier: "path-only-match-addition",
						},
					},
				},
			},
		},
		{
			Path:     "/addition-header-match",
			PathType: dataplane.PathTypeExact,
			MatchRules: []dataplane.MatchRule{
				{
					Match: dataplane.Match{
						Method: helpers.GetPointer("GET"),
					},
					BackendGroup: fooGroup,
					Additions: []dataplane.Addition{
						{
							Bytes:      []byte("match-addition"),
							Identifier: "match-addition",
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
			Additions: []dataplane.Addition{
				{
					Bytes:      []byte("server-addition-1"),
					Identifier: "server-addition-1",
				},
				{
					Bytes:      []byte("server-addition-2"),
					Identifier: "server-addition-2",
				},
			},
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
			Additions: []dataplane.Addition{
				{
					Bytes:      []byte("server-addition-1"),
					Identifier: "server-addition-1",
				},
				{
					Bytes:      []byte("server-addition-3"),
					Identifier: "server-addition-3",
				},
			},
		},
	}

	expMatchPairs := httpMatchPairs{
		"1_0": {
			{Method: "POST", RedirectPath: "@rule0-route0"},
			{Method: "PATCH", RedirectPath: "@rule0-route1"},
			{RedirectPath: "@rule0-route2", Any: true},
		},
		"1_1": {
			{
				Method:       "GET",
				Headers:      []string{"Version:V1", "test:foo", "my-header:my-value"},
				QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
				RedirectPath: "@rule1-route0",
			},
		},
		"1_6": {
			{RedirectPath: "@rule6-route0", Headers: []string{"redirect:this"}},
		},
		"1_8": {
			{
				Headers:      []string{"rewrite:this"},
				RedirectPath: "@rule8-route0",
			},
		},
		"1_10": {
			{
				Headers:      []string{"filter:this"},
				RedirectPath: "@rule10-route0",
			},
		},
		"1_12": {
			{
				Method:       "GET",
				RedirectPath: "@rule12-route0",
				Headers:      nil,
				QueryParams:  nil,
				Any:          false,
			},
		},
		"1_17": {
			{
				Method:       "GET",
				RedirectPath: "@rule17-route0",
			},
		},
	}

	allExpMatchPair := make(httpMatchPairs)
	maps.Copy(allExpMatchPair, expMatchPairs)
	modifiedMatchPairs := modifyMatchPairs(expMatchPairs)
	maps.Copy(allExpMatchPair, modifiedMatchPairs)

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

	getExpectedLocations := func(isHTTPS bool) []http.Location {
		port := 8080
		ssl := ""
		if isHTTPS {
			port = 8443
			ssl = "SSL"
		}

		return []http.Location{
			{
				Path:            "@rule0-route0",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:            "@rule0-route1",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:            "@rule0-route2",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:         "/",
				HTTPMatchKey: ssl + "1_0",
			},
			{
				Path:            "@rule1-route0",
				ProxyPass:       "http://$test__route1_rule1$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:         "/test/",
				HTTPMatchKey: ssl + "1_1",
			},
			{
				Path:            "/path-only/",
				ProxyPass:       "http://invalid-backend-ref$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:            "= /path-only",
				ProxyPass:       "http://invalid-backend-ref$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:            "/backend-tls-policy/",
				ProxyPass:       "https://test_btp_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
				ProxySSLVerify: &http.ProxySSLVerify{
					Name:               "test-btp.example.com",
					TrustedCertificate: "/etc/nginx/secrets/test-btp.crt",
				},
			},
			{
				Path:            "= /backend-tls-policy",
				ProxyPass:       "https://test_btp_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
				ProxySSLVerify: &http.ProxySSLVerify{
					Name:               "test-btp.example.com",
					TrustedCertificate: "/etc/nginx/secrets/test-btp.crt",
				},
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
				Path: "@rule6-route0",
				Return: &http.Return{
					Body: "$scheme://foo.example.com:8080$request_uri",
					Code: 302,
				},
			},
			{
				Path:         "/redirect-with-headers/",
				HTTPMatchKey: ssl + "1_6",
			},
			{
				Path:         "= /redirect-with-headers",
				HTTPMatchKey: ssl + "1_6",
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
				Path:            "@rule8-route0",
				Rewrites:        []string{"^/rewrite-with-headers(.*)$ /prefix-replacement$1 break"},
				ProxyPass:       "http://test_foo_80",
				ProxySetHeaders: rewriteProxySetHeaders,
			},
			{
				Path:         "/rewrite-with-headers/",
				HTTPMatchKey: ssl + "1_8",
			},
			{
				Path:         "= /rewrite-with-headers",
				HTTPMatchKey: ssl + "1_8",
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
				Path: "@rule10-route0",
				Return: &http.Return{
					Code: http.StatusInternalServerError,
				},
			},
			{
				Path:         "/invalid-filter-with-headers/",
				HTTPMatchKey: ssl + "1_10",
			},
			{
				Path:         "= /invalid-filter-with-headers",
				HTTPMatchKey: ssl + "1_10",
			},
			{
				Path:            "= /exact",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:            "@rule12-route0",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
			},
			{
				Path:         "= /test",
				HTTPMatchKey: ssl + "1_12",
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
				ResponseHeaders: http.ResponseHeaders{
					Add: []http.Header{
						{
							Name:  "my-header-response",
							Value: "some-value-response-123",
						},
					},
					Set:    []http.Header{},
					Remove: []string{},
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
				ResponseHeaders: http.ResponseHeaders{
					Add: []http.Header{
						{
							Name:  "my-header-response",
							Value: "some-value-response-123",
						},
					},
					Set:    []http.Header{},
					Remove: []string{},
				},
			},
			{
				Path:            "= /grpc/method",
				ProxyPass:       "grpc://test_foo_80",
				GRPC:            true,
				ProxySetHeaders: grpcBaseHeaders,
			},
			{
				Path:      "= /grpc-with-backend-tls-policy/method",
				ProxyPass: "grpcs://test_btp_80",
				ProxySSLVerify: &http.ProxySSLVerify{
					Name:               "test-btp.example.com",
					TrustedCertificate: "/etc/nginx/secrets/test-btp.crt",
				},
				GRPC:            true,
				ProxySetHeaders: grpcBaseHeaders,
			},
			{
				Path:            "= /addition-path-only-match",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
				Includes: []string{
					includesFolder + "/path-only-match-addition.conf",
				},
			},
			{
				Path:            "@rule17-route0",
				ProxyPass:       "http://test_foo_80$request_uri",
				ProxySetHeaders: httpBaseHeaders,
				Includes: []string{
					includesFolder + "/match-addition.conf",
				},
			},
			{
				Path:         "= /addition-header-match",
				HTTPMatchKey: ssl + "1_17",
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
			GRPC:       true,
			Includes: []string{
				includesFolder + "/server-addition-1.conf",
				includesFolder + "/server-addition-2.conf",
			},
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
			GRPC:      true,
			Includes: []string{
				includesFolder + "/server-addition-1.conf",
				includesFolder + "/server-addition-3.conf",
			},
		},
	}

	g := NewWithT(t)

	result, httpMatchPair := createServers(httpServers, sslServers)

	g.Expect(httpMatchPair).To(Equal(allExpMatchPair))
	g.Expect(helpers.Diff(expectedServers, result)).To(BeEmpty())
}

func modifyMatchPairs(matchPairs httpMatchPairs) httpMatchPairs {
	modified := make(httpMatchPairs)
	for k, v := range matchPairs {
		modifiedKey := "SSL" + k
		modified[modifiedKey] = v
	}

	return modified
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
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "= /coffee",
					ProxyPass:       "http://test_bar_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
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
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/coffee/",
					ProxyPass:       "http://test_bar_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
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
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "= /coffee",
					ProxyPass:       "http://test_baz_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
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

			result, _ := createServers(httpServers, []dataplane.VirtualServer{})
			g.Expect(helpers.Diff(expectedServers, result)).To(BeEmpty())
		})
	}
}

func TestCreateLocationsRootPath(t *testing.T) {
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

	getPathRules := func(rootPath bool, grpc bool) []dataplane.PathRule {
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

		if grpc {
			rules = append(rules, dataplane.PathRule{
				Path: "/grpc",
				GRPC: true,
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
		grpc         bool
	}{
		{
			name:      "path rules with no root path should generate a default 404 root location",
			pathRules: getPathRules(false /* rootPath */, false /* grpc */),
			expLocations: []http.Location{
				{
					Path:            "/path-1",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/path-2",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
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
			name:      "path rules with grpc & with no root path should generate a default 404 root location and GRPC true",
			pathRules: getPathRules(false /* rootPath */, true /* grpc */),
			grpc:      true,
			expLocations: []http.Location{
				{
					Path:            "/path-1",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/path-2",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/grpc",
					ProxyPass:       "grpc://test_foo_80",
					GRPC:            true,
					ProxySetHeaders: grpcBaseHeaders,
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
			pathRules: getPathRules(true /* rootPath */, false /* grpc */),
			expLocations: []http.Location{
				{
					Path:            "/path-1",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/path-2",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
				},
				{
					Path:            "/",
					ProxyPass:       "http://test_foo_80$request_uri",
					ProxySetHeaders: httpBaseHeaders,
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
			g := NewWithT(t)

			locs, httpMatchPair, grpc := createLocations(&dataplane.VirtualServer{
				PathRules: test.pathRules,
				Port:      80,
			}, 1)
			g.Expect(locs).To(Equal(test.expLocations))
			g.Expect(httpMatchPair).To(BeEmpty())
			g.Expect(grpc).To(Equal(test.grpc))
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
	}{
		{
			filter:       nil,
			expected:     nil,
			listenerPort: listenerPortCustom,
			msg:          "filter is nil",
		},
		{
			filter:       &dataplane.HTTPRequestRedirectFilter{},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: http.StatusFound,
				Body: "$scheme://$host:123$request_uri",
			},
			msg: "all fields are empty",
		},
		{
			filter: &dataplane.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("foo.example.com"),
				Port:       helpers.GetPointer[int32](2022),
				StatusCode: helpers.GetPointer(301),
			},
			listenerPort: listenerPortCustom,
			expected: &http.Return{
				Code: 301,
				Body: "https://foo.example.com:2022$request_uri",
			},
			msg: "all fields are set",
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
			msg: "listenerPort is custom, scheme is set, no port",
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
			msg: "no scheme, listenerPort https, no port is set",
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
			msg: "scheme is https, listenerPort https, no port is set",
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
			msg: "scheme is http, listenerPort http, no port is set",
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
			msg: "scheme is http, port http",
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
			msg: "scheme is https, port https",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := createReturnValForRedirectFilter(test.filter, test.listenerPort)
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

func TestCreateRouteMatch(t *testing.T) {
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
		expected routeMatch
	}{
		{
			match: dataplane.Match{},
			expected: routeMatch{
				Any:          true,
				RedirectPath: testPath,
			},
			msg: "path only match",
		},
		{
			match: dataplane.Match{
				Method: testMethodMatch, // A path match with a method should not set the Any field to true
			},
			expected: routeMatch{
				Method:       "PUT",
				RedirectPath: testPath,
			},
			msg: "method only match",
		},
		{
			match: dataplane.Match{
				Headers: testHeaderMatches,
			},
			expected: routeMatch{
				RedirectPath: testPath,
				Headers:      expectedHeaders,
			},
			msg: "headers only match",
		},
		{
			match: dataplane.Match{
				QueryParams: testQueryParamMatches,
			},
			expected: routeMatch{
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
			expected: routeMatch{
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
			expected: routeMatch{
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
			expected: routeMatch{
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
			expected: routeMatch{
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
			expected: routeMatch{
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "duplicate header names",
		},
	}
	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			result := createRouteMatch(tc.match, testPath)
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
		GRPC     bool
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
		{
			expected: "grpc://10.0.0.1:80",
			grp: dataplane.BackendGroup{
				Backends: []dataplane.Backend{
					{
						UpstreamName: "10.0.0.1:80",
						Valid:        true,
						Weight:       1,
					},
				},
			},
			GRPC: true,
		},
	}

	for _, tc := range tests {
		result := createProxyPass(tc.grp, tc.rewrite, generateProtocolString(nil, tc.GRPC), tc.GRPC)
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
		GRPC            bool
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
		{
			msg:  "header filter with gRPC",
			GRPC: true,
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
					Name:  "Authority",
					Value: "$gw_api_compliant_host",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			headers := generateProxySetHeaders(tc.filters, tc.GRPC)
			g.Expect(headers).To(Equal(tc.expectedHeaders))
		})
	}
}

func TestConvertBackendTLSFromGroup(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		expected *http.ProxySSLVerify
		msg      string
		grp      []dataplane.Backend
	}{
		{
			msg: "tls enabled, one backend",
			grp: []dataplane.Backend{
				{
					UpstreamName: "my-upstream",
					Valid:        true,
					Weight:       1,
					VerifyTLS: &dataplane.VerifyTLS{
						CertBundleID: "default-my-cert",
						Hostname:     "my-hostname",
					},
				},
			},
			expected: &http.ProxySSLVerify{
				TrustedCertificate: "/etc/nginx/secrets/default-my-cert.crt",
				Name:               "my-hostname",
			},
		},
		{
			msg: "tls disabled",
			grp: []dataplane.Backend{
				{
					UpstreamName: "my-upstream",
					Valid:        true,
					Weight:       1,
					VerifyTLS:    nil,
				},
			},
			expected: nil,
		},
		{
			msg: "tls enabled, multiple backends",
			grp: []dataplane.Backend{
				{
					UpstreamName: "my-upstream",
					Valid:        true,
					Weight:       1,
					VerifyTLS: &dataplane.VerifyTLS{
						CertBundleID: "default-my-cert",
						Hostname:     "my-hostname",
					},
				},
				{
					UpstreamName: "my-upstream",
					Valid:        true,
					Weight:       2,
				},
			},
			expected: &http.ProxySSLVerify{
				TrustedCertificate: "/etc/nginx/secrets/default-my-cert.crt",
				Name:               "my-hostname",
			},
		},
		{
			msg: "tls enabled, system certs enabled",
			grp: []dataplane.Backend{
				{
					UpstreamName: "my-upstream",
					Valid:        true,
					Weight:       1,
					VerifyTLS: &dataplane.VerifyTLS{
						Hostname:   "my-hostname",
						RootCAPath: "/etc/ssl/certs/ca-certificates.crt",
					},
				},
			},
			expected: &http.ProxySSLVerify{
				TrustedCertificate: "/etc/ssl/certs/ca-certificates.crt",
				Name:               "my-hostname",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(_ *testing.T) {
			result := createProxyTLSFromBackends(tc.grp)
			g.Expect(result).To(Equal(tc.expected))
		})
	}
}

func TestGenerateResponseHeaders(t *testing.T) {
	tests := []struct {
		filters         *dataplane.HTTPFilters
		msg             string
		expectedHeaders http.ResponseHeaders
	}{
		{
			msg: "no filter set",
			filters: &dataplane.HTTPFilters{
				RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{},
			},
			expectedHeaders: http.ResponseHeaders{},
		},
		{
			msg: "set filters correctly",
			filters: &dataplane.HTTPFilters{
				ResponseHeaderModifiers: &dataplane.HTTPHeaderFilter{
					Add: []dataplane.HTTPHeader{
						{
							Name:  "Accept-Encoding",
							Value: "gzip",
						},
						{
							Name:  "Authorization",
							Value: "my-auth",
						},
					},
					Set: []dataplane.HTTPHeader{
						{
							Name:  "Accept-Encoding",
							Value: "my-new-overwritten-value",
						},
					},
					Remove: []string{"Transfer-Encoding"},
				},
			},
			expectedHeaders: http.ResponseHeaders{
				Add: []http.Header{
					{
						Name:  "Accept-Encoding",
						Value: "gzip",
					},
					{
						Name:  "Authorization",
						Value: "my-auth",
					},
				},
				Set: []http.Header{
					{
						Name:  "Accept-Encoding",
						Value: "my-new-overwritten-value",
					},
				},
				Remove: []string{"Transfer-Encoding"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)

			headers := generateResponseHeaders(tc.filters)
			g.Expect(headers).To(Equal(tc.expectedHeaders))
		})
	}
}

func TestCreateIncludes(t *testing.T) {
	tests := []struct {
		name      string
		additions []dataplane.Addition
		includes  []string
	}{
		{
			name:      "no additions",
			additions: nil,
			includes:  nil,
		},
		{
			name: "additions",
			additions: []dataplane.Addition{
				{
					Bytes:      []byte("one"),
					Identifier: "one",
				},
				{
					Bytes:      []byte("two"),
					Identifier: "two",
				},
				{
					Bytes:      []byte("three"),
					Identifier: "three",
				},
			},
			includes: []string{
				includesFolder + "/one.conf",
				includesFolder + "/two.conf",
				includesFolder + "/three.conf",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			includes := createIncludes(test.additions)
			g.Expect(includes).To(Equal(test.includes))
		})
	}
}

func TestCreateAdditionFileResults(t *testing.T) {
	conf := dataplane.Configuration{
		HTTPServers: []dataplane.VirtualServer{
			{
				Additions: []dataplane.Addition{
					{
						Identifier: "include-1",
						Bytes:      []byte("include-1"),
					},
					{
						Identifier: "include-2",
						Bytes:      []byte("include-2"),
					},
				},
				PathRules: []dataplane.PathRule{
					{
						MatchRules: []dataplane.MatchRule{
							{
								Additions: []dataplane.Addition{
									{
										Identifier: "include-3",
										Bytes:      []byte("include-3"),
									},
									{
										Identifier: "include-4",
										Bytes:      []byte("include-4"),
									},
								},
							},
						},
					},
				},
			},
			{
				Additions: []dataplane.Addition{
					{
						Identifier: "include-1", // dupe
						Bytes:      []byte("include-1"),
					},
					{
						Identifier: "include-2", // dupe
						Bytes:      []byte("include-2"),
					},
				},
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				Additions: []dataplane.Addition{
					{
						Identifier: "include-1", // dupe
						Bytes:      []byte("include-1"),
					},
					{
						Identifier: "include-2", // dupe
						Bytes:      []byte("include-2"),
					},
				},
				PathRules: []dataplane.PathRule{
					{
						MatchRules: []dataplane.MatchRule{
							{
								Additions: []dataplane.Addition{
									{
										Identifier: "include-3",
										Bytes:      []byte("include-3"), // dupe
									},
									{
										Identifier: "include-5",
										Bytes:      []byte("include-5"), // dupe
									},
									{
										Identifier: "include-6",
										Bytes:      []byte("include-6"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results := createAdditionFileResults(conf)

	expResults := []executeResult{
		{
			dest: includesFolder + "/" + "include-1.conf",
			data: []byte("include-1"),
		},
		{
			dest: includesFolder + "/" + "include-2.conf",
			data: []byte("include-2"),
		},
		{
			dest: includesFolder + "/" + "include-3.conf",
			data: []byte("include-3"),
		},
		{
			dest: includesFolder + "/" + "include-4.conf",
			data: []byte("include-4"),
		},
		{
			dest: includesFolder + "/" + "include-5.conf",
			data: []byte("include-5"),
		},
		{
			dest: includesFolder + "/" + "include-6.conf",
			data: []byte("include-6"),
		},
	}

	g := NewWithT(t)

	g.Expect(results).To(ConsistOf(expResults))
}

func TestAdditionFilename(t *testing.T) {
	g := NewWithT(t)

	name := createAdditionFileName(dataplane.Addition{Identifier: "my-addition"})
	g.Expect(name).To(Equal(includesFolder + "/" + "my-addition.conf"))
}

func TestGetIPFamily(t *testing.T) {
	test := []struct {
		msg            string
		baseHTTPConfig dataplane.BaseHTTPConfig
		expected       http.IPFamily
	}{
		{
			msg:            "ipv4",
			baseHTTPConfig: dataplane.BaseHTTPConfig{IPFamily: dataplane.IPv4},
			expected:       http.IPFamily{IPv4: true, IPv6: false},
		},
		{
			msg:            "ipv6",
			baseHTTPConfig: dataplane.BaseHTTPConfig{IPFamily: dataplane.IPv6},
			expected:       http.IPFamily{IPv4: false, IPv6: true},
		},
		{
			msg:            "dual",
			baseHTTPConfig: dataplane.BaseHTTPConfig{IPFamily: dataplane.Dual},
			expected:       http.IPFamily{IPv4: true, IPv6: true},
		},
	}

	for _, tc := range test {
		t.Run(tc.msg, func(t *testing.T) {
			g := NewWithT(t)
			result := getIPFamily(tc.baseHTTPConfig)
			g.Expect(result).To(Equal(tc.expected))
		})
	}
}
