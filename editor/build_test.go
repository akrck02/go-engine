package editor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

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
