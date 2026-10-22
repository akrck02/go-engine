package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// A phone held upright, in layout pixels, with a status bar and a gesture bar.
var (
	phoneScreen = draw2d.Rectangle{Width: 400, Height: 800}
	phoneSafe   = draw2d.Rectangle{Y: 40, Width: 400, Height: 720}
)

func TestAFingerOnATouchButtonHoldsItsKey(test *testing.T) {
	window := input.State{}
	var keys input.State
	buttons := touchLayoutFor(phoneScreen, phoneSafe, 16.0/9).buttons
	for _, button := range buttons {
		window.Feed().SetTouch(1, float64(button.area.X+button.area.Width/2), float64(button.area.Y+button.area.Height/2))
		if touchKeys(&keys, &window, buttons, [4]bool{}); !keys.Down(button.key) || !keys.Pressed(button.key) {
			test.Fatalf("a finger on the %v button does not press its key", button.icon)
		}
		if touchKeys(&keys, &window, buttons, [4]bool{}); !keys.Down(button.key) || keys.Pressed(button.key) {
			test.Fatalf("a finger kept on the %v button does not hold its key", button.icon)
		}
		window.Feed().LiftTouch(1)
		if touchKeys(&keys, &window, buttons, [4]bool{}); keys.Down(button.key) {
			test.Fatalf("the %v button's key stays down after the finger is lifted", button.icon)
		}
	}
}

func TestTheKeyboardStillPlaysWithTouchButtons(test *testing.T) {
	window := input.State{}
	window.Feed().SetKey(input.KeySpace, true)
	var keys input.State
	if touchKeys(&keys, &window, touchLayoutFor(phoneScreen, phoneSafe, 0).buttons, [4]bool{}); !keys.Down(input.KeySpace) {
		test.Fatal("a key held on a keyboard is not held with touch buttons")
	}
}

func TestAnUprightPhoneShowsTheGameAboveItsButtons(test *testing.T) {
	layout := touchLayoutFor(phoneScreen, phoneSafe, 2)
	if layout.scene != (draw2d.Rectangle{Y: 40, Width: 400, Height: 200}) {
		test.Fatalf("the game is shown at %v, not across the top of the safe part at its shape", layout.scene)
	}
	for _, button := range layout.buttons {
		if button.area.Y < layout.scene.Y+layout.scene.Height || button.area.Y+button.area.Height > phoneSafe.Y+phoneSafe.Height {
			test.Fatalf("the %v button at %v is not between the game and the gesture bar", button.icon, button.area)
		}
	}
}

func TestASidewaysPhoneShowsTheGameUnderItsButtons(test *testing.T) {
	sideways := draw2d.Rectangle{Width: 800, Height: 400}
	layout := touchLayoutFor(sideways, draw2d.Rectangle{X: 40, Width: 720, Height: 380}, 16.0/9)
	if layout.scene != sideways {
		test.Fatalf("the game is shown at %v, not filling the screen", layout.scene)
	}
	if bottom := layout.buttons[0].area.Y + touchButtonSide; bottom > 380 {
		test.Fatalf("the buttons reach %v, under the gesture bar", bottom)
	}
}

func TestEveryKeyOfAKeyboardReachesTheGameWithTouchButtons(test *testing.T) {
	window := input.State{}
	window.Feed().SetKey(input.KeyW, true)
	var keys input.State
	if touchKeys(&keys, &window, touchLayoutFor(phoneScreen, phoneSafe, 0).buttons, [4]bool{}); !keys.Down(input.KeyW) {
		test.Fatal("a key with no touch button is lost with touch buttons")
	}
}

// readerFrame runs one frame of an interface with a screen reader on, as a phone with TalkBack.
func readerFrame(test *testing.T, context *ui.Context, state *input.State, build func()) {
	test.Helper()
	context.Begin(state)
	build()
	context.End()
	state.Feed().EndFrame()
}

func TestAScreenReaderReadsAndPressesTheTouchButtons(test *testing.T) {
	context, err := ui.New()
	if err != nil {
		test.Fatal(err)
	}
	context.SetScreenReader(true)
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 400, 800
	state.FramebufferWidth, state.FramebufferHeight = 400, 800
	var console consoleInput
	console.buttons = touchLayoutFor(phoneScreen, phoneSafe, 16.0/9).buttons
	readerFrame(test, context, state, func() { console.describeTouchButtons(context) })
	var jump ui.AccessibleNode
	for _, node := range context.ScreenReaderNodes() {
		if node.Label == "A" {
			jump = node
		}
	}
	if jump.Role != ui.ButtonRole || !jump.Pressable || jump.Area != console.buttons[3].area {
		test.Fatalf("the screen reader reads %+v, not the A button where it is drawn", context.ScreenReaderNodes())
	}
	context.ScreenReaderAct(jump.ID, ui.ReaderPress)
	readerFrame(test, context, state, func() { console.describeTouchButtons(context) })
	var keys input.State
	window := input.State{}
	if touchKeys(&keys, &window, console.buttons, console.readerHeld()); !keys.Pressed(input.KeySpace) {
		test.Fatal("a double tap on the A button does not press its key")
	}
	for range readerPressFrames {
		touchKeys(&keys, &window, console.buttons, console.readerHeld())
	}
	if keys.Down(input.KeySpace) {
		test.Fatal("a double tap on the A button holds its key for good")
	}
}
