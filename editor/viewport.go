package editor

import (
	"math"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenes"
	"github.com/akrck02/go-engine/shapes"
	"github.com/akrck02/go-engine/space"
)

// The 3D view: the scene seen from the editor's own camera, which orbits a point (drag to turn,
// right-drag to pan, scroll to zoom), over a grid of one-metre cells with the axes drawn in their
// colours (X red, Y green, Z blue). A click chooses a node; dragging the chosen node moves it over
// the ground, at its height.

// ViewMode is how the editor's camera shows the scene.
type ViewMode int

const (
	ViewPerspective ViewMode = iota
	ViewOrthographic
	ViewIsometric
)

// clickDistance is how far the pointer may move between a press and its release for them to be a
// click, in window pixels; and how far before a press on the chosen node drags it.
const clickDistance = 4

// viewport is the 3D view's state.
type viewport struct {
	orbit  *camera.Orbit
	camera models.Camera
	mode   ViewMode
	// through is a camera node of the scene to see through instead of the editor's camera.
	through *engine.Node

	gizmos    *scene.Node // the grid and the axes
	selection *scene.Node // the edges of the chosen node's box
	line      *scene.Mesh

	flat *flatView // the view of a 2D scene

	pressX, pressY float64
	pressing       bool
	dragging       *engine.Node
	dragMoved      bool
	dragHeight     float64
	dragOffset     space.Vec3
}

func newViewport() *viewport {
	view := &viewport{orbit: camera.NewOrbit(graphics.Vec4{0, 0, 0, 1}, 14), flat: newFlatView()}
	view.orbit.Yaw, view.orbit.Pitch = 0.6, 0.55
	view.orbit.MinimumDistance, view.orbit.MaximumDistance = 0.3, 500
	view.line = shapes.Cube()
	view.line.Material = &scene.Material{Name: "line", Kind: scene.KindUnlit, BaseColor: graphics.Vec4{1, 1, 1, 1}}
	view.gizmos = view.makeGizmos()
	view.selection = &scene.Node{Name: "selection", Transform: graphics.Identity()}
	return view
}

// lineNode is a thin box from one point to another along an axis, in a colour.
func (view *viewport) lineNode(centre, size space.Vec3, colour graphics.Vec4) *scene.Node {
	mesh := *view.line
	mesh.Material = &scene.Material{Name: "line", Kind: scene.KindUnlit, BaseColor: colour}
	mesh.GPU = nil
	transform := graphics.Translate(graphics.Vec4{float32(centre.X), float32(centre.Y), float32(centre.Z), 0}).
		Multiply(graphics.Scale(graphics.Vec4{float32(size.X), float32(size.Y), float32(size.Z), 0}))
	return &scene.Node{Name: "line", Transform: transform, Mesh: &mesh}
}

// gridColour and the axes' colours.
var (
	gridColour  = graphics.Vec4{0.35, 0.37, 0.4, 1}
	gridFive    = graphics.Vec4{0.22, 0.24, 0.27, 1}
	axisX       = graphics.Vec4{0.86, 0.2, 0.2, 1}
	axisY       = graphics.Vec4{0.25, 0.7, 0.3, 1}
	axisZ       = graphics.Vec4{0.2, 0.4, 0.9, 1}
	chosenColor = graphics.Vec4{1, 0.6, 0.1, 1}
)

// gridExtent is how far the grid reaches from the origin, in metres.
const gridExtent = 20

func (view *viewport) makeGizmos() *scene.Node {
	gizmos := &scene.Node{Name: "gizmos", Transform: graphics.Identity()}
	const thin, lift = 0.012, 0.002 // a hair above the ground, so a plane at 0 does not hide it
	for step := -gridExtent; step <= gridExtent; step++ {
		if step == 0 {
			continue
		}
		colour, width := gridColour, thin
		if step%5 == 0 {
			colour, width = gridFive, thin*1.6
		}
		offset := float64(step)
		gizmos.Children = append(gizmos.Children,
			view.lineNode(space.V(offset, lift, 0), space.V(width, width, 2*gridExtent), colour),
			view.lineNode(space.V(0, lift, offset), space.V(2*gridExtent, width, width), colour))
	}
	const axis = 0.03
	gizmos.Children = append(gizmos.Children,
		view.lineNode(space.V(0, lift, 0), space.V(2*gridExtent, axis, axis), axisX),
		view.lineNode(space.V(0, lift, 0), space.V(axis, axis, 2*gridExtent), axisZ),
		view.lineNode(space.V(0, 1, 0), space.V(axis, 2, axis), axisY))
	return gizmos
}

// showSelection draws the edges of the chosen node's box.
func (view *viewport) showSelection(chosen *engine.Node) {
	view.selection.Children = view.selection.Children[:0]
	if chosen == nil {
		return
	}
	box := chosen.World().Position
	bounds := chosen.Bounds()
	if bounds.IsEmpty() {
		bounds = space.Box{Minimum: box.Sub(space.V(0.15, 0.15, 0.15)), Maximum: box.Add(space.V(0.15, 0.15, 0.15))}
	}
	size, low, high := bounds.Size(), bounds.Minimum, bounds.Maximum
	thick := math.Max(0.015, math.Max(size.X, math.Max(size.Y, size.Z))*0.006)
	for _, first := range []float64{0, 1} {
		for _, second := range []float64{0, 1} {
			y, z := low.Y+first*size.Y, low.Z+second*size.Z
			x := low.X + first*size.X
			view.selection.Children = append(view.selection.Children,
				view.lineNode(space.V((low.X+high.X)/2, y, z), space.V(size.X+thick, thick, thick), chosenColor),
				view.lineNode(space.V(x, (low.Y+high.Y)/2, low.Z+second*size.Z), space.V(thick, size.Y+thick, thick), chosenColor),
				view.lineNode(space.V(low.X+second*size.X, y, (low.Z+high.Z)/2), space.V(thick, thick, size.Z+thick), chosenColor))
		}
	}
}

// step follows the mouse in the 3D view and places the camera; it returns what to draw.
func (view *viewport) step(editor *Editor, state *input.State, seconds float64) (*scene.Scene, models.Camera) {
	if played, playing := editor.PlayingHere(); playing {
		if played.Is2D() {
			played.Update(seconds, state)
			return played.World(), view.camera
		}
		return view.play(played, state, seconds)
	}
	if editor.Scene.Is2D() {
		return view.flat.step(editor, state)
	}
	world := editor.Scene.World()
	view.attach(world)
	mouseTaken := view.followMouse(editor, state)
	switch {
	case mouseTaken:
		view.orbit.Update(nil, seconds, &view.camera)
	case view.mode == ViewIsometric:
		view.orbit.Update(view.isometricInput(state), seconds, &view.camera)
	default:
		view.orbit.Update(state, seconds, &view.camera)
	}
	if view.mode == ViewIsometric {
		view.fixIsometricAngle()
	}
	view.camera.OrthographicHeight = 0
	switch view.mode {
	case ViewOrthographic, ViewIsometric:
		view.camera.OrthographicHeight = float32(view.orbit.Distance * 0.75)
		view.camera.Near, view.camera.Far = -1000, 2000
	default:
		view.camera.Near, view.camera.Far = 0.05, 2000
	}
	view.showSelection(editor.Selected)
	if view.through != nil && view.through.Scene() == editor.Scene && view.through.Parent() != nil {
		return world, view.through.View()
	}
	view.through = nil
	return world, view.camera
}

// attach puts the grid and the selection into the scene's world, once per scene.
func (view *viewport) attach(world *scene.Scene) {
	for _, node := range world.Nodes {
		if node == view.gizmos {
			return
		}
	}
	world.Nodes = append(world.Nodes, view.gizmos, view.selection)
}

// followMouse chooses with a click and drags the chosen node; it tells whether the camera must
// leave the mouse alone this frame.
func (view *viewport) followMouse(editor *Editor, state *input.State) bool {
	if state == nil {
		return false
	}
	x, y := state.CursorX, state.CursorY
	if state.ButtonPressed(input.MouseLeft) {
		view.pressX, view.pressY, view.pressing, view.dragMoved = x, y, true, false
		view.dragging = nil
		if chosen := editor.Selected; chosen != nil && view.through == nil {
			if origin, direction, found := view.ray(state, x, y); found {
				if _, hit := chosen.Bounds().Hit(origin, direction); hit {
					view.startDrag(chosen, origin, direction)
				}
			}
		}
	}
	moved := math.Hypot(x-view.pressX, y-view.pressY)
	if view.dragging != nil && state.Button(input.MouseLeft) {
		if moved > clickDistance {
			view.dragMoved = true
		}
		if view.dragMoved {
			if origin, direction, found := view.ray(state, x, y); found {
				if point, onPlane := pointAtHeight(origin, direction, view.dragHeight); onPlane {
					editor.MoveTo(view.dragging, point.Add(view.dragOffset))
				}
			}
		}
		return true
	}
	if state.ButtonReleased(input.MouseLeft) && view.pressing {
		view.pressing = false
		wasDragging := view.dragging != nil && view.dragMoved
		view.dragging = nil
		if !wasDragging && moved <= clickDistance {
			if origin, direction, found := view.ray(state, x, y); found {
				picked, _ := editor.Pick(origin, direction)
				editor.Select(picked)
			}
		}
		return wasDragging
	}
	return false
}

func (view *viewport) startDrag(chosen *engine.Node, origin, direction space.Vec3) {
	position := chosen.World().Position
	point, onPlane := pointAtHeight(origin, direction, position.Y)
	if !onPlane {
		return
	}
	view.dragging, view.dragHeight, view.dragOffset = chosen, position.Y, position.Sub(point)
}

// ray is the ray from the editor's camera through a point of the window.
func (view *viewport) ray(state *input.State, x, y float64) (origin, direction space.Vec3, found bool) {
	width, height := state.WindowWidth, state.WindowHeight
	through, valid := camera.RayThrough(view.camera, width, height, x, y)
	if !valid {
		return space.Zero, space.Zero, false
	}
	origin = space.V(float64(through.Origin[0]), float64(through.Origin[1]), float64(through.Origin[2]))
	direction = space.V(float64(through.Direction[0]), float64(through.Direction[1]), float64(through.Direction[2])).Normalized()
	return origin, direction, true
}

// pointAtHeight is where a ray crosses the level plane at a height.
func pointAtHeight(origin, direction space.Vec3, height float64) (space.Vec3, bool) {
	if math.Abs(direction.Y) < 1e-9 {
		return space.Zero, false
	}
	distance := (height - origin.Y) / direction.Y
	if distance < 0 {
		return space.Zero, false
	}
	return origin.Add(direction.Scale(distance)), true
}

// frame turns the camera to the chosen node, near enough to see it whole.
func (view *viewport) frame(chosen *engine.Node) {
	if chosen == nil {
		view.orbit.Target, view.orbit.Distance = graphics.Vec4{0, 0, 0, 1}, 14
		return
	}
	bounds := chosen.Bounds()
	centre, size := chosen.World().Position, 1.0
	if !bounds.IsEmpty() {
		centre, size = bounds.Centre(), bounds.Size().Length()
	}
	view.orbit.Target = graphics.Vec4{float32(centre.X), float32(centre.Y), float32(centre.Z), 1}
	view.orbit.Distance = math.Max(size*1.8, 2)
}

// setMode changes how the editor's camera shows the scene; isometric turns it to the classic angle,
// which it keeps.
func (view *viewport) setMode(mode ViewMode) {
	view.mode = mode
	if mode == ViewIsometric {
		view.fixIsometricAngle()
	}
}

// fixIsometricAngle keeps the isometric view at its angle: 45° round, looking down 35.26°.
func (view *viewport) fixIsometricAngle() {
	view.orbit.Yaw, view.orbit.Pitch = math.Pi/4, space.Radians(-scenes.IsometricAngles[0])
}

// isometricInput is the input for the camera in the isometric view, whose angle does not turn: a
// drag with the left button pans, as a drag with the right one does (the camera pans with Shift
// held), and the arrows are left out (they turn it).
func (view *viewport) isometricInput(state *input.State) *input.State {
	fixed := state.Unclaimed()
	if fixed.Button(input.MouseLeft) {
		fixed.SetKey(input.KeyShift, true)
	}
	for _, arrow := range []input.Key{input.KeyUp, input.KeyDown, input.KeyLeft, input.KeyRight} {
		fixed.SetKey(arrow, false)
	}
	return &fixed
}

// play moves the scene played in the view on a frame, with the keys and the mouse the interface
// left, and sees it through its main camera (or the editor's, if it has none).
func (view *viewport) play(played *engine.Scene, state *input.State, seconds float64) (*scene.Scene, models.Camera) {
	played.Update(seconds, state)
	if camera, found := played.MainCamera(); found {
		return played.World(), camera.View()
	}
	view.orbit.Update(nil, seconds, &view.camera)
	return played.World(), view.camera
}
