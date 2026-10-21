package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// The controls a game reads whole rather than as keys: the sticks, the gyroscope and the touch
// screen of a 3DS, and the mouse's movement on a computer. The 3DS's shoulders and START are keys
// too, the ones that stand for them on a keyboard.

// The keys the 3DS's shoulder buttons and START are read as, and press on a keyboard.
const (
	ButtonL     = input.KeyQ
	ButtonR     = input.KeyE
	ButtonZL    = input.KeyZ
	ButtonZR    = input.KeyC
	ButtonStart = input.KeyEnter
)

// Stick is the circle pad (or a game pad's left stick) from -1 to 1 on each axis, up positive.
func (keys *Input) Stick() (x, y float64) {
	if keys.state == nil {
		return 0, 0
	}
	return keys.state.StickX, -keys.state.StickY
}

// CStick is a New 3DS's C-stick from -1 to 1 on each axis, up positive; 0 elsewhere.
func (keys *Input) CStick() (x, y float64) { return keys.console.CStick() }

// Gyroscope is how fast the 3DS turns, in degrees a second; zero elsewhere, or in an emulator.
func (keys *Input) Gyroscope() ctr.Gyroscope { return keys.console.Gyroscope }

// Touch is where a finger or the stylus is on the 3DS's bottom screen (320 × 240 pixels), and
// whether one is on it at all; on other systems, never.
func (keys *Input) Touch() (x, y int, touching bool) {
	console := keys.console
	return console.TouchX, console.TouchY, console.Held&ctr.ButtonTouch != 0
}

// MouseMovement is how far the mouse moved since the last frame, in window pixels: it keeps
// counting while the pointer is captured (SetPointerCaptured).
func (keys *Input) MouseMovement() (x, y float64) {
	if keys.state == nil {
		return 0, 0
	}
	return float64(keys.state.CursorDeltaX), float64(keys.state.CursorDeltaY)
}

// pointerCaptured holds and hides the desktop's pointer while the game plays (SetPointerCaptured).
var pointerCaptured bool

// SetPointerCaptured holds and hides the pointer, for a game that turns with the mouse, or lets
// it go (a pause menu): MouseMovement still reads the mouse.
func SetPointerCaptured(captured bool) { pointerCaptured = captured }
