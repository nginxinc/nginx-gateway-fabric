package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	pidFile            = "/var/run/nginx/nginx.pid"
	pidFileTimeout     = 10000 * time.Millisecond
	childProcsTimeout  = 1000 * time.Millisecond
	nginxReloadTimeout = 60000 * time.Millisecond
)

type (
	readFileFunc  func(string) ([]byte, error)
	checkFileFunc func(string) (fs.FileInfo, error)
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

// Manager manages the runtime of NGINX.
type Manager interface {
	// Reload reloads NGINX configuration. It is a blocking operation.
	Reload(ctx context.Context, configVersion int) error
}

// ManagerImpl implements Manager.
type ManagerImpl struct {
	verifyClient *verifyClient
}

// NewManagerImpl creates a new ManagerImpl.
func NewManagerImpl() *ManagerImpl {
	return &ManagerImpl{
		verifyClient: newVerifyClient(nginxReloadTimeout),
	}
}

func (m *ManagerImpl) Reload(ctx context.Context, configVersion int) error {
	// We find the main NGINX PID on every reload because it will change if the NGINX container is restarted.
	pid, err := findMainProcess(ctx, os.Stat, os.ReadFile, pidFileTimeout)
	if err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}

	childProcFile := fmt.Sprintf("/proc/%[1]v/task/%[1]v/children", pid)
	previousChildProcesses, err := os.ReadFile(childProcFile)
	if err != nil {
		return err
	}

	// send HUP signal to the NGINX main process reload configuration
	// See https://nginx.org/en/docs/control.html
	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send the HUP signal to NGINX main: %w", err)
	}

	newProcsStarted, err := ensureNewNginxWorkers(
		ctx, childProcFile, previousChildProcesses, os.ReadFile, childProcsTimeout)
	if !newProcsStarted {
		return fmt.Errorf("reload unsuccessful: no new NGINX worker processes started: %w", err)
	}

	if err = m.verifyClient.WaitForCorrectVersion(ctx, configVersion); err != nil {
		return fmt.Errorf("could not get newest config version: %w", err)
	}

	return nil
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
		func(ctx context.Context) (bool, error) {
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

func ensureNewNginxWorkers(
	ctx context.Context,
	childProcFile string,
	previousContents []byte,
	readFile readFileFunc,
	timeout time.Duration,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := wait.PollUntilContextCancel(
		ctx,
		25*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			content, err := readFile(childProcFile)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(previousContents, content) {
				return true, nil
			}
			return false, nil
		}); err != nil {
		return false, err
	}

	return true, nil
}
