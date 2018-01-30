package tarfs

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func Extract(src io.Reader, dest string) error {
	tarReader := tar.NewReader(src)

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if hdr.Name == "." {
			continue
		}

		err = extractTarArchiveFile(hdr, dest, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTarArchiveFile(header *tar.Header, dest string, input io.Reader) error {
	filePath := filepath.Join(dest, header.Name)
	fileInfo := header.FileInfo()

	if fileInfo.IsDir() {
		err := os.MkdirAll(filePath, fileInfo.Mode())
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return err
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return os.Symlink(header.Linkname, filePath)
		}

		if fileInfo.Mode().IsRegular() {
			fileCopy, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
			if err != nil {
				return err
			}
			defer fileCopy.Close()

			_, err = io.Copy(fileCopy, input)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
