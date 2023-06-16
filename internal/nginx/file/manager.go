package file

import (
	"fmt"
	"os"
)

const confdFolder = "/etc/nginx/conf.d"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Manager

type File struct {
	Path        string
	Content     []byte
	Permissions os.FileMode
}

// Manager manages NGINX configuration files.
type Manager interface {
	// ReplaceFiles replaces the files on the file system with the given files.
	ReplaceFiles(files []File) error
}

// ManagerImpl is an implementation of Manager.
type ManagerImpl struct {
	lastWrittenPaths []string
}

// NewManagerImpl creates a new NewManagerImpl.
func NewManagerImpl() *ManagerImpl {
	return &ManagerImpl{}
}

func (m *ManagerImpl) ReplaceFiles(files []File) error {
	for _, path := range m.lastWrittenPaths {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to delete file %s: %w", path, err)
		}
	}

	// In some cases, NGINX reads files in runtime, like JWT secrets
	// In that case, removal will lead to errors.
	// However, we don't have such files yet. so not a problem

	m.lastWrittenPaths = make([]string, 0, len(files))

	for _, file := range files {
		f, err := os.Create(file.Path)
		if err != nil {
			return fmt.Errorf("failed to create server config %s: %w", file.Path, err)
		}
		defer f.Close()

		_, err = f.Write(file.Content)
		if err != nil {
			return fmt.Errorf("failed to write server config %s: %w", file.Path, err)
		}

		m.lastWrittenPaths = append(m.lastWrittenPaths, file.Path)
	}

	return nil
}
