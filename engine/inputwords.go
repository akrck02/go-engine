package engine

import (
	"strconv"
	"strings"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// The words engine.json binds actions with, read for one target: "Space", "Mouse Left",
// "Pad Start", "Stick X", "Back"; on a 3DS, its own buttons ("A", "Start", "ZL"). A button word with
// "-" or "+" after it ("Left-", "D+") moves an axis instead of pressing.

// inputKind is what an input reads.
type inputKind int

const (
	keyInput inputKind = iota
	mouseInput
	padInput
	consoleInput3DS
	backInput
	stickInput
	touchInput // a touch control the game draws (touchcontrols.go)
)

// stickAxis is one axis of one stick.
type stickAxis int

const (
	mainStickX stickAxis = iota
	mainStickY
	rightStickX
	rightStickY
)

// boundInput is one input an action is bound to: what it reads, and for a button moving an axis,
// which way (-1 or 1; 0 presses).
type boundInput struct {
	kind    inputKind
	code    int // the key, mouse button, pad button, 3DS button bit or stick axis
	towards float64
	name    string // a touch control's
}

// The words for a stick's axes, the same on every target: a 3DS's circle pad is its main stick and
// its C-stick its right one.
var stickWords = map[string]stickAxis{"Stick X": mainStickX, "Stick Y": mainStickY, "Right Stick X": rightStickX, "Right Stick Y": rightStickY}

// mouseWords are the mouse's buttons.
var mouseWords = map[string]input.MouseButton{"Mouse Left": input.MouseLeft, "Mouse Right": input.MouseRight, "Mouse Middle": input.MouseMiddle}

// padWords are a game pad's buttons, after "Pad ".
var padWords = map[string]input.PadButton{"A": input.PadA, "B": input.PadB, "X": input.PadX, "Y": input.PadY,
	"Left Shoulder": input.PadLeftShoulder, "Right Shoulder": input.PadRightShoulder, "Back": input.PadBack,
	"Start": input.PadStart, "Left Thumb": input.PadLeftThumb, "Right Thumb": input.PadRightThumb,
	"Up": input.PadUp, "Right": input.PadRight, "Down": input.PadDown, "Left": input.PadLeft}

// consoleWords are a 3DS's buttons.
var consoleWords = map[string]ctr.Buttons{"A": ctr.ButtonA, "B": ctr.ButtonB, "X": ctr.ButtonX, "Y": ctr.ButtonY,
	"L": ctr.ButtonL, "R": ctr.ButtonR, "ZL": ctr.ButtonZL, "ZR": ctr.ButtonZR, "Start": ctr.ButtonStart,
	"Select": ctr.ButtonSelect, "Up": ctr.ButtonUp, "Down": ctr.ButtonDown, "Left": ctr.ButtonLeft, "Right": ctr.ButtonRight}

// keyWords are the keyboard's keys by their names.
var keyWords = keyNames()

// keyNames names every key: letters, digits, F1 to F12 and the rest.
func keyNames() map[string]input.Key {
	names := map[string]input.Key{"Up": input.KeyUp, "Down": input.KeyDown, "Left": input.KeyLeft, "Right": input.KeyRight,
		"Space": input.KeySpace, "Tab": input.KeyTab, "Enter": input.KeyEnter, "Escape": input.KeyEscape,
		"Shift": input.KeyShift, "Control": input.KeyControl, "Alt": input.KeyAlt, "Plus": input.KeyPlus,
		"Minus": input.KeyMinus, "Page Up": input.KeyPageUp, "Page Down": input.KeyPageDown,
		"Backspace": input.KeyBackspace, "Delete": input.KeyDelete, "Home": input.KeyHome, "End": input.KeyEnd}
	for letter := range 26 {
		names[string(rune('A'+letter))] = input.KeyA + input.Key(letter)
	}
	for digit := range 10 {
		names[strconv.Itoa(digit)] = input.Key0 + input.Key(digit)
	}
	for function := range 12 {
		names["F"+strconv.Itoa(function+1)] = input.KeyF1 + input.Key(function)
	}
	return names
}

// readWord is the input a word names on a target, or why it names none there.
func readWord(word, target string) (boundInput, error) {
	if axis, found := stickWords[word]; found {
		return boundInput{kind: stickInput, code: int(axis)}, nil
	}
	button, towards := word, 0.0
	if strings.HasSuffix(word, "-") || strings.HasSuffix(word, "+") {
		button, towards = word[:len(word)-1], 1
		if strings.HasSuffix(word, "-") {
			towards = -1
		}
	}
	bound, found := readButton(button, target)
	if !found {
		return boundInput{}, failure.Wrap(InputUnknown, word+" on "+target, nil)
	}
	bound.towards = towards
	return bound, nil
}

// readButton is the button a word names on a target.
func readButton(word, target string) (boundInput, bool) {
	if target == Nintendo3DSTarget {
		button, found := consoleWords[word]
		return boundInput{kind: consoleInput3DS, code: int(button)}, found
	}
	if target == AndroidTarget {
		if touch, found := readTouchWord(word); found {
			return touch, true
		}
	}
	if word == "Back" && target == AndroidTarget {
		return boundInput{kind: backInput}, true
	}
	if button, found := mouseWords[word]; found {
		return boundInput{kind: mouseInput, code: int(button)}, true
	}
	if name, isPad := strings.CutPrefix(word, "Pad "); isPad {
		button, found := padWords[name]
		return boundInput{kind: padInput, code: int(button)}, found
	}
	key, found := keyWords[word]
	return boundInput{kind: keyInput, code: int(key)}, found
}

// InputTargets are the targets an action can be bound on, in the order the editor shows them.
var InputTargets = []string{DesktopTarget, Nintendo3DSTarget, AndroidTarget, WebTarget}

// CheckInputs tells which of some words a target does not know, each by its code (InputUnknown).
func CheckInputs(words []string, target string) []error {
	var problems []error
	for _, word := range words {
		if _, err := readWord(word, target); err != nil {
			problems = append(problems, err)
		}
	}
	return problems
}
