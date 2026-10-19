package editor

import (
	"strconv"
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// The inspector's sections for the parts of a 2D scene: a node's body, its sprite's sheet and
// animations, and what a camera follows.

var (
	bodySizeLimits     = numberLimits{0, 1000, 0.05}
	gravityScaleLimits = numberLimits{-10, 10, 0.05}
	zonePullLimits     = numberLimits{-100, 100, 0.5}
	sheetCellLimits    = numberLimits{1, 256, 1}
	paceLimits         = numberLimits{1, 120, 1}
	smoothingLimits    = numberLimits{0, 10, 0.05}
)

// newCameraBounds is the box a camera keeps its view in when it is first given one of its own:
// twenty metres a side round the origin.
var newCameraBounds = [4]float64{-10, -10, 10, 10}

// bodyKindNotes say what each kind of body does, in the order of scenefile.BodyKinds.
var bodyKindNotes = []string{
	"Solid: never moves; moving bodies stand on it and stop against it.",
	"Moving: falls with gravity, moves with its velocity, and is stopped by solids.",
	"Trigger: stops nothing; its behaviours hear (Touched) when a moving body begins touching it.",
	"Pushable: falls as a moving body does; moving bodies stand on it, and push it by walking into it.",
	"Platform: moves at the velocity its behaviours give it, through solids and with no gravity, carrying what stands on it.",
}

func (view *View) bodyInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	body, found := node.BodySettings()
	context.Separator()
	context.Row(func() {
		context.Heading("Body")
		context.Spacer()
		if found && context.FlatIconButton(iconRemove, "") {
			editor.TakeBody()
		}
		if found {
			context.Tooltip("Remove the body")
		}
	})
	if !found {
		note(context, "A body puts the node in the scene's physics: ground to stand on, something that falls, or a trigger that notices what touches it.")
		if context.Button("Give it a body") {
			editor.GiveBody(scenefile.Solid)
		}
		return
	}
	kind := max(0, indexOf(scenefile.BodyKinds, body.Kind))
	if context.Tabs("Body kind", &kind, "Solid", "Moving", "Trigger", "Pushable", "Platform") {
		editor.GiveBody(scenefile.BodyKinds[kind])
	}
	note(context, bodyKindNotes[kind])
	view.bodySizeFields(context, body)
	gravity := body.Gravity()
	if body.Kind == scenefile.Moving || body.Kind == scenefile.Pushable {
		context.Caption("Gravity: 1 pulls fully, 0 not at all")
		if numberField(context, "Gravity scale", &gravity, gravityScaleLimits, "%.2f×") {
			editor.SetGravityScale(gravity)
		}
	}
	if body.Kind == scenefile.Trigger {
		view.zonePullFields(context, body)
	}
}

// zonePullFields give a trigger a gravity of its own, or leave the world's.
func (view *View) zonePullFields(context *ui.Context, body *scenefile.Body) {
	editor := view.Editor
	pulls := body.Pull != nil
	if context.Checkbox("A gravity of its own, for what falls inside", &pulls) {
		if pulls {
			editor.SetZonePull(&[2]float64{0, engine.DefaultGravity})
		} else {
			editor.SetZonePull(nil)
		}
	}
	if body.Pull == nil {
		return
	}
	pull := *body.Pull
	context.Caption("Pull across and up, in metres a second each second (the world's is 0, -9.81)")
	across := numberField(context, "Pull across", &pull[0], zonePullLimits, "→ %.1f")
	up := numberField(context, "Pull up", &pull[1], zonePullLimits, "↑ %.1f")
	if across || up {
		editor.SetZonePull(&pull)
	}
}

// bodySizeFields edit a body's width and height; zero is the node's own size.
func (view *View) bodySizeFields(context *ui.Context, body *scenefile.Body) {
	context.Caption("Size in metres; 0 is the node's own")
	context.Push("Body size")
	context.Grid(2, func() {
		if numberField(context, "Width", &body.Size[0], bodySizeLimits, "Width %.2f") {
			view.Editor.Changed()
		}
		if numberField(context, "Height", &body.Size[1], bodySizeLimits, "Height %.2f") {
			view.Editor.Changed()
		}
	})
	context.Pop()
}

func indexOf[Item comparable](items []Item, wanted Item) int {
	for index, item := range items {
		if item == wanted {
			return index
		}
	}
	return -1
}

// forgetAnimationTexts drops what was being typed in the animations' fields.
func (view *View) forgetAnimationTexts() {
	for key := range view.fieldTexts {
		if strings.HasPrefix(key, "animation ") {
			delete(view.fieldTexts, key)
		}
	}
}

func (view *View) animationNameField(context *ui.Context, index int, name string) {
	key := "animation name " + strconv.Itoa(index)
	text, editing := view.fieldTexts[key]
	if !editing {
		text = name
	}
	if context.TextField(key, &text) {
		view.fieldTexts[key] = text
		view.Editor.RenameAnimation(index, text)
	}
}

// followInspector chooses what a 2D camera follows, how closely, and the box its view keeps in.
func (view *View) followInspector(context *ui.Context, camera *engine.Node, settings *scenefile.Camera) {
	editor := view.Editor
	options := append([]string{"Nothing"}, editor.Followable(camera)...)
	followed := max(0, indexOf(options, settings.Follow))
	context.Caption("Follows")
	if context.Select("Follows", &followed, options) {
		settings.Follow = ""
		if followed > 0 {
			settings.Follow = options[followed]
		}
		editor.Changed()
	}
	if settings.Follow == "" {
		return
	}
	context.Caption("Seconds it takes to catch up; 0 stays right on it")
	if numberField(context, "Smoothing", &settings.SmoothingSeconds, smoothingLimits, "%.2f s") {
		editor.Changed()
	}
	view.followBoundsFields(context, settings)
}

// followBoundsFields keep a following camera's view inside the tile maps, or a box of its own.
func (view *View) followBoundsFields(context *ui.Context, settings *scenefile.Camera) {
	editor := view.Editor
	own := 0
	if settings.Bounds != nil {
		own = 1
	}
	context.Caption("Its view keeps inside")
	if context.Tabs("Keep inside", &own, "The tile maps", "A box of its own") {
		settings.Bounds = nil
		if own == 1 {
			bounds := newCameraBounds // a copy: each camera edits its own
			settings.Bounds = &bounds
		}
		editor.Changed()
	}
	if settings.Bounds == nil {
		if context.Checkbox("The top of the tile maps is a ceiling", &settings.Ceiling) {
			editor.Changed()
		}
		note(context, "Without a ceiling the sky stays open: the camera rises with what it follows.")
		return
	}
	context.Push("Bounds")
	context.Grid(2, func() {
		names := []string{"Lowest X", "Lowest Y", "Highest X", "Highest Y"}
		for index, name := range names {
			if numberField(context, name, &settings.Bounds[index], positionLimits, name+" %.2f") {
				editor.Changed()
			}
		}
	})
	context.Pop()
}
