package runtime

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
	"k8s.io/apimachinery/pkg/util/wait"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	PidFile            = "/var/run/nginx/nginx.pid"
	pidFileTimeout     = 10000 * time.Millisecond
	NginxReloadTimeout = 60000 * time.Millisecond
)

type (
	ReadFileFunc  func(string) ([]byte, error)
	CheckFileFunc func(string) (fs.FileInfo, error)
)

var childProcPathFmt = "/proc/%[1]v/task/%[1]v/children"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . nginxPlusClient

type nginxPlusClient interface {
	UpdateHTTPServers(
		upstream string,
		servers []ngxclient.UpstreamServer,
	) (
		added []ngxclient.UpstreamServer,
		deleted []ngxclient.UpstreamServer,
		updated []ngxclient.UpstreamServer,
		err error,
	)
	GetUpstreams() (*ngxclient.Upstreams, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ProcessHandler

type ProcessHandler interface {
	FindMainProcess(
		ctx context.Context,
		checkFile CheckFileFunc,
		readFile ReadFileFunc,
		timeout time.Duration,
	) (int, error)
	ReadFile(file string) ([]byte, error)
	Kill(pid int) error
	EnsureNginxRunning(ctx context.Context) error
}

type ProcessHandlerImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

// Manager manages the runtime of NGINX.
type Manager interface {
	// Reload reloads NGINX configuration. It is a blocking operation.
	Reload(ctx context.Context, configVersion int) error
	// IsPlus returns whether or not we are running NGINX plus.
	IsPlus() bool
	// UpdateHTTPServers uses the NGINX Plus API to update HTTP servers.
	// Only usable if running NGINX Plus.
	UpdateHTTPServers(string, []ngxclient.UpstreamServer) error
	// GetUpstreams uses the NGINX Plus API to get the upstreams.
	// Only usable if running NGINX Plus.
	GetUpstreams() (ngxclient.Upstreams, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . MetricsCollector

// MetricsCollector is an interface for the metrics of the NGINX runtime manager.
type MetricsCollector interface {
	IncReloadCount()
	IncReloadErrors()
	ObserveLastReloadTime(ms time.Duration)
}

// ManagerImpl implements Manager.
type ManagerImpl struct {
	processHandler   ProcessHandler
	metricsCollector MetricsCollector
	verifyClient     verifyClient
	ngxPlusClient    nginxPlusClient
	logger           logr.Logger
}

// NewManagerImpl creates a new ManagerImpl.
func NewManagerImpl(
	ngxPlusClient nginxPlusClient,
	collector MetricsCollector,
	logger logr.Logger,
	processHandler ProcessHandler,
	verifyClient verifyClient,
) *ManagerImpl {
	return &ManagerImpl{
		processHandler:   processHandler,
		metricsCollector: collector,
		verifyClient:     verifyClient,
		ngxPlusClient:    ngxPlusClient,
		logger:           logger,
	}
}

// IsPlus returns whether or not we are running NGINX plus.
func (m *ManagerImpl) IsPlus() bool {
	return m.ngxPlusClient != nil
}

func (m *ManagerImpl) Reload(ctx context.Context, configVersion int) error {
	start := time.Now()
	// We find the main NGINX PID on every reload because it will change if the NGINX container is restarted.
	pid, err := m.processHandler.FindMainProcess(ctx, os.Stat, os.ReadFile, pidFileTimeout)
	if err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}

	childProcFile := fmt.Sprintf(childProcPathFmt, pid)
	previousChildProcesses, err := m.processHandler.ReadFile(childProcFile)
	if err != nil {
		return err
	}

	// send HUP signal to the NGINX main process reload configuration
	// See https://nginx.org/en/docs/control.html
	if err := m.processHandler.Kill(pid); err != nil {
		m.metricsCollector.IncReloadErrors()
		return fmt.Errorf("failed to send the HUP signal to NGINX main: %w", err)
	}

	if err = m.verifyClient.WaitForCorrectVersion(
		ctx,
		configVersion,
		childProcFile,
		previousChildProcesses,
		os.ReadFile,
	); err != nil {
		m.metricsCollector.IncReloadErrors()
		return err
	}
	m.metricsCollector.IncReloadCount()

	finish := time.Now()
	m.metricsCollector.ObserveLastReloadTime(finish.Sub(start))
	return nil
}

// UpdateHTTPServers uses the NGINX Plus API to update HTTP upstream servers.
// Only usable if running NGINX Plus.
func (m *ManagerImpl) UpdateHTTPServers(upstream string, servers []ngxclient.UpstreamServer) error {
	if !m.IsPlus() {
		panic("cannot update HTTP upstream servers: NGINX Plus not enabled")
	}

	added, deleted, updated, err := m.ngxPlusClient.UpdateHTTPServers(upstream, servers)
	m.logger.V(1).Info("Added upstream servers", "count", len(added))
	m.logger.V(1).Info("Deleted upstream servers", "count", len(deleted))
	m.logger.V(1).Info("Updated upstream servers", "count", len(updated))

	return err
}

// GetUpstreams uses the NGINX Plus API to get the upstreams.
// Only usable if running NGINX Plus.
func (m *ManagerImpl) GetUpstreams() (ngxclient.Upstreams, error) {
	if !m.IsPlus() {
		panic("cannot get HTTP upstream servers: NGINX Plus not enabled")
	}

	upstreams, err := m.ngxPlusClient.GetUpstreams()
	if err != nil {
		return nil, err
	}

	if upstreams == nil {
		return nil, errors.New("GET upstreams returned nil value")
	}

	return *upstreams, nil
}

// EnsureNginxRunning ensures NGINX is running by locating the main process.
func (p *ProcessHandlerImpl) EnsureNginxRunning(ctx context.Context) error {
	if _, err := p.FindMainProcess(ctx, os.Stat, os.ReadFile, pidFileTimeout); err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}
	return nil
}

func (p *ProcessHandlerImpl) FindMainProcess(
	ctx context.Context,
	checkFile CheckFileFunc,
	readFile ReadFileFunc,
	timeout time.Duration,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			_, err := checkFile(PidFile)
			if err == nil {
				return true, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return false, err
			}
			return false, nil
		})
	if err != nil {
		return 0, err
	}

	content, err := readFile(PidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid file content %q: %w", content, err)
	}

	return pid, nil
}

func (p *ProcessHandlerImpl) ReadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

func (p *ProcessHandlerImpl) Kill(pid int) error {
	return syscall.Kill(pid, syscall.SIGHUP)
}
