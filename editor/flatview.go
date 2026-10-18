package editor

import (
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// The view of a 2D scene: flat, from the front, over a grid of one-metre cells with the X axis in
// red and the Y axis in green. Drag the background to move around, scroll to zoom; click a node to
// choose it, drag the chosen one to move it. Cameras show the frame they see.

// flatView is the 2D view's state.
type flatView struct {
	view     engine.View2D
	drawing  draw2d.List // the list being painted into, while paint runs
	area     draw2d.Rect // where the view was last painted
	pressing bool
	panning  bool
	dragging *engine.Node
	moved    bool
	pressX   float64
	pressY   float64
	grabbed  space.Vector // the dragged node's position minus the point pressed
	lastX    float64
	lastY    float64
	pointer  space.Vector // the point of the world under the pointer
}

func newFlatView() *flatView { return &flatView{view: engine.View2D{Height: 14}} }

// Colours of the 2D view's grid and marks.
var (
	flatBackground = draw2d.Hex("#3a3f4b")
	flatGrid       = draw2d.Color{0, 0, 0, 0.1}
	flatGridFive   = draw2d.Color{0, 0, 0, 0.2}
	flatAxisX      = draw2d.Color{0.9, 0.3, 0.3, 0.9}
	flatAxisY      = draw2d.Color{0.35, 0.8, 0.4, 0.9}
	flatChosen     = draw2d.Color{1, 0.6, 0.1, 1}
	flatCamera     = draw2d.Color{1, 1, 1, 0.55}
)

// step follows the mouse in the 2D view, over the area it was last painted in; it returns what
// go-renderer shows under the interface: the scene's world, empty in 2D.
func (flat *flatView) step(editor *Editor, state *input.State) (*scene.Scene, models.Camera) {
	if !flat.area.IsEmpty() {
		flat.followMouse(editor, state)
	}
	return editor.Scene.World(), models.Camera{Position: [4]float32{0, 0, 10, 1}, Up: [4]float32{0, 1, 0, 0}}
}

// paint draws the 2D view into its area of the window (go-ui's Paint, in the middle of the
// editor): the background, the grid, the scene and the editor's marks.
func (flat *flatView) paint(editor *Editor, area draw2d.Rect, list *draw2d.List) {
	flat.area = area
	saved := flat.drawing
	flat.drawing = *list
	background := flatBackground
	if editor.Scene.Sky != "" {
		background = draw2d.Hex(editor.Scene.Sky)
	}
	flat.drawing.Rectangle(area, 0, background, draw2d.Color{}, 0)
	flat.drawing.PushClip(area)
	flat.drawGrid()
	editor.Scene.DrawContent2D(&flat.drawing, area, flat.view)
	flat.drawMarks(editor)
	flat.drawing.PopClip()
	*list = flat.drawing
	flat.drawing = saved
}

// paintPlaying draws a 2D scene being played, through its camera, into the view's area.
func paintPlaying(played *engine.Scene, area draw2d.Rect, list *draw2d.List) {
	list.Rectangle(area, 0, flatBackground, draw2d.Color{}, 0)
	played.Draw2D(list, area, played.View2D())
}

func (flat *flatView) followMouse(editor *Editor, state *input.State) {
	x, y := state.CursorX, state.CursorY
	if state.ScrollDelta != 0 {
		flat.zoomAt(x, y, state.ScrollDelta)
	}
	point := flat.view.FromScreen(flat.area, float32(x), float32(y))
	flat.pointer = point
	if editor.PaintingHere() && paintTiles(editor, state, point) {
		return
	}
	if state.ButtonPressed(input.MouseLeft) || state.ButtonPressed(input.MouseRight) || state.ButtonPressed(input.MouseMiddle) {
		flat.pressing, flat.moved, flat.panning, flat.dragging = true, false, !state.Button(input.MouseLeft), nil
		flat.pressX, flat.pressY, flat.lastX, flat.lastY = x, y, x, y
		if chosen := editor.Selected; state.Button(input.MouseLeft) && chosen != nil {
			if picked, found := editor.PickFlat(point, flat.reach()); found && picked == chosen {
				flat.dragging, flat.grabbed = chosen, chosen.World().Position.Subtract(point)
			}
		}
		if state.Button(input.MouseLeft) && flat.dragging == nil {
			flat.panning = true // the background: the press may become a click, or a pan
		}
	}
	held := state.Button(input.MouseLeft) || state.Button(input.MouseRight) || state.Button(input.MouseMiddle)
	if flat.pressing && held {
		if math.Hypot(x-flat.pressX, y-flat.pressY) > clickDistance {
			flat.moved = true
		}
		switch {
		case flat.dragging != nil && flat.moved:
			target := point.Add(flat.grabbed)
			target.Z = flat.dragging.World().Position.Z
			editor.MoveTo(flat.dragging, target)
		case flat.panning && flat.moved:
			scale := flat.view.Height / float64(max(flat.area.Height, 1))
			flat.view.Centre = flat.view.Centre.Add(space.NewVector(-(x-flat.lastX)*scale, (y-flat.lastY)*scale, 0))
		}
		flat.lastX, flat.lastY = x, y
		return
	}
	if flat.pressing && !held {
		flat.pressing = false
		if !flat.moved && state.ButtonReleased(input.MouseLeft) {
			picked, _ := editor.PickFlat(point, flat.reach())
			editor.Select(picked)
		}
	}
}

// zoomAt zooms around the point under the pointer, which stays under it.
func (flat *flatView) zoomAt(x, y, scrolled float64) {
	before := flat.view.FromScreen(flat.area, float32(x), float32(y))
	flat.view.Height = math.Min(math.Max(flat.view.Height*math.Pow(0.88, scrolled), 0.5), 2000)
	after := flat.view.FromScreen(flat.area, float32(x), float32(y))
	flat.view.Centre = flat.view.Centre.Add(before.Subtract(after))
}

// paintTiles paints the chosen tile map while the left button is held, and erases while the right
// one is; it tells whether it took the mouse.
func paintTiles(editor *Editor, state *input.State, point space.Vector) bool {
	switch {
	case state.Button(input.MouseLeft):
		editor.PaintTileAt(point, editor.Brush)
	case state.Button(input.MouseRight):
		editor.PaintTileAt(point, scenefile.EmptyTile)
	default:
		return false
	}
	return true
}

// reach is how near, in metres, a click must fall to choose a camera or an empty node: a dozen
// pixels.
func (flat *flatView) reach() float64 {
	return 12 * flat.view.Height / float64(max(flat.area.Height, 1))
}

// frame centres the view on the chosen node.
func (flat *flatView) frame(chosen *engine.Node) {
	if chosen == nil {
		flat.view = engine.View2D{Height: 14}
		return
	}
	flat.view.Centre = chosen.World().Position
	if box := chosen.Bounds(); !box.IsEmpty() {
		flat.view.Height = math.Max(box.Size().Y*3, 4)
	}
}

func (flat *flatView) drawGrid() {
	low := flat.view.FromScreen(flat.area, flat.area.X, flat.area.Y+flat.area.Height)
	high := flat.view.FromScreen(flat.area, flat.area.X+flat.area.Width, flat.area.Y)
	step := 1.0
	for (high.X-low.X)/step > 120 {
		step *= 5 // far out, a line every five metres, then every twenty-five
	}
	for x := math.Floor(low.X/step) * step; x <= high.X; x += step {
		screen, _ := flat.view.ToScreen(flat.area, space.NewVector(x, 0, 0))
		colour := flatGrid
		if math.Mod(math.Abs(x), step*5) < 1e-9 {
			colour = flatGridFive
		}
		if math.Abs(x) < 1e-9 {
			colour = flatAxisY
		}
		flat.drawing.Rectangle(draw2d.Rect{X: screen, Y: flat.area.Y, Width: 1, Height: flat.area.Height}, 0, colour, draw2d.Color{}, 0)
	}
	for y := math.Floor(low.Y/step) * step; y <= high.Y; y += step {
		_, screen := flat.view.ToScreen(flat.area, space.NewVector(0, y, 0))
		colour := flatGrid
		if math.Mod(math.Abs(y), step*5) < 1e-9 {
			colour = flatGridFive
		}
		if math.Abs(y) < 1e-9 {
			colour = flatAxisX
		}
		flat.drawing.Rectangle(draw2d.Rect{X: flat.area.X, Y: screen, Width: flat.area.Width, Height: 1}, 0, colour, draw2d.Color{}, 0)
	}
}

// drawMarks draws what only the editor shows: the frame each camera sees, a dot for each empty
// node, and an outline around the chosen node.
func (flat *flatView) drawMarks(editor *Editor) {
	aspect := float64(flat.area.Width) / float64(max(flat.area.Height, 1))
	editor.Scene.Root.Walk(func(node *engine.Node) bool {
		if node == editor.Scene.Root || !node.Visible {
			return node == editor.Scene.Root
		}
		position := node.World().Position
		if settings, isCamera := node.Camera(); isCamera {
			height := settings.Height
			if height <= 0 {
				height = 10
			}
			half := space.NewVector(height*aspect/2, height/2, 0)
			flat.outline(space.Box{Minimum: position.Subtract(half), Maximum: position.Add(half)}, flatCamera, 1)
		} else if node.Bounds().IsEmpty() {
			x, y := flat.view.ToScreen(flat.area, position)
			flat.drawing.Rectangle(draw2d.Rect{X: x - 5, Y: y - 5, Width: 10, Height: 10}, 5, draw2d.Color{0.95, 0.6, 0.1, 0.9}, draw2d.Color{}, 0)
		}
		return true
	})
	chosen := editor.Selected
	if chosen == nil {
		return
	}
	if column, row, inside := chosen.CellAt(flat.pointer); editor.PaintingHere() && inside {
		flat.outline(chosen.CellBox(column, row), hoveredCell, 2)
	}
	box := chosen.Bounds()
	if box.IsEmpty() {
		position := chosen.World().Position
		box = space.Box{Minimum: position.Subtract(space.NewVector(0.2, 0.2, 0)), Maximum: position.Add(space.NewVector(0.2, 0.2, 0))}
	}
	flat.outline(box, flatChosen, 2)
}

func (flat *flatView) outline(box space.Box, colour draw2d.Color, width float32) {
	left, top := flat.view.ToScreen(flat.area, space.NewVector(box.Minimum.X, box.Maximum.Y, 0))
	right, bottom := flat.view.ToScreen(flat.area, space.NewVector(box.Maximum.X, box.Minimum.Y, 0))
	flat.drawing.Rectangle(draw2d.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}, 0, draw2d.Color{}, colour, width)
}
