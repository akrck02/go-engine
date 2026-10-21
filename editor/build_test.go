package editor

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/install"
	"github.com/akrck02/go-engine/project"
)

func TestABuildsResultIsKeptPerTarget(test *testing.T) {
	editor := open(test)
	program := filepath.Join(test.TempDir(), "lights.3dsx")
	os.WriteFile(program, make([]byte, 2048), 0o644)
	editor.builders[project.Nintendo3DS] = func() (project.Built, error) { return project.Built{Program: program}, nil }
	editor.builders[project.Desktop] = func() (project.Built, error) { return project.Built{}, errors.New("no compiler here") }
	editor.recordBuild(project.Nintendo3DS, editor.buildNow(project.Nintendo3DS))
	editor.recordBuild(project.Desktop, editor.buildNow(project.Desktop))
	console, found := editor.LastBuild(project.Nintendo3DS)
	if !found || console.Program != program || console.Bytes != 2048 || console.Err != nil || console.At.IsZero() {
		test.Fatalf("the console's build is %+v", console)
	}
	computer, _ := editor.LastBuild(project.Desktop)
	if computer.Err == nil || len(computer.Log) == 0 {
		test.Fatalf("the computer's failed build is %+v", computer)
	}
}

func TestSavingTheToolsKeepsThemAndLooksAgain(test *testing.T) {
	editor := open(test)
	for _, variable := range []string{"XDG_CONFIG_HOME", "HOME", "APPDATA", "LOCALAPPDATA"} {
		test.Setenv(variable, test.TempDir()) // not the person's tools, nor their ~/go/bin/tinygo
	}
	test.Setenv("TINYGO", "")
	test.Setenv("PATH", test.TempDir())
	editor.Project.Tools = build.Tools{}
	if tinygo := editor.ToolsFound()[0]; tinygo.Name != "TinyGo" || tinygo.Err == nil {
		test.Fatalf("TinyGo is found where there is none: %+v", tinygo)
	}
	if runtime.GOOS == "windows" {
		test.Skip("the fake tinygo is a shell script")
	}
	root := test.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "targets"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "targets", "nintendo3ds.json"), nil, 0o644); err != nil {
		test.Fatal(err)
	}
	program := filepath.Join(test.TempDir(), "tinygo")
	if err := os.WriteFile(program, []byte("#!/bin/sh\necho "+root+"\n"), 0o755); err != nil {
		test.Fatal(err)
	}
	chosen := editor.Project.Tools
	chosen.TinyGo = program
	if err := editor.SaveTools(chosen); err != nil {
		test.Fatal(err)
	}
	if tinygo := editor.ToolsFound()[0]; tinygo.Found != program || tinygo.Err != nil {
		test.Fatalf("TinyGo set in the editor is found as %+v", tinygo)
	}
	if kept, err := build.ReadTools(); err != nil || kept.TinyGo != program {
		test.Fatalf("the tools kept are %+v, %v", kept, err)
	}
}

func TestTheToolsDoNotChangeWhileTheGameBuilds(test *testing.T) {
	editor := open(test)
	for _, variable := range []string{"XDG_CONFIG_HOME", "HOME", "APPDATA", "LOCALAPPDATA"} {
		test.Setenv(variable, test.TempDir())
	}
	editor.running = true
	if err := editor.SaveTools(build.Tools{TinyGo: "/opt/tinygo"}); !errors.Is(err, ToolsInUse) {
		test.Fatalf("the tools are saved while the game builds: %v", err)
	}
}

func TestTypedPlacesAreCheckedBeforeSaving(test *testing.T) {
	editor := open(test)
	editor.LookForTools(build.Tools{TinyGo: "/nowhere/tinygo"})
	if tinygo := editor.ToolsFound()[0]; !errors.Is(tinygo.Err, build.ToolPlaceWrong) {
		test.Fatalf("a TinyGo typed where nothing is is found as %+v", tinygo)
	}
	if editor.Project.Tools.TinyGo == "/nowhere/tinygo" {
		test.Fatal("checking a place typed changed the tools the builds use")
	}
}

func TestBrowsingForAToolTypesThePlaceChosen(test *testing.T) {
	view := NewView(open(test))
	for index, part := range buildToolParts {
		view.browseForTool(index, ToolFound{Name: part.name})
		if view.picking == nil || view.picking.options.Folder != part.folder {
			test.Fatalf("browsing for %s asks for a folder (%v): %+v", part.name, part.folder, view.picking)
		}
		view.picking.chosen("/opt/" + part.name)
		view.picking = nil
		if place := *part.place(&view.build.tools); place != "/opt/"+part.name {
			test.Fatalf("the place chosen for %s is typed as %q", part.name, place)
		}
	}
	if view.Editor.Project.Tools.TinyGo == "/opt/TinyGo" {
		test.Fatal("a place chosen is kept before it is saved")
	}
}

func TestAModelChosenIsCopiedIntoTheProject(test *testing.T) {
	editor := open(test)
	model := filepath.Join(test.TempDir(), "boat.glb")
	if err := os.WriteFile(model, []byte("glTF"), 0o644); err != nil {
		test.Fatal(err)
	}
	editor.ImportModel(model)
	if _, err := os.Stat(editor.Project.Path("assets/models/boat.glb")); err != nil {
		test.Fatalf("the model chosen is not in the project: %v", err)
	}
}

// waitForTheTask applies the background task's events until it ends.
func waitForTheTask(test *testing.T, editor *Editor) {
	test.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for editor.Running() {
		if time.Now().After(deadline) {
			test.Fatal("the task did not end")
		}
		editor.TakeEvents()
		time.Sleep(time.Millisecond)
	}
}

func TestAToolInstalledIsSetAndKept(test *testing.T) {
	editor := open(test)
	for _, variable := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "HOME", "APPDATA", "LOCALAPPDATA"} {
		test.Setenv(variable, test.TempDir())
	}
	editor.Project.Tools = build.Tools{}
	var program string
	editor.installTool("TinyGo", func(installer install.Installer) (string, error) {
		program = filepath.Join(installer.Folder, "tinygo", "bin", "tinygo")
		installer.Progress(50, 100)
		if err := os.MkdirAll(filepath.Dir(program), 0o755); err != nil {
			return "", err
		}
		return program, os.WriteFile(program, nil, 0o755)
	}, func(tools *build.Tools, place string) { tools.TinyGo = place })
	if tool, _, _, busy := editor.Installing(); !busy || tool != "TinyGo" {
		test.Fatal("the editor does not say TinyGo is being installed")
	}
	waitForTheTask(test, editor)
	if _, _, _, busy := editor.Installing(); busy {
		test.Fatal("the install still shows once it ended")
	}
	if kept, err := build.ReadTools(); err != nil || kept.TinyGo != program || editor.ToolsFound()[0].Found != program {
		test.Fatalf("TinyGo installed is kept as %+v (%v) and found as %+v", kept, err, editor.ToolsFound()[0])
	}
}

func TestAndroidsLicenceWaitsToBeAcceptedOrDeclined(test *testing.T) {
	editor := open(test)
	editor.androidLicence = "Use the lighthouse kindly."
	if licence, waiting := editor.AndroidLicence(); !waiting || licence != "Use the lighthouse kindly." {
		test.Fatal("the licence read does not wait to be accepted")
	}
	editor.DeclineAndroidLicence()
	if _, waiting := editor.AndroidLicence(); waiting {
		test.Fatal("a licence declined still waits")
	}
}

func TestEveryToolThatCanBeInstalledIsOne(test *testing.T) {
	names := map[string]bool{}
	for _, part := range buildToolParts {
		names[part.name] = true
	}
	for name := range toolInstalls {
		if !names[name] {
			test.Errorf("%q can be installed, but is no tool", name)
		}
	}
	if len(toolInstalls) != len(buildToolParts) {
		test.Errorf("%d tools, %d can be installed", len(buildToolParts), len(toolInstalls))
	}
}

func TestNothingIsInstalledForAndroidWithoutTheLicenceRead(test *testing.T) {
	editor := open(test)
	editor.AcceptAndroidLicence()
	if _, _, _, busy := editor.Installing(); busy || editor.Running() {
		test.Fatal("Android is installed with no licence read")
	}
}
