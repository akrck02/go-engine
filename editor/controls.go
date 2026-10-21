package editor

import (
	"slices"
	"strings"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace's work, without its interface: the game's actions as they are edited
// (engine.json's actions), which of them move an axis, the target whose device is shown, the
// button whose actions are being chosen, and catching the next input pressed.

// controlsView is what the Controls workspace remembers between frames.
type controlsView struct {
	actions   map[string]map[string][]string // being edited: each action's words per target
	loaded    bool                           // actions was filled from the project's settings
	changed   bool                           // the actions differ from the ones saved
	axes      map[string]bool                // the actions that move an axis (-1 to 1), not press
	target    int                            // the target shown, by its place in engine.InputTargets
	chosen    string                         // the action a caught input goes to
	catching  bool                           // the next input pressed is added to chosen
	opened    string                         // the button whose actions are being chosen
	newAction string
}

// shownTarget is the target the workspace shows.
func (controls *controlsView) shownTarget() string { return engine.InputTargets[controls.target] }

// load fills the actions from the project's settings, the first time they are shown.
func (controls *controlsView) load(saved map[string]map[string][]string) {
	if controls.loaded {
		return
	}
	controls.actions, controls.axes = map[string]map[string][]string{}, map[string]bool{}
	for name, targets := range saved {
		controls.actions[name] = map[string][]string{}
		for target, words := range targets {
			controls.actions[name][target] = slices.Clone(words)
		}
		controls.axes[name] = actionIcon(targets) == iconAxis
	}
	controls.loaded = true
}

// names are the actions' names, in order.
func (controls *controlsView) names() []string {
	names := make([]string, 0, len(controls.actions))
	for name := range controls.actions {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// addAction adds an action with no inputs yet, and chooses it.
func (controls *controlsView) addAction(name string) {
	name = strings.TrimSpace(name)
	if name == "" || controls.actions[name] != nil {
		return
	}
	controls.actions[name] = map[string][]string{}
	controls.chosen, controls.changed = name, true
}

// bindingsOf are the bindings of a button on the target shown, as its drawing names them: the
// action, and for an axis the way it moves it ("turn −").
func (controls *controlsView) bindingsOf(word string) []string {
	var bound []string
	for _, name := range controls.names() {
		for _, each := range controls.actions[name][controls.shownTarget()] {
			if strings.TrimRight(each, "-+") == word {
				bound = append(bound, strings.TrimSpace(name+" "+signOf(each)))
			}
		}
	}
	return bound
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

// toggleBinding binds an action to a word on the target shown, or unbinds it when it is bound: a
// button pressing it ("A"), or moving it one way ("A-", "A+"). The button's other ways for the
// action go.
func (controls *controlsView) toggleBinding(name, word string) {
	target := controls.shownTarget()
	words := controls.actions[name][target]
	had := slices.Contains(words, word)
	words = slices.DeleteFunc(words, func(each string) bool { return strings.TrimRight(each, "-+") == strings.TrimRight(word, "-+") })
	if !had {
		words = append(words, word)
	}
	controls.actions[name][target] = words
	controls.changed = true
}

// removeAction removes an action from every target.
func (controls *controlsView) removeAction(name string) {
	delete(controls.actions, name)
	controls.changed = true
}

// addInput binds the chosen action to a word on the target shown, once.
func (controls *controlsView) addInput(word string) {
	inputs := controls.actions[controls.chosen]
	if inputs == nil || slices.Contains(inputs[controls.shownTarget()], word) {
		return
	}
	inputs[controls.shownTarget()] = append(inputs[controls.shownTarget()], word)
	controls.changed = true
}

// catchInput adds the input pressed this frame to the chosen action, while the workspace waits for
// one; Escape stops waiting. It tells whether the frame's input was taken.
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
	if word, found := engine.PressedInput(state, controls.shownTarget()); found {
		controls.addInput(word)
		controls.catching = false
	}
	return true
}

// SaveControls keeps the actions edited in the project's settings.
func (editor *Editor) SaveControls(actions map[string]map[string][]string) error {
	editor.Project.Settings.Actions = actions
	return editor.Project.SaveSettings()
}
