package engine

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// A phone held sideways, its safe part 1000 by 500 layout pixels.
var sidewaysSafe = draw2d.Rect{X: 40, Width: 1000, Height: 500}

// someTouchControls are a stick on the left and a button on the right.
func someTouchControls() []TouchControl {
	return []TouchControl{
		{Name: "Move", Kind: TouchStick, X: 0.2, Y: 0.7, Size: 0.4},
		{Name: "Jump", Kind: TouchButton, X: 0.8, Y: 0.7, Size: 0.2},
	}
}

func TestATouchControlIsPlacedAsSharesOfTheScreen(test *testing.T) {
	useTouchControls(someTouchControls())
	placeTouchControls(sidewaysSafe)
	want := draw2d.Rect{X: 40 + 800 - 50, Y: 350 - 50, Width: 100, Height: 100}
	if touchPanel.areas[1] != want {
		test.Fatalf("the button is at %v, not %v", touchPanel.areas[1], want)
	}
}

func TestTouchControlsAreBoundByTheirWords(test *testing.T) {
	controls := someTouchControls()
	if words := TouchWords(controls[0]); !slices.Equal(words, []string{"Touch Move X", "Touch Move Y"}) {
		test.Fatalf("a stick's words are %v", words)
	}
	if problems := CheckInputs(TouchWords(controls[1]), AndroidTarget); len(problems) != 0 {
		test.Fatalf("a phone does not know a button's word: %v", problems)
	}
	if problems := CheckInputs([]string{"Touch Jump"}, DesktopTarget); len(problems) != 1 {
		test.Fatal("a computer knows a touch control's word")
	}
}

func TestAFingerOnATouchButtonHoldsItsAction(test *testing.T) {
	useTouchControls(someTouchControls())
	placeTouchControls(sidewaysSafe)
	BindActions(map[string]map[string][]string{"jump": {AndroidTarget: {"Touch Jump"}}}, AndroidTarget)
	keys := &Input{state: &input.State{}}
	readTouchControls([]input.Touch{{Finger: 1, X: 840, Y: 350}})
	if keys.updateActions(); !keys.Action("jump").Pressed {
		test.Fatal("a finger on the button does not press its action")
	}
	readTouchControls(nil)
	if keys.updateActions(); !keys.Action("jump").Released {
		test.Fatal("lifting the finger does not let the action go")
	}
}

func TestDraggingATouchStickMovesItsAxes(test *testing.T) {
	useTouchControls(someTouchControls())
	placeTouchControls(sidewaysSafe)
	BindActions(map[string]map[string][]string{"turn": {AndroidTarget: {"Touch Move X"}}, "climb": {AndroidTarget: {"Touch Move Y"}}}, AndroidTarget)
	keys := &Input{state: &input.State{}}
	// the stick's centre is at (240, 350) and its radius 100: a finger right of it and above
	readTouchControls([]input.Touch{{Finger: 1, X: 240 + 100, Y: 350 - 50}})
	keys.updateActions()
	if turn, climb := keys.Action("turn").Value, keys.Action("climb").Value; turn != 1 || climb != 0.5 {
		test.Fatalf("the stick dragged right and up gives turn %v and climb %v", turn, climb)
	}
	readTouchControls([]input.Touch{{Finger: 1, X: 900, Y: 50}}) // far from it
	if keys.updateActions(); keys.Action("turn").Value != 0 {
		test.Fatal("a finger far from the stick moves it")
	}
}

func TestReadingTouchControlsAllocatesNothing(test *testing.T) {
	useTouchControls(someTouchControls())
	placeTouchControls(sidewaysSafe)
	touches := []input.Touch{{Finger: 1, X: 840, Y: 350}}
	if allocations := testing.AllocsPerRun(100, func() { readTouchControls(touches) }); allocations != 0 {
		test.Fatalf("reading the touch controls allocates %v times", allocations)
	}
}

func TestATouchControlIsAsBigUprightAndStaysOnTheScreen(test *testing.T) {
	useTouchControls(someTouchControls())
	upright := draw2d.Rect{Width: 500, Height: 1000}
	placeTouchControls(upright)
	stick := touchPanel.areas[0]
	if stick.Width != 0.4*500 {
		test.Fatalf("the stick held upright is %v wide, not a share of the shorter side", stick.Width)
	}
	if stick.X < 0 || stick.X+stick.Width > upright.Width || stick.Y+stick.Height > upright.Height {
		test.Fatalf("the stick at %v leaves the screen", stick)
	}
}
