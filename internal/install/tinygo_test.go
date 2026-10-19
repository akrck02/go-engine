package install

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// releaseServer serves a tinygo-3ds release for Linux: its archive and SHA256SUMS, whose sum is
// the archive's unless the test asks for a wrong one.
func releaseServer(test *testing.T, wrongSum bool) *httptest.Server {
	archive := filepath.Join(test.TempDir(), "release.tar.gz")
	tarGzip(test, archive, []tar.Header{{Name: "tinygo/bin/tinygo", Mode: 0o755, Typeflag: tar.TypeReg}}, []string{"#!/bin/sh\n"})
	data, err := os.ReadFile(archive)
	if err != nil {
		test.Fatal(err)
	}
	sum := sha256.Sum256(data)
	written := hex.EncodeToString(sum[:])
	if wrongSum {
		written = "00" + written[2:]
	}
	name := "tinygo" + tinyGoVersion + ".linux-amd64.tar.gz"
	files := map[string][]byte{
		"/" + TinyGoRelease + "/" + name:          data,
		"/" + TinyGoRelease + "/" + checksumsFile: []byte(written + "  " + name + "\n"),
	}
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		content, found := files[request.URL.Path]
		if !found {
			http.NotFound(writer, request)
			return
		}
		writer.Write(content)
	}))
}

func TestAReleaseIsDownloadedCheckedAndItsProgramGiven(test *testing.T) {
	server := releaseServer(test, false)
	defer server.Close()
	installer := Installer{Folder: test.TempDir(), TinyGoReleases: server.URL + "/", System: "linux-amd64"}
	program, err := installer.TinyGo()
	if err != nil {
		test.Fatal(err)
	}
	if want := filepath.Join(installer.Folder, "tinygo", TinyGoRelease, "tinygo", "bin", "tinygo"); program != want {
		test.Fatalf("TinyGo is at %q, not %q", program, want)
	}
	if _, err := os.Stat(program); err != nil {
		test.Fatalf("the program is not there: %v", err)
	}
}

func TestAReleaseThatDoesNotMatchItsChecksumIsRefused(test *testing.T) {
	server := releaseServer(test, true)
	defer server.Close()
	installer := Installer{Folder: test.TempDir(), TinyGoReleases: server.URL + "/", System: "linux-amd64"}
	if _, err := installer.TinyGo(); !errors.Is(err, ChecksumWrong) {
		test.Fatalf("a release with another checksum gives %v", err)
	}
	if _, err := os.Stat(filepath.Join(installer.Folder, "tinygo", TinyGoRelease, "tinygo")); err == nil {
		test.Fatal("a release refused was unpacked")
	}
}

func TestASystemWithNoBuildIsSaidByItsCode(test *testing.T) {
	if _, err := (Installer{System: "plan9-386"}).TinyGo(); !errors.Is(err, NoBuildForSystem) {
		test.Fatalf("a system with no build gives %v", err)
	}
}
