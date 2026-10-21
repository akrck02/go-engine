package editor

import (
	"unicode/utf8"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Painting a device in lines, in the editor's colours: its outlines thin and grey, a control that
// does an action in the action's soft colour, and the action's name at the drawing's edge, joined
// to its control by a dotted line (under a key, whose neighbours leave no room for one).

// The drawings' measures, in pixels.
const (
	drawingLine      = 1.5  // an outline's width
	boundLine        = 2.5  // the outline of a control that does an action
	drawingMargin    = 12   // around a drawing, inside its room
	largestScale     = 1.4  // a drawing grows no larger, in pixels for each of its units
	keyLabelShare    = 0.3  // a key's label, as a share of its height
	tagSize          = 14   // an action's name at the drawing's edge
	keyTagSize       = 11   // an action's name under a key
	tagLineGap       = 6    // between a dotted line's end and the name it leads to
	leaderStep       = 6    // between a dotted line's dots
	leaderDot        = 1.5  // a dot's side
	letterWidthShare = 0.56 // a character's width, as a share of its text's size, to centre text
	fadedAlpha       = 0.3  // a control that does not do the action highlighted
	tagInkShare      = 0.45 // how much of the ink an action's colour takes, for its name to be read
)

// actionColors are the actions' colours, soft, in the order of their names, again from the first
// after the last.
var actionColors = []draw2d.Color{
	draw2d.Hex("#93b4f5"), draw2d.Hex("#fbb98a"), draw2d.Hex("#86dcb8"), draw2d.Hex("#d3aef7"),
	draw2d.Hex("#f5a3a3"), draw2d.Hex("#f2d27a"), draw2d.Hex("#8fd6e6"), draw2d.Hex("#f4a9cf"),
}

// actionColor is an action's colour.
func (controls *controlsView) actionColor(name string) draw2d.Color {
	index := max(0, indexOf(controls.names(), name))
	return actionColors[index%len(actionColors)]
}

// fitted is where a drawing's units go in an area: the scale, and where its corner lands.
func (device drawing) fitted(area draw2d.Rect) (scale, x, y float32) {
	room := draw2d.Rect{X: area.X + drawingMargin, Y: area.Y + drawingMargin,
		Width: area.Width - 2*drawingMargin, Height: area.Height - 2*drawingMargin}
	scale = max(min(room.Width/device.width, room.Height/device.height, largestScale), 0.01)
	x = room.X + (room.Width-device.width*scale)/2 - device.left*scale
	y = room.Y + (room.Height-device.height*scale)/2 - device.top*scale
	return scale, x, y
}

// placed is a rectangle of a drawing's units in the window.
func (device drawing) placed(area draw2d.Rect, x, y, width, height float32) draw2d.Rect {
	scale, originX, originY := device.fitted(area)
	return draw2d.Rect{X: originX + x*scale, Y: originY + y*scale, Width: width * scale, Height: height * scale}
}

// unitsAt is a point of the window in a drawing's units.
func (device drawing) unitsAt(area draw2d.Rect, x, y float32) (float32, float32) {
	scale, originX, originY := device.fitted(area)
	return (x - originX) / scale, (y - originY) / scale
}

// controlAt is the control of a drawing under a point of the window, if one is; controls with no
// words only draw, and are not found.
func (device drawing) controlAt(area draw2d.Rect, x, y float32) (int, bool) {
	for index, each := range device.controls {
		shape := device.placed(area, each.x, each.y, each.width, each.height)
		if len(each.words) > 0 && shape.Contains(x, y) {
			return index, true
		}
	}
	return 0, false
}

// paintDrawing paints a device in an area: its outline, then its controls with the actions they do.
func (view *View) paintDrawing(theme *ui.Theme, list *draw2d.List, area draw2d.Rect, device drawing, hovered int) {
	scale, originX, originY := device.fitted(area)
	list.Stroke(device.outline, originX, originY, scale, drawingLine, lineColor(theme))
	for index, each := range device.controls {
		view.paintControl(theme, list, area, device, each, index == hovered, index == view.controls.chosen)
	}
}

// lineColor is the drawings' grey.
func lineColor(theme *ui.Theme) draw2d.Color { return theme.InkQuiet.Mix(theme.Surface, 0.35) }

// paintControl paints a control: its outline, in its action's colour when it does one, ringed in
// the theme's main colour when chosen; its label; and the actions it does, named.
func (view *View) paintControl(theme *ui.Theme, list *draw2d.List, area draw2d.Rect, device drawing, each control, hovered, chosen bool) {
	controls := &view.controls
	scale, _, _ := device.fitted(area)
	shape := device.placed(area, each.x, each.y, each.width, each.height)
	first, _ := controls.firstActionOf(each)
	fill, border, width := draw2d.Transparent, draw2d.Transparent, float32(drawingLine)
	if each.live {
		border = lineColor(theme)
	}
	alpha := float32(1)
	if controls.highlighted != "" && !controls.doesAction(each, controls.highlighted) {
		alpha = fadedAlpha
	}
	if first != "" {
		color := controls.actionColor(first)
		fill, border, width = color.WithAlpha(0.22), color, boundLine
	}
	if hovered && len(each.words) > 0 {
		fill = theme.Sunken.WithAlpha(0.8)
	}
	if chosen {
		fill, border, width = theme.Primary.WithAlpha(0.14), theme.Primary, boundLine
	}
	list.Rectangle(shape, each.radius*scale, fill.WithAlpha(fill[3]*alpha), border.WithAlpha(border[3]*alpha), width)
	if each.stick && (each.live || border != draw2d.Transparent) {
		knob := shape.Width * stickKnobShare
		inner := draw2d.Rect{X: shape.X + (shape.Width-knob)/2, Y: shape.Y + (shape.Height-knob)/2, Width: knob, Height: knob}
		list.Rectangle(inner, knob/2, draw2d.Transparent, border.WithAlpha(border[3]*alpha), width)
	}
	labelSize, labelInk := min(max(shape.Height*0.42, 9), 15), theme.InkQuiet
	if each.tags == belowSide {
		labelSize, labelInk = min(max(shape.Height*keyLabelShare, 8), 12), lineColor(theme)
	}
	centredText(list, each.label, shape, labelSize, labelInk.WithAlpha(alpha), draw2d.Medium)
	view.paintTags(theme, list, area, device, each, shape, alpha)
}

// paintTags names the actions a control does: at the drawing's edge with a dotted line to it, or
// under it for a key.
func (view *View) paintTags(theme *ui.Theme, list *draw2d.List, area draw2d.Rect, device drawing, each control, shape draw2d.Rect, alpha float32) {
	controls := &view.controls
	row := 0
	for _, word := range each.words {
		name, label := controls.actionOn(word)
		if name == "" {
			continue
		}
		if each.stick {
			label += axisMark(word)
		}
		ink := controls.actionColor(name).Mix(theme.Ink, tagInkShare).WithAlpha(alpha)
		if each.tags == belowSide {
			below := draw2d.Rect{X: shape.X, Y: shape.Y + shape.Height + 2 + float32(row)*keyTagSize, Width: shape.Width, Height: keyTagSize}
			centredText(list, label, below, keyTagSize, ink, draw2d.Medium)
			row++
			continue
		}
		edge := device.placed(area, device.tagsLeft, 0, 0, 0).X
		if each.tags == rightSide {
			edge = device.placed(area, device.tagsRight, 0, 0, 0).X
		}
		centreY := shape.Y + shape.Height/2 + float32(row)*(tagSize+4)
		paintLeader(list, shape, edge, centreY, each.tags, ink.WithAlpha(0.5*alpha))
		textX := edge + tagLineGap
		if each.tags == leftSide {
			textX = edge - tagLineGap - textWidth(label, tagSize)
		}
		list.Text(label, textX, centreY+tagSize*0.35, tagSize, draw2d.Medium, ink)
		row++
	}
}

// axisMark is how a stick's axis is shown after its action's name: across or up and down.
func axisMark(word string) string {
	if len(word) > 0 && word[len(word)-1] == 'Y' {
		return " ↕"
	}
	return " ↔"
}

// paintLeader paints a dotted line from a control's side to the drawing's edge.
func paintLeader(list *draw2d.List, shape draw2d.Rect, edge, y float32, where side, ink draw2d.Color) {
	from, to := shape.X+shape.Width+tagLineGap, edge
	if where == leftSide {
		from, to = edge, shape.X-tagLineGap
	}
	for x := from; x < to; x += leaderStep {
		dot := draw2d.Rect{X: x, Y: y - leaderDot/2, Width: leaderDot, Height: leaderDot}
		list.Rectangle(dot, leaderDot/2, ink, draw2d.Transparent, 0)
	}
}

// textWidth is a text's width guessed from its characters.
func textWidth(text string, size float32) float32 {
	return float32(utf8.RuneCountInString(text)) * size * letterWidthShare
}

// centredText writes a line in the middle of an area.
func centredText(list *draw2d.List, text string, area draw2d.Rect, size float32, color draw2d.Color, weight draw2d.Weight) {
	if text == "" {
		return
	}
	list.Text(text, area.X+(area.Width-textWidth(text, size))/2, area.Y+area.Height/2+size*0.35, size, weight, color)
}

// firstActionOf is the first action a control does on the device shown, and whether it does any.
func (controls *controlsView) firstActionOf(each control) (string, bool) {
	for _, word := range each.words {
		if name, _ := controls.actionOn(word); name != "" {
			return name, true
		}
	}
	return "", false
}

// doesAction tells whether a control does an action on the device shown.
func (controls *controlsView) doesAction(each control, name string) bool {
	for _, word := range each.words {
		if bound, _ := controls.actionOn(word); bound == name {
			return true
		}
	}
	return false
}
