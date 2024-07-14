package runtime_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime/runtimefakes"
)

var _ = Describe("NGINX Runtime Manager", func() {
	It("returns whether or not we're using NGINX Plus", func() {
		mgr := runtime.NewManagerImpl(nil, nil, zap.New(), nil, nil)
		Expect(mgr.IsPlus()).To(BeFalse())

		mgr = runtime.NewManagerImpl(&ngxclient.NginxClient{}, nil, zap.New(), nil, nil)
		Expect(mgr.IsPlus()).To(BeTrue())
	})

	var (
		err             error
		manager         runtime.Manager
		upstreamServers []ngxclient.UpstreamServer
		ngxPlusClient   *runtimefakes.FakeNginxPlusClient
		process         *runtimefakes.FakeProcessHandler

		metrics      *runtimefakes.FakeMetricsCollector
		verifyClient *runtimefakes.FakeVerifyClient
	)

	BeforeEach(func() {
		upstreamServers = []ngxclient.UpstreamServer{
			{},
		}
	})

	Context("Reload", func() {
		BeforeEach(func() {
			ngxPlusClient = &runtimefakes.FakeNginxPlusClient{}
			process = &runtimefakes.FakeProcessHandler{}
			metrics = &runtimefakes.FakeMetricsCollector{}
			verifyClient = &runtimefakes.FakeVerifyClient{}
			manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)
		})

		It("Is successful", func() {
			Expect(manager.Reload(context.Background(), 1)).To(Succeed())

			Expect(process.FindMainProcessCallCount()).To(Equal(1))
			Expect(process.ReadFileCallCount()).To(Equal(1))
			Expect(process.KillCallCount()).To(Equal(1))
			Expect(metrics.IncReloadCountCallCount()).To(Equal(1))
			Expect(verifyClient.WaitForCorrectVersionCallCount()).To(Equal(1))
			Expect(metrics.ObserveLastReloadTimeCallCount()).To(Equal(1))
			Expect(metrics.IncReloadErrorsCallCount()).To(Equal(0))
		})

		When("MetricsCollector is nil", func() {
			It("panics", func() {
				metrics = nil
				manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)

				reload := func() {
					err = manager.Reload(context.Background(), 0)
				}

				Expect(reload).To(Panic())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("VerifyClient is nil", func() {
			It("panics", func() {
				metrics = &runtimefakes.FakeMetricsCollector{}
				verifyClient = nil
				manager = runtime.NewManagerImpl(ngxPlusClient, metrics, zap.New(), process, verifyClient)

				reload := func() {
					err = manager.Reload(context.Background(), 0)
				}

				Expect(reload).To(Panic())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	When("running NGINX plus", func() {
		BeforeEach(func() {
			ngxPlusClient = &runtimefakes.FakeNginxPlusClient{}
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, zap.New(), nil, nil)
		})

		It("successfully updates HTTP server upstream", func() {
			Expect(manager.UpdateHTTPServers("test", upstreamServers)).To(Succeed())
		})

		It("returns no upstreams from NGINX Plus API when upstreams are nil", func() {
			upstreams, err := manager.GetUpstreams()

			Expect(err).To(HaveOccurred())
			Expect(upstreams).To(BeEmpty())
		})

		It("returns an error when GetUpstreams fails", func() {
			ngxPlusClient.GetUpstreamsReturns(nil, errors.New("failed to get upstreams"))

			upstreams, err := manager.GetUpstreams()

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("failed to get upstreams"))
			Expect(upstreams).To(BeNil())
		})
	})

	When("not running NGINX plus", func() {
		BeforeEach(func() {
			ngxPlusClient = nil
			manager = runtime.NewManagerImpl(ngxPlusClient, nil, zap.New(), nil, nil)
		})

		It("should panic when updating HTTP upstream servers", func() {
			updateServers := func() {
				err = manager.UpdateHTTPServers("test", upstreamServers)
			}

			Expect(updateServers).To(Panic())
			Expect(err).ToNot(HaveOccurred())
		})

		It("should panic when fetching HTTP upstream servers", func() {
			upstreams := func() {
				_, err = manager.GetUpstreams()
			}

			Expect(upstreams).To(Panic())
			Expect(err).ToNot(HaveOccurred())
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
			p := runtime.NewProcessHandlerImpl(
				test.readFile,
				test.checkFile)
			result, err := p.FindMainProcess(test.ctx, 2*time.Millisecond)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(result).To(Equal(test.expected))
			}
		})
	}
}
