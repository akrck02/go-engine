package editor

import (
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The Controls workspace: the game's actions on the left, each a card in its colour, and in the
// middle the device chosen in its tabs, drawn in lines. A control of the device names the actions
// it does at the drawing's edge; clicking it opens a panel beside it to choose one, or to create
// one. On a phone the controls on the screen are the game's own, designed here too.

// The Controls workspace's measures, in pixels.
const (
	actionsWidth   = 300
	panelWidth     = 280
	panelAway      = 14  // between a control and its panel
	panelAbove     = 40  // how far the panel's top is above the control's middle
	panelTallest   = 420 // the room a panel is given to grow in
	cardRadius     = 10
	cardLine       = 1.5
	cardLineHover  = 2.5
	choiceDotSize  = 10
	touchNameWidth = 120
)

func (view *View) controlsWorkspace(context *ui.Context) {
	view.controls.load(view.Editor.Project.Settings)
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Actions", ui.Fixed(actionsWidth), func() { view.actionsColumn(context) })
		context.Separator()
		panel(context, "Device", ui.Grow(), func() { view.deviceColumn(context) })
	})
}

// actionsColumn is every action as a card, and a button to add one.
func (view *View) actionsColumn(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Heading("Actions") })
		if context.FlatIconButton(iconAdd, "") {
			name := controls.newActionName()
			controls.addAction(name, false)
			controls.renaming, controls.draft = name, name
			context.FocusField("rename")
			view.saveControls()
		}
		context.Tooltip("Add an action")
	})
	note(context, "What the player does, by name (Input.Action in the code). Click a control of the device to give it one.")
	names := controls.names()
	if len(names) == 0 {
		note(context, "No actions yet: add one with +.")
	}
	device := controls.drawingShown()
	for _, name := range names {
		view.actionCard(context, device, name)
	}
}

// actionCard is an action: its name, buttons to rename it, to switch its kind and to remove it,
// the hint a player of the device shown reads, and the controls of the device shown that do it. Clicking it makes its controls stand out.
func (view *View) actionCard(context *ui.Context, device drawing, name string) {
	controls := &view.controls
	context.Push(name)
	defer context.Pop()
	color := controls.actionColor(name)
	normal := ui.Look{Fill: context.Theme.Surface, Border: color, BorderWidth: cardLine, Radius: cardRadius}
	hovered := normal
	hovered.BorderWidth = cardLineHover
	if controls.highlighted == name {
		normal = hovered
	}
	looks := ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(12, 10), Gap: 6}
	if context.Clickable("card", layout, looks, func() {
		view.actionTitle(context, name)
		actionHint(context, controls.hintOf(name))
		view.actionChips(context, device, name)
	}) {
		controls.highlighted = toggled(controls.highlighted, name)
	}
}

// actionHint is what a player of the device shown reads for an action (Input.ActionHint), when
// the device has an input for it.
func actionHint(context *ui.Context, hint string) {
	if hint == "" {
		return
	}
	theme := context.Theme
	context.Text("Hint: "+hint, ui.TextStyle{Color: theme.InkSoft, Size: theme.SmallTextSize})
}

// toggled is a name, or none when it was the one already.
func toggled(current, name string) string {
	if current == name {
		return ""
	}
	return name
}

// actionTitle is an action's name, or the field to rename it, and its buttons.
func (view *View) actionTitle(context *ui.Context, name string) {
	controls := &view.controls
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 2, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			if controls.renaming != name {
				context.Text(name, ui.TextStyle{Weight: weightFor(true)})
				return
			}
			if fieldEnded(context, "rename", &controls.renameFocused) {
				controls.renameAction(name, controls.draft)
				controls.renaming = ""
				view.saveControls()
			}
			context.TextFieldWith("rename", &controls.draft, ui.TextFieldOptions{Placeholder: "Name it"})
		})
		view.actionButtons(context, name)
	})
}

// fieldEnded tells whether a text field stopped having the keyboard since the last frame (Enter,
// or a click elsewhere): an edit done. Call it before the field.
func fieldEnded(context *ui.Context, name string, wasFocused *bool) bool {
	focused := context.FieldFocused(name)
	ended := *wasFocused && !focused
	*wasFocused = focused
	return ended
}

// actionButtons rename an action, switch its kind, catch an input for it and remove it.
func (view *View) actionButtons(context *ui.Context, name string) {
	controls := &view.controls
	if context.FlatIconButton(iconRename, "") {
		controls.renaming, controls.draft = name, name
		context.FocusField("rename")
	}
	context.Tooltip("Rename")
	kindIcon, kindTip := iconPress, "Presses: click to make it move an axis"
	if controls.axes[name] {
		kindIcon, kindTip = iconAxis, "Moves an axis: click to make it press"
	}
	if context.FlatIconButton(kindIcon, "") {
		controls.switchKind(name)
		view.saveControls()
	}
	context.Tooltip(kindTip)
	if controls.tab < 2 && context.FlatIconButton(iconCatch, "") {
		controls.highlighted, controls.catching = name, true
	}
	if controls.tab < 2 {
		context.Tooltip("Press a key or a pad's button to give it this action")
	}
	if context.FlatIconButton(iconDelete, "") {
		controls.removeAction(name)
		view.saveControls()
	}
	context.Tooltip("Remove " + name)
}

// actionChips are the controls of the device shown that do an action, each opening its panel.
func (view *View) actionChips(context *ui.Context, device drawing, name string) {
	controls := &view.controls
	theme := context.Theme
	if controls.catching && controls.highlighted == name {
		context.Chip(iconCatch, "Press an input… (Escape stops)", ui.ChipLook{Fill: theme.Primary, Ink: theme.OnPrimary})
		context.RequestFrame(0) // the input is caught in the frame's step
	}
	found := false
	context.Box(ui.Layout{Width: ui.Grow(), Wrap: true, Gap: 4}, func() {
		for index, each := range device.controls {
			label, does := controls.chipFor(each, name)
			if !does {
				continue
			}
			found = true
			context.Push(strconv.Itoa(index))
			if context.Chip("", label, ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft}) {
				controls.chosen, controls.query, controls.design.on = index, "", false
			}
			context.Pop()
		}
	})
	if !found {
		context.Text("Not on this device yet", ui.TextStyle{Color: theme.InkQuiet, Size: theme.SmallTextSize})
	}
	for _, problem := range engine.CheckInputs(controls.actions[name][controls.shownTab().targets[0]], controls.shownTab().targets[0]) {
		wrapped(context, explained(problem), ui.TextStyle{Color: theme.Danger, Size: theme.SmallTextSize})
	}
}

// chipFor is how an action's card names a control that does it ("Left stick ↔"), and whether it
// does it.
func (controls *controlsView) chipFor(each control, name string) (string, bool) {
	var marks []string
	for _, word := range each.words {
		if bound, _ := controls.actionOn(word); bound == name {
			marks = append(marks, strings.TrimSpace(axisMarkOf(each, word)))
		}
	}
	if len(marks) == 0 {
		return "", false
	}
	return strings.TrimSpace(each.name + " " + strings.Join(marks, " ")), true
}

// axisMarkOf is a stick's axis mark, and nothing for a button.
func axisMarkOf(each control, word string) string {
	if !each.stick {
		return ""
	}
	return axisMark(word)
}

// deviceColumn is the devices' tabs, the phone's layout bar, and the device shown.
func (view *View) deviceColumn(context *ui.Context) {
	controls := &view.controls
	context.Row(func() {
		for index, tab := range deviceTabs {
			if context.Selectable(tab.icon, tab.label, controls.tab == index) {
				controls.showTab(index)
			}
		}
	})
	if controls.tab == mobileTab {
		view.layoutBar(context)
	}
	device := controls.drawingShown()
	view.drawingBox(context, device)
	if controls.chosen >= 0 && controls.chosen < len(device.controls) && !controls.design.on {
		view.controlPanel(context, device, device.controls[controls.chosen])
	}
}

// drawingBox is the device drawn as large as its room; a click chooses the control under the
// pointer, and while the phone's layout is designed a drag moves one.
func (view *View) drawingBox(context *ui.Context, device drawing) {
	controls := &view.controls
	pointerX, pointerY := context.Pointer()
	hovered, found := device.controlAt(controls.area, pointerX, pointerY)
	if !found {
		hovered = noControl
	}
	paint := func() {
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(room draw2d.Rect, list *draw2d.List) {
			controls.area = room
			view.paintDrawing(&context.Theme, list, room, device, hovered)
			if controls.design.on {
				view.paintDesign(&context.Theme, list, room, device)
			}
		})
	}
	if controls.design.on {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, paint)
		view.dragTouch(context, device)
		return
	}
	if context.Clickable("drawing", ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, ui.Looks{}, paint) {
		controls.query = ""
		controls.chosen = noControl
		if found && hovered != controls.chosen {
			controls.chosen = hovered
		}
	}
}

// controlPanel floats beside the chosen control: what it is, a field to find or name an action,
// and the actions it can do, one chosen.
func (view *View) controlPanel(context *ui.Context, device drawing, chosen control) {
	controls := &view.controls
	shape := device.placed(controls.area, chosen.x, chosen.y, chosen.width, chosen.height)
	x := shape.X + shape.Width + panelAway
	if x+panelWidth > controls.area.X+controls.area.Width {
		x = shape.X - panelAway - panelWidth
	}
	y := max(controls.area.Y, shape.Y+shape.Height/2-panelAbove)
	theme := context.Theme
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: cardRadius, Shadow: true}
	context.Within(draw2d.Rect{X: x, Y: y, Width: panelWidth, Height: panelTallest}, func() {
		context.Overlay("control", ui.LayerOptions{AlignX: ui.Start, AlignY: ui.Start, Width: ui.Fixed(panelWidth)}, func() {
			context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(12), Gap: 6}, look, func() {
				view.controlPanelContent(context, chosen)
			})
		})
	})
}

// controlPanelContent is the chosen control's panel: a button's actions, or a stick's two axes'.
func (view *View) controlPanelContent(context *ui.Context, chosen control) {
	controls := &view.controls
	context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(chosen.name, ui.TextStyle{Weight: weightFor(true)}) })
		if context.FlatIconButton(iconClose, "") {
			controls.chosen = noControl
		}
	})
	note(context, chosen.about)
	context.TextFieldWith("find", &controls.query, ui.TextFieldOptions{Placeholder: "Find or create an action"})
	if chosen.stick {
		context.Caption("Left–right")
		view.choices(context, chosen.words[0], true)
		context.Caption("Up–down")
		view.choices(context, chosen.words[1], true)
	} else {
		view.choices(context, chosen.words[0], false)
	}
	view.createChoice(context, chosen)
}

// choices are the actions a word can do, of one kind, that match what is looked for, and None.
func (view *View) choices(context *ui.Context, word string, axis bool) {
	controls := &view.controls
	context.Push(word)
	defer context.Pop()
	current, _ := controls.actionOn(word)
	if view.choice(context, "None", draw2d.Transparent, current == "") {
		controls.assign(word, "")
		view.saveControls()
	}
	query := strings.ToLower(strings.TrimSpace(controls.query))
	for _, name := range controls.names() {
		if controls.axes[name] != axis || !strings.Contains(strings.ToLower(name), query) {
			continue
		}
		if view.choice(context, name, controls.actionColor(name), current == name) {
			controls.assign(word, name)
			view.saveControls()
		}
	}
}

// choice is a row of the panel: an action's colour, its name, a tick when chosen.
func (view *View) choice(context *ui.Context, name string, color draw2d.Color, chosen bool) bool {
	theme := context.Theme
	normal := ui.Look{Radius: 6}
	if chosen {
		normal.Fill = theme.Sunken
	}
	looks := ui.Looks{Normal: normal, Hovered: ui.Look{Radius: 6, Fill: theme.Sunken}, Pressed: ui.Look{Radius: 6, Fill: theme.Line}}
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(8, 5), Gap: 8, AlignY: ui.Center}
	return context.Clickable(name, layout, looks, func() {
		context.Paint(ui.Layout{Width: ui.Fixed(choiceDotSize), Height: ui.Fixed(choiceDotSize)}, func(area draw2d.Rect, list *draw2d.List) {
			list.Rectangle(area, choiceDotSize/2, color, theme.ControlLine, 1)
		})
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Label(name) })
		if chosen {
			context.IconSized(iconCheck, theme.IconSize, theme.Primary)
		}
	})
}

// createChoice creates the action typed in the panel's field, of the control's kind, and gives
// it to the control; a stick gets it across.
func (view *View) createChoice(context *ui.Context, chosen control) {
	controls := &view.controls
	name := strings.TrimSpace(controls.query)
	if name == "" || controls.actions[name] != nil {
		return
	}
	if context.IconButton(iconAdd, "Create “"+name+"” and give it to "+chosen.name) {
		controls.addAction(name, chosen.stick)
		controls.assign(chosen.words[0], name)
		view.saveControls()
	}
}
