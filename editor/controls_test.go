package editor

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

func TestAButtonTogglesAnActionAndOneWayOfAnAxis(test *testing.T) {
	var controls controlsView
	controls.load(map[string]map[string][]string{"turn": {"desktop": {"Stick X"}}, "fire": {}})
	if !controls.axes["turn"] || controls.axes["fire"] {
		test.Fatal("the actions' kinds are not read from their inputs")
	}
	controls.toggleBinding("turn", "A-")
	controls.toggleBinding("turn", "A+") // the other way replaces the first
	controls.toggleBinding("fire", "Space")
	if words := controls.actions["turn"]["desktop"]; !slices.Equal(words, []string{"Stick X", "A+"}) {
		test.Fatalf("turn is bound to %v", words)
	}
	if bound := controls.bindingsOf("A"); !slices.Equal(bound, []string{"turn +"}) {
		test.Fatalf("A says it does %v", bound)
	}
	controls.toggleBinding("fire", "Space")
	if len(controls.actions["fire"]["desktop"]) != 0 || !controls.changed {
		test.Fatal("clicking a bound action again does not unbind it")
	}
}

func TestAButtonOffersEachActionAsItCanDoIt(test *testing.T) {
	if choices := choicesFor("A", "turn", true); len(choices) != 2 || choices[0].word != "A-" {
		test.Fatalf("a button offers an axis as %v", choices)
	}
	if choices := choicesFor("Stick X", "fire", false); len(choices) != 0 {
		test.Fatalf("a stick offers a press as %v", choices)
	}
}

func TestAnInputPressedIsCaughtForTheActionChosen(test *testing.T) {
	editor := open(test)
	view := NewView(editor)
	editor.ShowWorkspace(ControlsWorkspace)
	view.controls.load(nil)
	view.controls.addAction("jump")
	view.controls.catching = true
	state := input.NewState()
	state.SetKey(input.Key1, true) // a key that would change the workspace
	view.keys(state)
	if words := view.controls.actions["jump"][engine.DesktopTarget]; !slices.Equal(words, []string{"1"}) || view.controls.catching {
		test.Fatalf("the key pressed is caught as %v, still catching %v", words, view.controls.catching)
	}
	if editor.ShownWorkspace() != ControlsWorkspace {
		test.Fatal("the key caught also changed the workspace")
	}
}
