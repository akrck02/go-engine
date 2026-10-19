package project

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/notices"
)

// A game is built for Android with go-ui-apk, go-ui-android's tool, run as a tool of the game's
// own module (go tool go-ui-apk) so it builds with the game's versions of go-ui and go-renderer.

// The tool, the version of it go-engine builds with, go-engine and go-ui.
const (
	androidTool        = "github.com/akrck02/go-ui-android/cmd/go-ui-apk"
	androidToolVersion = "v0.0.0-20261015201303-015f7fb03632"
	engineModule       = "github.com/akrck02/go-engine"
	uiModule           = "github.com/akrck02/go-ui"
	// uiAskedVersion is the go-ui go-ui-android asks for: v0.0.0, as it builds against a local
	// copy. It is above every version of go-ui yet (they are all pre-releases of v0.0.0), so the
	// game replaces it with the go-ui its go-engine asks for (keepUIReplace).
	uiAskedVersion = "v0.0.0"
)

// androidProcessors are the processors the package carries a library for: phones' and the
// emulator's on a PC.
const androidProcessors = "arm64,x86_64"

// applicationIDPrefix starts an application id made from the game's name.
const applicationIDPrefix = "io.goengine."

// applicationIDForm is an application id's form: two parts or more joined by dots, each starting
// with a letter.
var applicationIDForm = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)

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
	sdk, err := project.Tools.FindAndroidSDK()
	if err != nil {
		return Built{}, err
	}
	ndk, err := project.Tools.FindAndroidNDK(sdk)
	if err != nil {
		return Built{}, err
	}
	java, err := project.Tools.FindJava()
	if err != nil {
		return Built{}, err
	}
	if err := project.addAndroidTool(); err != nil {
		return Built{}, err
	}
	if err := project.Tidy(); err != nil {
		return Built{}, err
	}
	if err := os.MkdirAll(project.Path("build"), 0o755); err != nil {
		return Built{}, err // go-ui-apk writes the package but makes no folder
	}
	if err := build.Run(project.Folder, androidEnvironment(sdk, ndk, java), build.GoTool(), project.androidArguments()...); err != nil {
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

// androidEnvironment is what go-ui-apk reads: where the SDK, the NDK and the JDK are, go on the
// path for the builds it starts, and the release tag that leaves the editor out, after the
// person's own go flags. (A local Go Engine's commit, EngineFlags, cannot go in GOFLAGS: it holds a space.)
func androidEnvironment(sdk, ndk, java string) []string {
	flags := strings.TrimSpace(os.Getenv("GOFLAGS") + " -tags=release")
	return []string{"ANDROID_HOME=" + sdk, "ANDROID_NDK_HOME=" + ndk, "JAVA_HOME=" + java, "GOFLAGS=" + flags, build.GoOnPath()}
}

// addAndroidTool adds go-ui-apk to the game's module when missing, with go-ui's replace it needs.
func (project *Project) addAndroidTool() error {
	module, err := project.readModule()
	if err != nil {
		return err
	}
	if err := project.keepUIReplace(module); err != nil {
		return err
	}
	if module.hasTool(androidTool) {
		return nil
	}
	return build.Run(project.Folder, nil, build.GoTool(), "get", "-tool", androidTool+"@"+androidToolVersion)
}

// keepUIReplace replaces the go-ui go-ui-android asks for with the one the game's go-engine asks
// for, written again on every build: a go-engine of another version asks for another go-ui. A
// go-ui the game replaces with a folder of its own is kept.
func (project *Project) keepUIReplace(module moduleFile) error {
	if module.replacedByFolder(uiModule) {
		return nil
	}
	graph, err := build.Output(project.Folder, nil, build.GoTool(), "mod", "graph")
	if err != nil {
		return err
	}
	version, found := uiVersionIn(graph)
	if !found {
		return nil // the game asks for no go-engine that asks for go-ui: nothing to keep
	}
	return build.Run(project.Folder, nil, build.GoTool(), "mod", "edit",
		"-dropreplace="+uiModule, "-dropreplace="+uiModule+"@"+uiAskedVersion,
		"-replace="+uiModule+"@"+uiAskedVersion+"="+uiModule+"@"+version)
}

// uiVersionIn is the go-ui go-engine asks for, in what go mod graph prints.
func uiVersionIn(graph string) (string, bool) {
	for line := range strings.Lines(graph) {
		asker, asked, found := strings.Cut(strings.TrimSpace(line), " ")
		version, isUI := strings.CutPrefix(asked, uiModule+"@")
		if found && isUI && strings.HasPrefix(asker, engineModule+"@") && version != uiAskedVersion {
			return version, true
		}
	}
	return "", false
}

// moduleFile is what go mod edit -json says of a go.mod.
type moduleFile struct {
	Replace []struct {
		Old, New struct{ Path, Version string }
	}
	Tool []struct{ Path string }
}

// readModule reads the game's go.mod.
func (project *Project) readModule() (moduleFile, error) {
	var module moduleFile
	printed, err := build.Output(project.Folder, nil, build.GoTool(), "mod", "edit", "-json")
	if err != nil {
		return module, err
	}
	return module, json.Unmarshal([]byte(printed), &module)
}

// hasTool tells whether the module has a tool.
func (module moduleFile) hasTool(tool string) bool {
	for _, each := range module.Tool {
		if each.Path == tool {
			return true
		}
	}
	return false
}

// replacedByFolder tells whether the module replaces another with a folder.
func (module moduleFile) replacedByFolder(replaced string) bool {
	for _, each := range module.Replace {
		if each.Old.Path == replaced && each.New.Version == "" {
			return true
		}
	}
	return false
}

// PlayOnAndroid builds the game for Android, installs it on the first device adb sees (a phone
// plugged in, or the emulator) and starts it. It returns the command that shows its log, not
// started, which ends when the game does: the editor shows what it says. StopOnAndroid stops it.
func (project *Project) PlayOnAndroid() (*exec.Cmd, error) {
	adb, err := project.Tools.FindADB()
	if err != nil {
		return nil, err
	}
	serial, err := androidDevice(project.Folder, adb)
	if err != nil {
		return nil, err // before building: a build takes long, and finding no device does not
	}
	built, err := project.BuildAndroid()
	if err != nil {
		return nil, err
	}
	return project.playOnAndroid(adb, serial, built.Program)
}

// playOnAndroid installs a package on a device and starts it.
func (project *Project) playOnAndroid(adb, serial, androidPackage string) (*exec.Cmd, error) {
	identifier := project.ApplicationID()
	steps := [][]string{
		{"install", "-r", androidPackage},
		{"shell", "am", "force-stop", identifier},
		{"shell", "monkey", "-p", identifier, "-c", "android.intent.category.LAUNCHER", "1"},
	}
	for _, step := range steps {
		if err := build.Run(project.Folder, nil, adb, append([]string{"-s", serial}, step...)...); err != nil {
			return nil, err
		}
	}
	return exec.Command(adb, "-s", serial, "shell", androidLogScript(identifier)), nil
}

// androidLogScript is a shell script, run on the device, that shows the game's log (all of it:
// a crash is logged under the system's tags, not the game's) while the game runs, and ends when
// the game does.
func androidLogScript(identifier string) string {
	return "for wait in 1 2 3 4 5 6 7 8 9 10; do process=$(pidof " + identifier + ") && break; sleep 0.5; done; " +
		"[ -n \"$process\" ] || exit 1; " +
		"logcat -v brief --pid=$process & log=$!; " +
		"while [ -d /proc/$process ]; do sleep 1; done; kill $log"
}

// StopOnAndroid stops the game on the first device adb sees.
func (project *Project) StopOnAndroid() error {
	adb, err := project.Tools.FindADB()
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
