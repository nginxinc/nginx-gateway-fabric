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
	pidFile            = "/var/run/nginx/nginx.pid"
	pidFileTimeout     = 10000 * time.Millisecond
	nginxReloadTimeout = 60000 * time.Millisecond
)

type (
	readFileFunc  func(string) ([]byte, error)
	checkFileFunc func(string) (fs.FileInfo, error)
)

var childProcPathFmt = "/proc/%[1]v/task/%[1]v/children"

<<<<<<< HEAD
//counterfeiter:generate . Manager
=======
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . NginxPlusClient

type NginxPlusClient interface {
	UpdateHTTPServers(upstream string, servers []ngxclient.UpstreamServer) (added []ngxclient.UpstreamServer, deleted []ngxclient.UpstreamServer, updated []ngxclient.UpstreamServer, err error)
	GetUpstreams() (*ngxclient.Upstreams, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager
>>>>>>> 0b42d66 (test update for ManagerImpl)

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

// MetricsCollector is an interface for the metrics of the NGINX runtime manager.
type MetricsCollector interface {
	IncReloadCount()
	IncReloadErrors()
	ObserveLastReloadTime(ms time.Duration)
}

// ManagerImpl implements Manager.
type ManagerImpl struct {
	verifyClient     *verifyClient
	metricsCollector MetricsCollector
	ngxPlusClient    NginxPlusClient
	logger           logr.Logger
}

// NewManagerImpl creates a new ManagerImpl.
func NewManagerImpl(
	ngxPlusClient NginxPlusClient,
	collector MetricsCollector,
	logger logr.Logger,
) *ManagerImpl {
	return &ManagerImpl{
		verifyClient:     newVerifyClient(nginxReloadTimeout),
		metricsCollector: collector,
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
	pid, err := findMainProcess(ctx, os.Stat, os.ReadFile, pidFileTimeout)
	if err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}

	childProcFile := fmt.Sprintf(childProcPathFmt, pid)
	previousChildProcesses, err := os.ReadFile(childProcFile)
	if err != nil {
		return err
	}

	// send HUP signal to the NGINX main process reload configuration
	// See https://nginx.org/en/docs/control.html
	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		m.metricsCollector.IncReloadErrors()
		return fmt.Errorf("failed to send the HUP signal to NGINX main: %w", err)
	}

	if err = m.verifyClient.waitForCorrectVersion(
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
func EnsureNginxRunning(ctx context.Context) error {
	if _, err := findMainProcess(ctx, os.Stat, os.ReadFile, pidFileTimeout); err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}
	return nil
}

func findMainProcess(
	ctx context.Context,
	checkFile checkFileFunc,
	readFile readFileFunc,
	timeout time.Duration,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(_ context.Context) (bool, error) {
			_, err := checkFile(pidFile)
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

	content, err := readFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid file content %q: %w", content, err)
	}

	return pid, nil
}
