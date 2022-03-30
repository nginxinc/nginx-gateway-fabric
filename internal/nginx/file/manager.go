package file

import (
	"fmt"
	"os"
	"path/filepath"
)

const confdFolder = "/etc/nginx/conf.d"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

// Manager manages NGINX configuration files.
type Manager interface {
	// WriteServerConfig writes the server config on the file system.
	// The name distinguishes this server config among all other server configs. For that, it must be unique.
	// Note that name is not the name of the corresponding configuration file.
	WriteServerConfig(name string, cfg []byte) error
	// DeleteServerConfig deletes the corresponding configuration file for the server from the file system.
	DeleteServerConfig(name string) error
}

// ManagerImpl is an implementation of Manager.
type ManagerImpl struct{}

// NewManagerImpl creates a new NewManagerImpl.
func NewManagerImpl() *ManagerImpl {
	return &ManagerImpl{}
}

func (m *ManagerImpl) WriteServerConfig(name string, cfg []byte) error {
	path := getPathForServerConfig(name)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create server config %s: %w", path, err)
	}

	defer file.Close()

	_, err = file.Write(cfg)
	if err != nil {
		return fmt.Errorf("failed to write server config %s: %w", path, err)
	}

	return nil
}

func (m *ManagerImpl) DeleteServerConfig(name string) error {
	path := getPathForServerConfig(name)

	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove server config %s: %w", path, err)
	}

	return nil
}

func getPathForServerConfig(name string) string {
	return filepath.Join(confdFolder, name+".conf")
}
