package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// The controls a game reads whole rather than as keys: two sticks, how fast the device turns, a
// second screen's touches and the mouse's movement. Each system gives what it has (a 3DS all but
// the mouse, a computer the mouse and a game pad's sticks) and the rest reads as still. Shoulder
// buttons, triggers and START are keys, the ones that stand for them on a keyboard.

// The keys shoulder buttons, triggers and START are read as (a 3DS's L, R, ZL, ZR and START), and
// that press them on a keyboard.
const (
	ButtonLeftShoulder  = input.KeyQ
	ButtonRightShoulder = input.KeyE
	ButtonLeftTrigger   = input.KeyZ
	ButtonRightTrigger  = input.KeyC
	ButtonStart         = input.KeyEnter
)

// Stick is the main stick (a 3DS's circle pad, a game pad's left stick) from -1 to 1 on each
// axis, up positive.
func (keys *Input) Stick() (x, y float64) {
	if keys.state == nil {
		return 0, 0
	}
	return keys.state.StickX, keys.state.StickY
}

// RightStick is the second stick (a New 3DS's C-stick) from -1 to 1 on each axis, up positive; 0
// without one.
func (keys *Input) RightStick() (x, y float64) { return keys.console.CStick() }

// Turn is how fast a device turns, in degrees a second: about the axis across its screen (Pitch),
// up through it (Yaw) and out of it (Roll).
type Turn struct{ Roll, Pitch, Yaw float64 }

// Gyroscope is how fast the device turns (a 3DS's gyroscope); still without one, or in an
// emulator.
func (keys *Input) Gyroscope() Turn {
	turn := keys.console.Gyroscope
	return Turn{Roll: turn.Roll, Pitch: turn.Pitch, Yaw: turn.Yaw}
}

// SecondScreenTouch is where a finger or a stylus is on the second screen (a 3DS's bottom one,
// 320 × 240 pixels), and whether one is on it at all; never without a second screen.
func (keys *Input) SecondScreenTouch() (x, y int, touching bool) {
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
