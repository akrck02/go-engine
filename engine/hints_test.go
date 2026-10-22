package engine

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/nintendo3ds"
)

func TestAnActionGivesTheWordsItIsBoundToOnTheTargetPlayed(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"Space", "Mouse Left", "Pad A"}) {
		test.Fatalf("fire's words on a computer are %q", words)
	}
	BindActions(gameActions, Nintendo3DSTarget)
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"A", "R"}) {
		test.Fatalf("fire's words on a 3DS are %q", words)
	}
}

func TestAReboundActionGivesItsNewWords(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.Rebind("fire", []string{"Enter", "Not A Key", "Pad B"})
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"Enter", "Pad B"}) {
		test.Fatalf("fire rebound gives %q", words)
	}
}

func TestAnActionNotBoundGivesNoWords(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	if words := keys.ActionWords("jump"); len(words) != 0 {
		test.Fatalf("an action not bound gives %q", words)
	}
	if words := keys.ActionWords("menu"); !slices.Equal(words, []string{"Tab", "F1"}) {
		test.Fatalf("menu gives %q", words)
	}
}

func TestReadingAnActionsWordsAllocatesNothing(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	allocations := testing.AllocsPerRun(100, func() { _ = keys.ActionWords("fire") })
	if allocations != 0 {
		test.Fatalf("reading an action's words allocates %v times", allocations)
	}
}

func TestTheLastKindOfInputUsedFollowsThePlayersHands(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.updateActions()
	if kind := keys.LastInputKind(); kind != NoInputYet {
		test.Fatalf("before any input the kind is %v", kind)
	}
	steps := []struct {
		press func(state *input.State)
		want  InputKind
	}{
		{func(state *input.State) { state.Feed().SetKey(input.KeyTab, true) }, KeyboardAndMouse},
		{func(state *input.State) { state.Feed().SetPadButton(input.PadStart, true) }, GamePad},
		{func(state *input.State) { state.Feed().SetTouch(1, 10, 10) }, TouchScreen},
		{func(state *input.State) { state.Feed().SetRightStick(0.8, 0) }, GamePad},
		{func(state *input.State) { state.Feed().SetButton(input.MouseLeft, true) }, KeyboardAndMouse},
	}
	for _, step := range steps {
		keys.state = input.NewState()
		step.press(keys.state)
		keys.updateActions()
		if kind := keys.LastInputKind(); kind != step.want {
			test.Fatalf("the kind is %v, want %v", kind, step.want)
		}
		keys.state = input.NewState() // nothing happens the next frame: the kind stays
		keys.updateActions()
		if kind := keys.LastInputKind(); kind != step.want {
			test.Fatalf("a frame without input changes the kind to %v", kind)
		}
	}
}

func TestA3DSsButtonsAreAConsolesOwn(test *testing.T) {
	BindActions(gameActions, Nintendo3DSTarget)
	keys := &Input{state: input.NewState(), console: nintendo3ds.Reading{Held: nintendo3ds.ButtonStart}}
	keys.updateActions()
	if kind := keys.LastInputKind(); kind != ConsoleButtons {
		test.Fatalf("a 3DS's Start gives %v", kind)
	}
}

// hintActions are bindings of a computer with a keyboard and a pad, and of a 3DS.
var hintActions = map[string]map[string][]string{
	"menu":  {"desktop": {"Tab", "Pad Start"}, "nintendo3ds": {"Start"}},
	"climb": {"desktop": {"Pad Up", "W"}},
}

func TestAHintNamesTheInputOfTheDeviceInThePlayersHands(test *testing.T) {
	BindActions(hintActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.updateActions()
	if hint := keys.ActionHint("menu"); hint != "Tab" {
		test.Fatalf("before any input menu's hint is %q", hint)
	}
	keys.state.Feed().SetPadButton(input.PadA, true)
	keys.updateActions()
	if hint := keys.ActionHint("menu"); hint != "Start" {
		test.Fatalf("with a pad menu's hint is %q", hint)
	}
	keys.state.Feed().SetPadLayout(input.PlayStationLayout)
	if hint := keys.ActionHint("menu"); hint != "Options" {
		test.Fatalf("with a PlayStation pad menu's hint is %q", hint)
	}
	keys.state = input.NewState()
	keys.state.Feed().SetKey(input.KeyEscape, true)
	keys.updateActions()
	if hint := keys.ActionHint("climb"); hint != "W" {
		test.Fatalf("at the keyboard climb's hint is %q", hint)
	}
	if hint := keys.ActionHint("jump"); hint != "" {
		test.Fatalf("an action not bound hints %q", hint)
	}
}

func TestAHintOnThe3DSNamesItsOwnButton(test *testing.T) {
	BindActions(hintActions, Nintendo3DSTarget)
	keys := &Input{state: input.NewState(), console: nintendo3ds.Reading{Held: nintendo3ds.ButtonA}}
	keys.updateActions()
	if hint := keys.ActionHint("menu"); hint != "Start" {
		test.Fatalf("menu's hint on a 3DS is %q", hint)
	}
}

func TestAHintWhenNoInputOfTheKindIsBoundIsTheFirst(test *testing.T) {
	BindActions(hintActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.state.Feed().SetTouch(1, 10, 10)
	keys.updateActions()
	if hint := keys.ActionHint("menu"); hint != "Tab" {
		test.Fatalf("with a touch, menu's hint is %q", hint)
	}
}

func TestReadingAHintAllocatesNothing(test *testing.T) {
	BindActions(hintActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.state.Feed().SetPadButton(input.PadA, true)
	allocations := testing.AllocsPerRun(100, func() {
		keys.updateActions()
		_ = keys.ActionHint("menu")
		_ = keys.ActionHint("climb")
	})
	if allocations != 0 {
		test.Fatalf("reading a hint allocates %v times", allocations)
	}
}

func TestAHintOfSomeWordsNamesTheFirstOfTheKindAsked(test *testing.T) {
	words := []string{"Tab", "Pad Start", "Not A Key"}
	if hint := HintOf(words, DesktopTarget, GamePad, input.NintendoLayout); hint != "+" {
		test.Fatalf("with a Nintendo pad the hint is %q", hint)
	}
	if hint := HintOf(words, DesktopTarget, TouchScreen, input.XboxLayout); hint != "Tab" {
		test.Fatalf("with no word of the kind the hint is %q", hint)
	}
	if hint := HintOf([]string{"Not A Key"}, DesktopTarget, GamePad, input.XboxLayout); hint != "" {
		test.Fatalf("with no word the target knows the hint is %q", hint)
	}
}
