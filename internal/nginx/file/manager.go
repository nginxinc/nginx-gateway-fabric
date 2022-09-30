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
	// WriteHTTPConfig writes the http config on the file system.
	// The name distinguishes this config among all other configs. For that, it must be unique.
	// Note that name is not the name of the corresponding configuration file.
	WriteHTTPConfig(name string, cfg []byte) error
}

// ManagerImpl is an implementation of Manager.
type ManagerImpl struct{}

// NewManagerImpl creates a new NewManagerImpl.
func NewManagerImpl() *ManagerImpl {
	return &ManagerImpl{}
}

func (m *ManagerImpl) WriteHTTPConfig(name string, cfg []byte) error {
	path := getPathForConfig(name)

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

func getPathForConfig(name string) string {
	return filepath.Join(confdFolder, name+".conf")
}
