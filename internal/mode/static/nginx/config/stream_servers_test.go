package config

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
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
	g := NewWithT(t)

	results := executeStreamServers(conf)
	g.Expect(results).To(HaveLen(1))
	result := results[0]

	fmt.Println(string(result.data))

	g.Expect(result.dest).To(Equal(streamConfigFile))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(strings.Count(string(result.data), expSubStr)).To(Equal(expCount))
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

	expectedStreamServers := []stream.Server{
		{
			Listen:     getSocketName(conf.TLSServers[0].Port, conf.TLSServers[0].Hostname),
			ProxyPass:  conf.TLSServers[0].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     getSocketName(conf.TLSServers[1].Port, conf.TLSServers[1].Hostname),
			ProxyPass:  conf.TLSServers[1].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     getSocketName(conf.TLSServers[2].Port, conf.TLSServers[2].Hostname),
			ProxyPass:  conf.TLSServers[2].UpstreamName,
			SSLPreread: false,
		},
		{
			Listen:     fmt.Sprint(8081),
			Pass:       getVariableName(8081),
			SSLPreread: true,
		},
		{
			Listen:     fmt.Sprint(8080),
			Pass:       getVariableName(8080),
			SSLPreread: true,
		},
	}

	for i := range streamServers {
		g.Expect(streamServers[i]).To(Equal(expectedStreamServers[i]))
	}
}
