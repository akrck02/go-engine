package project

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/notices"
)

// A game is built for Android with go-ui-apk, go-ui-android's tool, run as a tool of the game's
// own module (go tool go-ui-apk) so it builds with the game's versions of go-ui and go-renderer.

// The tool, the version of it go-engine builds with, and go-ui, which go-ui-android asks for at
// v0.0.0 (it builds against a local copy): the game replaces it with the version go-engine uses.
const (
	androidTool        = "github.com/akrck02/go-ui-android/cmd/go-ui-apk"
	androidToolVersion = "v0.0.0-20261015201303-015f7fb03632"
	uiModule           = "github.com/akrck02/go-ui"
)

// androidProcessors are the processors the package carries a library for: phones' and the
// emulator's on a PC.
const androidProcessors = "arm64,x86_64"

// applicationIDPrefix starts an application id made from the game's name.
const applicationIDPrefix = "io.goengine."

// ApplicationID is the game's name on Android, as io.example.game: the settings' AndroidID, or one
// made from the game's name.
func (project *Project) ApplicationID() string {
	if project.Settings.AndroidID != "" {
		return project.Settings.AndroidID
	}
	var word strings.Builder
	for _, letter := range strings.ToLower(project.Settings.Name) {
		if letter >= 'a' && letter <= 'z' || letter >= '0' && letter <= '9' {
			word.WriteRune(letter)
		}
	}
	name := word.String()
	if name == "" || name[0] < 'a' {
		name = "game" + name // each part of an id starts with a letter
	}
	return applicationIDPrefix + name
}

// AndroidPackage is where the Android package goes.
func (project *Project) AndroidPackage() string {
	return project.Path(path.Join("build", build.FileName(project.Settings.Name)+".apk"))
}

// BuildAndroid builds the game for Android into build/, a package signed for trying (not for a
// store), without the editor.
func (project *Project) BuildAndroid() (Built, error) {
	sdk, err := build.AndroidSDK()
	if err != nil {
		return Built{}, err
	}
	ndk, err := build.AndroidNDK(sdk)
	if err != nil {
		return Built{}, err
	}
	if err := project.Tidy(); err != nil {
		return Built{}, err
	}
	if err := project.addAndroidTool(); err != nil {
		return Built{}, err
	}
	if err := os.MkdirAll(project.Path("build"), 0o755); err != nil {
		return Built{}, err // go-ui-apk writes the package but makes no folder
	}
	if err := build.Run(project.Folder, androidEnvironment(sdk, ndk), build.GoTool(), project.androidArguments()...); err != nil {
		return Built{}, err
	}
	noticesFile, err := writeNotices(project.Path("build"), notices.Android)
	return Built{Program: project.AndroidPackage(), Notices: noticesFile}, err
}

// androidArguments are the go command's to build the package.
func (project *Project) androidArguments() []string {
	return []string{"tool", "go-ui-apk", "-name", project.Settings.Name, "-id", project.ApplicationID(),
		"-abi", androidProcessors, "-o", project.AndroidPackage(), "."}
}

// androidEnvironment is what go-ui-apk reads: where the SDK and the NDK are, go on the path for the
// builds it starts, and the release tag that leaves the editor out.
func androidEnvironment(sdk, ndk string) []string {
	return []string{"ANDROID_HOME=" + sdk, "ANDROID_NDK_HOME=" + ndk, "GOFLAGS=-tags=release", build.GoOnPath()}
}

// addAndroidTool adds go-ui-apk to the game's module, and go-ui's replace it needs, when missing.
func (project *Project) addAndroidTool() error {
	data, err := os.ReadFile(project.Path("go.mod"))
	if err != nil {
		return err
	}
	if !hasLine(string(data), "replace "+uiModule+" ") {
		version, err := build.Output(project.Folder, nil, build.GoTool(), "list", "-m", "-f", "{{.Version}}", uiModule)
		if err != nil {
			return err
		}
		replace := "-replace=" + uiModule + "=" + uiModule + "@" + strings.TrimSpace(version)
		if err := build.Run(project.Folder, nil, build.GoTool(), "mod", "edit", replace); err != nil {
			return err
		}
	}
	if hasLine(string(data), "tool "+androidTool) {
		return nil
	}
	return build.Run(project.Folder, nil, build.GoTool(), "get", "-tool", androidTool+"@"+androidToolVersion)
}

// hasLine tells whether a go.mod has a line that starts so.
func hasLine(module, start string) bool {
	for line := range strings.Lines(module) {
		if strings.HasPrefix(strings.TrimSpace(line), start) {
			return true
		}
	}
	return false
}

// PlayOnAndroid builds the game for Android, installs it on the first device adb sees (a phone
// plugged in, or the emulator) and starts it; it returns the command that reads its log, not
// started: the editor shows what it says, and stopping it is StopOnAndroid.
func (project *Project) PlayOnAndroid() (*exec.Cmd, error) {
	built, err := project.BuildAndroid()
	if err != nil {
		return nil, err
	}
	adb, err := build.ADBTool()
	if err != nil {
		return nil, err
	}
	return project.playOnAndroid(adb, built.Program)
}

// playOnAndroid installs a package on the first device and starts it, with the log read from
// then on.
func (project *Project) playOnAndroid(adb, androidPackage string) (*exec.Cmd, error) {
	serial, err := androidDevice(project.Folder, adb)
	if err != nil {
		return nil, err
	}
	steps := [][]string{
		{"install", "-r", androidPackage},
		{"logcat", "-c"}, // what the game said before is not this game's
		{"shell", "am", "force-stop", project.ApplicationID()},
		{"shell", "monkey", "-p", project.ApplicationID(), "-c", "android.intent.category.LAUNCHER", "1"},
	}
	for _, step := range steps {
		if err := build.Run(project.Folder, nil, adb, append([]string{"-s", serial}, step...)...); err != nil {
			return nil, err
		}
	}
	return exec.Command(adb, "-s", serial, "logcat", "-v", "brief", "-s", androidLogTag), nil
}

// androidLogTag is the tag go-renderer's Android window writes the program's log under.
const androidLogTag = "Go"

// StopOnAndroid stops the game on the first device adb sees.
func (project *Project) StopOnAndroid() error {
	adb, err := build.ADBTool()
	if err != nil {
		return err
	}
	serial, err := androidDevice(project.Folder, adb)
	if err != nil {
		return err
	}
	return build.Run(project.Folder, nil, adb, "-s", serial, "shell", "am", "force-stop", project.ApplicationID())
}

// androidDevice is the first device adb sees, ready to be played on.
func androidDevice(folder, adb string) (string, error) {
	listing, err := build.Output(folder, nil, adb, "devices")
	if err != nil {
		return "", err
	}
	serial, found := firstDevice(listing)
	if !found {
		return "", NoAndroidDevice
	}
	return serial, nil
}

// firstDevice is the first ready device in what adb devices prints: a line of a serial and
// "device" (not "unauthorized" or "offline").
func firstDevice(listing string) (string, bool) {
	for line := range strings.Lines(listing) {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == "device" {
			return fields[0], true
		}
	}
	return "", false
}
