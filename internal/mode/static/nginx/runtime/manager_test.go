package runtime

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-plus-go-client/client"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX Runtime Manager", func() {
	It("returns whether or not we're using NGINX Plus", func() {
		mgr := NewManagerImpl(nil, nil, logr.New(GinkgoLogr.GetSink()))
		Expect(mgr.IsPlus()).To(BeFalse())

		mgr = NewManagerImpl(&ngxclient.NginxClient{}, nil, logr.New(GinkgoLogr.GetSink()))
		Expect(mgr.IsPlus()).To(BeTrue())
	})

	Describe("ManagerImpl", Ordered, func() {
		var (
			mockedManager Manager
			serverMocks   []ngxclient.UpstreamServer
			serversMock   ngxclient.UpstreamServer

			clientMock *ngxclient.NginxClient
		)

		var (
			MaxConnsMock = 150
			MaxFailsMock = 20
			WeightMock   = 10
			BackupMock   = false
			DownMock     = false
		)

		BeforeAll(func() {

			serversMock = ngxclient.UpstreamServer{
				ID:          1,
				Server:      "unknown",
				MaxConns:    &MaxConnsMock,
				MaxFails:    &MaxFailsMock,
				FailTimeout: "test",
				SlowStart:   "test",
				Route:       "test",
				Backup:      &BackupMock,
				Down:        &DownMock,
				Drain:       false,
				Weight:      &WeightMock,
				Service:     "",
			}

			serverMocks = []ngxclient.UpstreamServer{
				serversMock,
			}

			httpClient := &http.Client{
				Transport:     http.DefaultTransport,
				CheckRedirect: http.DefaultClient.CheckRedirect,
				Jar:           http.DefaultClient.Jar,
				Timeout:       time.Second * 4,
			}

			clientMock, _ = ngxclient.NewNginxClient("test", client.WithHTTPClient(httpClient))
			logrMock := logr.New(GinkgoLogr.GetSink())
			mockedManager = NewManagerImpl(clientMock, nil, logrMock)

		})

		It("UpdateHTTPServers fails upon unknown HTTP server upstream", func() {
			err := mockedManager.UpdateHTTPServers("unknown", serverMocks)
			Expect(err).ToNot(BeNil())
		})

		Context("GetUpstreams returns empty map of upstreams", func() {
			It("returns no upstreams from NGINX Plus API", func() {

				upstreams, err := mockedManager.GetUpstreams()
				Expect(err).To(HaveOccurred())
				Expect(upstreams).To(BeEmpty())
			})
		})
	})
})

func TestEnsureNginxRunning(t *testing.T) {
	ctx := context.Background()
	cancellingCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(1*time.Millisecond, cancel)
	tests := []struct {
		ctx         context.Context
		name        string
		expectError bool
	}{
		{
			ctx:         ctx,
			name:        "context exceeded",
			expectError: true,
		},
		{
			ctx:         cancellingCtx,
			name:        "context cancelled",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := EnsureNginxRunning(test.ctx)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestFindMainProcess(t *testing.T) {
	readFileFuncGen := func(content []byte) readFileFunc {
		return func(name string) ([]byte, error) {
			if name != pidFile {
				return nil, errors.New("error")
			}
			return content, nil
		}
	}
	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	checkFileFuncGen := func(content fs.FileInfo) checkFileFunc {
		return func(name string) (fs.FileInfo, error) {
			if name != pidFile {
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
		readFile    readFileFunc
		checkFile   checkFileFunc
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

			result, err := findMainProcess(test.ctx, test.checkFile, test.readFile, 2*time.Millisecond)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(result).To(Equal(test.expected))
			}
		})
	}
}
