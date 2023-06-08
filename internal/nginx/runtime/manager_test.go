package runtime

import (
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

	tests := []struct {
		readFile    readFileFunc
		checkFile   checkFileFunc
		msg         string
		expected    int
		expectError bool
	}{
		{
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    1,
			expectError: false,
			msg:         "normal case",
		},
		{
			readFile:    readFileFuncGen([]byte("")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "empty file content",
		},
		{
			readFile:    readFileFuncGen([]byte("not a number")),
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "bad file content",
		},
		{
			readFile:    readFileError,
			checkFile:   checkFileFuncGen(testFileInfo),
			expected:    0,
			expectError: true,
			msg:         "cannot read file",
		},
		{
			readFile:    readFileFuncGen([]byte("1\n")),
			checkFile:   checkFileError,
			expected:    0,
			expectError: true,
			msg:         "cannot file pid file",
		},
	}

	for _, test := range tests {
		result, err := findMainProcess(test.checkFile, test.readFile, 1*time.Microsecond)

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
