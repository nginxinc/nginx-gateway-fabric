package config

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteStreamServers(t *testing.T) {
	conf := dataplane.Configuration{
		TLSPassthroughServers: []dataplane.Layer4VirtualServer{
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
		StreamUpstreams: []dataplane.Upstream{
			{
				Name: "backend1",
			},
			{
				Name: "backend2",
			},
		},
	}

	expSubStrings := map[string]int{
		"pass $dest8081;": 1,
		"pass $dest8080;": 1,
		"ssl_preread on;": 2,
		"proxy_pass":      3,
	}
	g := NewWithT(t)

	results := executeStreamServers(conf)
	g.Expect(results).To(HaveLen(1))
	result := results[0]

	g.Expect(result.dest).To(Equal(streamConfigFile))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(strings.Count(string(result.data), expSubStr)).To(Equal(expCount))
	}
}

func TestCreateStreamServers(t *testing.T) {
	conf := dataplane.Configuration{
		TLSPassthroughServers: []dataplane.Layer4VirtualServer{
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
			{
				Hostname:     "wrong.example.com",
				Port:         8081,
				UpstreamName: "",
			},
		},
		StreamUpstreams: []dataplane.Upstream{
			{
				Name: "backend1",
			},
			{
				Name: "backend2",
			},
		},
	}

	streamServers := createStreamServers(conf)

	g := NewWithT(t)

	expectedStreamServers := []stream.Server{
		{
			Listen:     getSocketNameTLS(conf.TLSPassthroughServers[0].Port, conf.TLSPassthroughServers[0].Hostname),
			ProxyPass:  conf.TLSPassthroughServers[0].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     getSocketNameTLS(conf.TLSPassthroughServers[1].Port, conf.TLSPassthroughServers[1].Hostname),
			ProxyPass:  conf.TLSPassthroughServers[1].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     getSocketNameTLS(conf.TLSPassthroughServers[2].Port, conf.TLSPassthroughServers[2].Hostname),
			ProxyPass:  conf.TLSPassthroughServers[2].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     fmt.Sprint(8081),
			Pass:       getTLSPassthroughVarName(8081),
			SSLPreread: true,
		},
		{
			Listen:     fmt.Sprint(8080),
			Pass:       getTLSPassthroughVarName(8080),
			SSLPreread: true,
		},
	}

	sort.Slice(expectedStreamServers, func(i, j int) bool {
		return expectedStreamServers[i].Listen < expectedStreamServers[j].Listen
	})
	sort.Slice(streamServers, func(i, j int) bool { return streamServers[i].Listen < streamServers[j].Listen })

	g.Expect(streamServers).To(Equal(expectedStreamServers))
}

func TestCreateStreamServersWithNone(t *testing.T) {
	conf := dataplane.Configuration{
		TLSPassthroughServers: nil,
	}

	streamServers := createStreamServers(conf)

	g := NewWithT(t)

	var expectedStreamServers []stream.Server

	g.Expect(streamServers).To(Equal(expectedStreamServers))
}
