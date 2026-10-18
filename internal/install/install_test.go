package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestADownloadArrivesWholeAndSaysItsProgress(test *testing.T) {
	content := bytes.Repeat([]byte("lighthouse "), 10000)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/lamp.tar.gz" {
			http.NotFound(writer, request)
			return
		}
		writer.Write(content)
	}))
	defer server.Close()
	file := filepath.Join(test.TempDir(), "lamp.tar.gz")
	var last int64
	if err := Download(server.URL+"/lamp.tar.gz", file, func(done, total int64) { last = done }); err != nil {
		test.Fatal(err)
	}
	if read, _ := os.ReadFile(file); !bytes.Equal(read, content) || last != int64(len(content)) {
		test.Fatalf("%d bytes arrived, %d said, of %d", len(read), last, len(content))
	}
	if err := Download(server.URL+"/nowhere", file+"2", nil); !errors.Is(err, DownloadFailed) {
		test.Fatalf("a link that is not there gives %v", err)
	}
	if _, err := os.Stat(file + "2.part"); err == nil {
		test.Fatal("a failed download leaves its part behind")
	}
}

func TestAFileIsCheckedAgainstItsChecksum(test *testing.T) {
	file := filepath.Join(test.TempDir(), "lamp")
	if err := os.WriteFile(file, []byte("lamp"), 0o644); err != nil {
		test.Fatal(err)
	}
	sum := sha256.Sum256([]byte("lamp"))
	if err := Check(file, SHA256, hex.EncodeToString(sum[:])); err != nil {
		test.Fatalf("the right checksum gives %v", err)
	}
	if err := Check(file, SHA256, "00"); !errors.Is(err, ChecksumWrong) {
		test.Fatalf("a wrong checksum gives %v", err)
	}
	if err := Check(file, SHA1, "b9b9b2b6c5d6a0e7d57d2b1b2f0a3e6a0f4c1d2e"); !errors.Is(err, ChecksumWrong) {
		test.Fatalf("a wrong sha1 gives %v", err)
	}
}

// tarGzip makes a .tar.gz of files, by name, with their modes.
func tarGzip(test *testing.T, file string, entries []tar.Header, contents []string) {
	var buffer bytes.Buffer
	zipped := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(zipped)
	for index, header := range entries {
		header.Size = int64(len(contents[index]))
		if err := writer.WriteHeader(&header); err != nil {
			test.Fatal(err)
		}
		writer.Write([]byte(contents[index]))
	}
	writer.Close()
	zipped.Close()
	if err := os.WriteFile(file, buffer.Bytes(), 0o644); err != nil {
		test.Fatal(err)
	}
}

func TestATarGzipIsUnpackedWithItsPrograms(test *testing.T) {
	if runtime.GOOS == "windows" {
		test.Skip("Windows has no program bit")
	}
	folder := test.TempDir()
	archive := filepath.Join(folder, "tool.tar.gz")
	tarGzip(test, archive, []tar.Header{
		{Name: "tool/bin/tool", Mode: 0o755, Typeflag: tar.TypeReg},
		{Name: "tool/README", Mode: 0o644, Typeflag: tar.TypeReg},
		{Name: "tool/bin/alias", Linkname: "tool", Typeflag: tar.TypeSymlink},
	}, []string{"#!/bin/sh\n", "read me", ""})
	target := filepath.Join(folder, "out")
	if err := Unpack(archive, target); err != nil {
		test.Fatal(err)
	}
	details, err := os.Stat(filepath.Join(target, "tool", "bin", "alias"))
	if err != nil || details.Mode().Perm()&0o100 == 0 {
		test.Fatalf("the program (through its link) is not one: %v, %v", details, err)
	}
}

func TestAnArchiveThatNamesAFileOutsideIsRefused(test *testing.T) {
	folder := test.TempDir()
	archive := filepath.Join(folder, "bad.tar.gz")
	tarGzip(test, archive, []tar.Header{{Name: "../escaped", Mode: 0o644, Typeflag: tar.TypeReg}}, []string{"x"})
	if err := Unpack(archive, filepath.Join(folder, "out")); !errors.Is(err, PathOutside) {
		test.Fatalf("../ in a tar gives %v", err)
	}
	linked := filepath.Join(folder, "link.tar.gz")
	tarGzip(test, linked, []tar.Header{{Name: "tool/etc", Linkname: "../../../etc", Typeflag: tar.TypeSymlink}}, []string{""})
	if err := Unpack(linked, filepath.Join(folder, "out2")); !errors.Is(err, PathOutside) {
		test.Fatalf("a link out of the folder gives %v", err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entry, _ := writer.Create("../../escaped.txt")
	entry.Write([]byte("x"))
	writer.Close()
	zipped := filepath.Join(folder, "bad.zip")
	if err := os.WriteFile(zipped, buffer.Bytes(), 0o644); err != nil {
		test.Fatal(err)
	}
	if err := Unpack(zipped, filepath.Join(folder, "out3")); !errors.Is(err, PathOutside) {
		test.Fatalf("../ in a zip gives %v", err)
	}
}

func TestAZipIsUnpacked(test *testing.T) {
	folder := test.TempDir()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entry, _ := writer.Create("tools/notes.txt")
	entry.Write([]byte("keeper"))
	writer.Close()
	archive := filepath.Join(folder, "tools.zip")
	if err := os.WriteFile(archive, buffer.Bytes(), 0o644); err != nil {
		test.Fatal(err)
	}
	if err := Unpack(archive, filepath.Join(folder, "out")); err != nil {
		test.Fatal(err)
	}
	if read, _ := os.ReadFile(filepath.Join(folder, "out", "tools", "notes.txt")); string(read) != "keeper" {
		test.Fatalf("the zip's file holds %q", read)
	}
}
