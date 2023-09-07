package runtime

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

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
			g := NewGomegaWithT(t)

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

func TestEnsureNewNginxWorkers(t *testing.T) {
	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	tempFileFunc := func(contents []byte) *os.File {
		tempFile, err := os.CreateTemp("", "tmpfile-")
		if err != nil {
			return nil
		}
		if _, err = tempFile.Write(contents); err != nil {
			return nil
		}
		return tempFile
	}

	previousContents := []byte("1 2 3")

	childFileSame := tempFileFunc(previousContents)
	childFileDifferent := tempFileFunc([]byte("4 5 6"))

	defer childFileSame.Close()
	defer os.Remove(childFileSame.Name())
	defer childFileDifferent.Close()
	defer os.Remove(childFileDifferent.Name())

	ctx := context.Background()
	cancellingCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(1*time.Millisecond, cancel)

	tests := []struct {
		ctx              context.Context
		readFile         readFileFunc
		childFile        string
		name             string
		previousContents []byte
		expectError      bool
	}{
		{
			ctx:              ctx,
			readFile:         os.ReadFile,
			childFile:        childFileDifferent.Name(),
			previousContents: previousContents,
			expectError:      false,
			name:             "normal case",
		},
		{
			ctx:              ctx,
			readFile:         readFileError,
			childFile:        childFileDifferent.Name(),
			previousContents: previousContents,
			expectError:      true,
			name:             "cannot read file",
		},
		{
			ctx:              ctx,
			readFile:         os.ReadFile,
			childFile:        childFileSame.Name(),
			previousContents: previousContents,
			expectError:      true,
			name:             "no new workers",
		},
		{
			ctx:              cancellingCtx,
			readFile:         os.ReadFile,
			childFile:        childFileSame.Name(),
			previousContents: previousContents,
			expectError:      true,
			name:             "context canceled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := ensureNewNginxWorkers(
				test.ctx,
				test.childFile,
				test.previousContents,
				test.readFile,
				2*time.Millisecond,
			)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
