package config

import (
	"testing"

	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/resolver"
)

func TestConvertEndpoints(t *testing.T) {
	endpoints := []resolver.Endpoint{
		{
			Address: "1.2.3.4",
			Port:    80,
		},
		{
			Address: "5.6.7.8",
			Port:    0,
		},
	}

	expUpstreams := []ngxclient.UpstreamServer{
		{
			Server: "1.2.3.4:80",
		},
		{
			Server: "5.6.7.8",
		},
	}

	g := NewWithT(t)
	g.Expect(ConvertEndpoints(endpoints)).To(Equal(expUpstreams))
}
