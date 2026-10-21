package editor

import (
	"slices"
	"strconv"
	"strings"

	"github.com/akrck02/go-engine/engine"
)

// The phone's layout, designed in the Controls workspace: the sticks and buttons the game draws
// on its screen, where they sit and how big they are. Places and sizes are shares of the screen
// (engine.TouchControl), so the layout fits every phone.

// touchDesign is the phone's layout being designed: whether it is, the control chosen and the one
// being dragged (by their places in the layout), and whether the pointer was pressed last frame.
type touchDesign struct {
	on          bool
	selected    int
	dragging    int
	wasPressing bool
	draft       string // the chosen control's name, being typed
	nameFocused bool   // its field had the keyboard last frame
}

// The touch controls' sizes, as shares of the screen's height.
const (
	touchStickSize  = 0.34
	touchButtonSize = 0.18
	touchSizeStep   = 0.02
	touchSmallest   = 0.08
	touchLargest    = 0.5
)

// defaultTouchControls are a layout to start from: a stick for the left thumb, two buttons for
// the right one.
func defaultTouchControls() []engine.TouchControl {
	return []engine.TouchControl{
		{Name: "Stick", Kind: engine.TouchStick, X: 0.18, Y: 0.7, Size: touchStickSize},
		{Name: "B", Kind: engine.TouchButton, X: 0.76, Y: 0.62, Size: touchButtonSize},
		{Name: "A", Kind: engine.TouchButton, X: 0.88, Y: 0.8, Size: touchButtonSize},
	}
}

// startLayout makes the phone's layout the game's own, from the default one; the engine's four
// buttons, which hold keys, go.
func (controls *controlsView) startLayout() {
	controls.remember()
	controls.touch = defaultTouchControls()
}

// addTouch adds a stick or a button in the middle of the screen, and chooses it.
func (controls *controlsView) addTouch(kind engine.TouchKind) {
	controls.remember()
	added := engine.TouchControl{Name: controls.freeTouchName(kind), Kind: kind, X: 0.5, Y: 0.5, Size: touchButtonSize}
	if kind == engine.TouchStick {
		added.Size = touchStickSize
	}
	controls.touch = append(controls.touch, added)
	controls.chooseTouch(len(controls.touch) - 1)
}

// freeTouchName is a name no touch control has: a stick's "Stick", "Stick 2"…, a button's a
// letter after the ones taken.
func (controls *controlsView) freeTouchName(kind engine.TouchKind) string {
	if kind == engine.TouchStick {
		name := "Stick"
		for number := 2; controls.touchIndex(name) >= 0; number++ {
			name = "Stick " + strconv.Itoa(number)
		}
		return name
	}
	for letter := 'A'; letter <= 'Z'; letter++ {
		if controls.touchIndex(string(letter)) < 0 {
			return string(letter)
		}
	}
	return "Button " + strconv.Itoa(len(controls.touch)+1)
}

// touchIndex is where a touch control is, by its name; -1 when none has it.
func (controls *controlsView) touchIndex(name string) int {
	return slices.IndexFunc(controls.touch, func(each engine.TouchControl) bool { return each.Name == name })
}

// chooseTouch chooses a touch control, its name ready to change.
func (controls *controlsView) chooseTouch(index int) {
	controls.design.selected = index
	controls.design.draft = ""
	if index >= 0 && index < len(controls.touch) {
		controls.design.draft = controls.touch[index].Name
	}
}

// moveTouch puts a touch control's centre at a place on the screen (shares of it, from 0 to 1),
// kept whole on the screen.
func (controls *controlsView) moveTouch(index int, x, y float32) {
	each := &controls.touch[index]
	halfAcross := each.Size * phoneScreenHeight / phoneScreenWidth / 2
	each.X = min(max(x, halfAcross), 1-halfAcross)
	each.Y = min(max(y, each.Size/2), 1-each.Size/2)
}

// resizeTouch makes the chosen touch control bigger, or smaller for a step below 0.
func (controls *controlsView) resizeTouch(step float32) {
	index := controls.design.selected
	if index < 0 || index >= len(controls.touch) {
		return
	}
	controls.remember()
	each := &controls.touch[index]
	each.Size = min(max(each.Size+step, touchSmallest), touchLargest)
	controls.moveTouch(index, each.X, each.Y)
}

// removeTouch removes the chosen touch control, and the actions' words for it.
func (controls *controlsView) removeTouch() {
	index := controls.design.selected
	if index < 0 || index >= len(controls.touch) {
		return
	}
	controls.remember()
	words := engine.TouchWords(controls.touch[index])
	for _, targets := range controls.actions {
		targets[engine.AndroidTarget] = slices.DeleteFunc(targets[engine.AndroidTarget], func(word string) bool {
			return slices.Contains(words, strings.TrimRight(word, "-+"))
		})
	}
	controls.touch = slices.Delete(controls.touch, index, index+1)
	controls.chooseTouch(noControl)
}

// renameTouch gives the chosen touch control the name typed, and the actions' words follow it. A
// name that is empty, taken, or ends as a stick's axis (" X", " Y") leaves it as it was.
func (controls *controlsView) renameTouch() {
	index, name := controls.design.selected, strings.TrimSpace(controls.design.draft)
	if index < 0 || index >= len(controls.touch) || name == controls.touch[index].Name {
		return
	}
	if name == "" || controls.touchIndex(name) >= 0 || isAxisWord(name) {
		controls.design.draft = controls.touch[index].Name
		return
	}
	controls.remember()
	before := engine.TouchWords(controls.touch[index])
	controls.touch[index].Name = name
	after := engine.TouchWords(controls.touch[index])
	for _, targets := range controls.actions {
		for place, word := range targets[engine.AndroidTarget] {
			if at := slices.Index(before, strings.TrimRight(word, "-+")); at >= 0 {
				targets[engine.AndroidTarget][place] = after[at] + strings.TrimPrefix(word, before[at])
			}
		}
	}
}
