package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

// Manager manages the runtime of NGINX.
type Manager interface {
	// Reload reloads NGINX configuration. It is a blocking operation.
	Reload(ctx context.Context) error
}

// ManagerImpl implements Manager.
type ManagerImpl struct {
}

// NewManagerImpl creates a new ManagerImpl.
func NewManagerImpl() Manager {
	return &ManagerImpl{}
}

func (m *ManagerImpl) Reload(ctx context.Context) error {
	// FIXME(pleshakov): Before reload attempt, make sure NGINX is running.
	// If the gateway container starts before NGINX container (which is possible), then it is possible that a reload can be attempted
	// when NGINX is not running yet. Make sure to prevent this case, so we don't get an error.

	// We find the NGINX master PID on every reload because it will change if the NGINX container is restarted.
	pid, err := findMasterProcess(readProcDir, os.ReadFile)
	if err != nil {
		return fmt.Errorf("failed to find NGINX master process: %w", err)
	}

	// send HUP signal to the NGINX master process reload configuration
	// See https://nginx.org/en/docs/control.html
	err = syscall.Kill(pid, syscall.SIGHUP)
	if err != nil {
		return fmt.Errorf("failed to send the HUP signal to NGINX master: %w", err)
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

func readProcDir() ([]os.DirEntry, error) {
	return os.ReadDir("/proc")
}

func findMasterProcess(readProcDir func() ([]os.DirEntry, error), readFile func(string) ([]byte, error)) (int, error) {
	procFolders, err := readProcDir()
	if err != nil {
		return 0, fmt.Errorf("unable to read process directory: %w", err)
	}

	for _, folder := range procFolders {
		pid, err := strconv.Atoi(folder.Name())
		if err != nil {
			continue
		}

		cmdlineFile := fmt.Sprintf("/proc/%v/cmdline", folder.Name())
		content, err := readFile(cmdlineFile)
		if err != nil {
			return 0, fmt.Errorf("unable to read file %s: %w", cmdlineFile, err)
		}

		if bytes.HasPrefix(content, []byte("nginx: master process")) {
			return pid, nil
		}
	}

	return 0, errors.New("NGINX master is not running")
}
