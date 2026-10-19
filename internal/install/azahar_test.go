package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// azaharServer serves a made-up latest release of Azahar for Linux, its AppImage's digest the
// right one or not.
func azaharServer(test *testing.T, rightDigest bool) *httptest.Server {
	image := []byte("an emulator")
	sum := sha256.Sum256(image)
	digest := hex.EncodeToString(sum[:])
	if !rightDigest {
		digest = "00" + digest[2:]
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/latest":
			response.Write([]byte(`{"tag_name": "2100.1", "assets": [{"name": "azahar.AppImage", "browser_download_url": "` +
				server.URL + `/azahar.AppImage", "digest": "sha256:` + digest + `"}]}`))
		case "/azahar.AppImage":
			response.Write(image)
		default:
			http.NotFound(response, request)
		}
	}))
	return server
}

func TestAzaharIsDownloadedCheckedAndMadeAProgram(test *testing.T) {
	server := azaharServer(test, true)
	defer server.Close()
	installer := Installer{Folder: test.TempDir(), AzaharLatest: server.URL + "/latest", System: "linux-amd64"}
	program, err := installer.Azahar()
	if err != nil {
		test.Fatal(err)
	}
	details, err := os.Stat(program)
	if err != nil || details.Mode().Perm()&0o100 == 0 {
		test.Fatalf("Azahar at %q is not a program: %v", program, err)
	}
}

func TestAzaharThatDoesNotMatchItsDigestIsRemoved(test *testing.T) {
	server := azaharServer(test, false)
	defer server.Close()
	installer := Installer{Folder: test.TempDir(), AzaharLatest: server.URL + "/latest", System: "linux-amd64"}
	working := filepath.Join(installer.Folder, "azahar", "2100.1", "azahar.AppImage")
	if err := os.MkdirAll(filepath.Dir(working), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(working, []byte("the one that works"), 0o755); err != nil {
		test.Fatal(err)
	}
	if _, err := installer.Azahar(); !errors.Is(err, ChecksumWrong) {
		test.Fatalf("an AppImage with another digest gives %v", err)
	}
	if read, _ := os.ReadFile(working); string(read) != "the one that works" {
		test.Fatal("a download refused replaced the Azahar that worked")
	}
}

func TestEachSystemGetsItsOwnAzahar(test *testing.T) {
	cases := map[string]string{"linux-amd64": "azahar.AppImage", "darwin-arm64": "azahar-macos-arm64-2100.1.zip", "windows-amd64": "azahar-windows-msvc-2100.1.zip"}
	for system, wanted := range cases {
		if name, err := azaharFile(system, "2100.1"); err != nil || name != wanted {
			test.Errorf("%s gets %q, %v", system, name, err)
		}
	}
	if _, err := azaharFile("plan9-386", "2100.1"); !errors.Is(err, NoBuildForSystem) {
		test.Errorf("a system with no Azahar gives %v", err)
	}
}
