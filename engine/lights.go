package engine

import (
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/scenefile"
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
func (owner *Scene) drawLights(list *draw2d.List, area draw2d.Rectangle, view View2D) {
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
func (owner *Scene) drawLight(list *draw2d.List, area draw2d.Rectangle, view View2D, node *Node) {
	light := node.light
	colour := owner.lightColour(light.Colour, light.Strength())
	if light.Kind == scenefile.SpriteLight {
		picture, found := owner.spritePicture(node)
		rect, shown := owner.shownRect(area, view, node)
		if found && shown && !rect.Intersect(area).IsEmpty() {
			list.SpriteLight(picture, rect, colour)
		}
		return
	}
	x, y := view.ToScreen(area, node.World().Position)
	inner, outer := node.LightRadii()
	scale := float32(view.scale(area))
	reach := draw2d.Rectangle{X: x - float32(outer)*scale, Y: y - float32(outer)*scale, Width: 2 * float32(outer) * scale, Height: 2 * float32(outer) * scale}
	if reach.Intersect(area).IsEmpty() {
		return // shines on nothing in sight
	}
	list.PointLight(x, y, float32(inner)*scale, float32(outer)*scale, colour)
}

// LightRadii are a point light's radii in metres, times its node's largest scale (across and up
// in 2D, which has no depth to scale): a lamp scaled up reaches further. A node that is not a
// light has none.
func (node *Node) LightRadii() (inner, outer float64) {
	if node.light == nil {
		return 0, 0
	}
	inner, outer = node.light.Radii()
	scale := node.World().Scale
	largest := max(math.Abs(scale.X), math.Abs(scale.Y))
	if !node.scene.Is2D() {
		largest = max(largest, math.Abs(scale.Z))
	}
	return inner * largest, outer * largest
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

// syncLights gives go-renderer a 3D scene's point lights, where their nodes are now: called by
// Sync, every frame the scene plays. A 2D scene draws its lights itself (drawLights).
func (owner *Scene) syncLights() {
	lights := owner.world.Lights[:0]
	if !owner.Is2D() {
		owner.lit = owner.collectLights(owner.lit[:0], owner.Root)
		for _, node := range owner.lit {
			if node.light.Kind != scenefile.PointLight {
				continue // a sprite light shines in 2D only
			}
			lights = append(lights, owner.pointLight(node))
		}
	}
	owner.world.Lights = append(lights, owner.ownLights...)
}

// SetPointLights gives a 3D scene point lights of the program's own, besides its lamp nodes: lights
// it places itself every frame (on things it draws on the world directly), in the world's
// coordinates, their colours linear and their alpha the intensity, as go-renderer takes them. The
// slice is kept and read every frame, so change its lights in place to move them: nothing is
// allocated. nil removes them.
func (owner *Scene) SetPointLights(lights []scene.PointLight) { owner.ownLights = lights }

// pointLight is a 3D node's light as go-renderer takes it: in the world, its colour linear.
func (owner *Scene) pointLight(node *Node) scene.PointLight {
	position := node.World().Position
	inner, outer := node.LightRadii()
	colour := owner.lightColour(node.light.Colour, node.light.Strength())
	return scene.PointLight{
		Position: graphics.Vector{float32(position.X), float32(position.Y), float32(position.Z), 1},
		Color:    graphics.Vector{linear(colour[0]), linear(colour[1]), linear(colour[2]), colour[3]},
		Inner:    float32(inner), Outer: float32(outer),
	}
}

// screenGamma is the gamma colours written as "#rrggbb" are in; 3D shading works in linear light.
const screenGamma = 2.2

// linear turns a channel of a screen colour into linear light.
func linear(channel float32) float32 { return float32(math.Pow(float64(channel), screenGamma)) }

// SetAmbient sets the scene's ambient light (nil for none): in 2D the light that reaches
// everything; in 3D how strong its sky light is, from 0 (none) to 2 (twice the default).
func (owner *Scene) SetAmbient(ambient *scenefile.Ambient) {
	owner.Ambient = ambient
	owner.applyAmbient()
}

// applyAmbient makes a 3D scene's sky light as strong as its ambient says.
func (owner *Scene) applyAmbient() {
	owner.world.Environment.Ambient = scene.DefaultEnvironment().Ambient
	if owner.Ambient != nil {
		owner.world.Environment.Ambient *= float32(owner.Ambient.Strength())
	}
}
