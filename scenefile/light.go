package scenefile

import "github.com/akrck02/go-engine/failure"

// A 2D scene can be lit, as Unity lights its 2D scenes: its ambient light lights everything a
// little, and lights on its nodes add to it where they shine. A scene with neither is drawn
// unlit, as it always was.

// Ambient is the light that reaches everything in a 2D scene: daylight is white at 1, a dim room
// is a dark blue at 0.4.
type Ambient struct {
	Colour string `json:"colour,omitempty"` // "#rrggbb"; empty is white
	// Intensity is how strong it is: 1 (the default) leaves the scene's colours as they are, less
	// darkens them, up to 2 brightens them.
	Intensity *float64 `json:"intensity,omitempty"`
}

// Light makes a node a light in a 2D scene.
type Light struct {
	Kind      LightKind `json:"kind"`
	Colour    string    `json:"colour,omitempty"`    // "#rrggbb"; empty is white
	Intensity *float64  `json:"intensity,omitempty"` // as the ambient's; nil is 1
	// Inner and Outer are a point light's radii in metres: it is full within Inner and fades to
	// nothing at Outer.
	Inner float64 `json:"inner,omitempty"`
	Outer float64 `json:"outer,omitempty"`
}

// LightKind is what a light shines from.
type LightKind string

const (
	// PointLight shines from the node's position all around, as a lamp.
	PointLight LightKind = "point"
	// SpriteLight shines through the node's sprite: its colours light where it is, and it is not
	// drawn itself. A glowing sign, a window.
	SpriteLight LightKind = "sprite"
)

// LightKinds are the kinds of light, in the order the editor offers them.
var LightKinds = []LightKind{PointLight, SpriteLight}

// The defaults of what a light or an ambient leaves unsaid.
const (
	DefaultIntensity  = 1.0
	DefaultLightOuter = 3.0 // metres
	// MostIntensity is the strongest a light or an ambient can be: twice as bright.
	MostIntensity = 2.0
)

// Strength is the ambient's intensity, or 1.
func (ambient *Ambient) Strength() float64 { return intensityOf(ambient.Intensity) }

// Strength is the light's intensity, or 1.
func (light *Light) Strength() float64 { return intensityOf(light.Intensity) }

// Radii are a point light's inner and outer radii in metres, the outer one 3 when it has none.
func (light *Light) Radii() (inner, outer float64) {
	outer = light.Outer
	if outer <= 0 {
		outer = DefaultLightOuter
	}
	return min(max(light.Inner, 0), outer), outer
}

func intensityOf(intensity *float64) float64 {
	if intensity == nil {
		return DefaultIntensity
	}
	return min(max(*intensity, 0), MostIntensity)
}

// Copy is a light that shares nothing with this one; nil for none.
func (light *Light) Copy() *Light {
	if light == nil {
		return nil
	}
	copied := *light
	copied.Intensity = copyNumber(light.Intensity)
	return &copied
}

// Copy is an ambient that shares nothing with this one; nil for none.
func (ambient *Ambient) Copy() *Ambient {
	if ambient == nil {
		return nil
	}
	copied := *ambient
	copied.Intensity = copyNumber(ambient.Intensity)
	return &copied
}

func copyNumber(number *float64) *float64 {
	if number == nil {
		return nil
	}
	copied := *number
	return &copied
}

// check tells whether a node's light is one the engine can draw.
func (light *Light) check(path string) error {
	if !isLightKind(light.Kind) {
		return failure.Wrap(UnknownLightKind, path+" is a "+string(light.Kind)+" light", nil)
	}
	if light.Inner < 0 || light.Outer < 0 || light.Outer > 0 && light.Inner > light.Outer {
		return failure.Wrap(LightRadiiWrong, path, nil)
	}
	return nil
}

func isLightKind(kind LightKind) bool {
	for _, each := range LightKinds {
		if each == kind {
			return true
		}
	}
	return false
}
