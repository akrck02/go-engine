// Package install downloads and unpacks the tools that build games for consoles and phones, into
// the user's data folder: a file over HTTP with its progress, checked against the checksum its
// publisher gives, and unpacked from a .tar.gz or a .zip. It uses the standard library only.
package install

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// Installer installs tools into a folder, saying the progress of what it downloads.
type Installer struct {
	Folder   string   // where tools go: DataFolder, or a test's own
	Progress Progress // hears each download's progress; may be nil
	// TinyGoReleases is where tinygo-3ds's releases are downloaded from: TinyGoReleases, or a
	// test's server.
	TinyGoReleases string
	// AndroidRepository is where Google's Android repository is: AndroidRepository, or a test's
	// server.
	AndroidRepository string
	// AzaharLatest is where GitHub says what Azahar's latest release holds: AzaharLatest, or a
	// test's server.
	AzaharLatest string
	System       string // the system and processor to install for, as "linux-amd64"; empty is this one
}

// system is the system and processor the tools are for.
func (installer Installer) system() string {
	if installer.System != "" {
		return installer.System
	}
	return runtime.GOOS + "-" + runtime.GOARCH
}

// Progress hears how much of a download has arrived, in bytes; total is 0 when the server does not
// say.
type Progress func(done, total int64)

// Download fetches a link into a file, saying its progress; the file appears only once complete.
func Download(link, file string, progress Progress) error {
	response, err := http.Get(link)
	if err != nil {
		return failure.Wrap(DownloadFailed, link, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return failure.Wrap(DownloadFailed, link+": "+response.Status, nil)
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return failure.Wrap(DownloadFailed, file, err)
	}
	partial := file + ".part"
	if err := copyWithProgress(partial, response.Body, response.ContentLength, progress); err != nil {
		os.Remove(partial) // what arrived is of no use; a failure to remove it changes nothing
		return failure.Wrap(DownloadFailed, link, err)
	}
	return os.Rename(partial, file)
}

// copyWithProgress writes what a reader gives into a file, telling the progress as it goes.
func copyWithProgress(file string, source io.Reader, total int64, progress Progress) error {
	output, err := os.Create(file)
	if err != nil {
		return err
	}
	counted := &countingReader{source: source, total: total, progress: progress}
	if _, err := io.Copy(output, counted); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

// countingReader tells a Progress how much was read.
type countingReader struct {
	source   io.Reader
	done     int64
	total    int64
	progress Progress
}

func (counted *countingReader) Read(buffer []byte) (int, error) {
	read, err := counted.source.Read(buffer)
	counted.done += int64(read)
	if counted.progress != nil && read > 0 {
		counted.progress(counted.done, counted.total)
	}
	return read, err
}

// Checksum algorithms, as publishers name them.
const (
	SHA256 = "sha256"
	SHA1   = "sha1"
)

// Check tells whether a file's checksum, by an algorithm, is the one wanted (in hexadecimal).
func Check(file, algorithm, wanted string) error {
	input, err := os.Open(file)
	if err != nil {
		return failure.Wrap(ChecksumWrong, file, err)
	}
	defer input.Close()
	var summer hash.Hash = sha256.New()
	if algorithm == SHA1 {
		summer = sha1.New()
	}
	if _, err := io.Copy(summer, input); err != nil {
		return failure.Wrap(ChecksumWrong, file, err)
	}
	if got := hex.EncodeToString(summer.Sum(nil)); !strings.EqualFold(got, wanted) {
		return failure.Wrap(ChecksumWrong, filepath.Base(file)+": "+got+", not "+wanted, nil)
	}
	return nil
}

// DataFolder is go-engine's folder in the user's data: where installed tools go.
func DataFolder() (string, error) {
	base, err := userDataFolder()
	if err != nil {
		return "", failure.Wrap(DataFolderNotFound, "", err)
	}
	return filepath.Join(base, "go-engine"), nil
}

// userDataFolder is where a system keeps a user's applications' data: $XDG_DATA_HOME or
// ~/.local/share on Linux, Application Support on a Mac, AppData\Local on Windows.
func userDataFolder() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.UserCacheDir() // %LocalAppData%
	case "darwin":
		return os.UserConfigDir() // ~/Library/Application Support
	}
	if folder := os.Getenv("XDG_DATA_HOME"); folder != "" {
		return folder, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}
