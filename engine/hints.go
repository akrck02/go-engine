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
	KeyboardAndMouse
	GamePad
	TouchScreen
	ConsoleButtons // a console's own buttons and sticks (a 3DS's)
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
	if keys.console.Held != 0 || keys.console.CircleX != 0 || keys.console.CircleY != 0 {
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
	case state.HasInput():
		return KeyboardAndMouse
	}
	return NoInputYet
}

// padUsed tells whether a pad's button went down this frame or a stick is out of its dead zone
// (Go Renderer reads a stick at rest as 0).
func padUsed(state *input.State) bool {
	if state.StickX != 0 || state.StickY != 0 || state.RightStickX != 0 || state.RightStickY != 0 {
		return true
	}
	for button := input.PadA; button <= input.PadLeft; button++ {
		if state.PadPressed(button) {
			return true
		}
	}
	return false
}
