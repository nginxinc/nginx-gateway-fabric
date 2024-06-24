package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteMaps(t *testing.T) {
	g := NewWithT(t)
	pathRules := []dataplane.PathRule{
		{
			MatchRules: []dataplane.MatchRule{
				{
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-add-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
				{
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-second-add-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
				{
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Set: []dataplane.HTTPHeader{
								{
									Name:  "my-set-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
			},
		},
	}

	conf := dataplane.Configuration{
		HTTPServers: []dataplane.VirtualServer{
			{
				PathRules: pathRules,
			},
			{
				PathRules: pathRules,
			},
			{
				IsDefault: true,
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				PathRules: pathRules,
			},
			{
				IsDefault: true,
			},
		},
	}

	expSubStrings := map[string]int{
		"map ${http_my_add_header} $my_add_header_header_var {": 1,
		"default '';":                2,
		"~.* ${http_my_add_header},": 1,
		"map ${http_my_second_add_header} $my_second_add_header_header_var {": 1,
		"~.* ${http_my_second_add_header},;":                                  1,
		"map ${http_my_set_header} $my_set_header_header_var {":               0,
	}

	mapResult := executeMaps(conf)
	g.Expect(mapResult).To(HaveLen(1))
	maps := string(mapResult[0].data)
	g.Expect(mapResult[0].dest).To(Equal(httpConfigFile))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(maps, expSubStr)))
	}
}

func TestBuildAddHeaderMaps(t *testing.T) {
	g := NewWithT(t)
	pathRules := []dataplane.PathRule{
		{
			MatchRules: []dataplane.MatchRule{
				{
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-add-header",
									Value: "some-value-123",
								},
								{
									Name:  "my-add-header",
									Value: "some-value-123",
								},
								{
									Name:  "my-second-add-header",
									Value: "some-value-123",
								},
							},
							Set: []dataplane.HTTPHeader{
								{
									Name:  "my-set-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
				{
					Filters: dataplane.HTTPFilters{
						RequestHeaderModifiers: &dataplane.HTTPHeaderFilter{
							Set: []dataplane.HTTPHeader{
								{
									Name:  "my-set-header",
									Value: "some-value-123",
								},
							},
							Add: []dataplane.HTTPHeader{
								{
									Name:  "my-add-header",
									Value: "some-value-123",
								},
							},
						},
					},
				},
			},
		},
	}

	testServers := []dataplane.VirtualServer{
		{
			PathRules: pathRules,
		},
		{
			PathRules: pathRules,
		},
		{
			IsDefault: true,
		},
	}
	expectedMap := []http.Map{
		{
			Source:   "${http_my_add_header}",
			Variable: "$my_add_header_header_var",
			Parameters: []http.MapParameter{
				{Value: "default", Result: "''"},
				{
					Value:  "~.*",
					Result: "${http_my_add_header},",
				},
			},
		},
		{
			Source:   "${http_my_second_add_header}",
			Variable: "$my_second_add_header_header_var",
			Parameters: []http.MapParameter{
				{Value: "default", Result: "''"},
				{
					Value:  "~.*",
					Result: "${http_my_second_add_header},",
				},
			},
		},
	}
	maps := buildAddHeaderMaps(testServers)

	g.Expect(maps).To(ConsistOf(expectedMap))
}

func TestExecuteStreamMaps(t *testing.T) {
	g := NewWithT(t)
	conf := dataplane.Configuration{
		TLSServers: []dataplane.Layer4Server{
			{
				Hostname:     "example.com",
				Port:         8081,
				UpstreamName: "backend1",
			},
			{
				Hostname:     "example.com",
				Port:         8080,
				UpstreamName: "backend1",
			},
			{
				Hostname:     "cafe.example.com",
				Port:         8080,
				UpstreamName: "backend2",
			},
		},
	}

	expSubStrings := map[string]int{
		"example.com unix:/var/run/nginx/example.com8081.sock;":           1,
		"example.com unix:/var/run/nginx/example.com8080.sock;":           1,
		"cafe.example.com unix:/var/run/nginx/cafe.example.com8080.sock;": 1,
	}

	results := executeStreamMaps(conf)
	g.Expect(results).To(HaveLen(1))
	result := results[0]

	g.Expect(result.dest).To(Equal(streamConfigFile))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(strings.Count(string(result.data), expSubStr)).To(Equal(expCount))
	}
}

func TestCreateStreamMaps(t *testing.T) {
	g := NewWithT(t)
	conf := dataplane.Configuration{
		TLSServers: []dataplane.Layer4Server{
			{
				Hostname:     "example.com",
				Port:         8081,
				UpstreamName: "backend1",
			},
			{
				Hostname:     "example.com",
				Port:         8080,
				UpstreamName: "backend1",
			},
			{
				Hostname:     "cafe.example.com",
				Port:         8080,
				UpstreamName: "backend2",
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				Hostname: "app.example.com",
				Port:     8080,
			},
		},
	}

	maps := createStreamMaps(conf)

	expectedMaps := []http.Map{
		{
			Source:   "$ssl_preread_server_name",
			Variable: getVariableName(8081),
			Parameters: []http.MapParameter{
				{Value: "example.com", Result: "unix:/var/run/nginx/example.com8081.sock"},
			},
		},
		{
			Source:   "$ssl_preread_server_name",
			Variable: getVariableName(8080),
			Parameters: []http.MapParameter{
				{Value: "example.com", Result: "unix:/var/run/nginx/example.com8080.sock"},
				{Value: "cafe.example.com", Result: "unix:/var/run/nginx/cafe.example.com8080.sock"},
				{Value: "app.example.com", Result: "unix:/var/run/nginx/app.example.com8080.sock"},
			},
		},
	}
	g.Expect(maps).To(Equal(expectedMaps))
}
