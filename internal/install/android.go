package install

import (
	"encoding/xml"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// The Android SDK is installed as Android Studio installs it: Google's command-line tools, read
// from Google's repository (which also holds each archive's checksum and the licence), and then
// their sdkmanager, which installs the packages a game's build needs once the person has accepted
// Android's licence. The tools are Java programs: a JDK must be installed.

// Google's repository, its licence, and the packages a game's Android build needs: the
// platform-tools (adb), build-tools (aapt2, d8, apksigner), the platform (android.jar) and the NDK
// go-ui-apk builds with.
const (
	AndroidRepository  = "https://dl.google.com/android/repository/"
	androidIndex       = "repository2-3.xml"
	androidLicence     = "android-sdk-license"
	androidCommandLine = "cmdline-tools;latest"
	androidSDKFolder   = "android-sdk" // the SDK's folder, in the installer's
)

// AndroidPackages are the SDK's packages installed for a game's Android build.
var AndroidPackages = []string{"platform-tools", "build-tools;34.0.0", "platforms;android-34", "ndk;28.2.13676358"}

// androidRepository is what Google's repository index holds that an install reads.
type androidRepository struct {
	Licences []struct {
		Name string `xml:"id,attr"`
		Text string `xml:",chardata"`
	} `xml:"license"`
	Packages []struct {
		Path     string           `xml:"path,attr"`
		Archives []androidArchive `xml:"archives>archive"`
	} `xml:"remotePackage"`
}

// androidArchive is one archive of a package, for a system.
type androidArchive struct {
	URL      string `xml:"complete>url"`
	Checksum string `xml:"complete>checksum"`
	System   string `xml:"host-os"`
	Machine  string `xml:"host-arch"`
}

// AndroidLicence reads Android's licence, which the person reads and accepts before the SDK is
// installed.
func (installer Installer) AndroidLicence() (string, error) {
	repository, err := installer.androidRepository()
	if err != nil {
		return "", err
	}
	for _, licence := range repository.Licences {
		if licence.Name == androidLicence {
			return strings.TrimSpace(licence.Text), nil
		}
	}
	return "", failure.Wrap(DownloadFailed, androidLicence+": not in Google's repository", nil)
}

// AndroidSDK installs Android's command-line tools and, with them, the packages a game's build
// needs; accepted is the licence's text the person read and accepted (AndroidLicence). sdkmanager
// asks for each licence the packages need: it is told yes only for that one, word for word, and
// no for any other. sdkmanager runs on the JDK in javaHome. It returns the SDK's folder (the NDK is
// in it).
func (installer Installer) AndroidSDK(accepted, javaHome string) (string, error) {
	if strings.TrimSpace(accepted) == "" {
		return "", failure.Wrap(LicenceNotAccepted, androidLicence, nil)
	}
	if javaHome == "" {
		return "", JavaNotFound
	}
	sdk := filepath.Join(installer.Folder, androidSDKFolder)
	manager, err := installer.androidCommandLine(sdk)
	if err != nil {
		return "", err
	}
	environment := append(os.Environ(), "JAVA_HOME="+javaHome)
	answers := licenceAnswers{licence: androidLicence, text: accepted}
	if err := runAndroidManager(manager, environment, &answers, append([]string{"--sdk_root=" + sdk}, AndroidPackages...)...); err != nil {
		return "", err
	}
	return sdk, answers.refusal()
}

// androidRepository downloads and reads Google's repository index.
func (installer Installer) androidRepository() (androidRepository, error) {
	var repository androidRepository
	index := filepath.Join(installer.Folder, androidSDKFolder, androidIndex)
	if err := Download(installer.androidLink(androidIndex), index, nil); err != nil {
		return repository, err
	}
	data, err := os.ReadFile(index)
	if err == nil {
		err = xml.Unmarshal(data, &repository)
	}
	if err != nil {
		return repository, failure.Wrap(DownloadFailed, androidIndex, err)
	}
	return repository, nil
}

// androidLink is where a file of Google's repository is.
func (installer Installer) androidLink(file string) string {
	if installer.AndroidRepository != "" {
		return installer.AndroidRepository + file
	}
	return AndroidRepository + file
}

// androidCommandLine installs the command-line tools into the SDK's cmdline-tools/latest, as
// sdkmanager wants them, and returns sdkmanager's place.
func (installer Installer) androidCommandLine(sdk string) (string, error) {
	repository, err := installer.androidRepository()
	if err != nil {
		return "", err
	}
	archive, err := repository.archiveFor(androidCommandLine, installer.system())
	if err != nil {
		return "", err
	}
	unpacked := filepath.Join(sdk, "cmdline-tools", "unpacked")
	os.RemoveAll(unpacked) // a half unpacked earlier try; when it is not there, nothing happens
	if err := installer.fetchAndUnpack(installer.androidLink(archive.URL), filepath.Join(sdk, archive.URL), SHA1, archive.Checksum, unpacked); err != nil {
		return "", err
	}
	latest := filepath.Join(sdk, "cmdline-tools", "latest")
	os.RemoveAll(latest) // the tools installed before, replaced by these
	if err := os.Rename(filepath.Join(unpacked, "cmdline-tools"), latest); err != nil {
		return "", failure.Wrap(UnpackFailed, archive.URL, err)
	}
	os.Remove(unpacked) // empty once its folder moved
	manager := filepath.Join(latest, "bin", "sdkmanager")
	if strings.HasPrefix(installer.system(), "windows") {
		manager += ".bat"
	}
	return manager, nil
}

// archiveFor is a package's archive for a system ("linux-amd64").
func (repository androidRepository) archiveFor(path, system string) (androidArchive, error) {
	operating, machine, _ := strings.Cut(system, "-")
	wantedSystem := map[string]string{"linux": "linux", "darwin": "macosx", "windows": "windows"}[operating]
	wantedMachine := map[string]string{"amd64": "x64", "arm64": "aarch64"}[machine]
	for _, remote := range repository.Packages {
		if remote.Path != path {
			continue
		}
		for _, archive := range remote.Archives {
			if archive.System == wantedSystem && (archive.Machine == "" || archive.Machine == wantedMachine) {
				return archive, nil
			}
		}
	}
	return androidArchive{}, failure.Wrap(NoBuildForSystem, path+", on "+system, nil)
}

// runAndroidManager runs sdkmanager, answering the licences it asks about as it asks.
func runAndroidManager(manager string, environment []string, answers *licenceAnswers, arguments ...string) error {
	command := exec.Command(manager, arguments...)
	command.Env = environment
	output, err := command.StdoutPipe()
	if err != nil {
		return failure.Wrap(AndroidManagerFailed, manager, err)
	}
	command.Stderr = command.Stdout // one stream: its questions and its errors, in order
	input, err := command.StdinPipe()
	if err != nil {
		return failure.Wrap(AndroidManagerFailed, manager, err)
	}
	if err := command.Start(); err != nil {
		return failure.Wrap(AndroidManagerFailed, manager, err)
	}
	said := answers.answer(output, input)
	input.Close() // no more answers: a question left would read the end of its input as no
	if err := command.Wait(); err != nil {
		lines := strings.Split(strings.TrimSpace(said), "\n")
		return failure.Wrap(AndroidManagerFailed, strings.Join(lines[max(len(lines)-8, 0):], "\n"), err)
	}
	return nil
}

// fileExists tells whether a path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// licenceAnswers answers sdkmanager's licence questions: yes to one licence, when the text it
// shows is the text the person accepted; no to the others, which are remembered.
type licenceAnswers struct {
	licence, text string
	refused       []string
}

// licenceQuestion is how sdkmanager asks, after a licence's name ("License …:") and text.
const licenceQuestion = "Accept? (y/N):"

// answer reads what sdkmanager says until it ends, answering each question; it returns what it
// said, for errors.
func (answers *licenceAnswers) answer(output io.Reader, input io.Writer) string {
	var said, waiting strings.Builder
	buffer := make([]byte, 4096)
	for {
		read, err := output.Read(buffer)
		said.Write(buffer[:read])
		waiting.Write(buffer[:read])
		if question := waiting.String(); strings.Contains(question, licenceQuestion) {
			asked, rest, _ := strings.Cut(question, licenceQuestion)
			io.WriteString(input, answers.reply(asked)) // a closed input ends sdkmanager, whose error is read
			waiting.Reset()
			waiting.WriteString(rest)
		}
		if err != nil {
			return said.String()
		}
	}
}

// reply is the answer to one question, from what sdkmanager said before it.
func (answers *licenceAnswers) reply(asked string) string {
	start := strings.LastIndex(asked, "License ")
	if start < 0 {
		answers.refused = append(answers.refused, "a licence with no name")
		return "n\n"
	}
	name, text, _ := strings.Cut(asked[start+len("License "):], ":")
	text = strings.ReplaceAll(text, "---------------------------------------", "")
	if name != answers.licence || strings.Join(strings.Fields(text), " ") != strings.Join(strings.Fields(answers.text), " ") {
		answers.refused = append(answers.refused, name)
		return "n\n"
	}
	return "y\n"
}

// refusal is LicenceNotAccepted when a licence was answered no.
func (answers *licenceAnswers) refusal() error {
	if len(answers.refused) == 0 {
		return nil
	}
	return failure.Wrap(LicenceNotAccepted, strings.Join(answers.refused, ", "), nil)
}
