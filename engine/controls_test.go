package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/nintendo3ds"
)

func TestThe3DSsShouldersAndStartAreReadAsTheirKeys(test *testing.T) {
	var console consoleInput
	window := input.NewState()
	cases := map[nintendo3ds.Buttons]Key{nintendo3ds.ButtonL: ButtonLeftShoulder, nintendo3ds.ButtonR: ButtonRightShoulder, nintendo3ds.ButtonZL: ButtonLeftTrigger, nintendo3ds.ButtonZR: ButtonRightTrigger, nintendo3ds.ButtonStart: ButtonStart}
	for button, key := range cases {
		keys := console.keysFrom(window, button)
		if !keys.Pressed(key) || !keys.Down(key) {
			test.Errorf("button %b is not read as key %v, pressed and held", button, key)
		}
		console.keysFrom(window, 0)
	}
}

func TestThe3DSsSticksGyroscopeAndTouchAreReadWhole(test *testing.T) {
	reading := nintendo3ds.Reading{RightStickX: 146, RightStickY: 0, Gyroscope: nintendo3ds.Gyroscope{Yaw: 30},
		Held: nintendo3ds.ButtonTouch, TouchX: 160, TouchY: 120}
	keys := Input{state: input.NewState(), console: reading}
	keys.state.Feed().SetRightStick(1, -0.5)
	if x, y := keys.RightStick(); x != 1 || y != -0.5 {
		test.Errorf("the right stick right and down reads %v, %v", x, y)
	}
	if keys.Gyroscope().Yaw != 30 {
		test.Errorf("the gyroscope reads %+v", keys.Gyroscope())
	}
	if x, y, touching := keys.SecondScreenTouch(); !touching || x != 160 || y != 120 {
		test.Errorf("a touch in the middle reads %v, %v, %v", x, y, touching)
	}
	keys.state.Feed().SetStick(0.5, 1)
	if x, y := keys.Stick(); x != 0.5 || y != 1 {
		test.Errorf("the circle pad up and right reads %v, %v", x, y)
	}
}

func TestTheMouseMovesAndThePointerIsCapturedWhenAsked(test *testing.T) {
	game := &runner{played: load(test)}
	state := input.NewState()
	state.CursorDeltaX, state.CursorDeltaY = 4, -2
	SetPointerCaptured(true)
	defer SetPointerCaptured(false)
	game.step(state, 1.0/60)
	if !state.CursorLocked() {
		test.Fatal("the pointer is not captured")
	}
	if x, y := game.played.input.MouseMovement(); x != 4 || y != -2 {
		test.Fatalf("the mouse moved %v, %v", x, y)
	}
}

func TestThe3DSsKeysCarryBothSticks(test *testing.T) {
	var console consoleInput
	window := input.NewState()
	window.Feed().SetStick(0.5, 0)
	window.Feed().SetRightStick(0, 1)
	keys := console.keysFrom(window, 0)
	stickX, _ := keys.Stick()
	_, rightY := keys.RightStick()
	if stickX != 0.5 || rightY != 1 {
		test.Fatalf("the game's keys have the sticks at %v and %v", stickX, rightY)
	}
}

func TestTheDesktopBuildRunsOnTheDesktop(test *testing.T) {
	if RunningOn() != DesktopTarget {
		test.Fatalf("the tests' build runs on %q", RunningOn())
	}
}
