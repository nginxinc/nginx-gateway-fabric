package runtime

import (
	"errors"
	"testing"
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

	tests := []struct {
		readFile    readFileFunc
		expected    int
		expectError bool
		msg         string
	}{
		{
			readFile:    readFileFuncGen([]byte("1\n")),
			expected:    1,
			expectError: false,
			msg:         "normal case",
		},
		{
			readFile:    readFileFuncGen([]byte("")),
			expected:    0,
			expectError: true,
			msg:         "empty file content",
		},
		{
			readFile:    readFileFuncGen([]byte("not a number")),
			expected:    0,
			expectError: true,
			msg:         "bad file content",
		},
		{
			readFile:    readFileError,
			expected:    0,
			expectError: true,
			msg:         "cannot read file",
		},
	}

	for _, test := range tests {
		result, err := findMainProcess(test.readFile)

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
