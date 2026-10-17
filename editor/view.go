package editor

import (
	"math"
	"path"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenes"
	"github.com/akrck02/go-engine/space"
)

// The editor's window: a toolbar (the scene, save, play, code, build), the scene's nodes and the
// project's models on the left, the 3D view in the middle, the inspector on the right, and a
// status line.

// View is the editor's window and what it remembers between frames.
type View struct {
	Editor   *Editor
	viewport *viewport

	// the inspector's text fields, kept while they are typed in, for the node they were filled for
	shownFor   *engine.Node
	shownScene *engine.Scene
	nameText   string
	fieldTexts map[string]string
	sceneName  string

	dark, themeSet bool
	// DarkByDefault starts the editor dark when no mode was kept (the window's -dark).
	DarkByDefault bool

	assetTree    project.AssetEntry
	assetsReadAt time.Time
	openFolders  map[string]bool
}

// NewView shows an editor.
func NewView(editor *Editor) *View {
	editor.Wake = window.Wake
	return &View{Editor: editor, viewport: newViewport(), fieldTexts: map[string]string{}, openFolders: map[string]bool{"assets": true}}
}

// Options are the window's, with the step that shows the scene and reads what the interface
// leaves of the mouse and the keys.
func (view *View) Options(options window.Options) window.Options {
	options.Title = "go-engine — " + view.Editor.Project.Settings.Name
	options.Scene = func(state *input.State, seconds float64) *window.SceneView {
		view.keys(state)
		world, camera := view.viewport.step(view.Editor, state, seconds)
		return &window.SceneView{World: world, Camera: camera}
	}
	return options
}

func (view *View) keys(state *input.State) {
	editor := view.Editor
	if _, playing := editor.PlayingHere(); playing {
		if state.Pressed(input.KeyEscape) {
			editor.StopHere()
		}
		return
	}
	control := state.Down(input.KeyControl)
	switch {
	case control && state.Pressed(input.KeyS):
		editor.Save()
	case control && state.Pressed(input.KeyD):
		editor.Duplicate()
	case state.Pressed(input.KeyDelete):
		editor.Delete()
	case state.Pressed(input.KeyF):
		view.frame()
	case state.Pressed(input.KeyEscape):
		editor.Select(nil)
	}
}

// Build describes the window for a frame.
func (view *View) Build(context *ui.Context) {
	if !view.themeSet {
		view.themeSet = true
		kept, found := window.Preference(themePreference)
		view.applyTheme(context, found && kept == "dark" || !found && (view.DarkByDefault || window.SystemPrefersDark()))
	}
	view.Editor.TakeEvents()
	if _, playing := view.Editor.PlayingHere(); playing {
		context.RequestFrame(0) // the game moves on every frame, input or not
	}
	if view.Editor.Running() {
		context.RequestFrame(250e6) // the status line follows the game's output
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		view.toolbar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.left(context)
			context.Separator()
			view.middle(context)
			context.Separator()
			view.inspector(context)
		})
		context.Separator()
		view.statusLine(context)
	})
}

func bar(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

func (view *View) toolbar(context *ui.Context) {
	editor := view.Editor
	bar(context, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(46), Padding: ui.PadXY(10, 0), Gap: 6, AlignY: ui.Center}, func() {
		context.IconSized("gamepad-2", 20, context.Theme.Primary)
		context.Heading(editor.Project.Settings.Name)
		context.Box(ui.Layout{Width: ui.Fixed(8)}, nil)
		context.MenuButton("map", path.Base(strings.TrimSuffix(editor.SceneFile, ".scene.json")), func() {
			for _, file := range editor.Project.Scenes() {
				if context.MenuItem("map", strings.TrimSuffix(path.Base(file), ".scene.json"), "") && file != editor.SceneFile {
					view.openScene(file)
				}
			}
			if context.MenuItem("box", "New 3D scene", "") {
				view.newScene(scenes.Mode3D)
			}
			if context.MenuItem("square", "New 2D scene", "") {
				view.newScene(scenes.Mode2D)
			}
		})
		if context.IconButton("save", "Save") {
			editor.Save()
		}
		context.Tooltip("Save the scene (Ctrl+S)")
		context.Spacer()
		if _, playing := editor.PlayingHere(); playing {
			if context.PrimaryIconButton("square", "Stop") {
				editor.StopHere()
			}
			context.Tooltip("Back to editing (Escape)")
		} else if context.PrimaryIconButton("play", "Play") {
			editor.PlayHere()
		}
		context.Tooltip("Play the scene in the view, with the project's code")
		if editor.Running() && editor.playing != nil {
			if context.IconButton("square", "Window") {
				editor.Stop()
			}
			context.Tooltip("Close the game's window")
		} else if context.FlatIconButton("external-link", "") {
			editor.PlayInWindow()
		}
		context.Tooltip("Build the game and play it in a window of its own")
		reload := context.FlatIconButton
		if editor.CodeChanged() {
			reload = context.IconButton
		}
		if reload("refresh-cw", "Reload code") {
			editor.Reload()
		}
		if editor.CodeChanged() {
			context.Tooltip("The project's Go code changed: build it again and reopen the editor with it")
		} else {
			context.Tooltip("Build the project's Go code again and reopen the editor with it")
		}
		if context.IconButton("folder-open", "Open project") {
			if err := editor.Project.OpenInZed(); err != nil {
				editor.say(err.Error(), true)
			} else {
				editor.say("Opened the project in Zed", false)
			}
		}
		context.Tooltip("Open the project's folder in Zed")
		context.MenuButton("hammer", "Build", func() {
			if context.MenuItem("monitor", "For this computer", "") {
				editor.Run("Building for this computer…", func() (string, error) {
					built, err := editor.Project.BuildDesktop()
					return "Built " + built.Program, err
				})
			}
			if context.MenuItem("gamepad-2", "For the Nintendo 3DS", "") {
				editor.Run("Building for the 3DS…", func() (string, error) {
					built, err := editor.Project.Build3DS()
					return "Built " + built.Program, err
				})
			}
		})
		themeIcon, themeTip := "moon", "Dark mode"
		if view.dark {
			themeIcon, themeTip = "sun", "Light mode"
		}
		if context.FlatIconButton(themeIcon, "") {
			view.setDark(context, !view.dark)
		}
		context.Tooltip(themeTip)
	})
}

// themePreference keeps the editor's light or dark mode between runs.
const themePreference = "go-engine.theme"

// setDark switches the editor between its dark and light modes, and keeps the choice.
func (view *View) setDark(context *ui.Context, dark bool) {
	view.applyTheme(context, dark)
	mode := "light"
	if dark {
		mode = "dark"
	}
	window.KeepPreference(themePreference, mode)
}

func (view *View) applyTheme(context *ui.Context, dark bool) {
	view.dark = dark
	if dark {
		context.Theme = ui.DarkTheme()
		return
	}
	context.Theme = ui.LightTheme()
}

func (view *View) openScene(file string) {
	if err := view.Editor.OpenScene(file); err != nil {
		view.Editor.say(err.Error(), true)
	}
	view.shownFor = nil
}

func (view *View) newScene(mode scenes.Mode) {
	editor := view.Editor
	name := "Scene " + FieldText(float64(len(editor.Project.Scenes())+1))
	file, err := editor.Project.NewScene(name, mode)
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	view.openScene(file)
}

// frame turns the view to the chosen node.
func (view *View) frame() {
	if view.Editor.Scene.Is2D() {
		view.viewport.flat.frame(view.Editor.Selected)
		return
	}
	view.viewport.frame(view.Editor.Selected)
}

// left is the scene's nodes, as a tree, over the project's models.
func (view *View) left(context *ui.Context) {
	editor := view.Editor
	bar(context, ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(270), Height: ui.Grow()}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 10, Top: 8, Right: 6}, AlignY: ui.Center}, func() {
			context.Heading("Scene")
			context.Spacer()
			context.MenuButton("plus", "", func() {
				for _, kind := range editor.KindsHere() {
					if context.MenuItem(kindIcon(kind), kind, "") {
						editor.Add(kind)
					}
				}
			})
			context.Tooltip("Add a node after the chosen one")
			view.treeButton(context, "copy", "Duplicate (Ctrl+D)", editor.Duplicate)
			view.treeButton(context, "trash-2", "Delete (Delete)", editor.Delete)
		})
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 10, Right: 6}, Gap: 2}, func() {
			view.treeButton(context, "arrow-up", "Earlier among its siblings", editor.MoveUp)
			view.treeButton(context, "arrow-down", "Later among its siblings", editor.MoveDown)
			view.treeButton(context, "indent-increase", "Hang it from the node above", editor.Nest)
			view.treeButton(context, "indent-decrease", "Hang it from its parent's parent", editor.Unnest)
		})
		tree := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.GrowBy(3), Padding: ui.PadXY(8, 4), Gap: 1}
		context.ScrollArea("Tree", tree, func() {
			for _, child := range editor.Scene.Root.Children() {
				view.treeNode(context, child, 0)
			}
		})
		context.Separator()
		view.assets(context)
	})
}

func (view *View) treeButton(context *ui.Context, icon, tip string, action func()) {
	if context.FlatIconButton(icon, "") {
		action()
	}
	context.Tooltip(tip)
}

func (view *View) treeNode(context *ui.Context, node *engine.Node, depth int) {
	context.Push(node.Name)
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 16}}, func() {
		if context.Selectable(nodeIcon(node), node.Name, node == view.Editor.Selected) {
			view.Editor.Select(node)
		}
	})
	for _, child := range node.Children() {
		view.treeNode(context, child, depth+1)
	}
	context.Pop()
}

func kindIcon(kind string) string {
	switch kind {
	case "Cube":
		return "box"
	case "Sphere":
		return "circle"
	case "Plane":
		return "square"
	case "Cylinder":
		return "cylinder"
	case "Camera":
		return "video"
	case "Square":
		return "square"
	case "Circle":
		return "circle"
	}
	return "crosshair"
}

func nodeIcon(node *engine.Node) string {
	if _, isCamera := node.Camera(); isCamera {
		return "video"
	}
	if node.Model() != "" {
		return "package"
	}
	if node.Sprite() != "" {
		return "image"
	}
	if node.Shape() != "" {
		return kindIcon(strings.ToUpper(node.Shape()[:1]) + node.Shape()[1:])
	}
	return "crosshair"
}

// assets shows the project's assets folder as a tree: whatever is copied there, from the file
// manager or the code editor, shows here. A model clicked is placed in the scene.
func (view *View) assets(context *ui.Context) {
	if time.Since(view.assetsReadAt) > 2*time.Second {
		view.assetTree, view.assetsReadAt = view.Editor.Project.AssetTree(), time.Now()
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.GrowBy(2), Padding: ui.Pad(10), Gap: 1}
	context.ScrollArea("Assets", layout, func() {
		context.Row(func() {
			context.Heading("Assets")
			context.Spacer()
			if context.FlatIconButton("refresh-cw", "") {
				view.assetsReadAt = time.Time{}
			}
			context.Tooltip("Read the assets folder again")
		})
		if len(view.assetTree.Children) == 0 {
			note(context, "The assets folder is empty: copy models (.glb) and pictures (.png) into "+view.Editor.Project.Path("assets"))
		}
		for _, child := range view.assetTree.Children {
			view.assetNode(context, child, 0)
		}
	})
}

func (view *View) assetNode(context *ui.Context, entry project.AssetEntry, depth int) {
	editor := view.Editor
	context.Push(entry.Path)
	defer context.Pop()
	indent := ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 14}}
	if entry.Folder {
		open := view.openFolders[entry.Path]
		icon := "folder"
		if open {
			icon = "folder-open"
		}
		context.Box(indent, func() {
			if context.Selectable(icon, entry.Name, false) {
				view.openFolders[entry.Path] = !open
			}
		})
		if open {
			for _, child := range entry.Children {
				view.assetNode(context, child, depth+1)
			}
		}
		return
	}
	kind := project.KindOf(entry.Path)
	context.Box(indent, func() {
		if context.Selectable(assetIcon(kind), entry.Name, false) {
			switch {
			case kind == project.ModelAsset && !editor.Scene.Is2D():
				editor.AddModel(entry.Path)
			case kind == project.ImageAsset && editor.Scene.Is2D():
				editor.AddSprite(entry.Path)
			case kind == project.ModelAsset:
				editor.say(entry.Path+": a model goes in a 3D scene", false)
			case kind == project.ImageAsset:
				editor.say(entry.Path+": a picture goes in a 2D scene, as a sprite", false)
			default:
				editor.say(entry.Path+": not something a scene shows yet", false)
			}
		}
	})
	switch kind {
	case project.ModelAsset, project.ImageAsset:
		context.Tooltip("Place " + entry.Path + " in the scene")
	default:
		context.Tooltip(entry.Path)
	}
}

func assetIcon(kind project.AssetKind) string {
	switch kind {
	case project.ModelAsset:
		return "package"
	case project.ImageAsset:
		return "image"
	case project.SoundAsset:
		return "music"
	}
	return "file"
}

// middle is the 3D view, seen through the window, with a strip of how to see it.
func (view *View) middle(context *ui.Context) {
	if played, playing := view.Editor.PlayingHere(); playing {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(8)}, func() {
			context.Frame(ui.Layout{Padding: ui.PadXY(10, 6), Gap: 6, AlignY: ui.Center}, ui.Look{Fill: context.Theme.Surface, Radius: 8, Shadow: true}, func() {
				context.IconSized("play", 16, context.Theme.Positive)
				context.Text("Playing", ui.TextStyle{Weight: draw2d.Medium})
				context.Caption("Escape stops")
			})
			if played.Is2D() {
				context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
					paintPlaying(played, area, list)
				})
			}
		})
		return
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(8)}, func() {
		context.Frame(ui.Layout{Padding: ui.PadXY(6, 4), Gap: 6, AlignY: ui.Center}, ui.Look{Fill: context.Theme.Surface, Radius: 8, Shadow: true}, func() {
			dimensions := 0
			if view.Editor.Scene.Is2D() {
				dimensions = 1
			}
			context.Box(ui.Layout{Width: ui.Fixed(110)}, func() {
				if context.Tabs("Dimensions", &dimensions, "3D", "2D") {
					view.Editor.SetMode([]scenes.Mode{scenes.Mode3D, scenes.Mode2D}[dimensions])
				}
			})
			context.Tooltip("Whether the scene is 3D (shapes and models) or 2D (sprites and flat shapes)")
			if !view.Editor.Scene.Is2D() {
				mode := int(view.viewport.mode)
				context.Box(ui.Layout{Width: ui.Fixed(330)}, func() {
					if context.Tabs("View", &mode, "Perspective", "Orthographic", "Isometric") {
						view.viewport.setMode(ViewMode(mode))
					}
				})
			}
			if context.FlatIconButton("scan", "") {
				view.frame()
			}
			context.Tooltip("Frame the chosen node (F)")
			if view.viewport.through != nil {
				if context.FlatIconButton("video-off", "") {
					view.viewport.through = nil
				}
				context.Tooltip("Back to the editor's camera")
			}
		})
		if view.Editor.Scene.Is2D() {
			context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
				view.viewport.flat.paint(view.Editor, area, list)
			})
			context.Caption("Drag the background to move · scroll to zoom · click to choose · drag the chosen to move it")
			return
		}
		context.Spacer()
		hint := "Drag to turn · right-drag to pan · scroll to zoom · click to choose · drag the chosen to move it"
		if view.viewport.mode == ViewIsometric {
			hint = "Isometric is a fixed angle: drag to pan · scroll to zoom · click to choose · drag the chosen to move it"
		}
		context.Text(hint,
			ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
	})
}

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
	if context.Checkbox("Daylight sky", ptr(sky == "")) {
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
			perMetre = scenes.DefaultPixelsPerMetre
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

func ptr(value bool) *bool { return &value }

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
	angles := node.Transform.Euler()
	turned := space.Vec3{X: angles.X, Y: angles.Y, Z: angles.Z}
	if !view.Editor.Scene.Is2D() && vectorRow(context, "Rotation", &turned, rotationLimits, "%.0f°") {
		node.Transform.SetEuler(space.Euler{X: turned.X, Y: turned.Y, Z: turned.Z})
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
			node.Transform.Scale = space.V(uniform, uniform, uniform)
			editor.Scene.Sync()
			editor.Changed()
		}
	})
	if world := node.World().Position; node.Parent() != editor.Scene.Root {
		context.Caption("In the world: " + formatVector(world) + " m")
	}
}

// vectorRow edits the three axes of a vector; it tells whether one changed.
func vectorRow(context *ui.Context, name string, value *space.Vec3, limits numberLimits, format string) bool {
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

func formatVector(vector space.Vec3) string {
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
	for index, each := range scenes.Shapes {
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
		node.SetShape(scenes.Shapes[shape], node.Colour())
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
			perMetre = scenes.DefaultPixelsPerMetre
		}
		note(context, trimmed(size.X)+" × "+trimmed(size.Y)+" m at "+trimmed(perMetre)+" pixels a metre, times its scale")
	}
	if context.Button("Read the picture again") {
		editor.Scene.Forget(node.Sprite())
		if err := node.SetSprite(node.Sprite()); err != nil {
			editor.say(err.Error(), true)
		}
	}
}

func (view *View) cameraInspector(context *ui.Context, node *engine.Node, settings *scenes.Camera) {
	editor := view.Editor
	context.Separator()
	context.Heading("Camera")
	projection := 0
	for index, each := range scenes.Projections {
		if each == settings.Projection {
			projection = index
		}
	}
	if editor.Scene.Is2D() {
		settings.Projection = scenes.Orthographic
		if settings.Height == 0 {
			settings.Height = 10
		}
		context.Caption("Metres seen, top to bottom")
		if numberField(context, "Height seen", &settings.Height, numberLimits{0, 1000, 0.5}, "%.1f m") {
			editor.Changed()
		}
		if context.Checkbox("The scene is seen through this camera", &settings.Main) {
			editor.Changed()
		}
		return
	}
	if context.Tabs("Projection", &projection, "Perspective", "Orthographic", "Isometric") {
		settings.Projection = scenes.Projections[projection]
		if settings.Projection == scenes.Isometric {
			node.Transform.SetEuler(space.Euler{X: scenes.IsometricAngles[0], Y: scenes.IsometricAngles[1]})
			editor.Scene.Sync()
		}
		editor.Changed()
	}
	switch settings.Projection {
	case scenes.Perspective:
		if settings.FieldOfView == 0 {
			settings.FieldOfView = 45
		}
		context.Caption("Field of view, top to bottom")
		if numberField(context, "Field of view", &settings.FieldOfView, numberLimits{10, 150, 1}, "%.0f°") {
			editor.Changed()
		}
	default:
		if settings.Height == 0 {
			settings.Height = 10
		}
		context.Caption("Metres seen, top to bottom")
		if numberField(context, "Height seen", &settings.Height, numberLimits{0, 1000, 0.5}, "%.1f m") {
			editor.Changed()
		}
		if settings.Projection == scenes.Isometric {
			note(context, "An isometric camera always looks from the classic angle, 45° round and 35.26° down; its rotation is not used, only its position and the height it sees.")
		}
	}
	if context.Checkbox("The scene is seen through this camera", &settings.Main) {
		editor.Changed()
	}
	if context.Button("Look through it") {
		view.viewport.through = node
	}
}

func (view *View) behaviourInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Separator()
	context.Row(func() {
		context.Heading("Behaviours")
		context.Spacer()
		context.MenuButton("plus", "Add", func() {
			for _, info := range editor.Behaviours() {
				if context.MenuItem("code", info.Name, info.File) {
					editor.AttachBehaviour(info.Name)
				}
			}
			if context.MenuItem("refresh-cw", "Read the code again", "") {
				editor.ReadBehaviours()
				editor.say("Read the project's code: "+FieldText(float64(len(editor.Behaviours())))+" behaviours", false)
			}
		})
	})
	if len(node.Behaviours) == 0 {
		note(context, "The project's Go code moves nodes through behaviours: add one, or write a type with an Update method (see spin.go).")
	}
	for index := range node.Behaviours {
		behaviour := &node.Behaviours[index]
		context.Push("behaviour " + behaviour.Name + FieldText(float64(index)))
		look := ui.Look{Fill: context.Theme.Background, Radius: 8, Border: context.Theme.Line, BorderWidth: 1}
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(8), Gap: 4}, look, func() {
			info, known := editor.BehaviourInfo(behaviour.Name)
			context.Row(func() {
				context.IconSized("code", 16, context.Theme.Primary)
				context.Text(behaviour.Name, ui.TextStyle{Weight: draw2d.Bold})
				context.Spacer()
				if context.FlatIconButton("x", "") {
					editor.DetachBehaviour(index)
				}
			})
			if !known {
				context.Text("Not in the project's code (yet?)", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
			} else if info.Doc != "" {
				note(context, info.Doc)
			}
			view.behaviourFields(context, behaviour, info.Fields)
		})
		context.Pop()
	}
}

// behaviourFields edits a behaviour's fields: the ones its code declares, then any others the
// scene holds.
func (view *View) behaviourFields(context *ui.Context, behaviour *scenes.Behaviour, declared []projectField) {
	shown := map[string]bool{}
	for _, field := range declared {
		shown[field.Name] = true
		view.behaviourField(context, behaviour, field.Name, field.Type+fieldComment(field.Comment))
	}
	for name := range behaviour.Fields {
		if !shown[name] {
			view.behaviourField(context, behaviour, name, "not in the code")
		}
	}
}

func fieldComment(comment string) string {
	if comment == "" {
		return ""
	}
	return " · " + comment
}

func (view *View) behaviourField(context *ui.Context, behaviour *scenes.Behaviour, name, hint string) {
	key := behaviour.Name + "." + name
	text, editing := view.fieldTexts[key]
	if !editing {
		text = FieldText(behaviour.Fields[name])
	}
	context.Caption(name + " (" + hint + ")")
	if context.TextField(key, &text) {
		view.fieldTexts[key] = text
		if behaviour.Fields == nil {
			behaviour.Fields = map[string]any{}
		}
		if strings.TrimSpace(text) == "" {
			delete(behaviour.Fields, name)
		} else {
			behaviour.Fields[name] = FieldValue(text)
		}
		view.Editor.Changed()
	}
}

// projectField is a behaviour's field, as the project's code declares it.
type projectField = project.FieldInfo

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

func (view *View) statusLine(context *ui.Context) {
	editor := view.Editor
	bar(context, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(28), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}, func() {
		style := ui.TextStyle{Color: context.Theme.InkSoft}
		if editor.Failed {
			style.Color = context.Theme.Danger
		}
		context.Text(editor.Message, style)
		if editor.output != nil && editor.Running() {
			if last := editor.output.Last(); last != "" {
				context.Text("· "+last, ui.TextStyle{Color: context.Theme.InkQuiet})
			}
		}
		context.Spacer()
		if editor.Saved {
			context.Text("Saved", ui.TextStyle{Color: context.Theme.Positive})
		} else {
			context.Text("Unsaved changes", ui.TextStyle{Color: context.Theme.Warning})
		}
	})
}
