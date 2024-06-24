package config

import (
	"fmt"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var upstreamsTemplate = gotemplate.Must(gotemplate.New("upstreams").Parse(upstreamsTemplateText))

const (
	// nginx502Server is used as a backend for services that cannot be resolved (have no IP address).
	nginx502Server = "unix:/var/run/nginx/nginx-502-server.sock"
	// nginx500Server is used as a server for the invalid backend ref upstream.
	nginx500Server = "unix:/var/run/nginx/nginx-500-server.sock"
	// invalidBackendRef is used as an upstream name for invalid backend references.
	invalidBackendRef = "invalid-backend-ref"
	// ossZoneSize is the upstream zone size for nginx open source.
	ossZoneSize = "512k"
	// plusZoneSize is the upstream zone size for nginx plus.
	plusZoneSize = "1m"
)

func (g GeneratorImpl) executeUpstreams(conf dataplane.Configuration) []executeResult {
	upstreams := g.createUpstreams(conf.Upstreams)

	result := executeResult{
		dest: httpConfigFile,
		data: helpers.MustExecuteTemplate(upstreamsTemplate, upstreams),
	}

	return []executeResult{result}
}

func (g GeneratorImpl) createUpstreams(upstreams []dataplane.Upstream) []http.Upstream {
	// capacity is the number of upstreams + 1 for the invalid backend ref upstream
	ups := make([]http.Upstream, 0, len(upstreams)+1)

	for _, u := range upstreams {
		ups = append(ups, g.createUpstream(u))
	}

	ups = append(ups, createInvalidBackendRefUpstream())

	return ups
}

func (g GeneratorImpl) createUpstream(up dataplane.Upstream) http.Upstream {
	zoneSize := ossZoneSize
	if g.plus {
		zoneSize = plusZoneSize
	}

	if len(up.Endpoints) == 0 {
		return http.Upstream{
			Name:     up.Name,
			ZoneSize: zoneSize,
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
		Name:     up.Name,
		ZoneSize: zoneSize,
		Servers:  upstreamServers,
	}
}

func createInvalidBackendRefUpstream() http.Upstream {
	// ZoneSize is omitted since we will only ever proxy to one destination/backend.
	return http.Upstream{
		Name: invalidBackendRef,
		Servers: []http.UpstreamServer{
			{
				Address: nginx500Server,
			},
		},
	}
}
