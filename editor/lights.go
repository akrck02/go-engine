package editor

import (
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// A light, in 2D or 3D, and a scene's ambient: the inspector's sections that edit them, and the
// changes they make on the chosen node or the scene. A change is synced at once, so the view (and
// a 3D scene's lamps) shows it as it is typed.

var (
	intensityLimits = numberLimits{0, scenefile.MostIntensity, 0.05}
	radiusLimits    = numberLimits{0, 1000, 0.1}
)

// The colour a new light and a new ambient start with: a warm lamp, and a dim blue night.
const (
	newLightColour   = "#ffd27a"
	newAmbientColour = "#203050"
)

// newAmbientIntensity is how strong a new ambient is: dim, so that lights show.
const newAmbientIntensity = 0.4

// lightKindNotes say what each kind of light does, in the order of scenefile.LightKinds.
var lightKindNotes = []string{
	"Point: shines all around the node, full within its inner radius and fading to nothing at its outer one, in metres.",
	"Sprite: shines through the node's sprite, in its place; the sprite itself is not drawn. 2D only.",
}

// newLampHeight is how high a new 3D lamp hangs, in metres: over what stands on the ground.
const newLampHeight = 2

// newPointLight is a new light: a warm lamp reaching the default distance.
func newPointLight() *scenefile.Light {
	return &scenefile.Light{Kind: scenefile.PointLight, Colour: newLightColour, Outer: scenefile.DefaultLightOuter}
}

// GiveLight makes the chosen node a light of a kind, or changes the kind of the one it has.
func (editor *Editor) GiveLight(kind scenefile.LightKind) {
	node := editor.Selected
	if node == nil {
		return
	}
	light, found := node.Light()
	if !found {
		light = newPointLight()
		node.SetLight(light)
	}
	light.Kind = kind
	editor.LightChanged()
}

// TakeLight makes the chosen node a light no more.
func (editor *Editor) TakeLight() {
	if editor.Selected == nil {
		return
	}
	editor.Selected.SetLight(nil)
	editor.LightChanged()
}

// LightChanged keeps a change to a light or the ambient, and shows it at once.
func (editor *Editor) LightChanged() {
	editor.Scene.Sync()
	editor.Changed()
}

// SetAmbient gives the scene an ambient light of its own, or takes it away (nil).
func (editor *Editor) SetAmbient(ambient *scenefile.Ambient) {
	editor.Scene.SetAmbient(ambient)
	editor.LightChanged()
}

// lightInspector edits the node's light, or offers to make it one.
func (view *View) lightInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	light, found := node.Light()
	context.Separator()
	context.Row(func() {
		context.Heading("Light")
		context.Spacer()
		if found && context.FlatIconButton(iconRemove, "") {
			editor.TakeLight()
		}
		if found {
			context.Tooltip("Make it a light no more")
		}
	})
	if !found {
		note(context, "A light shines on what is around the node: a lamp, a glowing sign. It moves with the node.")
		if context.Button("Make it a light") {
			editor.GiveLight(scenefile.PointLight)
		}
		return
	}
	view.lightKindTabs(context, light)
	view.lightColourFields(context, light)
	if light.Kind == scenefile.PointLight {
		view.lightRadiusFields(context, light)
	}
}

// lightKindTabs choose a point or a sprite light; a 3D scene has point lights only.
func (view *View) lightKindTabs(context *ui.Context, light *scenefile.Light) {
	kind := max(0, indexOf(scenefile.LightKinds, light.Kind))
	if !view.Editor.Scene.Is2D() {
		note(context, lightKindNotes[0])
		return
	}
	if context.Tabs("Light kind", &kind, "Point", "Sprite") {
		view.Editor.GiveLight(scenefile.LightKinds[kind])
	}
	note(context, lightKindNotes[kind])
}

// lightColourFields edit a light's colour and intensity.
func (view *View) lightColourFields(context *ui.Context, light *scenefile.Light) {
	colour := light.Colour
	if colour == "" {
		colour = "#ffffff"
	}
	context.Caption("Colour")
	if colourPicker(context, "Light colour", &colour) {
		light.Colour = colour
		view.Editor.LightChanged()
	}
	intensity := light.Strength()
	context.Caption("Intensity: 1 a plain lamp, up to 2")
	if numberField(context, "Light intensity", &intensity, intensityLimits, "%.2f×") {
		light.Intensity = &intensity
		view.Editor.LightChanged()
	}
}

// lightRadiusFields edit a point light's radii.
func (view *View) lightRadiusFields(context *ui.Context, light *scenefile.Light) {
	inner, outer := light.Radii()
	context.Caption("Full within, and gone at, in metres")
	context.Push("Light radii")
	context.Grid(2, func() {
		if numberField(context, "Inner", &inner, radiusLimits, "Inner %.1f m") {
			light.Inner = min(inner, outer)
			view.Editor.LightChanged()
		}
		if numberField(context, "Outer", &outer, radiusLimits, "Outer %.1f m") {
			light.Outer = max(outer, light.Inner)
			view.Editor.LightChanged()
		}
	})
	context.Pop()
}

// ambientInspector edits the scene's ambient light: in 2D its colour and intensity, in 3D how
// strong its sky light is.
func (view *View) ambientInspector(context *ui.Context) {
	editor := view.Editor
	ambient := editor.Scene.Ambient
	context.Caption("Ambient light: what reaches everything")
	ownAmbient := ambient != nil
	if context.Checkbox("An ambient light of its own", &ownAmbient) {
		if ownAmbient {
			intensity := newAmbientIntensity
			editor.SetAmbient(&scenefile.Ambient{Colour: newAmbientColour, Intensity: &intensity})
		} else {
			editor.SetAmbient(nil)
		}
	}
	if ambient == nil && editor.Scene.Is2D() {
		note(context, "None: daylight when the scene has lights, and no lighting at all when it has none.")
		return
	}
	if ambient == nil {
		note(context, "None: the sky lights the scene as it is.")
		return
	}
	if editor.Scene.Is2D() {
		colour := ambient.Colour
		if colour == "" {
			colour = "#ffffff"
		}
		if colourPicker(context, "Ambient colour", &colour) {
			ambient.Colour = colour
			editor.LightChanged()
		}
	}
	if !editor.Scene.Is2D() {
		note(context, "In 3D, how strong the sky's light is: 1 as it is, less for a darker scene.")
	}
	intensity := ambient.Strength()
	if numberField(context, "Ambient intensity", &intensity, intensityLimits, "%.2f×") {
		ambient.Intensity = &intensity
		editor.SetAmbient(ambient)
	}
}
