package editor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/project"
)

func TestABuildsResultIsKeptPerTarget(test *testing.T) {
	editor := open(test)
	program := filepath.Join(test.TempDir(), "lights.3dsx")
	os.WriteFile(program, make([]byte, 2048), 0o644)
	editor.builders[ConsoleTarget] = func() (project.Built, error) { return project.Built{Program: program}, nil }
	editor.builders[ComputerTarget] = func() (project.Built, error) { return project.Built{}, errors.New("no compiler here") }
	editor.recordBuild(ConsoleTarget, editor.buildNow(ConsoleTarget))
	editor.recordBuild(ComputerTarget, editor.buildNow(ComputerTarget))
	console, found := editor.LastBuild(ConsoleTarget)
	if !found || console.Program != program || console.Bytes != 2048 || console.Err != nil || console.At.IsZero() {
		test.Fatalf("the console's build is %+v", console)
	}
	computer, _ := editor.LastBuild(ComputerTarget)
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
	program := filepath.Join(test.TempDir(), "tinygo")
	if err := os.WriteFile(program, nil, 0o755); err != nil {
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
