package engine

import (
	"errors"
	"testing"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// gameActions are a made-up game's bindings.
var gameActions = map[string]map[string][]string{
	"fire": {"desktop": {"Space", "Mouse Left", "Pad A"}, "nintendo3ds": {"A", "R"}},
	"menu": {"desktop": {"Tab", "F1"}, "android": {"Back"}},
	"turn": {"desktop": {"A-", "D+", "Stick X"}, "nintendo3ds": {"Left-", "Right+", "Stick X"}},
}

// frameWith works the actions out for one frame of an input.
func frameWith(state *input.State, console ctr.Reading) *Input {
	keys := &Input{state: state, console: console}
	keys.updateActions()
	return keys
}

func TestAnActionIsPressedHeldAndLetGoByAnyOfItsInputs(test *testing.T) {
	if problems := BindActions(gameActions, DesktopTarget); len(problems) != 0 {
		test.Fatal(problems)
	}
	for _, press := range []func(state *input.State){
		func(state *input.State) { state.SetKey(input.KeySpace, true) },
		func(state *input.State) { state.SetButton(input.MouseLeft, true) },
		func(state *input.State) { state.SetPadButton(input.PadA, true) },
	} {
		state := input.NewState()
		press(state)
		if fire := frameWith(state, ctr.Reading{}).Action("fire"); !fire.Pressed || !fire.Down || fire.Value != 1 {
			test.Fatalf("fire is %+v", fire)
		}
		if fire := frameWith(state, ctr.Reading{}).Action("fire"); fire.Pressed || !fire.Down {
			test.Fatalf("fire held is %+v", fire)
		}
		if fire := frameWith(input.NewState(), ctr.Reading{}).Action("fire"); !fire.Released || fire.Down {
			test.Fatalf("fire let go is %+v", fire)
		}
	}
}

func TestAnAxisActionAddsItsButtonsAndSticks(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	state := input.NewState()
	state.SetKey(input.KeyA, true)
	if turn := frameWith(state, ctr.Reading{}).Action("turn"); turn.Value != -1 || !turn.Down {
		test.Fatalf("A held turns %+v", turn)
	}
	stick := input.NewState()
	stick.StickX = 0.3
	if turn := frameWith(stick, ctr.Reading{}).Action("turn"); turn.Value != 0.3 || turn.Down {
		test.Fatalf("the stick a little right turns %+v", turn)
	}
}

func TestA3DSReadsItsOwnButtons(test *testing.T) {
	if problems := BindActions(gameActions, Nintendo3DSTarget); len(problems) != 0 {
		test.Fatal(problems)
	}
	frameWith(input.NewState(), ctr.Reading{})
	if fire := frameWith(input.NewState(), ctr.Reading{Held: ctr.ButtonR}).Action("fire"); !fire.Pressed {
		test.Fatalf("R on a 3DS fires %+v", fire)
	}
	if turn := frameWith(input.NewState(), ctr.Reading{Held: ctr.ButtonRight}).Action("turn"); turn.Value != 1 {
		test.Fatalf("the cross right on a 3DS turns %+v", turn)
	}
}

func TestAnInputATargetDoesNotHaveIsReportedByItsCode(test *testing.T) {
	problems := BindActions(map[string]map[string][]string{"fire": {"nintendo3ds": {"Mouse Left", "A"}}}, Nintendo3DSTarget)
	if len(problems) != 1 || !errors.Is(problems[0], InputUnknown) {
		test.Fatalf("a mouse on a 3DS gives %v", problems)
	}
	if len(actionTable.actions[0].inputs) != 1 {
		test.Fatal("the inputs the 3DS has are not bound")
	}
}

func TestReadingAnActionAllocatesNothing(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	allocations := testing.AllocsPerRun(100, func() {
		keys.updateActions()
		_ = keys.Action("fire")
	})
	if allocations != 0 {
		test.Fatalf("a frame of actions allocates %v times", allocations)
	}
}

func TestABindingChangedWhilePlayingIsKeptAndReadBack(test *testing.T) {
	for _, variable := range []string{"XDG_CONFIG_HOME", "HOME", "APPDATA", "LOCALAPPDATA"} {
		test.Setenv(variable, test.TempDir())
	}
	settings := Settings{Name: "Lighthouse", Actions: gameActions}
	bindGame(settings, DesktopTarget)
	keys := &Input{state: input.NewState()}
	if problems := keys.Rebind("fire", []string{"Enter"}); len(problems) != 0 {
		test.Fatal(problems)
	}
	if err := SaveBindings(); err != nil {
		test.Fatal(err)
	}
	bindGame(settings, DesktopTarget) // the next time the game plays
	state := input.NewState()
	state.SetKey(input.KeyEnter, true)
	if fire := frameWith(state, ctr.Reading{}).Action("fire"); !fire.Pressed {
		test.Fatalf("Enter, rebound to fire and kept, fires %+v", fire)
	}
	space := input.NewState()
	space.SetKey(input.KeySpace, true)
	frameWith(input.NewState(), ctr.Reading{})
	if fire := frameWith(space, ctr.Reading{}).Action("fire"); fire.Down {
		test.Fatal("Space still fires after fire was rebound to Enter")
	}
}

func TestEveryInputChoiceIsAWordItsTargetKnows(test *testing.T) {
	for _, target := range InputTargets {
		for _, choice := range InputChoices(target) {
			if problems := CheckInputs([]string{choice.Word}, target); len(problems) != 0 {
				test.Errorf("%s offers %q, which it does not know", target, choice.Word)
			}
		}
	}
}

func TestAKeyPressedIsCaughtAsItsWord(test *testing.T) {
	state := input.NewState()
	state.SetPadButton(input.PadStart, true)
	if word, found := PressedInput(state, DesktopTarget); !found || word != "Pad Start" {
		test.Fatalf("the pad's Start is caught as %q, %v", word, found)
	}
	state.EndFrame()
	state.SetKey(input.KeyF11, true)
	if word, _ := PressedInput(state, DesktopTarget); word != "F11" {
		test.Fatalf("F11 is caught as %q", word)
	}
}

func TestAKeyWithAModifierPressesOnlyWithBothHeld(test *testing.T) {
	BindActions(map[string]map[string][]string{"save": {DesktopTarget: {"Control+S"}}}, DesktopTarget)
	state := input.NewState()
	state.SetKey(input.KeyS, true)
	if frameWith(state, ctr.Reading{}).Action("save").Down {
		test.Fatal("S alone saves")
	}
	state.SetKey(input.KeyControl, true)
	if !frameWith(state, ctr.Reading{}).Action("save").Pressed {
		test.Fatal("Control and S do not save")
	}
}

func TestAKeyTappedBetweenTwoFramesStillPressesItsAction(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	frameWith(input.NewState(), ctr.Reading{})
	state := input.NewState()
	state.SetKey(input.KeySpace, true)
	state.SetKey(input.KeySpace, false) // let go before the frame: a quick tap
	if fire := frameWith(state, ctr.Reading{}).Action("fire"); !fire.Pressed {
		test.Fatalf("a tap is %+v", fire)
	}
}
