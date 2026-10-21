package editor

import (
	"slices"
	"strings"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace's work, without its interface: the game's actions as they are edited
// (engine.json's actions), the target shown, the action that palette clicks and caught inputs go
// to, and catching the next input pressed.

// controlsView is what the Controls workspace remembers between frames.
type controlsView struct {
	actions   map[string]map[string][]string // being edited: each action's words per target
	loaded    bool                           // actions was filled from the project's settings
	changed   bool                           // the actions differ from the ones saved
	target    int                            // the target shown, by its place in engine.InputTargets
	chosen    string                         // the action inputs are added to
	catching  bool                           // the next input pressed is added to chosen
	newAction string
}

// shownTarget is the target the workspace shows.
func (controls *controlsView) shownTarget() string { return engine.InputTargets[controls.target] }

// load fills the actions from the project's settings, the first time they are shown.
func (controls *controlsView) load(saved map[string]map[string][]string) {
	if controls.loaded {
		return
	}
	controls.actions = map[string]map[string][]string{}
	for name, targets := range saved {
		controls.actions[name] = map[string][]string{}
		for target, words := range targets {
			controls.actions[name][target] = slices.Clone(words)
		}
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

// removeInput unbinds an action from one of its words on the target shown.
func (controls *controlsView) removeInput(name string, index int) {
	target := controls.shownTarget()
	controls.actions[name][target] = slices.Delete(controls.actions[name][target], index, index+1)
	controls.changed = true
}

// turnInput changes how a button moves an action: pressing it, then towards -1, then towards 1.
func (controls *controlsView) turnInput(name string, index int) {
	words := controls.actions[name][controls.shownTarget()]
	word := words[index]
	switch {
	case strings.HasSuffix(word, "-"):
		words[index] = strings.TrimSuffix(word, "-") + "+"
	case strings.HasSuffix(word, "+"):
		words[index] = strings.TrimSuffix(word, "+")
	default:
		words[index] = word + "-"
	}
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
