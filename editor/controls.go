package editor

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace's work, without its interface: the game's actions as they are edited
// (engine.json's actions), which of them move an axis, the phone's touch controls, the device
// shown, the control whose actions are being chosen, and catching the next input pressed. Every
// change is kept in engine.json at once, and can be undone.

// controlsView is what the Controls workspace remembers between frames.
type controlsView struct {
	actions map[string]map[string][]string // each action's words per target
	axes    map[string]bool                // the actions that move an axis (-1 to 1), not press
	touch   []engine.TouchControl          // the phone's touch controls
	loaded  bool                           // the above were read from the project's settings

	tab           int    // the device shown, by its place in deviceTabs
	chosen        int    // the control whose actions are chosen, by its place in the drawing; -1 for none
	query         string // what the chosen control's panel looks for, or names a new action
	renaming      string // the action whose name is being typed
	draft         string // its new name
	renameFocused bool   // the field of the new name had the keyboard last frame
	highlighted   string // the action whose controls stand out; the rest fade
	catching      bool   // the next input pressed is added to the highlighted action

	design   touchDesign // the phone's layout, being designed
	drawings [4]drawing  // each tab's device, made the first time it is shown
	area     draw2d.Rect // where the drawing was painted last frame
	undo     []controlsSnapshot
	redo     []controlsSnapshot
}

// deviceTab is a device the workspace shows, and the targets its bindings are kept for: a pad and
// a keyboard play the same game on a computer and in a browser.
type deviceTab struct {
	label, icon string
	targets     []string
}

// deviceTabs are the devices, in the order the tabs show them.
var deviceTabs = []deviceTab{
	{"Gamepad", iconGamepad, []string{engine.DesktopTarget, engine.WebTarget}},
	{"Keyboard and mouse", iconKeyboard, []string{engine.DesktopTarget, engine.WebTarget}},
	{"Nintendo 3DS", iconConsole, []string{engine.Nintendo3DSTarget}},
	{"Mobile", iconPhone, []string{engine.AndroidTarget}},
}

// mobileTab is the phone's tab, the only one with a layout to design.
const mobileTab = 3

// noControl is no control chosen.
const noControl = -1

// load fills the workspace from the project's settings, the first time it is shown.
func (controls *controlsView) load(settings engine.Settings) {
	if controls.loaded {
		return
	}
	controls.actions, controls.axes = map[string]map[string][]string{}, map[string]bool{}
	for name, targets := range settings.Actions {
		controls.actions[name] = map[string][]string{}
		for target, words := range targets {
			controls.actions[name][target] = slices.Clone(words)
		}
		controls.axes[name] = movesAnAxis(targets)
	}
	controls.touch = slices.Clone(settings.TouchControls)
	controls.chosen, controls.design.selected, controls.design.dragging = noControl, noControl, noControl
	controls.loaded = true
}

// movesAnAxis tells whether an action's inputs move an axis: a stick's, or a button's one way.
func movesAnAxis(targets map[string][]string) bool {
	for _, words := range targets {
		for _, word := range words {
			if isAxisWord(word) || strings.HasSuffix(word, "-") || strings.HasSuffix(word, "+") {
				return true
			}
		}
	}
	return false
}

// isAxisWord tells whether a word is an axis of a stick, a pad's or a screen's.
func isAxisWord(word string) bool {
	return strings.HasSuffix(word, " X") || strings.HasSuffix(word, " Y")
}

// names are the actions' names, in order.
func (controls *controlsView) names() []string {
	return slices.Sorted(maps.Keys(controls.actions))
}

// shownTab is the device shown.
func (controls *controlsView) shownTab() deviceTab { return deviceTabs[controls.tab] }

// drawingShown is the device shown, drawn. Each drawing is made once and kept, so its outline is
// drawn into a picture once; the phone's touch controls are placed on it each time.
func (controls *controlsView) drawingShown() drawing {
	kept := &controls.drawings[controls.tab]
	if kept.outline == nil {
		*kept = [...]func() drawing{padDrawing, keyboardDrawing, consoleDrawing, phoneDrawing}[controls.tab]()
	}
	if controls.tab == mobileTab {
		return withTouchControls(*kept, controls.touch)
	}
	return *kept
}

// showTab shows a device, with no control chosen.
func (controls *controlsView) showTab(tab int) {
	controls.tab, controls.chosen, controls.query, controls.catching = tab, noControl, "", false
	controls.design = touchDesign{selected: noControl, dragging: noControl}
}

// actionOn is the action a word does on the device shown, and how it is named beside its
// control ("turn −" for a button that moves an axis one way); empty for none.
func (controls *controlsView) actionOn(word string) (name, label string) {
	target := controls.shownTab().targets[0]
	for _, each := range controls.names() {
		for _, bound := range controls.actions[each][target] {
			if strings.TrimRight(bound, "-+") == word && (bound == word || !isAxisWord(word)) {
				return each, strings.TrimSpace(each + " " + signOf(bound))
			}
		}
	}
	return "", ""
}

// signOf is the way a word moves an axis, as a sign; empty for a press.
func signOf(word string) string {
	switch {
	case strings.HasSuffix(word, "-"):
		return "−"
	case strings.HasSuffix(word, "+"):
		return "+"
	}
	return ""
}

// assign makes a word do one action on the device shown, or none for an empty name: whatever it
// did before, on each of the device's targets, it stops doing.
func (controls *controlsView) assign(word, name string) {
	controls.remember()
	for _, target := range controls.shownTab().targets {
		for each, targets := range controls.actions {
			targets[target] = slices.DeleteFunc(targets[target], func(bound string) bool {
				return strings.TrimRight(bound, "-+") == word && (bound == word || !isAxisWord(word))
			})
			if each == name {
				targets[target] = append(targets[target], word)
			}
		}
	}
	controls.query = ""
}

// addAction adds an action with no inputs yet; it tells whether it was added (a name that is
// empty or taken is not).
func (controls *controlsView) addAction(name string, axis bool) bool {
	name = strings.TrimSpace(name)
	if name == "" || controls.actions[name] != nil {
		return false
	}
	controls.remember()
	controls.actions[name], controls.axes[name] = map[string][]string{}, axis
	return true
}

// newActionName is a name no action has yet: "new action", "new action 2"…
func (controls *controlsView) newActionName() string {
	name := "new action"
	for number := 2; controls.actions[name] != nil; number++ {
		name = "new action " + strconv.Itoa(number)
	}
	return name
}

// renameAction gives an action another name, keeping its inputs; a name that is empty or taken
// leaves it as it was.
func (controls *controlsView) renameAction(from, to string) {
	to = strings.TrimSpace(to)
	if to == from || to == "" || controls.actions[to] != nil || controls.actions[from] == nil {
		return
	}
	controls.remember()
	controls.actions[to], controls.axes[to] = controls.actions[from], controls.axes[from]
	delete(controls.actions, from)
	delete(controls.axes, from)
	if controls.highlighted == from {
		controls.highlighted = to
	}
}

// switchKind makes an action move an axis instead of pressing, or the other way; the inputs it
// had no longer fit, and go.
func (controls *controlsView) switchKind(name string) {
	controls.remember()
	controls.axes[name] = !controls.axes[name]
	controls.actions[name] = map[string][]string{}
}

// removeAction removes an action from every target.
func (controls *controlsView) removeAction(name string) {
	controls.remember()
	delete(controls.actions, name)
	delete(controls.axes, name)
	if controls.highlighted == name {
		controls.highlighted = ""
	}
}

// addInput binds the highlighted action to a word on the device shown, once.
func (controls *controlsView) addInput(word string) {
	inputs := controls.actions[controls.highlighted]
	if inputs == nil {
		return
	}
	controls.remember()
	for _, target := range controls.shownTab().targets {
		if !slices.Contains(inputs[target], word) {
			inputs[target] = append(inputs[target], word)
		}
	}
}

// catchInput adds the input pressed this frame to the highlighted action, while the workspace
// waits for one; Escape stops waiting. It tells whether the frame's input was taken.
func (view *View) catchInput(state *input.State) bool {
	controls := &view.controls
	if !controls.catching || view.Editor.ShownWorkspace() != ControlsWorkspace {
		controls.catching = false
		return false
	}
	if state.Pressed(input.KeyEscape) {
		controls.catching = false
		return true
	}
	if word, found := engine.PressedInput(state, controls.shownTab().targets[0]); found {
		controls.addInput(word)
		controls.catching = false
		view.saveControls()
	}
	return true
}
