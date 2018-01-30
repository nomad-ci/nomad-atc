package tarfs_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/concourse/go-archive/archivetest"
	"github.com/concourse/go-archive/tarfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extract", func() {
	var extractionSrc io.Reader
	var extractionDest string

	archiveFiles := archivetest.Archive{
		{
			Name: "./",
			Dir:  true,
		},
		{
			Name: "./some-file",
			Body: "some-file-contents",
		},
		{
			Name: "./empty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/file-in-dir",
			Body: "file-in-dir-contents",
		},
		{
			Name: "./legit-exe-not-a-virus.bat",
			Mode: 0644,
			Body: "rm -rf /",
		},
		{
			Name: "./some-symlink",
			Link: "some-file",
			Mode: 0755,
		},
	}

	BeforeEach(func() {
		var err error

		extractionDest, err = ioutil.TempDir("", "extracted")
		Expect(err).NotTo(HaveOccurred())

		extractionSrc, err = archiveFiles.TarStream()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(extractionDest)
	})

	JustBeforeEach(func() {
		err := tarfs.Extract(extractionSrc, extractionDest)
		Expect(err).NotTo(HaveOccurred())
	})

	extractionTest := func() {
		fileContents, err := ioutil.ReadFile(filepath.Join(extractionDest, "some-file"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("some-file-contents"))

		fileContents, err = ioutil.ReadFile(filepath.Join(extractionDest, "nonempty-dir", "file-in-dir"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("file-in-dir-contents"))

		executable, err := os.Open(filepath.Join(extractionDest, "legit-exe-not-a-virus.bat"))
		Expect(err).NotTo(HaveOccurred())

		executableInfo, err := executable.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(executableInfo.Mode()).To(Equal(os.FileMode(0644)))

		emptyDir, err := os.Open(filepath.Join(extractionDest, "empty-dir"))
		Expect(err).NotTo(HaveOccurred())

		emptyDirInfo, err := emptyDir.Stat()
		Expect(err).NotTo(HaveOccurred())

		Expect(emptyDirInfo.IsDir()).To(BeTrue())

		target, err := os.Readlink(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal("some-file"))

		symlinkInfo, err := os.Lstat(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())

		Expect(symlinkInfo.Mode() & 0755).To(Equal(os.FileMode(0755)))
	}

	Context("when 'tar' is on the PATH", func() {
		BeforeEach(func() {
			_, err := exec.LookPath("tar")
			Expect(err).NotTo(HaveOccurred())
		})

		It("extracts the TGZ's files, generating directories, and honoring file permissions and symlinks", extractionTest)
	})

	Context("when 'tar' is not in the PATH", func() {
		var oldPATH string

		BeforeEach(func() {
			oldPATH = os.Getenv("PATH")
			Expect(os.Setenv("PATH", "/dev/null")).To(Succeed())

			_, err := exec.LookPath("tar")
			Expect(err).To(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.Setenv("PATH", oldPATH)).To(Succeed())
		})

		It("extracts the TGZ's files, generating directories, and honoring file permissions and symlinks", extractionTest)
	})
})
