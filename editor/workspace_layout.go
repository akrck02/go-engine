package editor

import (
	"github.com/akrck02/go-renderer/draw2d"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The phone's layout designer, in the Controls workspace's Mobile tab: Actions gives the controls
// on the screen their actions, as on any device; Layout adds sticks and buttons, drags them where
// the thumbs rest, resizes, renames and removes them.

// The designer's measures, in pixels.
const (
	designRingAway = 4 // between a control and the ring that shows it chosen
)

// layoutBar switches between giving actions and designing the layout, and holds the designer's
// tools.
func (view *View) layoutBar(context *ui.Context) {
	controls := &view.controls
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
		if context.Selectable(iconActions, "Actions", !controls.design.on) {
			controls.design.on = false
		}
		if context.Selectable(iconLayout, "Layout", controls.design.on) {
			controls.design.on, controls.chosen = true, noControl
		}
		if !controls.design.on {
			return
		}
		context.Separator()
		if len(controls.touch) == 0 {
			view.startLayoutButton(context)
			return
		}
		view.designTools(context)
	})
}

// startLayoutButton gives the game a layout of its own, in place of the engine's four buttons.
func (view *View) startLayoutButton(context *ui.Context) {
	note(context, "The game shows the engine's four buttons, which hold keys.")
	if context.PrimaryIconButton(iconAdd, "Design the game's own") {
		view.controls.startLayout()
		view.saveControls()
	}
}

// designTools add a stick or a button, and rename, resize or remove the control chosen.
func (view *View) designTools(context *ui.Context) {
	controls := &view.controls
	if context.IconButton(iconStick, "Stick") {
		controls.addTouch(engine.TouchStick)
		view.saveControls()
	}
	if context.IconButton(iconButton, "Button") {
		controls.addTouch(engine.TouchButton)
		view.saveControls()
	}
	context.Box(ui.Layout{Width: ui.Grow()}, nil)
	selected := controls.design.selected
	if selected < 0 || selected >= len(controls.touch) {
		context.Text("Drag the controls where the thumbs rest", ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
		return
	}
	if fieldEnded(context, "touch name", &controls.design.nameFocused) {
		controls.renameTouch()
		view.saveControls()
	}
	context.TextFieldWith("touch name", &controls.design.draft, ui.TextFieldOptions{Width: ui.Fixed(touchNameWidth)})
	view.sizeButtons(context)
}

// sizeButtons make the chosen control smaller or bigger, and remove it.
func (view *View) sizeButtons(context *ui.Context) {
	controls := &view.controls
	if context.FlatIconButton(iconSmaller, "") {
		controls.resizeTouch(-touchSizeStep)
		view.saveControls()
	}
	context.Tooltip("Smaller")
	if context.FlatIconButton(iconAdd, "") {
		controls.resizeTouch(touchSizeStep)
		view.saveControls()
	}
	context.Tooltip("Bigger")
	if context.FlatIconButton(iconDelete, "") {
		controls.removeTouch()
		view.saveControls()
	}
	context.Tooltip("Remove it")
}

// touchOffset is where the touch controls start among the phone drawing's controls: after the
// system's back.
const touchOffset = 1

// dragTouch follows the pointer: pressing on a touch control chooses it and starts dragging it,
// moving it while pressed, and keeps the layout when let go. Pressing elsewhere on the phone
// chooses none.
func (view *View) dragTouch(context *ui.Context, device drawing) {
	controls := &view.controls
	design := &controls.design
	pressing := context.Pressing()
	pointerX, pointerY := context.Pointer()
	started := pressing && !design.wasPressing && controls.area.Contains(pointerX, pointerY)
	design.wasPressing = pressing
	if started {
		index, found := device.controlAt(controls.area, pointerX, pointerY)
		design.dragging = noControl
		controls.chooseTouch(noControl)
		if found && index >= touchOffset {
			controls.remember()
			design.dragging = index - touchOffset
			controls.chooseTouch(design.dragging)
		}
	}
	if design.dragging < 0 {
		return
	}
	if !pressing {
		design.dragging = noControl
		view.saveControls()
		return
	}
	x, y := device.unitsAt(controls.area, pointerX, pointerY)
	controls.moveTouch(design.dragging, (x-phoneScreenX)/phoneScreenWidth, (y-phoneScreenY)/phoneScreenHeight)
}

// paintDesign rings the touch control chosen.
func (view *View) paintDesign(theme *ui.Theme, list *draw2d.List, area draw2d.Rect, device drawing) {
	index := view.controls.design.selected + touchOffset
	if index < touchOffset || index >= len(device.controls) {
		return
	}
	each := device.controls[index]
	shape := device.placed(area, each.x, each.y, each.width, each.height)
	ring := draw2d.Rect{X: shape.X - designRingAway, Y: shape.Y - designRingAway,
		Width: shape.Width + 2*designRingAway, Height: shape.Height + 2*designRingAway}
	list.Rectangle(ring, ring.Width/2, theme.Primary.WithAlpha(0.1), theme.Primary, drawingLine)
}
