package engine

import (
	"strings"

	"github.com/akrck02/go-renderer/input"
)

// A word of engine.json ("Escape", "Pad A", "Touch Jump") is what an author writes; a player reads
// what the key cap or the button says ("Esc", "✕", "Jump"). Every name here is a constant or a
// piece of the word, so naming one allocates nothing.

// keyLabels are the keys whose caps say something shorter than their word; the rest say their word.
var keyLabels = map[string]string{"Escape": "Esc", "Page Up": "PgUp", "Page Down": "PgDn",
	"Control": "Ctrl", "Delete": "Del"}

// otherLabels are the mouse's buttons and the sticks, named the same on every pad.
var otherLabels = map[string]string{"Mouse Left": "Left click", "Mouse Right": "Right click",
	"Mouse Middle": "Middle click", "Stick X": "Stick", "Stick Y": "Stick",
	"Right Stick X": "Right stick", "Right Stick Y": "Right stick"}

// padLayoutCount is how many pad layouts padLabels names buttons for.
const padLayoutCount = 3

// padLabels are a pad's buttons (after "Pad ") as each layout prints them, in input.PadLayout's
// order: Xbox, PlayStation, Nintendo. The face buttons go bottom, right, left, top; a Nintendo pad
// prints B at the bottom and A on the right. Buttons not here (the cross) say their word.
var padLabels = map[string][padLayoutCount]string{
	"A":              {"A", "✕", "B"},
	"B":              {"B", "○", "A"},
	"X":              {"X", "□", "Y"},
	"Y":              {"Y", "△", "X"},
	"Left Shoulder":  {"LB", "L1", "L"},
	"Right Shoulder": {"RB", "R1", "R"},
	"Back":           {"Back", "Share", "−"},
	"Start":          {"Start", "Options", "+"},
	"Left Thumb":     {"LS", "L3", "LS"},
	"Right Thumb":    {"RS", "R3", "RS"},
}

// InputLabel is what a key cap or a button says for a word of engine.json, on a pad of some
// layout: "Escape" is "Esc", "Pad A" is "✕" on a PlayStation pad, "Touch Jump" is "Jump". A
// direction's "-" or "+" is dropped ("Left-" is "Left"); a combination stays whole ("Alt+Enter").
// A word it has no shorter name for (a 3DS's "Start") is itself.
func InputLabel(word string, layout input.PadLayout) string {
	if len(word) > 1 && (strings.HasSuffix(word, "-") || strings.HasSuffix(word, "+")) {
		word = word[:len(word)-1]
	}
	if label, found := otherLabels[word]; found {
		return label
	}
	if button, isPad := strings.CutPrefix(word, "Pad "); isPad {
		return padLabel(button, layout)
	}
	if control, isTouch := strings.CutPrefix(word, touchWordPrefix); isTouch {
		return touchLabel(control)
	}
	if label, found := keyLabels[word]; found {
		return label
	}
	return word
}

// padLabel is what a pad's button says on a layout; a layout it does not know is read as Xbox's.
func padLabel(button string, layout input.PadLayout) string {
	labels, found := padLabels[button]
	if !found {
		return button
	}
	if layout < 0 || layout >= padLayoutCount {
		layout = input.XboxLayout
	}
	return labels[layout]
}

// touchLabel is a touch control's name, without the axis of a stick ("Stick X" is "Stick").
func touchLabel(control string) string {
	if stick, isX := strings.CutSuffix(control, touchAxisX); isX {
		return stick
	}
	if stick, isY := strings.CutSuffix(control, touchAxisY); isY {
		return stick
	}
	return control
}
