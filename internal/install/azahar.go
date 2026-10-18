package install

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// Azahar, the 3DS emulator, is installed from its latest release on GitHub, which gives each
// file's SHA-256: the AppImage on Linux, the zip of its program on a Mac and on Windows.

// AzaharLatest is where GitHub says what Azahar's latest release holds.
const AzaharLatest = "https://api.github.com/repos/azahar-emu/azahar/releases/latest"

// azaharRelease is what an install reads of Azahar's latest release.
type azaharRelease struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Name   string `json:"name"`
		Link   string `json:"browser_download_url"`
		Digest string `json:"digest"` // "sha256:…"
	} `json:"assets"`
}

// Azahar downloads Azahar's latest release for this system, checks it and returns its program.
func (installer Installer) Azahar() (string, error) {
	release, err := installer.azaharRelease()
	if err != nil {
		return "", err
	}
	name, err := azaharFile(installer.system(), release.Tag)
	if err != nil {
		return "", err
	}
	for _, asset := range release.Assets {
		if asset.Name == name {
			return installer.installAzahar(asset.Link, asset.Name, strings.TrimPrefix(asset.Digest, "sha256:"), release.Tag)
		}
	}
	return "", failure.Wrap(NoBuildForSystem, "Azahar "+release.Tag+": no "+name, nil)
}

// azaharRelease asks GitHub what Azahar's latest release holds.
func (installer Installer) azaharRelease() (azaharRelease, error) {
	var release azaharRelease
	link := installer.AzaharLatest
	if link == "" {
		link = AzaharLatest
	}
	response, err := http.Get(link)
	if err != nil {
		return release, failure.Wrap(DownloadFailed, link, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return release, failure.Wrap(DownloadFailed, link+": "+response.Status, nil)
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return release, failure.Wrap(DownloadFailed, link, err)
	}
	return release, nil
}

// azaharFile is the name of Azahar's file for a system, in a release.
func azaharFile(system, tag string) (string, error) {
	switch system {
	case "linux-amd64":
		return "azahar.AppImage", nil
	case "darwin-arm64":
		return "azahar-macos-arm64-" + tag + ".zip", nil
	case "darwin-amd64":
		return "azahar-macos-x86_64-" + tag + ".zip", nil
	case "windows-amd64":
		return "azahar-windows-msvc-" + tag + ".zip", nil
	}
	return "", failure.Wrap(NoBuildForSystem, "Azahar, on "+system, nil)
}

// installAzahar downloads and checks Azahar's file, unpacks it when it is a zip, and returns its
// program.
func (installer Installer) installAzahar(link, name, wanted, tag string) (string, error) {
	folder := filepath.Join(installer.Folder, "azahar", tag)
	file := filepath.Join(folder, name)
	if !strings.HasSuffix(name, ".zip") {
		if err := Download(link, file, installer.Progress); err != nil {
			return "", err
		}
		if err := Check(file, SHA256, wanted); err != nil {
			os.Remove(file) // a file that is not the one published must not run
			return "", err
		}
		return file, os.Chmod(file, 0o755)
	}
	if err := installer.fetchAndUnpack(link, file, SHA256, wanted, folder); err != nil {
		return "", err
	}
	return findProgram(folder, "azahar"+executableSuffix(installer.system()))
}

// findProgram finds a program by its name in a folder unpacked (in a Mac's .app, it is in
// Contents/MacOS).
func findProgram(folder, name string) (string, error) {
	found := ""
	filepath.WalkDir(folder, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.EqualFold(entry.Name(), name) {
			found = path
			return fs.SkipAll
		}
		return nil // an unreadable folder is walked past
	})
	if found == "" {
		return "", failure.Wrap(UnpackFailed, name+": not in what was unpacked", nil)
	}
	return found, nil
}
