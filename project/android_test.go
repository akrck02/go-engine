package project

import (
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
	if !slices.Contains(androidEnvironment("/sdk", "/ndk"), "GOFLAGS=-tags=release") {
		test.Error("an Android build carries the editor: it is not built with the release tag")
	}
}

func TestAModuleKnowsWhetherItHasTheAndroidTool(test *testing.T) {
	without := "module game\n\ngo 1.25\n"
	with := without + "\ntool " + androidTool + "\n\nreplace " + uiModule + " => " + uiModule + " v0.0.0-1\n"
	if hasLine(without, "tool "+androidTool) || !hasLine(with, "tool "+androidTool) {
		test.Error("the tool line is not found where it is, or found where it is not")
	}
	if hasLine(without, "replace "+uiModule+" ") || !hasLine(with, "replace "+uiModule+" ") {
		test.Error("go-ui's replace is not found where it is, or found where it is not")
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
	logcat, err := project.playOnAndroid(adb, "/games/harbour.apk")
	if err != nil {
		test.Fatal(err)
	}
	said, _ := os.ReadFile(calls)
	for _, wanted := range []string{"-s emulator-5554 install -r /games/harbour.apk", "-s emulator-5554 shell monkey -p io.goengine.harbourlights"} {
		if !strings.Contains(string(said), wanted) {
			test.Errorf("adb was not asked %q; it was asked:\n%s", wanted, said)
		}
	}
	if arguments := strings.Join(logcat.Args, " "); !strings.Contains(arguments, "-s emulator-5554 logcat") {
		test.Errorf("the log is read with %q", arguments)
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
