package config

import (
	"fmt"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var upstreamsTemplate = gotemplate.Must(gotemplate.New("upstreams").Parse(upstreamsTemplateText))

const (
	// nginx503Server is used as a backend for services that cannot be resolved (have no IP address).
	nginx503Server = "unix:/var/run/nginx/nginx-503-server.sock"
	// nginx500Server is used as a server for the invalid backend ref upstream.
	nginx500Server = "unix:/var/run/nginx/nginx-500-server.sock"
	// invalidBackendRef is used as an upstream name for invalid backend references.
	invalidBackendRef = "invalid-backend-ref"
	// ossZoneSize is the upstream zone size for nginx open source.
	ossZoneSize = "512k"
	// plusZoneSize is the upstream zone size for nginx plus.
	plusZoneSize = "1m"
	// ossZoneSize is the upstream zone size for nginx open source.
	ossZoneSizeStream = "512k"
	// plusZoneSize is the upstream zone size for nginx plus.
	plusZoneSizeStream = "1m"
	// stateDir is the directory for storing state files.
	stateDir = "/var/lib/nginx/state"
)

func (g GeneratorImpl) executeUpstreams(conf dataplane.Configuration) []executeResult {
	upstreams := g.createUpstreams(conf.Upstreams)

	result := executeResult{
		dest: httpConfigFile,
		data: helpers.MustExecuteTemplate(upstreamsTemplate, upstreams),
	}

	return []executeResult{result}
}

func (g GeneratorImpl) executeStreamUpstreams(conf dataplane.Configuration) []executeResult {
	upstreams := g.createStreamUpstreams(conf.StreamUpstreams)

	result := executeResult{
		dest: streamConfigFile,
		data: helpers.MustExecuteTemplate(upstreamsTemplate, upstreams),
	}

	return []executeResult{result}
}

func (g GeneratorImpl) createStreamUpstreams(upstreams []dataplane.Upstream) []stream.Upstream {
	ups := make([]stream.Upstream, 0, len(upstreams))

	for _, u := range upstreams {
		if len(u.Endpoints) != 0 {
			ups = append(ups, g.createStreamUpstream(u))
		}
	}

	return ups
}

func (g GeneratorImpl) createStreamUpstream(up dataplane.Upstream) stream.Upstream {
	var stateFile string
	zoneSize := ossZoneSizeStream
	if g.plus {
		zoneSize = plusZoneSizeStream
		stateFile = fmt.Sprintf("%s/%s.conf", stateDir, up.Name)
	}

	upstreamServers := make([]stream.UpstreamServer, len(up.Endpoints))
	for idx, ep := range up.Endpoints {
		format := "%s:%d"
		if ep.IPv6 {
			format = "[%s]:%d"
		}
		upstreamServers[idx] = stream.UpstreamServer{
			Address: fmt.Sprintf(format, ep.Address, ep.Port),
		}
	}

	return stream.Upstream{
		Name:      up.Name,
		ZoneSize:  zoneSize,
		StateFile: stateFile,
		Servers:   upstreamServers,
	}
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
	var stateFile string
	zoneSize := ossZoneSize
	if g.plus {
		zoneSize = plusZoneSize
		stateFile = fmt.Sprintf("%s/%s.conf", stateDir, up.Name)
	}

	if len(up.Endpoints) == 0 {
		return http.Upstream{
			Name:      up.Name,
			ZoneSize:  zoneSize,
			StateFile: stateFile,
			Servers: []http.UpstreamServer{
				{
					Address: nginx503Server,
				},
			},
		}
	}

	upstreamServers := make([]http.UpstreamServer, len(up.Endpoints))
	for idx, ep := range up.Endpoints {
		format := "%s:%d"
		if ep.IPv6 {
			format = "[%s]:%d"
		}
		upstreamServers[idx] = http.UpstreamServer{
			Address: fmt.Sprintf(format, ep.Address, ep.Port),
		}
	}

	return http.Upstream{
		Name:      up.Name,
		ZoneSize:  zoneSize,
		StateFile: stateFile,
		Servers:   upstreamServers,
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
