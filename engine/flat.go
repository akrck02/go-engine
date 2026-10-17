package engine

import (
	_ "image/jpeg" // sprites may be JPEG pictures
	_ "image/png"
	"math"
	"sort"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// A 2D scene is drawn flat, from the front, by an orthographic camera: X to the right, Y up, in
// metres as in 3D. Sprites are pictures of the project's, as large as their pixels at the scene's
// pixels per metre (a sprite 32 pixels high stands a metre high at 32), times their node's scale;
// shapes are flat (a cube or a plane is a square, a sphere a circle, a metre across). Nodes with a
// greater Z are drawn over the others; with the same Z, in the scene's order. Turning a sprite is
// not drawn yet: draw2d draws pictures upright.

// Is2D tells whether the scene is 2D.
func (owner *Scene) Is2D() bool { return owner.Mode == scenefile.Mode2D }

// pixelsPerMetre is the scene's, or the default.
func (owner *Scene) pixelsPerMetre() float64 {
	if owner.PixelsPerMetre > 0 {
		return owner.PixelsPerMetre
	}
	return scenefile.DefaultPixelsPerMetre
}

// View2D is what a 2D camera sees: the point at the middle of the screen and how many metres fit
// between the screen's top and bottom.
type View2D struct {
	Centre space.Vector
	Height float64
}

// View2D is what the scene's main camera sees, or, without one, ten metres around the origin.
func (owner *Scene) View2D() View2D {
	camera, found := owner.MainCamera()
	if !found {
		return View2D{Height: 10}
	}
	height := 10.0
	if settings, _ := camera.Camera(); settings != nil && settings.Height > 0 {
		height = settings.Height
	}
	return View2D{Centre: camera.World().Position, Height: height}
}

// scale is how many screen pixels a metre takes in an area.
func (view View2D) scale(area draw2d.Rect) float64 {
	if view.Height <= 0 {
		return 1
	}
	return float64(area.Height) / view.Height
}

// ToScreen is where a point of the world falls in an area of the screen.
func (view View2D) ToScreen(area draw2d.Rect, point space.Vector) (x, y float32) {
	scale := view.scale(area)
	x = area.X + area.Width/2 + float32((point.X-view.Centre.X)*scale)
	y = area.Y + area.Height/2 - float32((point.Y-view.Centre.Y)*scale)
	return x, y
}

// FromScreen is the point of the world under a point of the screen, at Z 0.
func (view View2D) FromScreen(area draw2d.Rect, x, y float32) space.Vector {
	scale := view.scale(area)
	return space.NewVector(view.Centre.X+float64(x-area.X-area.Width/2)/scale, view.Centre.Y-float64(y-area.Y-area.Height/2)/scale, 0)
}

// MetresAcross is how many metres an area shows from side to side.
func (view View2D) MetresAcross(area draw2d.Rect) float64 {
	return float64(area.Width) / view.scale(area)
}

// flatBounds is the box a node fills in a 2D scene: its sprite or its shape, around its world
// position, and a centimetre deep.
func (node *Node) flatBounds() space.Box {
	width, height, found := node.flatSize()
	if !found {
		return space.EmptyBox()
	}
	world := node.World()
	half := space.NewVector(width/2, height/2, 0.005)
	return space.Box{Minimum: world.Position.Subtract(half), Maximum: world.Position.Add(half)}
}

// flatSize is how large a node is drawn in a 2D scene, in metres, if it shows anything.
func (node *Node) flatSize() (width, height float64, found bool) {
	scale := node.World().Scale
	switch {
	case node.sprite != "":
		picture, err := node.scene.assets.image(node.sprite)
		if err != nil {
			return 0, 0, false
		}
		pixelsWide, pixelsHigh := picture.Size()
		perMetre := node.scene.pixelsPerMetre()
		return math.Abs(float64(pixelsWide) / perMetre * scale.X), math.Abs(float64(pixelsHigh) / perMetre * scale.Y), true
	case node.shape != "":
		return math.Abs(scale.X), math.Abs(scale.Y), true
	}
	return 0, 0, false
}

// Draw2D draws a 2D scene into an area of the screen, as a view sees it: the sky's colour, then
// its sprites and shapes from the furthest to the nearest.
func (owner *Scene) Draw2D(list *draw2d.List, area draw2d.Rect, view View2D) {
	if owner.Sky != "" {
		list.Rectangle(area, 0, draw2d.Hex(owner.Sky), draw2d.Color{}, 0)
	}
	owner.DrawContent2D(list, area, view)
}

// DrawContent2D draws a 2D scene's sprites and shapes without its sky: over something of the
// program's own (the editor's grid).
func (owner *Scene) DrawContent2D(list *draw2d.List, area draw2d.Rect, view View2D) {
	list.PushClip(area)
	defer list.PopClip()
	for _, node := range owner.drawOrder() {
		owner.drawFlat(list, area, view, node)
	}
}

// drawOrder is the visible nodes that show something, furthest first.
func (owner *Scene) drawOrder() []*Node {
	var shown []*Node
	var visit func(node *Node)
	visit = func(node *Node) {
		if !node.Visible {
			return
		}
		if node.sprite != "" || node.shape != "" {
			shown = append(shown, node)
		}
		for _, child := range node.children {
			visit(child)
		}
	}
	for _, child := range owner.Root.children {
		visit(child)
	}
	sort.SliceStable(shown, func(first, second int) bool {
		return shown[first].World().Position.Z < shown[second].World().Position.Z
	})
	return shown
}

func (owner *Scene) drawFlat(list *draw2d.List, area draw2d.Rect, view View2D, node *Node) {
	box := node.flatBounds()
	if box.IsEmpty() {
		return
	}
	left, top := view.ToScreen(area, space.NewVector(box.Minimum.X, box.Maximum.Y, 0))
	right, bottom := view.ToScreen(area, space.NewVector(box.Maximum.X, box.Minimum.Y, 0))
	rect := draw2d.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
	if node.sprite != "" {
		if picture, err := owner.assets.image(node.sprite); err == nil {
			list.Image(picture, rect, 0, draw2d.Color{1, 1, 1, 1})
		}
		return
	}
	colour := draw2d.Hex(firstColour(node.colour))
	radius := float32(0)
	if node.shape == "sphere" {
		radius = min(rect.Width, rect.Height) / 2
	}
	list.Rectangle(rect, radius, colour, draw2d.Color{}, 0)
}

func firstColour(colour string) string {
	if colour == "" {
		return "#c8c8c8"
	}
	return colour
}
