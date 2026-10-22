package editor

import (
	"cmp"
	"math"
	"slices"
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
	draw2d.MustHex("#93b4f5"), draw2d.MustHex("#fbb98a"), draw2d.MustHex("#86dcb8"), draw2d.MustHex("#d3aef7"),
	draw2d.MustHex("#f5a3a3"), draw2d.MustHex("#f2d27a"), draw2d.MustHex("#8fd6e6"), draw2d.MustHex("#f4a9cf"),
}

// actionColor is an action's colour.
func (controls *controlsView) actionColor(name string) draw2d.Color {
	index := max(0, indexOf(controls.names(), name))
	return actionColors[index%len(actionColors)]
}

// fitted is where a drawing's units go in an area: the scale, and where its corner lands.
func (device drawing) fitted(area draw2d.Rectangle) (scale, x, y float32) {
	room := draw2d.Rectangle{X: area.X + drawingMargin, Y: area.Y + drawingMargin,
		Width: area.Width - 2*drawingMargin, Height: area.Height - 2*drawingMargin}
	scale = max(min(room.Width/device.width, room.Height/device.height, largestScale), 0.01)
	x = room.X + (room.Width-device.width*scale)/2 - device.left*scale
	y = room.Y + (room.Height-device.height*scale)/2 - device.top*scale
	return scale, x, y
}

// placed is a rectangle of a drawing's units in the window.
func (device drawing) placed(area draw2d.Rectangle, x, y, width, height float32) draw2d.Rectangle {
	scale, originX, originY := device.fitted(area)
	return draw2d.Rectangle{X: originX + x*scale, Y: originY + y*scale, Width: width * scale, Height: height * scale}
}

// unitsAt is a point of the window in a drawing's units.
func (device drawing) unitsAt(area draw2d.Rectangle, x, y float32) (float32, float32) {
	scale, originX, originY := device.fitted(area)
	return (x - originX) / scale, (y - originY) / scale
}

// controlAt is the control of a drawing under a point of the window, if one is; controls with no
// words only draw, and are not found.
func (device drawing) controlAt(area draw2d.Rectangle, x, y float32) (int, bool) {
	for index, each := range device.controls {
		shape := device.placed(area, each.x, each.y, each.width, each.height)
		if len(each.words) > 0 && shape.Contains(x, y) {
			return index, true
		}
	}
	return 0, false
}

// paintDrawing paints a device in an area: its outline, then its controls with the actions they do.
func (view *View) paintDrawing(theme *ui.Theme, list *draw2d.List, area draw2d.Rectangle, device drawing, hovered int) {
	scale, originX, originY := device.fitted(area)
	list.Stroke(device.outline, originX, originY, scale, drawingLine, lineColor(theme))
	controls := &view.controls
	controls.tags = controls.tags[:0]
	for index, each := range device.controls {
		view.paintControl(theme, list, area, device, each, index == hovered, index == controls.chosen)
	}
	spreadTags(controls.tags)
	for _, tag := range controls.tags {
		paintEdgeTag(list, tag)
	}
}

// lineColor is the drawings' grey.
func lineColor(theme *ui.Theme) draw2d.Color { return theme.InkQuiet.Mix(theme.Surface, 0.35) }

// paintControl paints a control: its outline, in its action's colour when it does one, ringed in
// the theme's main colour when chosen; its label; and the actions it does, named.
func (view *View) paintControl(theme *ui.Theme, list *draw2d.List, area draw2d.Rectangle, device drawing, each control, hovered, chosen bool) {
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
		inner := draw2d.Rectangle{X: shape.X + (shape.Width-knob)/2, Y: shape.Y + (shape.Height-knob)/2, Width: knob, Height: knob}
		list.Rectangle(inner, knob/2, draw2d.Transparent, border.WithAlpha(border[3]*alpha), width)
	}
	labelSize, labelInk := min(max(shape.Height*0.42, 9), 15), theme.InkQuiet
	if each.tags == belowSide {
		labelSize, labelInk = min(max(shape.Height*keyLabelShare, 8), 12), lineColor(theme)
	}
	centredText(list, each.label, shape, labelSize, labelInk.WithAlpha(alpha), draw2d.Medium)
	view.paintTags(theme, list, area, device, each, shape, alpha)
}

// edgeTag is an action's name at a drawing's edge: its text and ink, where its control's side is,
// and the height it is written at, moved down from its control's when another name was there.
type edgeTag struct {
	label        string
	ink          draw2d.Color
	where        side
	fromX, fromY float32 // the control's side, where the dotted line starts
	edge, y      float32 // where the name is written
}

// paintTags names the actions a control does: under it for a key, written at once; at the
// drawing's edge for the rest, gathered for paintDrawing to spread and write.
func (view *View) paintTags(theme *ui.Theme, list *draw2d.List, area draw2d.Rectangle, device drawing, each control, shape draw2d.Rectangle, alpha float32) {
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
			below := draw2d.Rectangle{X: shape.X, Y: shape.Y + shape.Height + 2 + float32(row)*keyTagSize, Width: shape.Width, Height: keyTagSize}
			width := textWidth(label, keyTagSize) + 6
			backing := draw2d.Rectangle{X: below.X + (below.Width-width)/2, Y: below.Y, Width: width, Height: below.Height}
			list.Rectangle(backing, keyTagSize/2, theme.Surface, draw2d.Transparent, 0) // over the key below, to be read
			centredText(list, label, below, keyTagSize, ink, draw2d.Medium)
			row++
			continue
		}
		tag := edgeTag{label: label, ink: ink, where: each.tags, fromX: shape.X + shape.Width + tagLineGap,
			fromY: shape.Y + shape.Height/2, edge: device.placed(area, device.tagsRight, 0, 0, 0).X}
		if each.tags == leftSide {
			tag.fromX, tag.edge = shape.X-tagLineGap, device.placed(area, device.tagsLeft, 0, 0, 0).X
		}
		tag.y = tag.fromY
		controls.tags = append(controls.tags, tag)
	}
}

// tagSpacing is the least height between two names on the same edge, in pixels.
const tagSpacing = tagSize + 4

// spreadTags moves names down where they would be written over one another on the same edge,
// keeping them in the order of their controls from top to bottom.
func spreadTags(tags []edgeTag) {
	slices.SortStableFunc(tags, func(first, second edgeTag) int {
		if first.where != second.where {
			return int(first.where) - int(second.where)
		}
		return cmp.Compare(first.fromY, second.fromY)
	})
	for index := 1; index < len(tags); index++ {
		above := tags[index-1]
		if above.where == tags[index].where {
			tags[index].y = max(tags[index].y, above.y+tagSpacing)
		}
	}
}

// paintEdgeTag writes a name at the drawing's edge, and a dotted line from its control to it.
func paintEdgeTag(list *draw2d.List, tag edgeTag) {
	paintLeader(list, tag.fromX, tag.fromY, tag.edge, tag.y, tag.ink.WithAlpha(0.5*tag.ink[3]))
	textX := tag.edge + tagLineGap
	if tag.where == leftSide {
		textX = tag.edge - tagLineGap - textWidth(tag.label, tagSize)
	}
	list.Text(tag.label, textX, tag.y+tagSize*0.35, tagSize, draw2d.Medium, tag.ink)
}

// axisMark is how a stick's axis is shown after its action's name: across or up and down.
func axisMark(word string) string {
	if len(word) > 0 && word[len(word)-1] == 'Y' {
		return " ↕"
	}
	return " ↔"
}

// paintLeader paints a dotted line between two points: from a control's side to its name.
func paintLeader(list *draw2d.List, fromX, fromY, toX, toY float32, ink draw2d.Color) {
	acrossX, downY := toX-fromX, toY-fromY
	length := float32(math.Hypot(float64(acrossX), float64(downY)))
	for along := float32(0); along < length; along += leaderStep {
		x, y := fromX+acrossX*along/length, fromY+downY*along/length
		dot := draw2d.Rectangle{X: x - leaderDot/2, Y: y - leaderDot/2, Width: leaderDot, Height: leaderDot}
		list.Rectangle(dot, leaderDot/2, ink, draw2d.Transparent, 0)
	}
}

// textWidth is a text's width guessed from its characters.
func textWidth(text string, size float32) float32 {
	return float32(utf8.RuneCountInString(text)) * size * letterWidthShare
}

// centredText writes a line in the middle of an area.
func centredText(list *draw2d.List, text string, area draw2d.Rectangle, size float32, color draw2d.Color, weight draw2d.Weight) {
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
