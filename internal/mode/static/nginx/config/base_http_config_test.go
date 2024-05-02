package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteBaseHttp(t *testing.T) {
	conf := dataplane.Configuration{
		BaseHTTPConfig: dataplane.BaseHTTPConfig{
			HTTP2: true,
		},
	}

	g := NewWithT(t)
	expSubStrings := map[string]int{
		"http2 on;": 1,
	}

	for expSubStr, expCount := range expSubStrings {
		res := executeBaseHTTPConfig(conf)
		g.Expect(res).To(HaveLen(1))
		g.Expect(expCount).To(Equal(strings.Count(string(res[0].data), expSubStr)))
	}
}

func TestExecuteBaseHttpEmpty(t *testing.T) {
	conf := dataplane.Configuration{
		BaseHTTPConfig: dataplane.BaseHTTPConfig{
			HTTP2: false,
		},
	}

	g := NewWithT(t)
	expSubStrings := map[string]int{
		"http2 on;": 0,
	}

	for expSubStr, expCount := range expSubStrings {
		res := executeBaseHTTPConfig(conf)
		g.Expect(res).To(HaveLen(1))
		g.Expect(expCount).To(Equal(strings.Count(string(res[0].data), expSubStr)))
	}
}
