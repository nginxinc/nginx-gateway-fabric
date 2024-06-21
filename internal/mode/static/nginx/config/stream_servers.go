package config

import (
	"fmt"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var streamServersTemplate = gotemplate.Must(gotemplate.New("servers").Parse(streamServersTemplateText))

func executeStreamServers(conf dataplane.Configuration) []executeResult {
	streamServers := createStreamServers(conf)

	streamServerResult := executeResult{
		dest: streamConfigFile,
		data: helpers.MustExecuteTemplate(streamServersTemplate, streamServers),
	}

	result := []executeResult{
		streamServerResult,
	}

	return result
}

func createStreamServers(conf dataplane.Configuration) []stream.Server {
	streamServers := make([]stream.Server, 0, len(conf.TLSServers)*2)
	for _, server := range conf.TLSServers {
		streamServers = append(streamServers, stream.Server{
			Listen:      "unix:/var/lib/nginx/" + server.Hostname + fmt.Sprint(server.Port) + ".sock",
			Destination: server.UpstreamName,
			ProxyPass:   true,
			SSLPreread:  false,
		})
	}

	ports := make(map[int32]bool, len(streamServers))

	for _, server := range conf.TLSServers {
		if ports[server.Port] {
			continue
		}
		ports[server.Port] = true
		streamServers = append(streamServers, stream.Server{
			Listen:      fmt.Sprint(server.Port),
			Destination: "$dest" + fmt.Sprint(server.Port),
			ProxyPass:   false,
			SSLPreread:  true,
		})
	}

	return streamServers
}
