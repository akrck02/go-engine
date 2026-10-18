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
	"time"

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

// The downloads' patience: how long a server may take to answer, and to send more of a file.
const (
	answerTimeout = 30 * time.Second
	stallTimeout  = 60 * time.Second
)

// client downloads, giving up on a server that does not answer.
var client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, ResponseHeaderTimeout: answerTimeout}}

// Download fetches a link into a file, saying its progress; the file appears only once complete.
func Download(link, file string, progress Progress) error {
	return DownloadChecked(link, file, "", "", progress)
}

// DownloadChecked fetches a link into a file, as Download, and when wanted is not empty checks it
// against that checksum before it takes the file's place: a file that is not the one published
// never replaces one that is.
func DownloadChecked(link, file, algorithm, wanted string, progress Progress) error {
	response, err := client.Get(link)
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
	stalled := time.AfterFunc(stallTimeout, func() { response.Body.Close() })
	defer stalled.Stop()
	body := &countingReader{source: response.Body, total: response.ContentLength, progress: progress, alive: func() { stalled.Reset(stallTimeout) }}
	if err := writeAll(partial, body); err != nil {
		os.Remove(partial) // what arrived is of no use; a failure to remove it changes nothing
		return failure.Wrap(DownloadFailed, link, err)
	}
	if wanted != "" {
		if err := Check(partial, algorithm, wanted); err != nil {
			os.Remove(partial) // not the file published: it must not stay, nor run
			return err
		}
	}
	if err := os.Rename(partial, file); err != nil {
		os.Remove(partial) // where it could not go, it is of no use
		return failure.Wrap(DownloadFailed, file, err)
	}
	return nil
}

// writeAll writes what a reader gives into a file.
func writeAll(file string, source io.Reader) error {
	output, err := os.Create(file)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, source); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

// countingReader tells a Progress how much was read, and alive that something was.
type countingReader struct {
	source   io.Reader
	done     int64
	total    int64
	progress Progress
	alive    func()
}

func (counted *countingReader) Read(buffer []byte) (int, error) {
	read, err := counted.source.Read(buffer)
	counted.done += int64(read)
	if read > 0 {
		counted.alive()
	}
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
