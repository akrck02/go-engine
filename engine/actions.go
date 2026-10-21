package engine

import (
	"math"
	"sort"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// A game reads what the player does as actions with names ("fire", "menu", "turn"), each bound in
// engine.json to the inputs of every target. The bindings are read once into a table; each frame
// every action's state is worked out from it, before the behaviours run, so reading one is a
// lookup that allocates nothing.

// ActionState is an action this frame: held, pressed or let go, and its value from -1 to 1 (an
// axis's, or 1 while a button holds it).
type ActionState struct {
	Down, Pressed, Released bool
	Value                   float64
}

// boundAction is an action's inputs on the target played, and its state.
type boundAction struct {
	inputs []boundInput
	state  ActionState
}

// actionTable is the actions bound for the target played, and where each is by its name.
var actionTable struct {
	actions []boundAction
	index   map[string]int
}

// actionDownAt is how far an axis goes before its action counts as held.
const actionDownAt = 0.5

// BindActions reads the actions' bindings (engine.json's actions) for a target, replacing those
// read before. Words the target does not know are left out, each reported.
func BindActions(bindings map[string]map[string][]string, target string) []error {
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names) // the same order every time, for the problems to read the same
	actionTable.actions, actionTable.index = make([]boundAction, 0, len(names)), make(map[string]int, len(names))
	var problems []error
	for _, name := range names {
		var bound boundAction
		for _, word := range bindings[name][target] {
			input, err := readWord(word, target)
			if err != nil {
				problems = append(problems, err)
				continue
			}
			bound.inputs = append(bound.inputs, input)
		}
		actionTable.index[name] = len(actionTable.actions)
		actionTable.actions = append(actionTable.actions, bound)
	}
	return problems
}

// Action is an action's state this frame; an action not bound is never held.
func (keys *Input) Action(name string) ActionState {
	index, found := actionTable.index[name]
	if !found {
		return ActionState{}
	}
	return actionTable.actions[index].state
}

// updateActions works every action's state out for this frame.
func (keys *Input) updateActions() {
	for index := range actionTable.actions {
		action := &actionTable.actions[index]
		wasDown := action.state.Down
		down, value := keys.readInputs(action.inputs)
		action.state = ActionState{Down: down, Pressed: down && !wasDown, Released: !down && wasDown, Value: value}
	}
}

// readInputs is whether any of some inputs holds an action, and the value they give it together.
func (keys *Input) readInputs(inputs []boundInput) (down bool, value float64) {
	for _, bound := range inputs {
		if bound.kind == stickInput {
			value += keys.stickAxis(stickAxis(bound.code))
			continue
		}
		if !keys.held(bound) {
			continue
		}
		if bound.towards != 0 {
			value += bound.towards
			continue
		}
		down, value = true, value+1
	}
	value = math.Max(-1, math.Min(1, value))
	return down || math.Abs(value) >= actionDownAt, value
}

// held tells whether a button input is held this frame (Android's back, only on the frame it came).
func (keys *Input) held(bound boundInput) bool {
	state := keys.state
	switch bound.kind {
	case consoleInput3DS:
		return keys.console.Held&ctr.Buttons(bound.code) != 0
	case backInput:
		return state != nil && state.TakeBack()
	}
	if state == nil {
		return false
	}
	switch bound.kind {
	case mouseInput:
		return state.Button(input.MouseButton(bound.code))
	case padInput:
		return state.PadDown(input.PadButton(bound.code))
	}
	return state.Down(input.Key(bound.code))
}

// stickAxis is one axis of a stick, up and right positive.
func (keys *Input) stickAxis(axis stickAxis) float64 {
	if keys.state == nil {
		return 0
	}
	switch axis {
	case mainStickY:
		return keys.state.StickY
	case rightStickX:
		return keys.state.RightStickX
	case rightStickY:
		return keys.state.RightStickY
	}
	return keys.state.StickX
}
