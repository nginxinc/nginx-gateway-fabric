package runtime_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime/runtimefakes"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("NGINX Runtime Manager", Ordered, func() {
	It("returns whether or not we're using NGINX Plus", func() {
		mgr := runtime.NewManagerImpl(nil, nil, zap.New(), nil, nil)
		Expect(mgr.IsPlus()).To(BeFalse())

		mgr = runtime.NewManagerImpl(&ngxclient.NginxClient{}, nil, zap.New(), nil, nil)
		Expect(mgr.IsPlus()).To(BeTrue())
	})

	var (
		manager         runtime.Manager
		upstreamServers []ngxclient.UpstreamServer
		upstreamServer  ngxclient.UpstreamServer
		ngxPlusClient   *runtimefakes.FakeNginxPlusClient
		process         *runtimefakes.FakeProcessHandler

		metrics      *runtimefakes.FakeMetricsCollector
		verifyClient *runtimefakes.FakeVerifyClient
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

	})

	Context("Reload", func() {
		BeforeAll(func() {
			ngxPlusClient = &runtimefakes.FakeNginxPlusClient{}
			process = &runtimefakes.FakeProcessHandler{}
			metrics = &runtimefakes.FakeMetricsCollector{}
			verifyClient = &runtimefakes.FakeVerifyClient{}
			manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)
		})

		It("sucessfully reloads NGINX configuration", func() {
			err := manager.Reload(context.Background(), 1)
			Expect(err).To(BeNil())
		})

		When("not sucessfully reloads NGINX configuration", func() {
			It("should panic if MetricsCollector not enabled", func() {
				metrics = nil
				manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)

				reload := func() {
					manager.Reload(context.Background(), 0)
				}

				Expect(reload).To(Panic())
			})

			It("should panic if VerifyClient not enabled", func() {
				metrics = &runtimefakes.FakeMetricsCollector{}
				verifyClient = nil
				manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)

				reload := func() {
					manager.Reload(context.Background(), 0)
				}

				Expect(reload).To(Panic())
			})
		})
	})

	When("running NGINX plus", func() {
		BeforeAll(func() {
			ngxPlusClient = &runtimefakes.FakeNginxPlusClient{}
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, zap.New(), nil, nil)
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
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, zap.New(), nil, nil)
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

func TestFindMainProcess(t *testing.T) {
	readFileFuncGen := func(content []byte) runtime.ReadFileFunc {
		return func(name string) ([]byte, error) {
			if name != runtime.PidFile {
				return nil, errors.New("error")
			}
			return content, nil
		}
	}
	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	checkFileFuncGen := func(content fs.FileInfo) runtime.CheckFileFunc {
		return func(name string) (fs.FileInfo, error) {
			if name != runtime.PidFile {
				return nil, errors.New("error")
			}
			return content, nil
		}
	}
	checkFileError := func(string) (fs.FileInfo, error) {
		return nil, errors.New("error")
	}
	var testFileInfo fs.FileInfo
	ctx := context.Background()
	cancellingCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(1*time.Millisecond, cancel)

	tests := []struct {
		ctx         context.Context
		readFile    runtime.ReadFileFunc
		checkFile   runtime.CheckFileFunc
		name        string
		expected    int
		expectError bool
	}{
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    1,
			expectError: false,
			name:        "normal case",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			name:        "empty file content",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("not a number")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			name:        "bad file content",
		},
		{
			ctx:         ctx,
			readFile:    readFileError,
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			name:        "cannot read file",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileError,
			expected:    0,
			expectError: true,
			name:        "cannot find pid file",
		},
		{
			ctx:         cancellingCtx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileError,
			expected:    0,
			expectError: true,
			name:        "context canceled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result, err := runtime.FindMainProcess(test.ctx, test.checkFile, test.readFile, 2*time.Millisecond)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(result).To(Equal(test.expected))
			}
		})
	}
}
