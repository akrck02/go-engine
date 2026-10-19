package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// windowCopy is what go-ui's window hands the scene each frame on the console: a fresh copy of
// its input, without the arrows and A its interface took.
func windowCopy() *input.State {
	copied := input.NewState()
	copied.StickX = 0.5
	return copied
}

func TestAConsoleButtonIsPressedOnceWhileItIsHeld(test *testing.T) {
	var console consoleInput
	keys := console.keysFrom(windowCopy(), ctr.ButtonUp|ctr.ButtonA)
	if !keys.Pressed(input.KeyUp) || !keys.Pressed(input.KeySpace) || keys.StickX != 0.5 {
		test.Fatal("the buttons that went down this frame are not pressed, or the stick is lost")
	}
	keys = console.keysFrom(windowCopy(), ctr.ButtonUp|ctr.ButtonA)
	if keys.Pressed(input.KeyUp) || keys.Pressed(input.KeySpace) || !keys.Down(input.KeySpace) {
		test.Fatal("held buttons were pressed again, or let go")
	}
	keys = console.keysFrom(windowCopy(), 0)
	if keys.Down(input.KeyUp) || keys.Down(input.KeySpace) {
		test.Fatal("let go, the keys are still down")
	}
}
