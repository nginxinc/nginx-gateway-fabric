package runtime

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"
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
		msg         string
		expected    int
		expectError bool
	}{
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    1,
			expectError: false,
			msg:         "normal case",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "empty file content",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("not a number")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "bad file content",
		},
		{
			ctx:         ctx,
			readFile:    readFileError,
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "cannot read file",
		},
		{
			ctx:         ctx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileError,
			expected:    0,
			expectError: true,
			msg:         "cannot find pid file",
		},
		{
			ctx:         cancellingCtx,
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileError,
			expected:    0,
			expectError: true,
			msg:         "context canceled",
		},
	}

	for _, test := range tests {
		result, err := findMainProcess(test.ctx, test.checkFile, test.readFile, 2*time.Millisecond)

		if result != test.expected {
			t.Errorf("findMainProcess() returned %d but expected %d for case %q", result, test.expected, test.msg)
		}

		if test.expectError {
			if err == nil {
				t.Errorf("findMainProcess() didn't return error for case %q", test.msg)
			}
		} else {
			if err != nil {
				t.Errorf("findMainProcess() returned unexpected error %v for case %q", err, test.msg)
			}
		}
	}
}
