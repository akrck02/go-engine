package editor

import (
	"os"
	"strings"
	"time"

	"github.com/akrck02/go-engine/project"
)

// The Build workspace's work, without its interface: building the game for each place it runs,
// in the background, and keeping each target's last result (what it made, how large, when, and
// what building said) until the next.

// BuildTarget is a place the game runs.
type BuildTarget int

const (
	// ComputerTarget is this computer.
	ComputerTarget BuildTarget = iota
	// ConsoleTarget is the Nintendo 3DS.
	ConsoleTarget
	// AndroidTarget is a phone or tablet with Android.
	AndroidTarget
)

// BuildTargets are the targets in the order of their cards.
var BuildTargets = []BuildTarget{ComputerTarget, ConsoleTarget, AndroidTarget}

// targetParts are each target's name, icon and what its build makes, in the order of BuildTargets.
var targetParts = []struct{ name, icon, makes string }{
	{"This computer", iconComputer, "a program for this computer"},
	{"Nintendo 3DS", iconConsole, "a .3dsx for the Homebrew Launcher"},
	{"Android", iconPhone, "an .apk to install on a phone, signed for trying"},
}

// Name is the target's name on its card.
func (target BuildTarget) Name() string { return targetParts[target].name }

// Icon is the target's icon on its card.
func (target BuildTarget) Icon() string { return targetParts[target].icon }

// Makes says what a build for the target makes.
func (target BuildTarget) Makes() string { return targetParts[target].makes }

// BuildResult is what the last build for a target made, or why it did not.
type BuildResult struct {
	Program string
	Bytes   int64
	At      time.Time
	Err     error
	Log     []string
}

// LastBuild is a target's last build in this session.
func (editor *Editor) LastBuild(target BuildTarget) (BuildResult, bool) {
	result, found := editor.builds[target]
	return result, found
}

// BuildFor builds the game for a target in the background; its result is kept when it ends.
// What is built is what is on disk: the scene and the file open in Art are saved first.
func (editor *Editor) BuildFor(target BuildTarget) {
	if editor.running {
		editor.say("Busy: wait for what is running to end, or stop the game", true)
		return
	}
	if !editor.Saved && editor.Save() != nil {
		return
	}
	if err := editor.saveArtIfChanged(); err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.Run("Building for "+target.Name()+"…", func() (string, error) {
		result := editor.buildNow(target)
		editor.send(func(editor *Editor) { editor.recordBuild(target, result) })
		if result.Err != nil {
			return "", result.Err
		}
		return "Built " + result.Program, nil
	})
}

// buildNow builds the game for a target and says how it went.
func (editor *Editor) buildNow(target BuildTarget) BuildResult {
	built, err := editor.builderFor(target)()
	result := BuildResult{Program: built.Program, At: time.Now(), Err: err}
	if err != nil {
		result.Log = strings.Split(strings.TrimSpace(err.Error()), "\n")
		return result
	}
	if file, statErr := os.Stat(built.Program); statErr == nil {
		result.Bytes = file.Size()
	}
	result.Log = []string{"Built " + built.Program, "Notices of what it carries: " + built.Notices}
	return result
}

// builderFor is how a target is built: the project's own builds, unless a test set another.
func (editor *Editor) builderFor(target BuildTarget) func() (project.Built, error) {
	if builder := editor.builders[target]; builder != nil {
		return builder
	}
	switch target {
	case ConsoleTarget:
		return editor.Project.Build3DS
	case AndroidTarget:
		return editor.Project.BuildAndroid
	}
	return editor.Project.BuildDesktop
}

// recordBuild keeps a target's last build.
func (editor *Editor) recordBuild(target BuildTarget, result BuildResult) {
	editor.builds[target] = result
}

// WriteReleaseWorkflow writes the workflow that builds the game for Windows, macOS and Linux on
// GitHub at each version tag, and says where it went or why it did not.
func (editor *Editor) WriteReleaseWorkflow() {
	workflow, err := editor.Project.WriteReleaseWorkflow()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.say("Wrote "+workflow+" and THIRD_PARTY_NOTICES.md: commit both, push a tag such as v1.0.0, and GitHub attaches the programs to its release", false)
}
