package config

import (
	"fmt"

	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

// ConvertEndpoints converts a list of Endpoints into a list of NGINX Plus SDK UpstreamServers.
func ConvertEndpoints(eps []resolver.Endpoint) []ngxclient.UpstreamServer {
	servers := make([]ngxclient.UpstreamServer, 0, len(eps))

	for _, ep := range eps {
		var port string
		if ep.Port != 0 {
			port = fmt.Sprintf(":%d", ep.Port)
		}

		format := "%s%s"
		if ep.IPv6 {
			format = "[%s]%s"
		}

		server := ngxclient.UpstreamServer{
			Server: fmt.Sprintf(format, ep.Address, port),
		}

		servers = append(servers, server)
	}

	return servers
}
