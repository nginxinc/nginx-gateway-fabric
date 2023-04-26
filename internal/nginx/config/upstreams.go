package config

import (
	"fmt"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

var upstreamsTemplate = gotemplate.Must(gotemplate.New("upstreams").Parse(upstreamsTemplateText))

const (
	// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
	nginx502Server = "unix:/var/lib/nginx/nginx-502-server.sock"
	// nginx500Server is used as a server for the invalid backend ref upstream.
	nginx500Server = "unix:/var/lib/nginx/nginx-500-server.sock"
	// invalidBackendRef is used as an upstream name for invalid backend references.
	invalidBackendRef = "invalid-backend-ref"
)

func executeUpstreams(confs []dataplane.Configuration) []byte {
	var upstreams []http.Upstream

	for _, conf := range confs {
		upstreams = append(upstreams, createUpstreams(conf.Key, conf.Upstreams)...)
	}

	return execute(upstreamsTemplate, upstreams)
}

func createUpstreams(key string, upstreams []dataplane.Upstream) []http.Upstream {
	// capacity is the number of upstreams + 1 for the invalid backend ref upstream
	ups := make([]http.Upstream, 0, len(upstreams)+1)

	for _, u := range upstreams {
		ups = append(ups, createUpstream(key, u))
	}

	ups = append(ups, createInvalidBackendRefUpstream(key))

	return ups
}

func createUpstream(key string, up dataplane.Upstream) http.Upstream {
	if len(up.Endpoints) == 0 {
		return http.Upstream{
			Name: fmt.Sprintf("%s__%s", key, up.Name),
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
		Name:    fmt.Sprintf("%s__%s", key, up.Name),
		Servers: upstreamServers,
	}
}

func createInvalidBackendRefUpstream(key string) http.Upstream {
	return http.Upstream{
		Name: fmt.Sprintf("%s__%s", key, invalidBackendRef),
		Servers: []http.UpstreamServer{
			{
				Address: nginx500Server,
			},
		},
	}
}
