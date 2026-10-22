package engine

import "github.com/akrck02/go-renderer/input"

// A game tells the player which input does what ("Tab menu", "Start to pause"). It reads the hint
// from the bindings, so a hint follows engine.json and the player's rebinding by itself, and names
// the device in the player's hands.

// ActionWords are the words an action is bound to on the target played, as engine.json's
// ("Tab", "Pad Start"), rebinding included; none for an action not bound. The slice is the
// engine's own, not a copy, so reading it allocates nothing: do not change it.
func (keys *Input) ActionWords(name string) []string {
	index, found := actionTable.index[name]
	if !found {
		return nil
	}
	return actionTable.actions[index].words
}

// InputKind is a kind of device a player uses: a hint names the inputs of the one in their hands.
type InputKind int

const (
	// NoInputYet is before the player used anything: a hint names an action's first input.
	NoInputYet InputKind = iota
	// KeyboardAndMouse is a computer's keys and mouse buttons.
	KeyboardAndMouse
	// GamePad is a pad's buttons and sticks, on a computer or in a browser.
	GamePad
	// TouchScreen is a phone's screen and its touch controls (and Android's back).
	TouchScreen
	// ConsoleButtons is a console's own buttons and sticks (a 3DS's).
	ConsoleButtons
)

// LastInputKind is the kind of the last input the player pressed or moved: it changes only when
// another kind is used, not when the player stops.
func (keys *Input) LastInputKind() InputKind { return keys.lastKind }

// noteInputKind remembers the kind of what the player used this frame, if anything.
func (keys *Input) noteInputKind() {
	if kind := keys.kindUsedNow(); kind != NoInputYet {
		keys.lastKind = kind
	}
}

// kindUsedNow is the kind of input used this frame, or NoInputYet when nothing was. Touches come
// before the mouse because a finger also moves the pointer.
func (keys *Input) kindUsedNow() InputKind {
	if keys.console.Held != 0 || keys.console.StickX != 0 || keys.console.StickY != 0 {
		return ConsoleButtons
	}
	state := keys.state
	switch {
	case state == nil:
		return NoInputYet
	case len(state.Touches()) > 0:
		return TouchScreen
	case padUsed(state):
		return GamePad
	case state.Feed().HasInput():
		return KeyboardAndMouse
	}
	return NoInputYet
}

// padUsed tells whether a pad's button went down this frame or a stick is out of its dead zone
// (Go Renderer reads a stick at rest as 0).
func padUsed(state *input.State) bool {
	stickX, stickY := state.Stick()
	rightX, rightY := state.RightStick()
	if stickX != 0 || stickY != 0 || rightX != 0 || rightY != 0 {
		return true
	}
	for button := input.PadA; button <= input.PadLeft; button++ {
		if state.PadPressed(button) {
			return true
		}
	}
	return false
}

// ActionHint is what to show a player for an action: the label (InputLabel) of its first input
// of the kind they used last ("Tab" at the keyboard, "Start" with a pad), or of its first input
// when none is of that kind; "" for an action not bound. A pad's buttons are named after the pad
// in their hands.
func (keys *Input) ActionHint(name string) string {
	index, found := actionTable.index[name]
	if !found || len(actionTable.actions[index].words) == 0 {
		return ""
	}
	action := &actionTable.actions[index]
	chosen := 0
	for position, bound := range action.inputs {
		if kindOfInput(bound.kind) == keys.lastKind {
			chosen = position
			break
		}
	}
	layout := input.XboxLayout
	if keys.state != nil {
		layout = keys.state.PadLayout()
	}
	return InputLabel(action.words[chosen], layout)
}

// kindOfInput is the kind of device an input is on.
func kindOfInput(kind inputKind) InputKind {
	switch kind {
	case keyInput, mouseInput:
		return KeyboardAndMouse
	case padInput, stickInput:
		return GamePad
	case touchInput, backInput:
		return TouchScreen
	}
	return ConsoleButtons
}

// HintOf is the hint of an action bound to some words on a target, for a player using a kind of
// input with a pad of some layout: as ActionHint, for an editor showing a device that is not
// played. Words the target does not know are passed over; with none it knows, the hint is "".
func HintOf(words []string, target string, kind InputKind, layout input.PadLayout) string {
	first := ""
	for _, word := range words {
		bound, err := readWord(word, target)
		if err != nil {
			continue // CheckInputs reports it; a hint names what the target has
		}
		if kindOfInput(bound.kind) == kind {
			return InputLabel(word, layout)
		}
		if first == "" {
			first = word
		}
	}
	return InputLabel(first, layout)
}
