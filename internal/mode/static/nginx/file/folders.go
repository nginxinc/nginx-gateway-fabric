package file

import (
	"fmt"
	"os"
	"path/filepath"
)

//counterfeiter:generate io/fs.DirEntry

//counterfeiter:generate . ClearFoldersOSFileManager

// ClearFoldersOSFileManager is an interface that exposes File I/O operations for ClearFolders.
// Used for unit testing.
type ClearFoldersOSFileManager interface {
	// ReadDir returns the directory entries for the directory.
	ReadDir(dirname string) ([]os.DirEntry, error)
	// Remove removes the file with given name.
	Remove(name string) error
}

// ClearFolders removes all files in the given folders and returns the removed files' full paths.
func ClearFolders(fileMgr ClearFoldersOSFileManager, paths []string) (removedFiles []string, e error) {
	for _, path := range paths {
		entries, err := fileMgr.ReadDir(path)
		if err != nil {
			return removedFiles, fmt.Errorf("failed to read directory %q: %w", path, err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if err := fileMgr.Remove(entryPath); err != nil {
				return removedFiles, fmt.Errorf("failed to remove %q: %w", entryPath, err)
			}

			removedFiles = append(removedFiles, entryPath)
		}
	}

	return removedFiles, nil
}
