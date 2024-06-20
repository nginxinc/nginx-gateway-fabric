package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteStreamServers(t *testing.T) {
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
		"pass $dest8081;": 1,
		"pass $dest8080;": 1,
		"ssl_preread on;": 2,
		"proxy_pass":      3,
	}

	type assertion func(g *WithT, data string)

	expectedResults := map[string]assertion{
		streamConfigFile: func(g *WithT, data string) {
			for expSubStr, expCount := range expSubStrings {
				g.Expect(strings.Count(data, expSubStr)).To(Equal(expCount))
			}
		},
	}
	g := NewWithT(t)

	results := executeStreamServers(conf)
	g.Expect(results).To(HaveLen(len(expectedResults)))

	for _, res := range results {
		g.Expect(expectedResults).To(HaveKey(res.dest), "executeStreamServers returned unexpected result destination")

		assertData := expectedResults[res.dest]
		assertData(g, string(res.data))
	}
}

func TestCreateStreamServers(t *testing.T) {
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

	streamServers := createStreamServers(conf)

	g := NewWithT(t)

	g.Expect(streamServers).To(HaveLen(5))

	SSLPrereadCount := 0
	ProxyPassCount := 0

	for _, streamServer := range streamServers {
		if streamServer.SSLPreread {
			SSLPrereadCount++
		}
		if streamServer.ProxyPass {
			ProxyPassCount++
		}
	}

	g.Expect(SSLPrereadCount).To(Equal(2))
	g.Expect(ProxyPassCount).To(Equal(3))
}
