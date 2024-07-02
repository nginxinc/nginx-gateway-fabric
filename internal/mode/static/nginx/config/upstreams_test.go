package config

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/stream"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

func TestExecuteUpstreams(t *testing.T) {
	gen := GeneratorImpl{}
	stateUpstreams := []dataplane.Upstream{
		{
			Name: "up1",
			Endpoints: []resolver.Endpoint{
				{
					Address: "10.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name: "up2",
			Endpoints: []resolver.Endpoint{
				{
					Address: "11.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name:      "up3",
			Endpoints: []resolver.Endpoint{},
		},
	}

	expectedSubStrings := []string{
		"upstream up1",
		"upstream up2",
		"upstream up3",
		"upstream invalid-backend-ref",
		"server 10.0.0.0:80;",
		"server 11.0.0.0:80;",
		"server unix:/var/run/nginx/nginx-502-server.sock;",
	}

	upstreamResults := gen.executeUpstreams(dataplane.Configuration{Upstreams: stateUpstreams})
	g := NewWithT(t)
	g.Expect(upstreamResults).To(HaveLen(1))
	upstreams := string(upstreamResults[0].data)

	g.Expect(upstreamResults[0].dest).To(Equal(httpConfigFile))
	for _, expSubString := range expectedSubStrings {
		g.Expect(upstreams).To(ContainSubstring(expSubString))
	}
}

func TestCreateUpstreams(t *testing.T) {
	gen := GeneratorImpl{}
	stateUpstreams := []dataplane.Upstream{
		{
			Name: "up1",
			Endpoints: []resolver.Endpoint{
				{
					Address: "10.0.0.0",
					Port:    80,
				},
				{
					Address: "10.0.0.1",
					Port:    80,
				},
				{
					Address: "10.0.0.2",
					Port:    80,
				},
			},
		},
		{
			Name: "up2",
			Endpoints: []resolver.Endpoint{
				{
					Address: "11.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name:      "up3",
			Endpoints: []resolver.Endpoint{},
		},
	}

	expUpstreams := []http.Upstream{
		{
			Name:     "up1",
			ZoneSize: ossZoneSize,
			Servers: []http.UpstreamServer{
				{
					Address: "10.0.0.0:80",
				},
				{
					Address: "10.0.0.1:80",
				},
				{
					Address: "10.0.0.2:80",
				},
			},
		},
		{
			Name:     "up2",
			ZoneSize: ossZoneSize,
			Servers: []http.UpstreamServer{
				{
					Address: "11.0.0.0:80",
				},
			},
		},
		{
			Name:     "up3",
			ZoneSize: ossZoneSize,
			Servers: []http.UpstreamServer{
				{
					Address: nginx502Server,
				},
			},
		},
		{
			Name: invalidBackendRef,
			Servers: []http.UpstreamServer{
				{
					Address: nginx500Server,
				},
			},
		},
	}

	g := NewWithT(t)
	result := gen.createUpstreams(stateUpstreams)
	g.Expect(result).To(Equal(expUpstreams))
}

func TestCreateUpstream(t *testing.T) {
	gen := GeneratorImpl{}
	tests := []struct {
		msg              string
		stateUpstream    dataplane.Upstream
		expectedUpstream http.Upstream
	}{
		{
			stateUpstream: dataplane.Upstream{
				Name:      "nil-endpoints",
				Endpoints: nil,
			},
			expectedUpstream: http.Upstream{
				Name:     "nil-endpoints",
				ZoneSize: ossZoneSize,
				Servers: []http.UpstreamServer{
					{
						Address: nginx502Server,
					},
				},
			},
			msg: "nil endpoints",
		},
		{
			stateUpstream: dataplane.Upstream{
				Name:      "no-endpoints",
				Endpoints: []resolver.Endpoint{},
			},
			expectedUpstream: http.Upstream{
				Name:     "no-endpoints",
				ZoneSize: ossZoneSize,
				Servers: []http.UpstreamServer{
					{
						Address: nginx502Server,
					},
				},
			},
			msg: "no endpoints",
		},
		{
			stateUpstream: dataplane.Upstream{
				Name: "multiple-endpoints",
				Endpoints: []resolver.Endpoint{
					{
						Address: "10.0.0.1",
						Port:    80,
					},
					{
						Address: "10.0.0.2",
						Port:    80,
					},
					{
						Address: "10.0.0.3",
						Port:    80,
					},
				},
			},
			expectedUpstream: http.Upstream{
				Name:     "multiple-endpoints",
				ZoneSize: ossZoneSize,
				Servers: []http.UpstreamServer{
					{
						Address: "10.0.0.1:80",
					},
					{
						Address: "10.0.0.2:80",
					},
					{
						Address: "10.0.0.3:80",
					},
				},
			},
			msg: "multiple endpoints",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)
			result := gen.createUpstream(test.stateUpstream)
			g.Expect(result).To(Equal(test.expectedUpstream))
		})
	}
}

func TestCreateUpstreamPlus(t *testing.T) {
	gen := GeneratorImpl{plus: true}

	stateUpstream := dataplane.Upstream{
		Name: "multiple-endpoints",
		Endpoints: []resolver.Endpoint{
			{
				Address: "10.0.0.1",
				Port:    80,
			},
		},
	}
	expectedUpstream := http.Upstream{
		Name:     "multiple-endpoints",
		ZoneSize: plusZoneSize,
		Servers: []http.UpstreamServer{
			{
				Address: "10.0.0.1:80",
			},
		},
	}

	result := gen.createUpstream(stateUpstream)

	g := NewWithT(t)
	g.Expect(result).To(Equal(expectedUpstream))
}

func TestExecuteStreamUpstreams(t *testing.T) {
	gen := GeneratorImpl{}
	stateUpstreams := []dataplane.Upstream{
		{
			Name: "up1",
			Endpoints: []resolver.Endpoint{
				{
					Address: "10.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name: "up2",
			Endpoints: []resolver.Endpoint{
				{
					Address: "11.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name:      "up3",
			Endpoints: []resolver.Endpoint{},
		},
	}

	expectedSubStrings := []string{
		"upstream up1",
		"upstream up2",
		"server 10.0.0.0:80;",
		"server 11.0.0.0:80;",
	}

	upstreamResults := gen.executeStreamUpstreams(dataplane.Configuration{StreamUpstreams: stateUpstreams})
	g := NewWithT(t)
	g.Expect(upstreamResults).To(HaveLen(1))
	upstreams := string(upstreamResults[0].data)

	g.Expect(upstreamResults[0].dest).To(Equal(streamConfigFile))
	for _, expSubString := range expectedSubStrings {
		g.Expect(upstreams).To(ContainSubstring(expSubString))
	}
}

func TestCreateStreamUpstreams(t *testing.T) {
	gen := GeneratorImpl{}
	stateUpstreams := []dataplane.Upstream{
		{
			Name: "up1",
			Endpoints: []resolver.Endpoint{
				{
					Address: "10.0.0.0",
					Port:    80,
				},
				{
					Address: "10.0.0.1",
					Port:    80,
				},
				{
					Address: "10.0.0.2",
					Port:    80,
				},
			},
		},
		{
			Name: "up2",
			Endpoints: []resolver.Endpoint{
				{
					Address: "11.0.0.0",
					Port:    80,
				},
			},
		},
		{
			Name:      "up3",
			Endpoints: []resolver.Endpoint{},
		},
	}

	expUpstreams := []stream.Upstream{
		{
			Name:     "up1",
			ZoneSize: ossZoneSize,
			Servers: []stream.UpstreamServer{
				{
					Address: "10.0.0.0:80",
				},
				{
					Address: "10.0.0.1:80",
				},
				{
					Address: "10.0.0.2:80",
				},
			},
		},
		{
			Name:     "up2",
			ZoneSize: ossZoneSize,
			Servers: []stream.UpstreamServer{
				{
					Address: "11.0.0.0:80",
				},
			},
		},
	}

	g := NewWithT(t)
	result := gen.createStreamUpstreams(stateUpstreams)
	g.Expect(result).To(Equal(expUpstreams))
}

func TestCreatStreamUpstream(t *testing.T) {
	gen := GeneratorImpl{}
	tests := []struct {
		msg              string
		stateUpstream    dataplane.Upstream
		expectedUpstream stream.Upstream
	}{
		{
			stateUpstream: dataplane.Upstream{
				Name: "multiple-endpoints",
				Endpoints: []resolver.Endpoint{
					{
						Address: "10.0.0.1",
						Port:    80,
					},
					{
						Address: "10.0.0.2",
						Port:    80,
					},
					{
						Address: "10.0.0.3",
						Port:    80,
					},
				},
			},
			expectedUpstream: stream.Upstream{
				Name:     "multiple-endpoints",
				ZoneSize: ossZoneSize,
				Servers: []stream.UpstreamServer{
					{
						Address: "10.0.0.1:80",
					},
					{
						Address: "10.0.0.2:80",
					},
					{
						Address: "10.0.0.3:80",
					},
				},
			},
			msg: "multiple endpoints",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)
			result := gen.createStreamUpstream(test.stateUpstream)
			g.Expect(result).To(Equal(test.expectedUpstream))
		})
	}
}

func TestCreateStreamUpstreamPlus(t *testing.T) {
	gen := GeneratorImpl{plus: true}

	stateUpstream := dataplane.Upstream{
		Name: "multiple-endpoints",
		Endpoints: []resolver.Endpoint{
			{
				Address: "10.0.0.1",
				Port:    80,
			},
		},
	}
	expectedUpstream := stream.Upstream{
		Name:     "multiple-endpoints",
		ZoneSize: plusZoneSize,
		Servers: []stream.UpstreamServer{
			{
				Address: "10.0.0.1:80",
			},
		},
	}

	result := gen.createStreamUpstream(stateUpstream)

	g := NewWithT(t)
	g.Expect(result).To(Equal(expectedUpstream))
}
