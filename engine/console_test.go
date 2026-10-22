package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/nintendo3ds"
)

// windowCopy is what go-ui's window hands the scene each frame on the console: a fresh copy of
// its input, without the arrows and A its interface took.
func windowCopy() *input.State {
	copied := input.NewState()
	copied.Feed().SetStick(0.5, 0)
	return copied
}

func TestA3DSButtonIsPressedOnceWhileItIsHeld(test *testing.T) {
	var console consoleInput
	keys := console.keysFrom(windowCopy(), nintendo3ds.ButtonUp|nintendo3ds.ButtonA)
	if !keys.Pressed(input.KeyUp) || !keys.Pressed(input.KeySpace) || stickXOf(keys) != 0.5 {
		test.Fatal("the buttons that went down this frame are not pressed, or the stick is lost")
	}
	keys = console.keysFrom(windowCopy(), nintendo3ds.ButtonUp|nintendo3ds.ButtonA)
	if keys.Pressed(input.KeyUp) || keys.Pressed(input.KeySpace) || !keys.Down(input.KeySpace) {
		test.Fatal("held buttons were pressed again, or let go")
	}
	keys = console.keysFrom(windowCopy(), 0)
	if keys.Down(input.KeyUp) || keys.Down(input.KeySpace) {
		test.Fatal("let go, the keys are still down")
	}
}

// stickXOf is the main stick's x in a state.
func stickXOf(state *input.State) float64 {
	x, _ := state.Stick()
	return x
}
