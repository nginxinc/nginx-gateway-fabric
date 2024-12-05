package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file/filefakes"
)

func TestCopyFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	src, err := os.CreateTemp(os.TempDir(), "testfile")
	g.Expect(err).ToNot(HaveOccurred())
	defer os.Remove(src.Name())

	dest, err := os.MkdirTemp(os.TempDir(), "testdir")
	g.Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dest)

	g.Expect(copyFile(file.NewStdLibOSFileManager(), src.Name(), dest)).To(Succeed())
	_, err = os.Stat(filepath.Join(dest, filepath.Base(src.Name())))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCopyFileErrors(t *testing.T) {
	t.Parallel()

	openErr := errors.New("open error")
	createErr := errors.New("create error")
	copyErr := errors.New("copy error")

	tests := []struct {
		fileMgr *filefakes.FakeOSFileManager
		expErr  error
		name    string
	}{
		{
			name: "can't open src file",
			fileMgr: &filefakes.FakeOSFileManager{
				OpenStub: func(_ string) (*os.File, error) {
					return nil, openErr
				},
			},
			expErr: openErr,
		},
		{
			name: "can't create dest file",
			fileMgr: &filefakes.FakeOSFileManager{
				CreateStub: func(_ string) (*os.File, error) {
					return nil, createErr
				},
			},
			expErr: createErr,
		},
		{
			name: "can't copy contents",
			fileMgr: &filefakes.FakeOSFileManager{
				CopyStub: func(_ io.Writer, _ io.Reader) error {
					return copyErr
				},
			},
			expErr: copyErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			g := NewWithT(t)
			err := copyFile(test.fileMgr, "source", "destDir")

			g.Expect(err).To(MatchError(test.expErr))
		})
	}
}
