package config_test

import (
	"fmt"
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

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
					KeyPairID: "test-keypair",
				},
				Port: 443,
			},
		},
		TLSPassthroughServers: []dataplane.Layer4VirtualServer{
			{
				Hostname:     "app.example.com",
				Port:         443,
				UpstreamName: "stream_up",
			},
		},
		Upstreams: []dataplane.Upstream{
			{
				Name:      "up",
				Endpoints: nil,
			},
		},
		StreamUpstreams: []dataplane.Upstream{
			{
				Name:      "stream_up",
				Endpoints: nil,
			},
		},
		BackendGroups: []dataplane.BackendGroup{bg},
		SSLKeyPairs: map[dataplane.SSLKeyPairID]dataplane.SSLKeyPair{
			"test-keypair": {
				Cert: []byte("test-cert"),
				Key:  []byte("test-key"),
			},
		},
		CertBundles: map[dataplane.CertBundleID]dataplane.CertBundle{
			"test-certbundle": []byte("test-cert"),
		},
		Telemetry: dataplane.Telemetry{
			Endpoint:    "1.2.3.4:123",
			ServiceName: "ngf:gw-ns:gw-name:my-name",
			Interval:    "5s",
			BatchSize:   512,
			BatchCount:  4,
		},
		BaseHTTPConfig: dataplane.BaseHTTPConfig{
			HTTP2: true,
		},
	}
	g := NewWithT(t)

	var plus bool
	generator := config.NewGeneratorImpl(plus)

	files := generator.Generate(conf)

	g.Expect(files).To(HaveLen(7))
	arrange := func(i, j int) bool {
		return files[i].Path < files[j].Path
	}
	sort.Slice(files, arrange)

	g.Expect(files[0].Type).To(Equal(file.TypeRegular))
	g.Expect(files[0].Path).To(Equal("/etc/nginx/conf.d/config-version.conf"))
	configVersion := string(files[0].Content)
	g.Expect(configVersion).To(ContainSubstring(fmt.Sprintf("return 200 %d", conf.Version)))

	g.Expect(files[1].Type).To(Equal(file.TypeRegular))
	g.Expect(files[1].Path).To(Equal("/etc/nginx/conf.d/http.conf"))
	httpCfg := string(files[1].Content) // converting to string so that on failure gomega prints strings not byte arrays
	// Note: this only verifies that Generate() returns a byte array with upstream, server, and split_client blocks.
	// It does not test the correctness of those blocks. That functionality is covered by other tests in this package.
	g.Expect(httpCfg).To(ContainSubstring("listen 80"))
	g.Expect(httpCfg).To(ContainSubstring("listen unix:/var/run/nginx/https443.sock"))
	g.Expect(httpCfg).To(ContainSubstring("upstream"))
	g.Expect(httpCfg).To(ContainSubstring("split_clients"))

	g.Expect(httpCfg).To(ContainSubstring("endpoint 1.2.3.4:123;"))
	g.Expect(httpCfg).To(ContainSubstring("interval 5s;"))
	g.Expect(httpCfg).To(ContainSubstring("batch_size 512;"))
	g.Expect(httpCfg).To(ContainSubstring("batch_count 4;"))
	g.Expect(httpCfg).To(ContainSubstring("otel_service_name ngf:gw-ns:gw-name:my-name;"))
	g.Expect(httpCfg).To(ContainSubstring("http2 on;"))

	g.Expect(files[2].Path).To(Equal("/etc/nginx/conf.d/matches.json"))

	g.Expect(files[2].Type).To(Equal(file.TypeRegular))
	expString := "{}"
	g.Expect(string(files[2].Content)).To(Equal(expString))

	g.Expect(files[3].Path).To(Equal("/etc/nginx/module-includes/load-modules.conf"))
	g.Expect(files[3].Content).To(Equal([]byte("load_module modules/ngx_otel_module.so;")))

	g.Expect(files[4].Path).To(Equal("/etc/nginx/secrets/test-certbundle.crt"))
	certBundle := string(files[4].Content)
	g.Expect(certBundle).To(Equal("test-cert"))

	g.Expect(files[5]).To(Equal(file.File{
		Type:    file.TypeSecret,
		Path:    "/etc/nginx/secrets/test-keypair.pem",
		Content: []byte("test-cert\ntest-key"),
	}))

	g.Expect(files[6].Path).To(Equal("/etc/nginx/stream-conf.d/stream.conf"))
	g.Expect(files[6].Type).To(Equal(file.TypeRegular))
	streamCfg := string(files[6].Content)
	g.Expect(streamCfg).To(ContainSubstring("listen unix:/var/run/nginx/app.example.com-443.sock"))
	g.Expect(streamCfg).To(ContainSubstring("listen 443"))
	g.Expect(streamCfg).To(ContainSubstring("app.example.com unix:/var/run/nginx/app.example.com-443.sock"))
	g.Expect(streamCfg).To(ContainSubstring("example.com unix:/var/run/nginx/https443.sock"))
}
