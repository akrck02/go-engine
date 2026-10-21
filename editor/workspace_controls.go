package editor

import (
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace: the game's actions on the left, and in the middle the device of the
// target chosen, drawn with its buttons where they sit (a 3DS, a phone, a computer's keyboard,
// mouse and pad). Each button says which actions it does; clicking it chooses them.

// The Controls workspace's measures, in pixels.
const actionsWidth = 300

// targetIcons are each target's icon on its tab.
var targetIcons = map[string]string{
	engine.DesktopTarget: iconComputer, engine.Nintendo3DSTarget: iconConsole,
	engine.AndroidTarget: iconPhone, engine.WebTarget: iconWeb,
}

func (view *View) controlsWorkspace(context *ui.Context) {
	view.controls.load(view.Editor.Project.Settings.Actions)
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Actions", ui.Fixed(actionsWidth), func() { view.actionsColumn(context) })
		context.Separator()
		panel(context, "Device", ui.Grow(), func() { view.deviceColumn(context) })
	})
}

// actionsColumn is every action with its kind, a field for a new one, and saving.
func (view *View) actionsColumn(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		context.IconSized(iconControls, cardIconSize, context.Theme.Primary)
		context.Heading("Actions")
	})
	note(context, "What the player does, by name (Input.Action in the code). Click a button of the device to choose what it does.")
	for _, name := range controls.names() {
		view.actionItem(context, name)
	}
	context.Separator()
	context.Row(func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.TextField("New action", &controls.newAction) })
		if context.IconButton(iconAdd, "") {
			controls.addAction(controls.newAction)
			controls.newAction = ""
		}
		context.Tooltip("Add the action named")
	})
	view.saveControlsButton(context)
}

// actionItem is one action: whether it presses or moves an axis, catching an input for it on a
// computer, and removing it.
func (view *View) actionItem(context *ui.Context, name string) {
	controls := &view.controls
	context.Push(name)
	defer context.Pop()
	context.Separator()
	context.Row(func() {
		context.IconSized(actionKindIcon(controls.axes[name]), stateIconSize, context.Theme.Primary)
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(name, ui.TextStyle{Weight: weightFor(true)}) })
		if context.FlatIconButton(iconCatch, "") {
			controls.chosen, controls.catching = name, true
		}
		context.Tooltip("Press a key or a pad's button to bind it (on a computer)")
		if context.FlatIconButton(iconDelete, "") {
			controls.removeAction(name)
		}
		context.Tooltip("Remove " + name)
	})
	kind := 0
	if controls.axes[name] {
		kind = 1
	}
	if context.Tabs("Kind", &kind, "Press", "Axis") {
		controls.axes[name] = kind == 1
	}
	if controls.catching && controls.chosen == name {
		context.Chip(iconCatch, "Press an input… (Escape stops)", ui.ChipLook{Fill: context.Theme.Primary, Ink: context.Theme.OnPrimary})
		context.RequestFrame(0) // the input is caught in the frame's step
	}
	target := controls.shownTarget()
	for _, problem := range engine.CheckInputs(controls.actions[name][target], target) {
		wrapped(context, explained(problem), ui.TextStyle{Color: context.Theme.Danger, Size: context.Theme.SmallTextSize})
	}
}

// actionKindIcon is an action's icon: an axis's, or a button's.
func actionKindIcon(axis bool) string {
	if axis {
		return iconAxis
	}
	return iconPress
}

// deviceColumn is the targets' tabs and the device of the target chosen.
func (view *View) deviceColumn(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		for index, target := range engine.InputTargets {
			if context.Selectable(targetIcons[target], target, controls.target == index) {
				controls.target, controls.catching, controls.opened = index, false, ""
			}
		}
	})
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), AlignX: ui.Center, Gap: deviceGap}, func() {
		switch controls.shownTarget() {
		case engine.Nintendo3DSTarget:
			view.consoleDevice(context)
		case engine.AndroidTarget:
			view.phoneDevice(context)
		default:
			view.computerDevices(context)
		}
	})
}

// saveControlsButton keeps the actions in engine.json, and says when they have changes not kept.
func (view *View) saveControlsButton(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		if context.PrimaryIconButton(iconSave, "Save controls") {
			if err := view.Editor.SaveControls(controls.actions); err != nil {
				view.Editor.say(err.Error(), true)
				return
			}
			controls.changed = false
			view.Editor.say("Saved the controls in engine.json", false)
		}
		if controls.changed {
			context.Text("Changes not saved", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
		}
	})
}

// actionIcon is an action's icon: an axis's when any of its inputs moves it, a button's otherwise.
func actionIcon(targets map[string][]string) string {
	for _, words := range targets {
		for _, word := range words {
			if isStick(word) || strings.HasSuffix(word, "-") || strings.HasSuffix(word, "+") {
				return iconAxis
			}
		}
	}
	return iconPress
}

// isStick tells whether a word is a stick's axis, which moves an action and cannot press it.
func isStick(word string) bool {
	return strings.HasPrefix(word, "Stick ") || strings.HasPrefix(word, "Right Stick ")
}
