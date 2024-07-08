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

	streamServerConfig := stream.ServerConfig{
		Servers:  streamServers,
		IPFamily: getIPFamily(conf.BaseHTTPConfig),
	}

	streamServerResult := executeResult{
		dest: streamConfigFile,
		data: helpers.MustExecuteTemplate(streamServersTemplate, streamServerConfig),
	}

	return []executeResult{
		streamServerResult,
	}
}

func createStreamServers(conf dataplane.Configuration) []stream.Server {
	if len(conf.TLSPassthroughServers) == 0 {
		return nil
	}

	streamServers := make([]stream.Server, 0, len(conf.TLSPassthroughServers)*2)
	portSet := make(map[int32]struct{})
	upstreams := make(map[string]dataplane.Upstream)

	for _, u := range conf.StreamUpstreams {
		upstreams[u.Name] = u
	}

	for _, server := range conf.TLSPassthroughServers {
		if u, ok := upstreams[server.UpstreamName]; ok && server.UpstreamName != "" {
			if server.Hostname != "" && len(u.Endpoints) > 0 {
				streamServers = append(streamServers, stream.Server{
					Listen:    getSocketNameTLS(server.Port, server.Hostname),
					ProxyPass: server.UpstreamName,
					IsSocket:  true,
				})
			}
		}

		if _, inPortSet := portSet[server.Port]; inPortSet {
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
