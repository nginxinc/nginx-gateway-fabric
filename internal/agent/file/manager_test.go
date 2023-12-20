package file_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	file2 "github.com/nginxinc/nginx-gateway-fabric/internal/agent/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/file/filefakes"
)

var _ = Describe("EventHandler", func() {
	Describe("Replace files", Ordered, func() {
		var (
			mgr                                  *file2.ManagerImpl
			tmpDir                               string
			regular1, regular2, regular3, secret file2.File
		)

		ensureFiles := func(files []file2.File) {
			entries, err := os.ReadDir(tmpDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).Should(HaveLen(len(files)))

			entriesMap := make(map[string]os.DirEntry)
			for _, entry := range entries {
				entriesMap[entry.Name()] = entry
			}

			for _, f := range files {
				_, ok := entriesMap[filepath.Base(f.Path)]
				Expect(ok).Should(BeTrue())

				info, err := os.Stat(f.Path)
				Expect(err).ToNot(HaveOccurred())

				Expect(info.IsDir()).To(BeFalse())

				if f.Type == file2.TypeRegular {
					Expect(info.Mode()).To(Equal(os.FileMode(0o644)))
				} else {
					Expect(info.Mode()).To(Equal(os.FileMode(0o640)))
				}

				bytes, err := os.ReadFile(f.Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(Equal(f.Content))
			}
		}

		ensureNotExist := func(files ...file2.File) {
			for _, f := range files {
				_, err := os.Stat(f.Path)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}
		}

		BeforeAll(func() {
			mgr = file2.NewManagerImpl(zap.New(), file2.NewStdLibOSFileManager())
			tmpDir = GinkgoT().TempDir()

			regular1 = file2.File{
				Type:    file2.TypeRegular,
				Path:    filepath.Join(tmpDir, "regular-1.conf"),
				Content: []byte("regular-1"),
			}
			regular2 = file2.File{
				Type:    file2.TypeRegular,
				Path:    filepath.Join(tmpDir, "regular-2.conf"),
				Content: []byte("regular-2"),
			}
			regular3 = file2.File{
				Type:    file2.TypeRegular,
				Path:    filepath.Join(tmpDir, "regular-3.conf"),
				Content: []byte("regular-3"),
			}
			secret = file2.File{
				Type:    file2.TypeSecret,
				Path:    filepath.Join(tmpDir, "secret.conf"),
				Content: []byte("secret"),
			}
		})

		It("should write initial config", func() {
			files := []file2.File{regular1, regular2, secret}

			err := mgr.ReplaceFiles(files)
			Expect(err).ToNot(HaveOccurred())

			ensureFiles(files)
		})

		It("should write subsequent config", func() {
			files := []file2.File{
				regular2, // overwriting
				regular3, // adding
				secret,   // overwriting
			}

			err := mgr.ReplaceFiles(files)
			Expect(err).ToNot(HaveOccurred())

			ensureFiles(files)
			ensureNotExist(regular1)
		})

		It("should remove all files", func() {
			err := mgr.ReplaceFiles(nil)
			Expect(err).ToNot(HaveOccurred())

			ensureNotExist(regular2, regular3, secret)
		})
	})

	When("file does not exist", func() {
		It("should not error", func() {
			fakeOSMgr := &filefakes.FakeOSFileManager{}
			mgr := file2.NewManagerImpl(zap.New(), fakeOSMgr)

			files := []file2.File{
				{
					Type:    file2.TypeRegular,
					Path:    "regular-1.conf",
					Content: []byte("regular-1"),
				},
			}

			Expect(mgr.ReplaceFiles(files)).ToNot(HaveOccurred())

			fakeOSMgr.RemoveReturns(os.ErrNotExist)
			Expect(mgr.ReplaceFiles(files)).ToNot(HaveOccurred())
		})
	})

	When("file type is not supported", func() {
		It("should panic", func() {
			mgr := file2.NewManagerImpl(zap.New(), nil)

			files := []file2.File{
				{
					Type: 123,
					Path: "unsupported.conf",
				},
			}

			replace := func() {
				_ = mgr.ReplaceFiles(files)
			}

			Expect(replace).Should(Panic())
		})
	})

	Describe("Edge cases with IO errors", func() {
		var (
			files = []file2.File{
				{
					Type:    file2.TypeRegular,
					Path:    "regular.conf",
					Content: []byte("regular"),
				},
				{
					Type:    file2.TypeSecret,
					Path:    "secret.conf",
					Content: []byte("secret"),
				},
			}
			testErr = errors.New("test error")
		)

		DescribeTable(
			"should return error on file IO error",
			func(fakeOSMgr *filefakes.FakeOSFileManager) {
				mgr := file2.NewManagerImpl(zap.New(), fakeOSMgr)

				// special case for Remove
				// to kick off removing, we need to successfully write files beforehand
				if fakeOSMgr.RemoveStub != nil {
					err := mgr.ReplaceFiles(files)
					Expect(err).ToNot(HaveOccurred())
				}

				err := mgr.ReplaceFiles(files)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(MatchError(testErr))
			},
			Entry(
				"Remove",
				&filefakes.FakeOSFileManager{
					RemoveStub: func(s string) error {
						return testErr
					},
				},
			),
			Entry(
				"Create",
				&filefakes.FakeOSFileManager{
					CreateStub: func(s string) (*os.File, error) {
						return nil, testErr
					},
				},
			),
			Entry(
				"Chmod",
				&filefakes.FakeOSFileManager{
					ChmodStub: func(os *os.File, mode os.FileMode) error {
						return testErr
					},
				},
			),
			Entry(
				"Write",
				&filefakes.FakeOSFileManager{
					WriteStub: func(os *os.File, bytes []byte) error {
						return testErr
					},
				},
			),
		)
	})
})
