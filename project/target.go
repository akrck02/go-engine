package project

import (
	"os/exec"

	"github.com/akrck02/go-engine/failure"
)

// The places a game runs, and what a project does for each with the same four verbs: build the
// game for it, play it there, play it in an emulator of it, and stop it. Each target does them its
// own way inside; one that cannot do something says so by its code.

// Target is a place a game runs.
type Target int

const (
	// Desktop is this computer.
	Desktop Target = iota
	// Nintendo3DS is the Nintendo 3DS, through the Homebrew Launcher.
	Nintendo3DS
	// Android is a phone or a tablet with Android.
	Android
)

// Targets are every target, in the order the editor shows them.
var Targets = []Target{Desktop, Nintendo3DS, Android}

// targetNames are each target's name, for errors and logs, in the order of Targets.
var targetNames = []string{"this computer", "the Nintendo 3DS", "Android"}

// String is the target's name, or "an unknown target" for a number that is none.
func (target Target) String() string {
	if target < 0 || int(target) >= len(targetNames) {
		return "an unknown target"
	}
	return targetNames[target]
}

// Build builds the game for a target into build/: a program, a .3dsx, an .apk.
func (project *Project) Build(target Target) (Built, error) {
	switch target {
	case Nintendo3DS:
		return project.build3DS()
	case Android:
		return project.buildAndroid()
	}
	return project.buildDesktop()
}

// PlayOn builds the game for a target and plays it there: in a window on this computer, sent to
// a 3DS's Homebrew Launcher, installed and started on the Android device adb sees. It returns the
// command that shows the game playing (the window, the device's log), not started, which ends
// when the game does; nil where nothing is left to watch (the 3DS, once sent).
func (project *Project) PlayOn(target Target) (*exec.Cmd, error) {
	switch target {
	case Nintendo3DS:
		return nil, project.sendTo3DS()
	case Android:
		return project.playOnAndroid()
	}
	return project.playCommand()
}

// PlayInEmulator builds the game for a target and returns the command that opens it in an
// emulator of the target, not started. Only the 3DS has one of its own: Android's emulator is a
// device to adb, and so PlayOn's.
func (project *Project) PlayInEmulator(target Target) (*exec.Cmd, error) {
	if target != Nintendo3DS {
		return nil, failure.Wrap(TargetCannot, target.String()+" has no emulator of its own", nil)
	}
	return project.emulatorCommand()
}

// StopOn stops the game where PlayOn played it, when stopping its command is not enough: on an
// Android device, where the game goes on after its log is closed. Elsewhere it does nothing.
func (project *Project) StopOn(target Target) error {
	if target != Android {
		return nil
	}
	return project.stopOnAndroid()
}
