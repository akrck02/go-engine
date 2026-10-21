package editor

import (
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace: the game's actions, one row each, with the inputs that press them on the
// target chosen as chips, and the target's inputs on the right to add with a click. On a computer
// an input can also be caught by pressing it.

// The Controls workspace's measures, in pixels.
const (
	paletteWidth   = 380
	actionRowGap   = 8
	actionRowRound = 10
)

// targetIcons are each target's icon on its tab, in the order of engine.InputTargets.
var targetIcons = map[string]string{
	engine.DesktopTarget: iconComputer, engine.Nintendo3DSTarget: iconConsole,
	engine.AndroidTarget: iconPhone, engine.WebTarget: iconWeb,
}

func (view *View) controlsWorkspace(context *ui.Context) {
	view.controls.load(view.Editor.Project.Settings.Actions)
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Actions", ui.Grow(), func() { view.actionsColumn(context) })
		context.Separator()
		panel(context, "Inputs", ui.Fixed(paletteWidth), func() { view.inputPalette(context) })
	})
}

// actionsColumn is the targets' tabs, every action, a field for a new one, and saving.
func (view *View) actionsColumn(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		context.IconSized(iconControls, cardIconSize, context.Theme.Primary)
		context.Heading("Controls")
	})
	note(context, "What the player does, by name (Input.Action in the code), and the inputs that do it on each target.")
	context.Row(func() {
		for index, target := range engine.InputTargets {
			if context.Selectable(targetIcons[target], target, controls.target == index) {
				controls.target, controls.catching = index, false
			}
		}
	})
	for _, name := range controls.names() {
		view.actionRow(context, name)
	}
	context.Row(func() {
		context.Box(ui.Layout{Width: ui.Fixed(220)}, func() { context.TextField("New action", &controls.newAction) })
		if context.IconButton(iconAdd, "New action") {
			controls.addAction(controls.newAction)
			controls.newAction = ""
		}
	})
	view.saveControlsButton(context)
}

// actionRow is one action: its name (chosen with a click), its inputs on the target shown, and a
// way to catch one.
func (view *View) actionRow(context *ui.Context, name string) {
	controls := &view.controls
	look := ui.Look{Fill: context.Theme.Background, Radius: actionRowRound, Border: context.Theme.Line, BorderWidth: 1}
	if controls.chosen == name {
		look.Border, look.BorderWidth = context.Theme.Primary, 2
	}
	context.Push(name)
	defer context.Pop()
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(10), Gap: actionRowGap}, look, func() {
		context.Row(func() {
			if context.Selectable(actionIcon(controls.actions[name]), name, controls.chosen == name) {
				controls.chosen = name
			}
			context.Tooltip("Inputs clicked on the right are added to the action chosen")
			if context.FlatIconButton(iconDelete, "") {
				controls.removeAction(name)
			}
			context.Tooltip("Remove " + name)
		})
		view.inputChips(context, name)
	})
}

// inputChips are an action's inputs on the target shown, and the button that catches one more.
func (view *View) inputChips(context *ui.Context, name string) {
	controls := &view.controls
	target := controls.shownTarget()
	words := controls.actions[name][target]
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true, AlignY: ui.Center}, func() {
		for index, word := range words {
			context.Push(word)
			clicked, removed := context.RemovableChip(inputIcon(word, target), inputLabel(word), ui.ChipLook{})
			context.Pop()
			switch {
			case removed:
				controls.removeInput(name, index)
				return // the words changed: the rest are shown next frame
			case clicked && !isStick(word):
				controls.turnInput(name, index)
			}
		}
		view.catchButton(context, name)
	})
	for _, problem := range engine.CheckInputs(words, target) {
		wrapped(context, explained(problem), ui.TextStyle{Color: context.Theme.Danger, Size: context.Theme.SmallTextSize})
	}
}

// catchButton waits for the next input pressed, for an action, where inputs can be pressed here.
func (view *View) catchButton(context *ui.Context, name string) {
	controls := &view.controls
	target := controls.shownTarget()
	if target != engine.DesktopTarget && target != engine.WebTarget {
		return // a 3DS's or a phone's own buttons are not on this computer: the palette has them
	}
	if controls.catching && controls.chosen == name {
		context.Chip(iconCatch, "Press a key or a pad's button… (Escape stops)", ui.ChipLook{Fill: context.Theme.Primary, Ink: context.Theme.OnPrimary})
		context.RequestFrame(0) // the input is caught in the frame's step
		return
	}
	if context.FlatIconButton(iconCatch, "Press to add") {
		controls.chosen, controls.catching = name, true
	}
	context.Tooltip("Add an input by pressing it")
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

// inputLabel is a word as its chip shows it: a pad's or a mouse's button without the device, and
// the way a button moves an axis as a sign.
func inputLabel(word string) string {
	base, sign := word, ""
	if strings.HasSuffix(word, "-") || strings.HasSuffix(word, "+") {
		base, sign = word[:len(word)-1], " "+strings.ReplaceAll(word[len(word)-1:], "-", "−")
	}
	base = strings.TrimPrefix(strings.TrimPrefix(base, "Pad "), "Mouse ")
	return base + sign
}

// inputIcon is the icon of a word's group on a target.
func inputIcon(word, target string) string {
	base := strings.TrimRight(word, "-+")
	for _, choice := range engine.InputChoices(target) {
		if choice.Word == base {
			return groupIcon(choice.Group)
		}
	}
	return iconMissing
}

// groupIcon is an input group's icon.
func groupIcon(group string) string {
	switch group {
	case engine.MouseInputs:
		return iconMouse
	case engine.PadInputs:
		return iconGamePad
	case engine.StickInputs:
		return iconStick
	case engine.ButtonInputs:
		return iconConsole
	case engine.SystemInputs:
		return iconBack
	}
	return iconKeyboard
}
