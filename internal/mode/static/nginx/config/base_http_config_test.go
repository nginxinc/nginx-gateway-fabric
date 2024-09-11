package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteBaseHttp(t *testing.T) {
	t.Parallel()
	confOn := dataplane.Configuration{
		BaseHTTPConfig: dataplane.BaseHTTPConfig{
			HTTP2: true,
		},
	}

	confOff := dataplane.Configuration{
		BaseHTTPConfig: dataplane.BaseHTTPConfig{
			HTTP2: false,
		},
	}

	expSubStr := "http2 on;"

	tests := []struct {
		name     string
		conf     dataplane.Configuration
		expCount int
	}{
		{
			name:     "http2 on",
			conf:     confOn,
			expCount: 1,
		},
		{
			name:     "http2 off",
			expCount: 0,
			conf:     confOff,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			res := executeBaseHTTPConfig(test.conf)
			g.Expect(res).To(HaveLen(1))
			g.Expect(test.expCount).To(Equal(strings.Count(string(res[0].data), expSubStr)))
			g.Expect(strings.Count(string(res[0].data), "map $http_host $gw_api_compliant_host {")).To(Equal(1))
			g.Expect(strings.Count(string(res[0].data), "map $http_upgrade $connection_upgrade {")).To(Equal(1))
			g.Expect(strings.Count(string(res[0].data), "map $request_uri $request_uri_path {")).To(Equal(1))
		})
	}
}
