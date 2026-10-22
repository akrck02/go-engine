package engine

import (
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/nintendo3ds"
)

// consoleButtons are the console's buttons a game reads as keys: the cross as the arrows, A as
// Space, B as Escape, X and Y as themselves, and the shoulders and START as the keys that stand for
// them on a keyboard (ButtonLeftShoulder and the rest). The sticks, the gyroscope and the touch screen are
// read whole (Input.Stick and the rest, controls.go).
var consoleButtons = [...]struct {
	button nintendo3ds.Buttons
	key    input.Key
}{
	{nintendo3ds.ButtonUp, input.KeyUp}, {nintendo3ds.ButtonDown, input.KeyDown}, {nintendo3ds.ButtonLeft, input.KeyLeft},
	{nintendo3ds.ButtonRight, input.KeyRight}, {nintendo3ds.ButtonA, input.KeySpace}, {nintendo3ds.ButtonB, input.KeyEscape},
	{nintendo3ds.ButtonX, input.KeyX}, {nintendo3ds.ButtonY, input.KeyY},
	{nintendo3ds.ButtonL, ButtonLeftShoulder}, {nintendo3ds.ButtonR, ButtonRightShoulder}, {nintendo3ds.ButtonZL, ButtonLeftTrigger}, {nintendo3ds.ButtonZR, ButtonRightTrigger},
	{nintendo3ds.ButtonStart, ButtonStart},
}

// consoleInput is a game's keys on the console, and on a touch screen (Android). go-ui's window
// hands the scene a fresh copy of its input every frame, without what its interface took (the
// arrows and A move between its controls and press them), so a key in it can never be pressed
// once and then held. The game's keys are kept here instead, from frame to frame, set from the
// buttons themselves.
type consoleInput struct {
	reading nintendo3ds.Reading // the console's controls this frame (readConsole)
	keys    input.State
	buttons [4]touchButton // where a touch screen's buttons were last drawn (Android)
	// readerFrames is, by button, the frames a screen reader's press still holds it for
	readerFrames [4]int
}

// keysFrom is the game's keys this frame: the console's buttons held now, and the window's stick.
func (console *consoleInput) keysFrom(window *input.State, held nintendo3ds.Buttons) *input.State {
	console.keys.Feed().EndFrame()
	for _, mapping := range consoleButtons {
		console.keys.Feed().SetKey(mapping.key, held&mapping.button != 0)
	}
	console.keys.Feed().SetStick(window.Stick())
	console.keys.Feed().SetRightStick(window.RightStick())
	return &console.keys
}
