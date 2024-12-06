package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/licensing/licensingfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file/filefakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestInitialize_OSS(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeFileMgr := &filefakes.FakeOSFileManager{}

	ic := initializeConfig{
		fileManager: fakeFileMgr,
		logger:      zap.New(),
		copy: copyFiles{
			destDirName:  "destDir",
			srcFileNames: []string{"src1", "src2"},
		},
		plus: false,
	}

	err := initialize(ic)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(fakeFileMgr.CreateCallCount()).To(Equal(2))
	g.Expect(fakeFileMgr.OpenCallCount()).To(Equal(2))
	g.Expect(fakeFileMgr.CopyCallCount()).To(Equal(2))
}

func TestInitialize_OSS_Error(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	openErr := errors.New("open error")
	fakeFileMgr := &filefakes.FakeOSFileManager{
		OpenStub: func(_ string) (*os.File, error) {
			return nil, openErr
		},
	}

	ic := initializeConfig{
		fileManager: fakeFileMgr,
		logger:      zap.New(),
		copy: copyFiles{
			destDirName:  "destDir",
			srcFileNames: []string{"src1", "src2"},
		},
		plus: false,
	}

	err := initialize(ic)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(openErr))
}

func TestInitialize_Plus(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeFileMgr := &filefakes.FakeOSFileManager{}
	fakeCollector := &licensingfakes.FakeCollector{}

	ic := initializeConfig{
		fileManager: fakeFileMgr,
		logger:      zap.New(),
		collector:   fakeCollector,
		copy: copyFiles{
			destDirName:  "destDir",
			srcFileNames: []string{"src1", "src2"},
		},
		plus: true,
	}

	err := initialize(ic)
	g.Expect(err).ToNot(HaveOccurred())
	// copies
	g.Expect(fakeFileMgr.OpenCallCount()).To(Equal(2))
	g.Expect(fakeFileMgr.CopyCallCount()).To(Equal(2))

	// 2 copies, 1 write deploy ctx
	g.Expect(fakeFileMgr.CreateCallCount()).To(Equal(3))
	// write deploy ctx
	g.Expect(fakeCollector.CollectCallCount()).To(Equal(1))
	g.Expect(fakeFileMgr.WriteCallCount()).To(Equal(1))
	g.Expect(fakeFileMgr.ChmodCallCount()).To(Equal(1))
}

func TestInitialize_Plus_Error(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	collectErr := errors.New("collect error")
	fakeFileMgr := &filefakes.FakeOSFileManager{}
	fakeCollector := &licensingfakes.FakeCollector{
		CollectStub: func(_ context.Context) (dataplane.DeploymentContext, error) {
			return dataplane.DeploymentContext{}, collectErr
		},
	}

	ic := initializeConfig{
		fileManager: fakeFileMgr,
		logger:      zap.New(),
		collector:   fakeCollector,
		copy: copyFiles{
			destDirName:  "destDir",
			srcFileNames: []string{"src1", "src2"},
		},
		plus: true,
	}

	err := initialize(ic)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(collectErr))
}

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
