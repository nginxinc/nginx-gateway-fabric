package runtime

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

type fakeEntry struct {
	name string
}

func (e *fakeEntry) Name() string {
	return e.name
}
func (*fakeEntry) IsDir() bool                { return false }
func (*fakeEntry) Type() fs.FileMode          { return 0 }
func (*fakeEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestFindMainProcess(t *testing.T) {
	readProcDir := func() ([]os.DirEntry, error) {
		return []os.DirEntry{
			&fakeEntry{name: "x"},
			&fakeEntry{name: "1"},
			&fakeEntry{name: "2"},
		}, nil
	}
	readProcDirError := func() ([]os.DirEntry, error) {
		return nil, errors.New("error")
	}

	readFile := func(name string) ([]byte, error) {
		results := map[string][]byte{
			"/proc/x/cmdline": []byte("test"),
			"/proc/1/cmdline": []byte("nginx: worker process"),
			"/proc/2/cmdline": []byte("nginx: master process /usr/sbin/nginx -c /etc/nginx/nginx.conf"),
		}

		res, exits := results[name]
		if !exits {
			return nil, errors.New("error")
		}

		return res, nil
	}
	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	tests := []struct {
		readProcDirFunc func() ([]os.DirEntry, error)
		readFileFunc    func(string) ([]byte, error)
		expected        int
		expectError     bool
		msg             string
	}{
		{
			readProcDirFunc: readProcDir,
			readFileFunc:    readFile,
			expected:        2,
			expectError:     false,
			msg:             "normal case",
		},
		{
			readProcDirFunc: readProcDirError,
			readFileFunc:    readFile,
			expected:        0,
			expectError:     true,
			msg:             "failed to read proc dir",
		},
		{
			readProcDirFunc: readProcDir,
			readFileFunc:    readFileError,
			expected:        0,
			expectError:     true,
			msg:             "failed to read file",
		},
	}

	for _, test := range tests {
		result, err := findMainProcess(test.readProcDirFunc, test.readFileFunc)

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
