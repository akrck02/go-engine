package project

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
)

func TestAnApplicationIDIsMadeFromTheGamesName(test *testing.T) {
	cases := []struct{ name, wanted string }{
		{"Harbour Lights", "io.goengine.harbourlights"},
		{"3 Lamps!", "io.goengine.game3lamps"},
		{"", "io.goengine.game"},
	}
	for _, each := range cases {
		project := &Project{Settings: engine.Settings{Name: each.name}}
		if made := project.ApplicationID(); made != each.wanted {
			test.Errorf("%q gives the id %q, not %q", each.name, made, each.wanted)
		}
	}
	project := &Project{Settings: engine.Settings{Name: "Harbour Lights", AndroidID: "org.example.harbour"}}
	if kept := project.ApplicationID(); kept != "org.example.harbour" {
		test.Errorf("the id the settings keep is not used: %q", kept)
	}
}

func TestAnAndroidBuildAsksForThePackage(test *testing.T) {
	project := &Project{Folder: "/games/harbour", Settings: engine.Settings{Name: "Harbour Lights"}}
	arguments := project.androidArguments()
	joined := strings.Join(arguments, " ")
	for _, wanted := range []string{"tool go-ui-apk", "-name Harbour Lights", "-id io.goengine.harbourlights", "-abi arm64,x86_64", "-o " + project.AndroidPackage()} {
		if !strings.Contains(joined, wanted) {
			test.Errorf("the Android build is asked for with %q, without %q", joined, wanted)
		}
	}
	if !strings.HasSuffix(project.AndroidPackage(), "harbour-lights.apk") || arguments[len(arguments)-1] != "." {
		test.Errorf("the package is %q, built from %q", project.AndroidPackage(), arguments[len(arguments)-1])
	}
	if !slices.Contains(androidEnvironment("/sdk", "/ndk", "/jdk"), "GOFLAGS=-tags=release") {
		test.Error("an Android build carries the editor: it is not built with the release tag")
	}
}

func TestAModuleKnowsItsToolsAndWhatAFolderReplaces(test *testing.T) {
	var module moduleFile
	printed := `{"Replace": [{"Old": {"Path": "` + uiModule + `"}, "New": {"Path": "../go-ui"}},
		{"Old": {"Path": "example.com/other"}, "New": {"Path": "example.com/other", "Version": "v1.0.0"}}],
		"Tool": [{"Path": "` + androidTool + `"}]}`
	if err := json.Unmarshal([]byte(printed), &module); err != nil {
		test.Fatal(err)
	}
	if !module.hasTool(androidTool) || module.hasTool("example.com/tool") {
		test.Error("the module's tools are not told apart")
	}
	if !module.replacedByFolder(uiModule) || module.replacedByFolder("example.com/other") {
		test.Error("a replace by a folder is not told from one by a version")
	}
}

func TestGoUIIsReplacedWithTheVersionGoEngineAsksFor(test *testing.T) {
	graph := "game " + engineModule + "@v0.0.0\n" +
		androidTool + "@v0.0.0-1 " + uiModule + "@v0.0.0\n" +
		engineModule + "@v0.0.0 " + uiModule + "@v0.0.0-20261017191932-0e94481e5103\n"
	if version, found := uiVersionIn(graph); !found || version != "v0.0.0-20261017191932-0e94481e5103" {
		test.Errorf("the go-ui go-engine asks for is %q, %v", version, found)
	}
	if _, found := uiVersionIn("game example.com/other@v1.0.0\n"); found {
		test.Error("a go-ui is found where go-engine asks for none")
	}
}

func TestAnApplicationIDOfTheWrongFormIsNotSaved(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	created.Settings.AndroidID = "my game"
	if err := created.SaveSettings(); !errors.Is(err, ApplicationIDWrong) {
		test.Fatalf("an id with a space is saved: %v", err)
	}
	created.Settings.AndroidID = "org.example.game"
	if err := created.SaveSettings(); err != nil {
		test.Fatalf("a good id is not saved: %v", err)
	}
}

func TestPlayingOnADeviceInstallsStartsAndReadsTheLog(test *testing.T) {
	if runtime.GOOS == "windows" {
		test.Skip("the fake adb is a shell script")
	}
	folder := test.TempDir()
	calls := filepath.Join(folder, "calls")
	adb := filepath.Join(folder, "adb")
	script := "#!/bin/sh\necho \"$@\" >> " + calls + "\n" +
		"if [ \"$1\" = devices ]; then printf 'List of devices attached\\nemulator-5554\\tdevice\\n'; fi\n"
	if err := os.WriteFile(adb, []byte(script), 0o755); err != nil {
		test.Fatal(err)
	}
	project := &Project{Folder: folder, Settings: engine.Settings{Name: "Harbour Lights"}}
	serial, err := androidDevice(folder, adb)
	if err != nil {
		test.Fatal(err)
	}
	logcat, err := project.playOnAndroid(adb, serial, "/games/harbour.apk")
	if err != nil {
		test.Fatal(err)
	}
	said, _ := os.ReadFile(calls)
	wanted := "devices\n" +
		"-s emulator-5554 install -r /games/harbour.apk\n" +
		"-s emulator-5554 shell am force-stop io.goengine.harbourlights\n" +
		"-s emulator-5554 shell monkey -p io.goengine.harbourlights -c android.intent.category.LAUNCHER 1\n"
	if string(said) != wanted {
		test.Errorf("adb was asked:\n%s\nnot:\n%s", said, wanted)
	}
	if arguments := strings.Join(logcat.Args, " "); !strings.Contains(arguments, "-s emulator-5554 shell") || !strings.Contains(arguments, "logcat") {
		test.Errorf("the log is read with %q", arguments)
	}
}

func TestNoDeviceIsSaidByItsCode(test *testing.T) {
	if runtime.GOOS == "windows" {
		test.Skip("the fake adb is a shell script")
	}
	adb := filepath.Join(test.TempDir(), "adb")
	if err := os.WriteFile(adb, []byte("#!/bin/sh\necho 'List of devices attached'\n"), 0o755); err != nil {
		test.Fatal(err)
	}
	if _, err := androidDevice(filepath.Dir(adb), adb); !errors.Is(err, NoAndroidDevice) {
		test.Fatalf("adb listing no device gives %v", err)
	}
}

func TestADeviceIsTheFirstReadyOneAdbLists(test *testing.T) {
	cases := []struct{ listing, wanted string }{
		{"List of devices attached\nR58N\tdevice\nemulator-5554\tdevice\n", "R58N"},
		{"List of devices attached\nR58N\tunauthorized\nemulator-5554\tdevice\n", "emulator-5554"},
		{"List of devices attached\n\n", ""},
	}
	for _, each := range cases {
		if serial, _ := firstDevice(each.listing); serial != each.wanted {
			test.Errorf("%q gives the device %q, not %q", each.listing, serial, each.wanted)
		}
	}
}
