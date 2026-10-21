package engine

import (
	"sort"

	"github.com/akrck02/go-renderer/input"
)

// The words a target's actions can be bound with, grouped for an editor to offer, and the one an
// input pressed now stands for, for an editor to catch.

// The groups of inputs, as an editor shows them.
const (
	KeyboardInputs = "Keyboard"
	MouseInputs    = "Mouse"
	PadInputs      = "Pad"
	StickInputs    = "Sticks"
	ButtonInputs   = "Buttons" // a device's own buttons (a 3DS's)
	SystemInputs   = "System"  // the system's back (Android's)
)

// InputChoice is a word an action can be bound with on a target, and its group.
type InputChoice struct{ Word, Group string }

// InputChoices are the words a target knows, group by group, each group in a reading order.
func InputChoices(target string) []InputChoice {
	var choices []InputChoice
	add := func(group string, words []string) {
		for _, word := range words {
			choices = append(choices, InputChoice{Word: word, Group: group})
		}
	}
	switch target {
	case Nintendo3DSTarget:
		add(ButtonInputs, consoleOrder)
	case AndroidTarget:
		add(SystemInputs, []string{"Back"})
		add(KeyboardInputs, keyOrder())
	default:
		add(KeyboardInputs, keyOrder())
		add(MouseInputs, []string{"Mouse Left", "Mouse Right", "Mouse Middle"})
		add(PadInputs, padOrder)
	}
	add(StickInputs, []string{"Stick X", "Stick Y", "Right Stick X", "Right Stick Y"})
	return choices
}

// consoleOrder are a 3DS's buttons as they sit on it: the cross, the face, the shoulders, the rest.
var consoleOrder = []string{"Up", "Down", "Left", "Right", "A", "B", "X", "Y", "L", "R", "ZL", "ZR", "Start", "Select"}

// padOrder are a game pad's buttons, faces first.
var padOrder = []string{"Pad A", "Pad B", "Pad X", "Pad Y", "Pad Left Shoulder", "Pad Right Shoulder",
	"Pad Start", "Pad Back", "Pad Up", "Pad Down", "Pad Left", "Pad Right", "Pad Left Thumb", "Pad Right Thumb"}

// keyOrder are the keys by their order on input's list: letters, digits, arrows and the rest.
func keyOrder() []string {
	words := make([]string, 0, len(keyWords))
	for word := range keyWords {
		words = append(words, word)
	}
	sort.Slice(words, func(first, second int) bool { return keyWords[words[first]] < keyWords[words[second]] })
	return words
}

// PressedInput is the word of an input pressed this frame on a target, if one was: a key, a mouse
// button or a pad's button (a 3DS's buttons and Android's back cannot be pressed from a computer).
func PressedInput(state *input.State, target string) (string, bool) {
	for _, choice := range InputChoices(target) {
		bound, err := readWord(choice.Word, target)
		if err != nil || !pressedNow(state, bound) {
			continue
		}
		return choice.Word, true
	}
	return "", false
}

// pressedNow tells whether a button input went down this frame.
func pressedNow(state *input.State, bound boundInput) bool {
	switch bound.kind {
	case keyInput:
		return state.Pressed(input.Key(bound.code))
	case mouseInput:
		return state.ButtonPressed(input.MouseButton(bound.code))
	case padInput:
		return state.PadPressed(input.PadButton(bound.code))
	}
	return false
}
