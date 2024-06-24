package config

import (
	"fmt"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var streamServersTemplate = gotemplate.Must(gotemplate.New("streamServers").Parse(streamServersTemplateText))

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
			Listen:     getSocketName(server.Port, server.Hostname),
			ProxyPass:  server.UpstreamName,
			SSLPreread: false,
		})
	}

	portSet := make(map[int32]struct{}, len(streamServers))

	for _, server := range conf.TLSServers {
		_, inPortSet := portSet[server.Port]
		if inPortSet {
			continue
		}
		portSet[server.Port] = struct{}{}
		streamServers = append(streamServers, stream.Server{
			Listen:     fmt.Sprint(server.Port),
			Pass:       getVariableName(server.Port),
			SSLPreread: true,
		})
	}

	return streamServers
}
