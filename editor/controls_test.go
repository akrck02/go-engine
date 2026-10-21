package editor

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

func TestInputsAreAddedOnceTurnedAndRemoved(test *testing.T) {
	var controls controlsView
	controls.load(map[string]map[string][]string{"turn": {"desktop": {"Stick X"}}})
	controls.chosen = "turn"
	controls.addInput("A")
	controls.addInput("A")
	if words := controls.actions["turn"]["desktop"]; !slices.Equal(words, []string{"Stick X", "A"}) {
		test.Fatalf("after adding A twice, turn has %v", words)
	}
	controls.turnInput("turn", 1)
	controls.turnInput("turn", 1)
	if words := controls.actions["turn"]["desktop"]; words[1] != "A+" || actionIcon(controls.actions["turn"]) != iconAxis {
		test.Fatalf("A turned twice is %q", words[1])
	}
	controls.removeInput("turn", 0)
	if words := controls.actions["turn"]["desktop"]; !slices.Equal(words, []string{"A+"}) || !controls.changed {
		test.Fatalf("after removing the stick, turn has %v", words)
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

func TestChipsSayTheirInputPlainly(test *testing.T) {
	for word, label := range map[string]string{"Pad A": "A", "Mouse Left": "Left", "Left-": "Left −", "D+": "D +", "Stick X": "Stick X"} {
		if got := inputLabel(word); got != label {
			test.Errorf("%q is shown as %q", word, got)
		}
	}
}
