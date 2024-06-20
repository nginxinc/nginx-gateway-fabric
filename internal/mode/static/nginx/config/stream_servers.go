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
	streamServers := make([]stream.Server, 0)
	for _, t := range conf.TLSServers {
		streamServers = append(streamServers, stream.Server{
			Listen:      "unix:/var/lib/nginx/" + t.Hostname + fmt.Sprint(t.Port) + ".sock",
			Destination: t.UpstreamName,
			ProxyPass:   true,
			SSLPreread:  false,
		})
	}

	ports := make(map[int32]bool)

	for _, t := range conf.TLSServers {
		if ports[t.Port] {
			continue
		}
		ports[t.Port] = true
		streamServers = append(streamServers, stream.Server{
			Listen:      fmt.Sprint(t.Port),
			Destination: "$dest" + fmt.Sprint(t.Port),
			ProxyPass:   false,
			SSLPreread:  true,
		})
	}

	return streamServers
}
