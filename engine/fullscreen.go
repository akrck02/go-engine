package engine

import (
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-renderer/input"
)

// On a computer and in a browser a game plays in a window or on the whole screen: engine.json's
// window says which it starts in on each target, the game changes it with SetFullscreen, and the
// player with the action fullscreen (F11 or Alt+Enter, unless the game binds it). A browser lets a
// page take the whole screen only when it answers a key or a click, so there it waits for the
// next one. Consoles and phones are always on their whole screen.

// The ways a game's window can start (engine.json's window).
const (
	Windowed         = "windowed"
	FullscreenWindow = "fullscreen"
)

// FullscreenAction is the action that puts the game on the whole screen and back.
const FullscreenAction = "fullscreen"

// fullscreenInputs are the fullscreen action's inputs on the targets with windows, when the game
// does not bind it.
var fullscreenInputs = []string{"F11", "Alt+Enter"}

// fullscreen is whether the game asks for the whole screen, and whether that changed since the
// window was told.
var fullscreen struct {
	wanted, changed bool
}

// SetFullscreen asks for the game on the whole screen, or in its window again.
func SetFullscreen(on bool) { fullscreen.wanted, fullscreen.changed = on, true }

// Fullscreen tells whether the game is asked to be on the whole screen.
func Fullscreen() bool { return fullscreen.wanted }

// startsFullscreen tells whether a game starts on the whole screen on a target, as engine.json's
// window says.
func startsFullscreen(settings Settings, target string) (bool, error) {
	switch way := settings.Window[target]; way {
	case "", Windowed:
		return false, nil
	case FullscreenWindow:
		return true, nil
	default:
		return false, failure.Wrap(WindowUnknown, way+" on "+target, nil)
	}
}

// bindFullscreen binds the fullscreen action to F11 and Alt+Enter on a target with windows, when
// the game does not bind it.
func bindFullscreen(actions map[string]map[string][]string, target string) {
	if _, bound := actions[FullscreenAction]; bound || !hasWindows(target) {
		return
	}
	bindOne(FullscreenAction, fullscreenInputs, target) // words the engine knows on these targets
}

// hasWindows tells whether a target shows games in windows.
func hasWindows(target string) bool { return target == DesktopTarget || target == WebTarget }

// toggleFullscreenIfAsked puts the game on the whole screen, or back, when the player presses the
// fullscreen action.
func toggleFullscreenIfAsked() {
	if actionPressed(FullscreenAction) {
		SetFullscreen(!fullscreen.wanted)
	}
}

// followFullscreen tells the window what the game asks for, when that changed; otherwise the game
// follows the window, which the player may have taken off the whole screen (a browser's Escape).
func followFullscreen(state *input.State) {
	if fullscreen.changed {
		state.SetFullscreen(fullscreen.wanted)
		fullscreen.changed = false
		return
	}
	fullscreen.wanted = state.Fullscreen()
}
