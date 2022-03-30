package runtime

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pidFile = "/etc/nginx/nginx.pid"

type readFileFunc func(string) ([]byte, error)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

// Manager manages the runtime of NGINX.
type Manager interface {
	// Reload reloads NGINX configuration. It is a blocking operation.
	Reload(ctx context.Context) error
}

// ManagerImpl implements Manager.
type ManagerImpl struct{}

// NewManagerImpl creates a new ManagerImpl.
func NewManagerImpl() *ManagerImpl {
	return &ManagerImpl{}
}

func (m *ManagerImpl) Reload(ctx context.Context) error {
	// FIXME(pleshakov): Before reload attempt, make sure NGINX is running.
	// If the gateway container starts before NGINX container (which is possible), then it is possible that a reload can be attempted
	// when NGINX is not running yet. Make sure to prevent this case, so we don't get an error.

	// We find the main NGINX PID on every reload because it will change if the NGINX container is restarted.
	pid, err := findMainProcess(os.ReadFile)
	if err != nil {
		return fmt.Errorf("failed to find NGINX main process: %w", err)
	}

	// send HUP signal to the NGINX main process reload configuration
	// See https://nginx.org/en/docs/control.html
	err = syscall.Kill(pid, syscall.SIGHUP)
	if err != nil {
		return fmt.Errorf("failed to send the HUP signal to NGINX main: %w", err)
	}

	// FIXME(pleshakov)
	// (1) ensure the reload actually happens.
	// (2) ensure that in case of an error, the error message can be seen by the admins.

	// for now, to prevent a subsequent reload starting before the in-flight reload finishes, we simply sleep.
	// Fixing (1) will make the sleep unnecessary.

	select {
	case <-ctx.Done():
		return nil
	case <-time.After(1 * time.Second):
	}

	return nil
}

func findMainProcess(readFile readFileFunc) (int, error) {
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
