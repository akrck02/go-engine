package install

import (
	"os"
	"path/filepath"
	"testing"
)

// makeFile makes an empty file, and the folders it is in.
func makeFile(test *testing.T, path string) string {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o755); err != nil {
		test.Fatal(err)
	}
	return path
}

func TestToolsInstalledButNotSavedAreFound(test *testing.T) {
	folder := test.TempDir()
	system := Installer{}.system()
	older := makeFile(test, tinyGoProgram(filepath.Join(folder, "tinygo", "v0.0.1-3ds.1"), system))
	wanted := makeFile(test, tinyGoProgram(filepath.Join(folder, "tinygo", TinyGoRelease), system))
	makeFile(test, tinyGoProgram(filepath.Join(folder, "tinygo", "v9.9.9-3ds.1"), system))
	azahar := makeFile(test, filepath.Join(folder, "azahar", "2100.1", "azahar.AppImage"))
	makeFile(test, filepath.Join(folder, "azahar", "2000.1", "azahar.AppImage"))
	makeFile(test, filepath.Join(folder, androidSDKFolder, "platform-tools", "adb"))
	installed := InstalledIn(folder)
	if installed.TinyGo != wanted || installed.TinyGo == older {
		test.Errorf("TinyGo installed is %q, not the release go-engine installs, %q", installed.TinyGo, wanted)
	}
	if installed.Azahar != azahar {
		test.Errorf("Azahar installed is %q, not the newest, %q", installed.Azahar, azahar)
	}
	if installed.AndroidSDK != filepath.Join(folder, androidSDKFolder) {
		test.Errorf("the SDK installed is %q", installed.AndroidSDK)
	}
}

func TestAnInstallThatStoppedHalfwayIsNotFound(test *testing.T) {
	folder := test.TempDir()
	makeFile(test, filepath.Join(folder, androidSDKFolder, "cmdline-tools", "latest", "bin", "sdkmanager"))
	if err := os.MkdirAll(filepath.Join(folder, "tinygo", TinyGoRelease), 0o755); err != nil {
		test.Fatal(err)
	}
	if installed := InstalledIn(folder); installed != (Installed{}) {
		test.Fatalf("an install that stopped halfway is found as %+v", installed)
	}
}
