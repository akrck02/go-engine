package install

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// TinyGo for the 3DS is tinygo-3ds's release: TinyGo with the nintendo3ds target and LLVM inside,
// built by its own workflow for each system, with the SHA-256 of each archive in SHA256SUMS.

// The tinygo-3ds release go-engine installs, and where its files are.
const (
	TinyGoRelease  = "v0.42.0-3ds.1"
	tinyGoVersion  = "0.42.0-3ds.1"
	TinyGoReleases = "https://github.com/akrck02/tinygo-3ds/releases/download/"
	checksumsFile  = "SHA256SUMS"
)

// tinyGoArchive is the name of tinygo-3ds's archive for a system.
func tinyGoArchive(system string) (string, error) {
	switch system {
	case "linux-amd64", "darwin-amd64", "darwin-arm64":
		return "tinygo" + tinyGoVersion + "." + system + ".tar.gz", nil
	case "windows-amd64":
		return "tinygo" + tinyGoVersion + "." + system + ".zip", nil
	}
	return "", failure.Wrap(NoBuildForSystem, "TinyGo for the 3DS, on "+system, nil)
}

// TinyGo downloads tinygo-3ds's release for this system, checks it and unpacks it; it returns
// the tinygo program's place.
func (installer Installer) TinyGo() (string, error) {
	archive, err := tinyGoArchive(installer.system())
	if err != nil {
		return "", err
	}
	releases := installer.TinyGoReleases
	if releases == "" {
		releases = TinyGoReleases
	}
	folder := filepath.Join(installer.Folder, "tinygo", TinyGoRelease)
	link := releases + TinyGoRelease + "/"
	sums := filepath.Join(folder, checksumsFile)
	if err := Download(link+checksumsFile, sums, nil); err != nil {
		return "", err
	}
	wanted, err := checksumOf(sums, archive)
	if err != nil {
		return "", err
	}
	if err := installer.fetchAndUnpack(link+archive, filepath.Join(folder, archive), SHA256, wanted, folder); err != nil {
		return "", err
	}
	return filepath.Join(folder, "tinygo", "bin", "tinygo"+executableSuffix(installer.system())), nil
}

// fetchAndUnpack downloads an archive, checks it, unpacks it into a folder and removes it.
func (installer Installer) fetchAndUnpack(link, archive, algorithm, wanted, folder string) error {
	if err := Download(link, archive, installer.Progress); err != nil {
		return err
	}
	defer os.Remove(archive) // once unpacked, or refused, the archive is of no use
	if err := Check(archive, algorithm, wanted); err != nil {
		return err
	}
	return Unpack(archive, folder)
}

// checksumOf finds a file's SHA-256 in a SHA256SUMS file (as sha256sum writes it).
func checksumOf(sums, file string) (string, error) {
	input, err := os.Open(sums)
	if err != nil {
		return "", failure.Wrap(ChecksumWrong, sums, err)
	}
	defer input.Close()
	lines := bufio.NewScanner(input)
	for lines.Scan() {
		fields := strings.Fields(lines.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == file {
			return fields[0], nil
		}
	}
	return "", failure.Wrap(ChecksumWrong, file+": not in "+checksumsFile, nil)
}

// executableSuffix is what programs end with on a system.
func executableSuffix(system string) string {
	if strings.HasPrefix(system, "windows") {
		return ".exe"
	}
	return ""
}
