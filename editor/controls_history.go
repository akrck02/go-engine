package editor

import (
	"maps"
	"slices"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace keeps every change in engine.json as it is made, with no button to
// save, and keeps its own history: Ctrl+Z there undoes a change of the controls, not of the scene.

// controlsHistoryLength is how many changes of the controls can be undone.
const controlsHistoryLength = 100

// controlsSnapshot is the controls as they were before a change.
type controlsSnapshot struct {
	actions map[string]map[string][]string
	axes    map[string]bool
	touch   []engine.TouchControl
}

// snapshot is a copy of the controls as they are.
func (controls *controlsView) snapshot() controlsSnapshot {
	actions := make(map[string]map[string][]string, len(controls.actions))
	for name, targets := range controls.actions {
		actions[name] = make(map[string][]string, len(targets))
		for target, words := range targets {
			actions[name][target] = slices.Clone(words)
		}
	}
	return controlsSnapshot{actions: actions, axes: maps.Clone(controls.axes), touch: slices.Clone(controls.touch)}
}

// restoreSnapshot puts the controls back as a snapshot has them.
func (controls *controlsView) restoreSnapshot(kept controlsSnapshot) {
	controls.actions, controls.axes, controls.touch = kept.actions, kept.axes, kept.touch
	controls.chosen, controls.design.selected, controls.design.dragging = noControl, noControl, noControl
}

// remember keeps the controls as they are, before a change, to undo it.
func (controls *controlsView) remember() {
	controls.undo = append(controls.undo, controls.snapshot())
	if len(controls.undo) > controlsHistoryLength {
		controls.undo = controls.undo[1:]
	}
	controls.redo = nil
}

// undoChange takes the controls back one change; it tells whether there was one.
func (controls *controlsView) undoChange() bool {
	if len(controls.undo) == 0 {
		return false
	}
	controls.redo = append(controls.redo, controls.snapshot())
	controls.restoreSnapshot(controls.undo[len(controls.undo)-1])
	controls.undo = controls.undo[:len(controls.undo)-1]
	return true
}

// redoChange takes the controls forward one change, after an undo; it tells whether there was one.
func (controls *controlsView) redoChange() bool {
	if len(controls.redo) == 0 {
		return false
	}
	controls.undo = append(controls.undo, controls.snapshot())
	controls.restoreSnapshot(controls.redo[len(controls.redo)-1])
	controls.redo = controls.redo[:len(controls.redo)-1]
	return true
}

// saveControls keeps the controls in engine.json, saying so only when it fails.
func (view *View) saveControls() {
	controls := &view.controls
	if err := view.Editor.SaveControls(controls.actions, controls.axisNames(), controls.touch); err != nil {
		view.Editor.say(err.Error(), true)
	}
}

// SaveControls keeps the actions, which of them move an axis, and the touch controls in the
// project's settings.
func (editor *Editor) SaveControls(actions map[string]map[string][]string, axes []string, touch []engine.TouchControl) error {
	settings := &editor.Project.Settings
	settings.Actions, settings.Axes, settings.TouchControls = actions, axes, touch
	return editor.Project.SaveSettings()
}

// axisNames are the actions that move an axis, by name, in order.
func (controls *controlsView) axisNames() []string {
	var names []string
	for _, name := range controls.names() {
		if controls.axes[name] {
			names = append(names, name)
		}
	}
	return names
}

// undo takes back the last change of the workspace shown: the controls' in Controls, the scene's
// elsewhere.
func (view *View) undo() {
	if view.Editor.ShownWorkspace() != ControlsWorkspace {
		view.Editor.Undo()
		return
	}
	if !view.controls.undoChange() {
		view.Editor.say("Nothing to undo", false)
		return
	}
	view.saveControls()
}

// redo does again the last change undone in the workspace shown.
func (view *View) redo() {
	if view.Editor.ShownWorkspace() != ControlsWorkspace {
		view.Editor.Redo()
		return
	}
	if !view.controls.redoChange() {
		view.Editor.say("Nothing to redo", false)
		return
	}
	view.saveControls()
}
