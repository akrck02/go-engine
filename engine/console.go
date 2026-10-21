package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// consoleButtons are the console's buttons a game reads as keys: the cross as the arrows, A as
// Space, B as Escape, X and Y as themselves, and the shoulders and START as the keys that stand for
// them on a keyboard (ButtonLeftShoulder and the rest). The sticks, the gyroscope and the touch screen are
// read whole (Input.Stick and the rest, controls.go).
var consoleButtons = [...]struct {
	button ctr.Buttons
	key    input.Key
}{
	{ctr.ButtonUp, input.KeyUp}, {ctr.ButtonDown, input.KeyDown}, {ctr.ButtonLeft, input.KeyLeft},
	{ctr.ButtonRight, input.KeyRight}, {ctr.ButtonA, input.KeySpace}, {ctr.ButtonB, input.KeyEscape},
	{ctr.ButtonX, input.KeyX}, {ctr.ButtonY, input.KeyY},
	{ctr.ButtonL, ButtonLeftShoulder}, {ctr.ButtonR, ButtonRightShoulder}, {ctr.ButtonZL, ButtonLeftTrigger}, {ctr.ButtonZR, ButtonRightTrigger},
	{ctr.ButtonStart, ButtonStart},
}

// consoleInput is a game's keys on the console, and on a touch screen (Android). go-ui's window
// hands the scene a fresh copy of its input every frame, without what its interface took (the
// arrows and A move between its controls and press them), so a key in it can never be pressed
// once and then held. The game's keys are kept here instead, from frame to frame, set from the
// buttons themselves.
type consoleInput struct {
	reading ctr.Reading // the console's controls this frame (readConsole)
	keys    input.State
	buttons [4]touchButton // where a touch screen's buttons were last drawn (Android)
	// readerFrames is, by button, the frames a screen reader's press still holds it for
	readerFrames [4]int
}

// keysFrom is the game's keys this frame: the console's buttons held now, and the window's stick.
func (console *consoleInput) keysFrom(window *input.State, held ctr.Buttons) *input.State {
	console.keys.EndFrame()
	for _, mapping := range consoleButtons {
		console.keys.SetKey(mapping.key, held&mapping.button != 0)
	}
	console.keys.StickX, console.keys.StickY = window.StickX, window.StickY
	console.keys.RightStickX, console.keys.RightStickY = window.RightStickX, window.RightStickY
	return &console.keys
}
