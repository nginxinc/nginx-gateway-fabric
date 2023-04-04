package config

import (
	"fmt"
	"text/template"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

var upstreamsTemplate = template.Must(template.New("upstreams").Parse(upstreamsTemplateText))

const (
	// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
	nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"
	// nginx500Server is used as a server for the invalid backend ref upstream.
	nginx500Server = "unix:/var/lib/nginx/nginx-500-server.sock"
	// invalidBackendRef is used as an upstream name for invalid backend references.
	invalidBackendRef = "invalid-backend-ref"
)

func executeUpstreams(conf dataplane.Configuration) []byte {
	upstreams := createUpstreams(conf.Upstreams)

	return execute(upstreamsTemplate, upstreams)
}

func createUpstreams(upstreams []dataplane.Upstream) []http.Upstream {
	// capacity is the number of upstreams + 1 for the invalid backend ref upstream
	ups := make([]http.Upstream, 0, len(upstreams)+1)

	for _, u := range upstreams {
		ups = append(ups, createUpstream(u))
	}

	ups = append(ups, createInvalidBackendRefUpstream())

	return ups
}

func createUpstream(up dataplane.Upstream) http.Upstream {
	if len(up.Endpoints) == 0 {
		return http.Upstream{
			Name: up.Name,
			Servers: []http.UpstreamServer{
				{
					Address: nginx502Server,
				},
			},
		}
	}

	upstreamServers := make([]http.UpstreamServer, len(up.Endpoints))
	for idx, ep := range up.Endpoints {
		upstreamServers[idx] = http.UpstreamServer{
			Address: fmt.Sprintf("%s:%d", ep.Address, ep.Port),
		}
	}

	return http.Upstream{
		Name:    up.Name,
		Servers: upstreamServers,
	}
}

func createInvalidBackendRefUpstream() http.Upstream {
	return http.Upstream{
		Name: invalidBackendRef,
		Servers: []http.UpstreamServer{
			{
				Address: nginx500Server,
			},
		},
	}
}
