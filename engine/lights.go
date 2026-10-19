package engine

import (
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// A 2D scene with lights is drawn as Unity draws its 2D lights: its sprites and shapes first,
// then its lights added together over its ambient into a picture of light, which multiplies them
// (draw2d's Lighting). A scene with no light and no ambient is drawn unlit, at no cost. A point
// light's radii are in metres, as the scene's; a sprite light shines through its node's sprite,
// in the node's place, and its sprite is not drawn.

// Light is the node's light, to edit in place; a node that is not one has none.
func (node *Node) Light() (*scenefile.Light, bool) { return node.light, node.light != nil }

// SetLight makes the node a light, or not one (nil).
func (node *Node) SetLight(light *scenefile.Light) { node.light = light }

// isSpriteLight tells whether the node shines through its sprite instead of showing it.
func (node *Node) isSpriteLight() bool {
	return node.light != nil && node.light.Kind == scenefile.SpriteLight
}

// white is the colour of a light or an ambient that names none.
const white = "#ffffff"

// drawLights draws the scene's lights and lights what was drawn before them with them; nothing
// when the scene has neither lights nor an ambient.
func (owner *Scene) drawLights(list *draw2d.List, area draw2d.Rect, view View2D) {
	owner.lit = owner.collectLights(owner.lit[:0], owner.Root)
	if len(owner.lit) == 0 && owner.Ambient == nil {
		return
	}
	for _, node := range owner.lit {
		owner.drawLight(list, area, view, node)
	}
	ambient := owner.Ambient
	if ambient == nil {
		ambient = &daylight
	}
	list.Lighting(owner.lightColour(ambient.Colour, ambient.Strength()))
}

// daylight is the ambient of a scene with lights and no ambient of its own: white, at 1.
var daylight = scenefile.Ambient{}

// collectLights lists the visible lights under a node.
func (owner *Scene) collectLights(nodes []*Node, node *Node) []*Node {
	for _, child := range node.children {
		if !child.Visible {
			continue
		}
		if child.light != nil {
			nodes = append(nodes, child)
		}
		nodes = owner.collectLights(nodes, child)
	}
	return nodes
}

// drawLight adds one light: a point light around the node's position, a sprite light over its
// sprite's place.
func (owner *Scene) drawLight(list *draw2d.List, area draw2d.Rect, view View2D, node *Node) {
	light := node.light
	colour := owner.lightColour(light.Colour, light.Strength())
	if light.Kind == scenefile.SpriteLight {
		picture, found := owner.spritePicture(node)
		rect, shown := owner.screenRect(area, view, node)
		if found && shown {
			list.SpriteLight(picture, rect, colour)
		}
		return
	}
	x, y := view.ToScreen(area, node.World().Position)
	inner, outer := light.Radii()
	scale := view.scale(area)
	list.PointLight(x, y, float32(inner*scale), float32(outer*scale), colour)
}

// lightColour is a light's colour with its intensity as the alpha, as draw2d's lights take it.
func (owner *Scene) lightColour(colour string, intensity float64) draw2d.Color {
	if colour == "" {
		colour = white
	}
	lit := owner.flatColour(colour)
	lit[3] = float32(intensity)
	return lit
}

// screenRect is where a node's sprite or shape falls on the screen, if it shows anything.
func (owner *Scene) screenRect(area draw2d.Rect, view View2D, node *Node) (draw2d.Rect, bool) {
	box := node.flatBounds()
	if box.IsEmpty() {
		return draw2d.Rect{}, false
	}
	left, top := view.ToScreen(area, space.NewVector(box.Minimum.X, box.Maximum.Y, 0))
	right, bottom := view.ToScreen(area, space.NewVector(box.Maximum.X, box.Minimum.Y, 0))
	return draw2d.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}, true
}
