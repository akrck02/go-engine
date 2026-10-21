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
