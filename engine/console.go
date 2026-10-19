package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// consoleButtons are the console's buttons a game reads as keys: the cross as the arrows, A as
// Space, B as Escape, X and Y as themselves. With the circle pad (Input.Move) they are all a game
// reads on the console: the touch screen and L and R belong to the bottom screen's interface.
var consoleButtons = [...]struct {
	button ctr.Buttons
	key    input.Key
}{
	{ctr.ButtonUp, input.KeyUp}, {ctr.ButtonDown, input.KeyDown}, {ctr.ButtonLeft, input.KeyLeft},
	{ctr.ButtonRight, input.KeyRight}, {ctr.ButtonA, input.KeySpace}, {ctr.ButtonB, input.KeyEscape},
	{ctr.ButtonX, input.KeyX}, {ctr.ButtonY, input.KeyY},
}

// consoleInput is a game's keys on the console, and on a touch screen (Android). go-ui's window
// hands the scene a fresh copy of its input every frame, without what its interface took (the
// arrows and A move between its controls and press them), so a key in it can never be pressed
// once and then held. The game's keys are kept here instead, from frame to frame, set from the
// buttons themselves.
type consoleInput struct {
	keys    input.State
	buttons [4]touchButton // where a touch screen's buttons were last drawn (Android)
}

// keysFrom is the game's keys this frame: the console's buttons held now, and the window's stick.
func (console *consoleInput) keysFrom(window *input.State, held ctr.Buttons) *input.State {
	console.keys.EndFrame()
	for _, mapping := range consoleButtons {
		console.keys.SetKey(mapping.key, held&mapping.button != 0)
	}
	console.keys.StickX, console.keys.StickY = window.StickX, window.StickY
	return &console.keys
}
