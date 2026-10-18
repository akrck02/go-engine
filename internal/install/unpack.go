package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// Unpack unpacks a .tar.gz or a .zip into a folder, keeping which files are programs. A name that
// would land outside the folder (../, an absolute path, a link out of it) stops it.
func Unpack(archive, folder string) error {
	var err error
	switch {
	case strings.HasSuffix(archive, ".zip"):
		err = unpackZip(archive, folder)
	case strings.HasSuffix(archive, ".tar.gz"), strings.HasSuffix(archive, ".tgz"):
		err = unpackTarGzip(archive, folder)
	default:
		return failure.Wrap(UnpackFailed, filepath.Base(archive)+": not a .tar.gz or a .zip", nil)
	}
	if err != nil && !failure.Is(err, PathOutside) {
		return failure.Wrap(UnpackFailed, filepath.Base(archive), err)
	}
	return err
}

// inside is where a name of an archive goes in a folder, or PathOutside.
func inside(folder, name string) (string, error) {
	place := filepath.Join(folder, filepath.FromSlash(name))
	relative, err := filepath.Rel(folder, place)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(name) {
		return "", failure.Wrap(PathOutside, name, nil)
	}
	return place, nil
}

// unpackZip unpacks a .zip.
func unpackZip(archive, folder string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, entry := range reader.File {
		if err := unpackZipEntry(entry, folder); err != nil {
			return err
		}
	}
	return nil
}

// unpackZipEntry writes one file or folder of a .zip.
func unpackZipEntry(entry *zip.File, folder string) error {
	place, err := inside(folder, entry.Name)
	if err != nil {
		return err
	}
	if entry.FileInfo().IsDir() {
		return os.MkdirAll(place, 0o755)
	}
	if entry.Mode()&fs.ModeSymlink != 0 {
		return failure.Wrap(PathOutside, entry.Name+": a link, which zips of tools do not hold", nil)
	}
	source, err := entry.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	return writeFile(place, source, entry.Mode())
}

// unpackTarGzip unpacks a .tar.gz.
func unpackTarGzip(archive, folder string) error {
	input, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer input.Close()
	unzipped, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	reader := tar.NewReader(unzipped)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := unpackTarEntry(header, reader, folder); err != nil {
			return err
		}
	}
}

// unpackTarEntry writes one file, folder or link of a tar.
func unpackTarEntry(header *tar.Header, reader io.Reader, folder string) error {
	place, err := inside(folder, header.Name)
	if err != nil {
		return err
	}
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(place, 0o755)
	case tar.TypeReg:
		return writeFile(place, reader, header.FileInfo().Mode())
	case tar.TypeSymlink:
		if _, err := inside(folder, filepath.Join(filepath.Dir(header.Name), header.Linkname)); err != nil || filepath.IsAbs(header.Linkname) {
			return failure.Wrap(PathOutside, header.Name+" → "+header.Linkname, nil)
		}
		if err := os.MkdirAll(filepath.Dir(place), 0o755); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, place)
	}
	return nil // other kinds (devices, pipes) are not in tools' archives
}

// writeFile writes a file of an archive, a program if the archive says so.
func writeFile(place string, source io.Reader, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(place), 0o755); err != nil {
		return err
	}
	output, err := os.OpenFile(place, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm()|0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, source); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}
