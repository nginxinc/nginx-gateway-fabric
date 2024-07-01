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
	if len(conf.TLSPassthroughServers) == 0 {
		return nil
	}

	streamServers := make([]stream.Server, 0, len(conf.TLSPassthroughServers)*2)
	portSet := make(map[int32]struct{})
	upstreamNameSet := make(map[string]struct{})

	for _, u := range conf.StreamUpstreams {
		upstreamNameSet[u.Name] = struct{}{}
	}

	for _, server := range conf.TLSPassthroughServers {
		if _, nameExists := upstreamNameSet[server.UpstreamName]; nameExists {
			streamServers = append(streamServers, stream.Server{
				Listen:    getSocketNameTLS(server.Port, server.Hostname),
				ProxyPass: server.UpstreamName,
			})
		}

		_, inPortSet := portSet[server.Port]

		if inPortSet {
			continue
		}

		portSet[server.Port] = struct{}{}
		streamServers = append(streamServers, stream.Server{
			Listen:     fmt.Sprint(server.Port),
			Pass:       getTLSPassthroughVarName(server.Port),
			SSLPreread: true,
		})
	}

	return streamServers
}
