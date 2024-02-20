package runtime_test

import (
	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime/runtimefakes"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX Runtime Manager", Ordered, func() {
	var (
		manager         runtime.Manager
		upstreamServers []ngxclient.UpstreamServer
		upstreamServer  ngxclient.UpstreamServer
		logger          logr.Logger
		ngxPlusClient   *runtimefakes.FakeNginxPlusClient
	)

	var (
		MaxConnsMock = 150
		MaxFailsMock = 20
		WeightMock   = 10
		BackupMock   = false
		DownMock     = false
	)

	BeforeAll(func() {
		upstreamServer = ngxclient.UpstreamServer{
			ID:          1,
			Server:      "10.0.0.1:80",
			MaxConns:    &MaxConnsMock,
			MaxFails:    &MaxFailsMock,
			FailTimeout: "test",
			SlowStart:   "test",
			Route:       "test",
			Backup:      &BackupMock,
			Down:        &DownMock,
			Drain:       false,
			Weight:      &WeightMock,
			Service:     "test",
		}

		upstreamServers = []ngxclient.UpstreamServer{
			upstreamServer,
		}

		logger = logr.New(GinkgoLogr.GetSink())
	})

	When("running NGINX plus", func() {
		BeforeAll(func() {
			ngxPlusClient = &runtimefakes.FakeNginxPlusClient{}
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, logger)
		})

		It("sucessfully updates HTTP server upstream", func() {
			err := manager.UpdateHTTPServers("test", upstreamServers)

			Expect(err).To(BeNil())
		})

		It("returns no upstreams from NGINX Plus API", func() {
			upstreams, err := manager.GetUpstreams()

			Expect(err).To(HaveOccurred())
			Expect(upstreams).To(BeEmpty())
		})
	})

	When("not running NGINX plus", func() {
		BeforeAll(func() {
			ngxPlusClient = nil
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, logger)
		})

		It("should panic when updating HTTP upstream servers", func() {
			updateServers := func() {
				manager.UpdateHTTPServers("test", upstreamServers)
			}

			Expect(updateServers).To(Panic())
		})

		It("should panic when fetching HTTP upstream servers", func() {
			upstreams := func() {
				manager.GetUpstreams()
			}

			Expect(upstreams).To(Panic())
		})
	})
})
