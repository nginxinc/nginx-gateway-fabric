package config_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

// Note: this test only verifies that Generate() returns a byte array with upstream, server, and split_client blocks.
// It does not test the correctness of those blocks. That functionality is covered by other tests in this package.
func TestGenerate(t *testing.T) {
	bg := dataplane.BackendGroup{
		Source:  types.NamespacedName{Namespace: "test", Name: "hr"},
		RuleIdx: 0,
		Backends: []dataplane.Backend{
			{UpstreamName: "test", Valid: true, Weight: 1},
			{UpstreamName: "test2", Valid: true, Weight: 1},
		},
	}

	conf := dataplane.Configuration{
		HTTPServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
				Port:      80,
			},
			{
				Hostname: "example.com",
				Port:     80,
			},
		},
		SSLServers: []dataplane.VirtualServer{
			{
				IsDefault: true,
				Port:      443,
			},
			{
				Hostname: "example.com",
				SSL: &dataplane.SSL{
					CertificatePath: "/etc/nginx/secrets/default",
				},
				Port: 443,
			},
		},
		Upstreams: []dataplane.Upstream{
			{
				Name:      "up",
				Endpoints: nil,
			},
		},
		BackendGroups: []dataplane.BackendGroup{bg},
	}
	g := NewGomegaWithT(t)

	generator := config.NewGeneratorImpl()
	cfg := string(generator.Generate(conf))

	g.Expect(cfg).To(ContainSubstring("listen 80"))
	g.Expect(cfg).To(ContainSubstring("listen 443"))
	g.Expect(cfg).To(ContainSubstring("upstream"))
	g.Expect(cfg).To(ContainSubstring("split_clients"))
}
