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
	t.Parallel()
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
		{
			Name: "up4-ipv6",
			Endpoints: []resolver.Endpoint{
				{
					Address: "2001:db8::1",
					Port:    80,
					IPv6:    true,
				},
			},
		},
	}

	expectedSubStrings := []string{
		"upstream up1",
		"upstream up2",
		"upstream up3",
		"upstream up4-ipv6",
		"upstream invalid-backend-ref",
		"server 10.0.0.0:80;",
		"server 11.0.0.0:80;",
		"server [2001:db8::1]:80",
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
	t.Parallel()
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
		{
			Name: "up4-ipv6",
			Endpoints: []resolver.Endpoint{
				{
					Address: "fd00:10:244:1::7",
					Port:    80,
					IPv6:    true,
				},
			},
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
			Name:     "up4-ipv6",
			ZoneSize: ossZoneSize,
			Servers: []http.UpstreamServer{
				{
					Address: "[fd00:10:244:1::7]:80",
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
	t.Parallel()
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
		{
			stateUpstream: dataplane.Upstream{
				Name: "endpoint-ipv6",
				Endpoints: []resolver.Endpoint{
					{
						Address: "fd00:10:244:1::7",
						Port:    80,
						IPv6:    true,
					},
				},
			},
			expectedUpstream: http.Upstream{
				Name:     "endpoint-ipv6",
				ZoneSize: ossZoneSize,
				Servers: []http.UpstreamServer{
					{
						Address: "[fd00:10:244:1::7]:80",
					},
				},
			},
			msg: "endpoint ipv6",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			result := gen.createUpstream(test.stateUpstream)
			g.Expect(result).To(Equal(test.expectedUpstream))
		})
	}
}

func TestCreateUpstreamPlus(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
				{
					Address: "2001:db8::1",
					IPv6:    true,
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
				{
					Address: "[2001:db8::1]:0",
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

func TestCreateStreamUpstream(t *testing.T) {
	t.Parallel()
	gen := GeneratorImpl{}
	up := dataplane.Upstream{
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
	}

	expectedUpstream := stream.Upstream{
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
	}

	g := NewWithT(t)
	result := gen.createStreamUpstream(up)
	g.Expect(result).To(Equal(expectedUpstream))
}

func TestCreateStreamUpstreamPlus(t *testing.T) {
	t.Parallel()
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
