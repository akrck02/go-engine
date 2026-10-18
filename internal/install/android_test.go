package install

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// androidServer serves a made-up Android repository: its index, with a licence, and the
// command-line tools for Linux, whose sdkmanager writes what it was asked to a file.
func androidServer(test *testing.T, calls string) *httptest.Server {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	header := &zip.FileHeader{Name: "cmdline-tools/bin/sdkmanager", Method: zip.Deflate}
	header.SetMode(0o755)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		test.Fatal(err)
	}
	entry.Write([]byte("#!/bin/sh\necho \"$@\" >> " + calls + "\ncat >> " + calls + "\n"))
	writer.Close()
	tools := buffer.Bytes()
	sum := sha1.Sum(tools)
	index := `<?xml version="1.0"?><sdk:sdk-repository xmlns:sdk="http://schemas.android.com/sdk/android/repo/repository2/03">
<license id="android-sdk-license" type="text">Use the lighthouse kindly.</license>
<remotePackage path="cmdline-tools;latest"><archives><archive><complete><size>1</size>
<checksum type="sha1">` + hex.EncodeToString(sum[:]) + `</checksum><url>tools-linux.zip</url></complete><host-os>linux</host-os></archive></archives></remotePackage>
</sdk:sdk-repository>`
	files := map[string][]byte{"/" + androidIndex: []byte(index), "/tools-linux.zip": tools}
	return httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		content, found := files[request.URL.Path]
		if !found {
			http.NotFound(response, request)
			return
		}
		response.Write(content)
	}))
}

// fakeJava makes a JDK of a made-up java, and sets $JAVA_HOME to it.
func fakeJava(test *testing.T) {
	home := test.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "java"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		test.Fatal(err)
	}
	test.Setenv("JAVA_HOME", home)
}

func TestAndroidsLicenceIsReadAndItsPackagesInstalled(test *testing.T) {
	if runtime.GOOS == "windows" {
		test.Skip("the fake sdkmanager is a shell script")
	}
	calls := filepath.Join(test.TempDir(), "calls")
	server := androidServer(test, calls)
	defer server.Close()
	fakeJava(test)
	installer := Installer{Folder: test.TempDir(), AndroidRepository: server.URL + "/", System: "linux-amd64"}
	if licence, err := installer.AndroidLicence(); err != nil || licence != "Use the lighthouse kindly." {
		test.Fatalf("the licence reads %q, %v", licence, err)
	}
	sdk, err := installer.AndroidSDK()
	if err != nil {
		test.Fatal(err)
	}
	said, _ := os.ReadFile(calls)
	for _, wanted := range []string{"--sdk_root=" + sdk + " --licenses\ny\n", "--sdk_root=" + sdk + " " + strings.Join(AndroidPackages, " ")} {
		if !strings.Contains(string(said), wanted) {
			test.Errorf("sdkmanager was not asked %q; it was asked:\n%s", wanted, said)
		}
	}
	if _, err := os.Stat(filepath.Join(sdk, "cmdline-tools", "latest", "bin", "sdkmanager")); err != nil {
		test.Fatalf("the command-line tools are not where sdkmanager wants them: %v", err)
	}
}

func TestAndroidWithoutJavaIsSaidByItsCode(test *testing.T) {
	test.Setenv("JAVA_HOME", "")
	test.Setenv("PATH", test.TempDir())
	if _, err := (Installer{Folder: test.TempDir()}).AndroidSDK(); !errors.Is(err, JavaNotFound) {
		test.Fatalf("no Java gives %v", err)
	}
}

func TestEachSystemGetsItsOwnCommandLineTools(test *testing.T) {
	var repository androidRepository
	repository.Packages = append(repository.Packages, struct {
		Path     string           `xml:"path,attr"`
		Archives []androidArchive `xml:"archives>archive"`
	}{Path: androidCommandLine, Archives: []androidArchive{
		{URL: "linux.zip", System: "linux"}, {URL: "mac-x64.zip", System: "macosx", Machine: "x64"},
		{URL: "mac-arm.zip", System: "macosx", Machine: "aarch64"}, {URL: "win.zip", System: "windows"},
	}})
	cases := map[string]string{"linux-amd64": "linux.zip", "darwin-amd64": "mac-x64.zip", "darwin-arm64": "mac-arm.zip", "windows-amd64": "win.zip"}
	for system, wanted := range cases {
		if archive, err := repository.archiveFor(androidCommandLine, system); err != nil || archive.URL != wanted {
			test.Errorf("%s gets %q, %v", system, archive.URL, err)
		}
	}
	if _, err := repository.archiveFor(androidCommandLine, "plan9-386"); !errors.Is(err, NoBuildForSystem) {
		test.Errorf("a system with no tools gives %v", err)
	}
}
