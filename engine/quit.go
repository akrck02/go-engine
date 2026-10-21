package engine

import (
	"sync/atomic"

	"github.com/akrck02/go-ui/window"
)

// A game ends itself from its menu ("Quit the game"): on a computer the window closes, on a 3DS it
// goes back to the Homebrew Launcher. In the editor's Play only the game played stops, not the
// editor.

// quitAsked is set by Quit, and taken by whoever plays the game.
var quitAsked atomic.Bool

// Quit ends the game once the frame being played ends. Call it from a behaviour or from the game's
// interface; it may be called from any goroutine.
func Quit() { quitAsked.Store(true) }

// TakeQuit tells whether the game asked to quit since the last time, and forgets it: one who plays
// the game elsewhere than Run (the editor's Play) stops it then.
func TakeQuit() bool { return quitAsked.Swap(false) }

// quitIfAsked closes the game's window when the game asked to quit.
func quitIfAsked() {
	if TakeQuit() {
		window.Quit()
	}
}
