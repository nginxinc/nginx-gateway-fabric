package runtime_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime/runtimefakes"
	"github.com/nginxinc/nginx-plus-go-client/client"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
)

var _ = Describe("ManagerImpl UpdateHTTPServers", func() {

	var (
		MaxConnsMock = 15
		MaxFailsMock = 2
		WeightMock   = 1
		BackupMock   = false
		DownMock     = false
	)

	BeforeEach(func() {
		serversMock := ngxclient.UpstreamServer{
			ID:          1,
			Server:      "10.0.0.1:8080",
			MaxConns:    &MaxConnsMock,
			MaxFails:    &MaxFailsMock,
			FailTimeout: "test",
			SlowStart:   "test",
			Route:       "test",
			Backup:      &BackupMock,
			Down:        &DownMock,
			Drain:       false,
			Weight:      &WeightMock,
			Service:     "server",
		}

		serverMocks := []ngxclient.UpstreamServer{
			serversMock,
		}

		nginxStatusSock := "http://10.0.0.1:8080"
		httpClient := &http.Client{}
		clientMock, _ := ngxclient.NewNginxClient(nginxStatusSock, client.WithHTTPClient(httpClient))
		logrMock := logr.New(GinkgoLogr.GetSink())

		mockedManager := runtime.NewManagerImpl(clientMock, nil, logrMock)
		err := mockedManager.UpdateHTTPServers("10.0.0.1:8080", serverMocks)
		Expect(err).To(BeNil())

	})
})

var _ = Describe("ManagerImpl GetUpstreams", func() {
	var (
		mockNginxClient *ngxclient.NginxClient
		manager         *runtimefakes.FakeManager
	)

	var (
		MaxConnsMock = 15
		MaxFailsMock = 2
		WeightMock   = 1
		BackupMock   = false
		DownMock     = false
	)

	BeforeEach(func() {
		nginxStatusSock := "http://10.0.0.1:8080"
		httpClient := &http.Client{}
		mockNginxClient, _ = ngxclient.NewNginxClient(nginxStatusSock, client.WithHTTPClient(httpClient))
		mockNginxClient.AddHTTPServer("10.0.0.1:8080", ngxclient.UpstreamServer{
			ID:          1,
			Server:      "10.0.0.1:8080",
			MaxConns:    &MaxConnsMock,
			MaxFails:    &MaxFailsMock,
			FailTimeout: "test",
			SlowStart:   "test",
			Route:       "test",
			Backup:      &BackupMock,
			Down:        &DownMock,
			Drain:       false,
			Weight:      &WeightMock,
			Service:     "server",
		})
		manager = &runtimefakes.FakeManager{}
	})

	Describe("GetUpstreams", func() {
		Context("when NGINX Plus is not enabled", func() {
			BeforeEach(func() {
				manager.IsPlusReturns(false)
			})

			It("returns no upstreams from NGINX Plus API", func() {

				_, err := manager.GetUpstreams()
				Expect(err).ToNot(HaveOccurred())
			})

		})
	})
})
