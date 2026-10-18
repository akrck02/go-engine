package install

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
// needs, once the person accepted the licence AndroidLicence gives; it returns the SDK's folder
// (the NDK is in it).
func (installer Installer) AndroidSDK() (string, error) {
	java, err := findJava()
	if err != nil {
		return "", err
	}
	sdk := filepath.Join(installer.Folder, "android-sdk")
	manager, err := installer.androidCommandLine(sdk)
	if err != nil {
		return "", err
	}
	environment := append(os.Environ(), "JAVA_HOME="+filepath.Dir(filepath.Dir(java)))
	answers := strings.Repeat("y\n", len(AndroidPackages)+1) // the licence was accepted in the editor
	if err := runAndroidManager(manager, environment, answers, "--sdk_root="+sdk, "--licenses"); err != nil {
		return "", err
	}
	if err := runAndroidManager(manager, environment, "", append([]string{"--sdk_root=" + sdk}, AndroidPackages...)...); err != nil {
		return "", err
	}
	return sdk, nil
}

// androidRepository downloads and reads Google's repository index.
func (installer Installer) androidRepository() (androidRepository, error) {
	var repository androidRepository
	index := filepath.Join(installer.Folder, "android-sdk", androidIndex)
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

// runAndroidManager runs sdkmanager, giving it answers to type.
func runAndroidManager(manager string, environment []string, answers string, arguments ...string) error {
	command := exec.Command(manager, arguments...)
	command.Env = environment
	command.Stdin = strings.NewReader(answers)
	output, err := command.CombinedOutput()
	if err != nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		return failure.Wrap(AndroidManagerFailed, strings.Join(lines[max(len(lines)-8, 0):], "\n"), err)
	}
	return nil
}

// findJava finds the java program: $JAVA_HOME's, or on the path.
func findJava() (string, error) {
	name := "java"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if home := os.Getenv("JAVA_HOME"); home != "" {
		if java := filepath.Join(home, "bin", name); fileExists(java) {
			return java, nil
		}
	}
	if java, err := exec.LookPath("java"); err == nil {
		return resolved(java), nil
	}
	return "", JavaNotFound
}

// resolved is a program's place with its links followed: java on the path is often a link to the
// JDK's own, whose folder is JAVA_HOME.
func resolved(program string) string {
	if real, err := filepath.EvalSymlinks(program); err == nil {
		return real
	}
	return program
}

// fileExists tells whether a path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
