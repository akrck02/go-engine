package editor

import (
	"math"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// A chosen 3D lamp shows how far it reaches: two level rings round it, where its light is full
// and where it ends, in its own colour, as the 2D view draws its circles.

// ringSegments is how many straight pieces draw a ring: round enough at any size seen.
const ringSegments = 48

// ringThickness is a ring's thickness, as a share of its radius, at least a thin line.
const (
	ringThickness      = 0.006
	thinnestRingMetres = 0.012
)

// showLampReach adds the chosen node's rings when it is a point light.
func (view *viewport) showLampReach(chosen *engine.Node) {
	light, isLight := chosen.Light()
	if !isLight || light.Kind != scenefile.PointLight {
		return
	}
	centre := chosen.World().Position
	inner, outer := chosen.LightRadii()
	colour := lampRingColour(light.Colour)
	for _, radius := range [2]float64{inner, outer} {
		if radius > 0 {
			view.addRing(centre, radius, colour)
		}
	}
}

// addRing adds a level ring of straight pieces round a point.
func (view *viewport) addRing(centre space.Vector, radius float64, colour graphics.Vec4) {
	thick := math.Max(thinnestRingMetres, radius*ringThickness)
	piece := 2 * math.Pi * radius / ringSegments
	for segment := range ringSegments {
		angle := 2 * math.Pi * (float64(segment) + 0.5) / ringSegments
		middle := centre.Add(space.NewVector(radius*math.Cos(angle), 0, radius*math.Sin(angle)))
		view.selection.Children = append(view.selection.Children, view.turnedLine(middle, angle, piece+thick, thick, colour))
	}
}

// turnedLine is a thin box of a length, standing at a point, turned round the vertical so that
// it runs along a ring at an angle.
func (view *viewport) turnedLine(middle space.Vector, angle, length, thick float64, colour graphics.Vec4) *scene.Node {
	line := view.lineNode(space.NewVector(0, 0, 0), space.NewVector(thick, thick, length), colour)
	place := graphics.Translate(graphics.Vec4{float32(middle.X), float32(middle.Y), float32(middle.Z), 0})
	line.Transform = place.Multiply(graphics.RotateY(-angle)).Multiply(line.Transform)
	return line
}

// lampRingColour is a light's colour as the rings show it: linear, opaque.
func lampRingColour(colour string) graphics.Vec4 {
	if colour == "" {
		colour = "#ffffff"
	}
	parsed, valid := scene.ColorFromHex(colour)
	if !valid {
		return chosenColor
	}
	return parsed
}
