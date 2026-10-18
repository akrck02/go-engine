package editor

import (
	"math"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

func (view *View) inspector(context *ui.Context) {
	editor := view.Editor
	if editor.Selected != view.shownFor || editor.Scene != view.shownScene {
		view.shownFor, view.shownScene = editor.Selected, editor.Scene
		view.fieldTexts = map[string]string{}
		if editor.Selected != nil {
			view.nameText = editor.Selected.Name
		}
		view.sceneName = editor.Scene.Name
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(320), Height: ui.Grow()}
	bar(context, layout, func() {
		inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(12), Gap: 6}
		context.ScrollArea("Inspector", inside, func() {
			if _, playing := editor.PlayingHere(); playing {
				context.Heading("Playing")
				note(context, "The view plays a copy of the scene with the project's code: what moves now is not saved. Stop (or Escape) goes back to editing.")
				return
			}
			if editor.Selected == nil {
				view.sceneInspector(context)
				return
			}
			view.nodeInspector(context, editor.Selected)
		})
	})
}

func (view *View) sceneInspector(context *ui.Context) {
	editor := view.Editor
	context.Heading("Scene")
	context.Caption(editor.SceneFile)
	context.Caption("Name")
	if context.TextField("Scene name", &view.sceneName) && strings.TrimSpace(view.sceneName) != "" {
		editor.Scene.Name = strings.TrimSpace(view.sceneName)
		editor.Changed()
	}
	context.Caption("Sky: a colour, or a daylight sky")
	sky := editor.Scene.Sky
	if context.Checkbox("Daylight sky", pointerTo(sky == "")) {
		if sky == "" {
			editor.Scene.SetSky("#20242c")
		} else {
			editor.Scene.SetSky("")
		}
		editor.Changed()
	}
	if editor.Scene.Sky != "" && colourPicker(context, "Sky", &sky) {
		editor.Scene.SetSky(sky)
		editor.Changed()
	}
	if editor.Scene.Is2D() {
		perMetre := editor.Scene.PixelsPerMetre
		if perMetre <= 0 {
			perMetre = scenefile.DefaultPixelsPerMetre
		}
		context.Caption("Pixels of a sprite that make a metre")
		if numberField(context, "Pixels per metre", &perMetre, numberLimits{1, 1024, 1}, "%.0f px") {
			editor.Scene.PixelsPerMetre = perMetre
			editor.Changed()
		}
	}
	context.Separator()
	note(context, "Coordinates: one unit is a metre; Y is up; X to the right and Z towards you (right-handed, as glTF). Rotations are in degrees: X tilts, Y turns, Z rolls.")
	note(context, "Choose a node in the view or in the list to see its transform.")
}

func pointerTo(value bool) *bool { return &value }

// note writes a quiet text that goes on to the next line where the panel ends: a word per box, in
// a box that wraps them.
func note(context *ui.Context, text string) {
	style := ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize}
	context.Box(ui.Layout{Width: ui.Grow(), Wrap: true, Gap: 3}, func() {
		for _, word := range strings.Fields(text) {
			context.Text(word, style)
		}
	})
}

func (view *View) nodeInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Row(func() {
		context.IconSized(nodeIcon(node), 18, context.Theme.Primary)
		context.Heading(node.Name)
	})
	context.Caption(node.Path())
	context.Caption("Name")
	if context.TextField("Node name", &view.nameText) {
		editor.Rename(view.nameText)
	}
	if context.Checkbox("Visible", &node.Visible) {
		editor.Scene.Sync()
		editor.Changed()
	}
	view.transformInspector(context, node)
	if node.Shape() != "" {
		view.shapeInspector(context, node)
	}
	if node.Model() != "" {
		context.Separator()
		context.Heading("Model")
		context.Caption(node.Model())
		if context.Button("Read the file again") {
			editor.Scene.Forget(node.Model())
			if err := node.SetModel(node.Model()); err != nil {
				editor.say(err.Error(), true)
			}
		}
	}
	if node.Sprite() != "" {
		view.spriteInspector(context, node)
	}
	if settings, isCamera := node.Camera(); isCamera {
		view.cameraInspector(context, node, settings)
	}
	if editor.Scene.Is2D() {
		view.bodyInspector(context, node)
	}
	view.behaviourInspector(context, node)
}

// transformInspector edits where the node is, how it is turned and how large, relative to its
// parent.
func (view *View) transformInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Separator()
	context.Heading("Transform")
	note(context, "Relative to its parent: position in metres, rotation in degrees, scale as times its size.")
	position := node.Transform.Position
	if vectorRow(context, "Position", &position, positionLimits, "%.2f") {
		node.Transform.Position = position
		editor.Scene.Sync()
		editor.Changed()
	}
	if view.Editor.Scene.Is2D() {
		note(context, "In 2D, Z is the depth: a greater Z is drawn in front. Turned sprites are not drawn turned yet.")
	}
	angles := node.Transform.Angles()
	turned := space.Vector{X: angles.X, Y: angles.Y, Z: angles.Z}
	if !view.Editor.Scene.Is2D() && vectorRow(context, "Rotation", &turned, rotationLimits, "%.0f°") {
		node.Transform.SetAngles(space.Angles{X: turned.X, Y: turned.Y, Z: turned.Z})
		editor.Scene.Sync()
		editor.Changed()
	}
	scale := node.Transform.Scale
	if vectorRow(context, "Scale", &scale, scaleLimits, "%.2f×") {
		node.Transform.Scale = scale
		editor.Scene.Sync()
		editor.Changed()
	}
	uniform := node.Transform.Scale.X
	context.Row(func() {
		context.Caption("All axes")
		if numberField(context, "Uniform scale", &uniform, scaleLimits, "%.2f×") {
			node.Transform.Scale = space.NewVector(uniform, uniform, uniform)
			editor.Scene.Sync()
			editor.Changed()
		}
	})
	if world := node.World().Position; node.Parent() != editor.Scene.Root {
		context.Caption("In the world: " + formatVector(world) + " m")
	}
}

// vectorRow edits the three axes of a vector; it tells whether one changed.
func vectorRow(context *ui.Context, name string, value *space.Vector, limits numberLimits, format string) bool {
	changed := false
	context.Caption(name)
	context.Push(name)
	context.Grid(3, func() {
		axes := []struct {
			label string
			value *float64
		}{{"X", &value.X}, {"Y", &value.Y}, {"Z", &value.Z}}
		for _, axis := range axes {
			if numberField(context, axis.label, axis.value, limits, axis.label+" "+format) {
				changed = true
			}
		}
	})
	context.Pop()
	return changed
}

// numberLimits are a number field's range and step. go-ui's NumberField keeps a value on the
// steps counted from the minimum, in float32: the minimums here are whole steps, and moderate, so
// a value read back is the value written.
type numberLimits struct{ minimum, maximum, step float32 }

var (
	positionLimits = numberLimits{-5000, 5000, 0.05}
	rotationLimits = numberLimits{-360, 360, 1}
	scaleLimits    = numberLimits{0, 1000, 0.05}
)

// numberField edits a number; it tells whether it changed by at least half a step, so the
// rounding of float32 is not taken for an edit.
func numberField(context *ui.Context, name string, value *float64, limits numberLimits, format string) bool {
	shown := float32(*value) + 0 // + 0: no -0 shown
	if !context.NumberField(name, &shown, limits.minimum, limits.maximum, limits.step, format) {
		return false
	}
	if math.Abs(float64(shown)-*value) < float64(limits.step)/2 {
		return false
	}
	*value = math.Round(float64(shown)/float64(limits.step)) * float64(limits.step)
	return true
}

func formatVector(vector space.Vector) string {
	return trimmed(vector.X) + ", " + trimmed(vector.Y) + ", " + trimmed(vector.Z)
}

func trimmed(value float64) string {
	text := FieldText(float64(int64(value*100)) / 100)
	return text
}

func (view *View) shapeInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Separator()
	context.Heading("Shape")
	shape := 0
	for index, each := range scenefile.Shapes {
		if each == node.Shape() {
			shape = index
		}
	}
	if editor.Scene.Is2D() {
		round := 0
		if node.Shape() == "sphere" {
			round = 1
		}
		if context.Tabs("Flat shape", &round, "Square", "Circle") {
			node.SetShape([]string{"cube", "sphere"}[round], node.Colour())
			editor.Changed()
		}
	} else if context.Tabs("Shape", &shape, "Cube", "Sphere", "Plane", "Cylinder") {
		node.SetShape(scenefile.Shapes[shape], node.Colour())
		editor.Changed()
	}
	colour := node.Colour()
	if colourPicker(context, "Colour", &colour) {
		node.SetShape(node.Shape(), colour)
		editor.Changed()
	}
}

func (view *View) spriteInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Separator()
	context.Heading("Sprite")
	context.Caption(node.Sprite())
	if size := node.Bounds().Size(); size.X > 0 {
		perMetre := editor.Scene.PixelsPerMetre
		if perMetre <= 0 {
			perMetre = scenefile.DefaultPixelsPerMetre
		}
		note(context, trimmed(size.X)+" × "+trimmed(size.Y)+" m at "+trimmed(perMetre)+" pixels a metre, times its scale")
	}
	if context.Button("Read the picture again") {
		editor.Scene.Forget(node.Sprite())
		if err := node.SetSprite(node.Sprite()); err != nil {
			editor.say(err.Error(), true)
		}
	}
	if editor.Scene.Is2D() {
		view.sheetInspector(context, node)
	}
}

func (view *View) cameraInspector(context *ui.Context, node *engine.Node, settings *scenefile.Camera) {
	context.Separator()
	context.Heading("Camera")
	if view.Editor.Scene.Is2D() {
		settings.Projection = scenefile.Orthographic
		view.heightSeenField(context, settings)
		view.mainCameraCheckbox(context, settings)
		view.followInspector(context, node, settings)
		return
	}
	view.projectionTabs(context, node, settings)
	if settings.Projection == scenefile.Perspective {
		view.fieldOfViewField(context, settings)
	} else {
		view.heightSeenField(context, settings)
	}
	if settings.Projection == scenefile.Isometric {
		note(context, "An isometric camera always looks from the classic angle, 45° round and 35.26° down; its rotation is not used, only its position and the height it sees.")
	}
	view.mainCameraCheckbox(context, settings)
	if context.Button("Look through it") {
		view.viewport.through = node
	}
}

// projectionTabs chooses a 3D camera's projection; an isometric one turns to its angle.
func (view *View) projectionTabs(context *ui.Context, node *engine.Node, settings *scenefile.Camera) {
	projection := 0
	for index, each := range scenefile.Projections {
		if each == settings.Projection {
			projection = index
		}
	}
	if !context.Tabs("Projection", &projection, "Perspective", "Orthographic", "Isometric") {
		return
	}
	settings.Projection = scenefile.Projections[projection]
	if settings.Projection == scenefile.Isometric {
		node.Transform.SetAngles(space.Angles{X: scenefile.IsometricAngles[0], Y: scenefile.IsometricAngles[1]})
		view.Editor.Scene.Sync()
	}
	view.Editor.Changed()
}

// fieldOfViewField sets how wide a perspective camera sees, top to bottom.
func (view *View) fieldOfViewField(context *ui.Context, settings *scenefile.Camera) {
	if settings.FieldOfView == 0 {
		settings.FieldOfView = 45
	}
	context.Caption("Field of view, top to bottom")
	if numberField(context, "Field of view", &settings.FieldOfView, numberLimits{10, 150, 1}, "%.0f°") {
		view.Editor.Changed()
	}
}

// heightSeenField sets how many metres an orthographic camera sees, top to bottom.
func (view *View) heightSeenField(context *ui.Context, settings *scenefile.Camera) {
	if settings.Height == 0 {
		settings.Height = 10
	}
	context.Caption("Metres seen, top to bottom")
	if numberField(context, "Height seen", &settings.Height, numberLimits{0, 1000, 0.5}, "%.1f m") {
		view.Editor.Changed()
	}
}

func (view *View) mainCameraCheckbox(context *ui.Context, settings *scenefile.Camera) {
	if context.Checkbox("The scene is seen through this camera", &settings.Main) {
		view.Editor.Changed()
	}
}

// Swatches are the colours offered besides typing one.
var swatches = []string{
	"#e07a5f", "#f2cc8f", "#81b29a", "#3d405b", "#7a9a6a", "#c8c8c8",
	"#264653", "#2a9d8f", "#e9c46a", "#f4a261", "#e76f51", "#20242c",
}

// colourPicker offers the swatches and a field to type a colour in; it tells whether it changed.
func colourPicker(context *ui.Context, name string, colour *string) bool {
	changed := false
	context.Push(name)
	context.Grid(6, func() {
		for _, each := range swatches {
			context.Push(each)
			look := ui.Look{Fill: draw2d.Hex(each), Radius: 4, Border: context.Theme.ControlLine, BorderWidth: 1}
			if each == *colour {
				look.Border, look.BorderWidth = context.Theme.Primary, 3
			}
			if context.Clickable("swatch", ui.Layout{Width: ui.Grow(), Height: ui.Fixed(22)}, ui.Looks{Normal: look, Hovered: look, Pressed: look}, nil) {
				*colour, changed = each, true
			}
			context.Pop()
		}
	})
	typed := *colour
	if context.TextField("Hex", &typed) {
		if _, err := draw2d.ParseHex(typed); err == nil {
			*colour, changed = typed, true
		}
	}
	context.Pop()
	return changed
}
