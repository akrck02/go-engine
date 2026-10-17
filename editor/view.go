package editor

import (
	"path/filepath"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/export"
	"github.com/akrck02/go-engine/game"
)

// The editor's window, top to bottom: a toolbar (the title, save, play, export), then the scenes
// and kinds on the left, the scene in the middle, the inspector on the right, and a status line.

// Swatches are the colours the inspector offers, besides typing one.
var swatches = []string{
	"#2f4858", "#33658a", "#86bbd8", "#a9d6e5", "#cfe1b9", "#6b705c",
	"#3d405b", "#8d99ae", "#b08968", "#f6bd60", "#e76f51", "#e63946",
	"#ffffff", "#edede9", "#adb5bd", "#495057", "#212529", "#000000",
}

// Icons are the icons a kind may wear, a few that read well in a small cell.
var Icons = []string{
	"", "user", "ghost", "cat", "bug", "fish", "heart", "star", "key", "lock", "door-open", "flag",
	"package", "box", "gem", "coins", "crown", "trophy", "lamp", "flame", "zap", "droplet", "leaf",
	"tree-pine", "mountain", "house", "anchor", "skull", "bomb", "sword", "shield", "sun", "moon",
}

// View is the editor's window: what Run shows, and what it remembers between frames.
type View struct {
	Editor *Editor

	tab       int // the inspector's tab: the scene or the kind
	nameTaken bool

	canvas        game.Grid // where the scene's cells were last drawn
	canvasShown   bool
	wasPressing   bool
	paintingDrag  bool // the press began on the scene: dragging paints
	hoverColumn   int
	hoverRow      int
	walking       bool
	exports       chan exportDone
	exportRunning bool
}

type exportDone struct {
	result export.Result
	err    error
}

// NewView shows an editor.
func NewView(editor *Editor) *View {
	return &View{Editor: editor, exports: make(chan exportDone, 1), hoverColumn: -1}
}

// Options are the window's, with the step that reads the keys the interface leaves: the arrows
// while playing, Ctrl+S to save, Escape to stop playing.
func (view *View) Options(options window.Options) window.Options {
	options.Title = "go-engine"
	options.Scene = func(state *input.State, seconds float64) *window.SceneView {
		view.keys(state, seconds)
		return nil
	}
	return options
}

func (view *View) keys(state *input.State, seconds float64) {
	editor := view.Editor
	if state.Down(input.KeyControl) && state.Pressed(input.KeyS) {
		editor.Save()
	}
	session, playing := editor.Playing()
	view.walking = false
	if !playing {
		return
	}
	if state.Pressed(input.KeyEscape) {
		editor.Stop()
		return
	}
	view.walking = session.Step(state, seconds)
}

// Build describes the window for a frame.
func (view *View) Build(context *ui.Context) {
	view.collectExport()
	if view.walking || view.exportRunning {
		context.RequestFrame(0)
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		view.toolbar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.sidebar(context)
			context.Separator()
			view.middle(context)
			context.Separator()
			view.inspector(context)
		})
		context.Separator()
		view.statusLine(context)
	})
}

// bar is a strip painted with the surface colour.
func bar(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

func (view *View) toolbar(context *ui.Context) {
	editor := view.Editor
	bar(context, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(48), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}, func() {
		context.IconSized("gamepad-2", 20, context.Theme.Primary)
		context.Box(ui.Layout{Width: ui.Fixed(240)}, func() {
			if context.TextField("Title", &editor.Game.Title) {
				editor.changed()
			}
		})
		if context.IconButton("save", "Save") {
			editor.Save()
		}
		context.Tooltip("Save the game (Ctrl+S) to " + editor.Path)
		context.Spacer()
		if _, playing := editor.Playing(); playing {
			if context.PrimaryIconButton("square", "Stop") {
				editor.Stop()
			}
			context.Tooltip("Back to editing (Escape)")
		} else if context.PrimaryIconButton("play", "Play") {
			editor.Play()
		}
		if context.IconButton("monitor", "Desktop") {
			view.startExport(export.Desktop)
		}
		context.Tooltip("Export a program for this computer")
		if context.IconButton("gamepad-2", "3DS") {
			view.startExport(export.Nintendo3DS)
		}
		context.Tooltip("Export a .3dsx for the Nintendo 3DS (needs TinyGo with the nintendo3ds target)")
	})
}

func (view *View) sidebar(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(230), Height: ui.Grow(), Padding: ui.Pad(10), Gap: 4}
	bar(context, ui.Layout{Width: ui.Fixed(230), Height: ui.Grow()}, func() {
		context.ScrollArea("Sidebar", layout, view.sidebarContent(context))
	})
}

// sidebarContent is the list of scenes and the palette of kinds.
func (view *View) sidebarContent(context *ui.Context) func() {
	editor := view.Editor
	return func() {
		context.Row(func() {
			context.Heading("Scenes")
			context.Spacer()
			if context.FlatIconButton("arrow-up", "") {
				editor.MoveScene(-1)
			}
			context.Tooltip("Earlier in the game")
			if context.FlatIconButton("arrow-down", "") {
				editor.MoveScene(1)
			}
			context.Tooltip("Later in the game")
			if context.FlatIconButton("plus", "") {
				editor.AddScene()
			}
			context.Tooltip("Add a scene after this one")
			if context.FlatIconButton("trash-2", "") {
				editor.RemoveScene()
			}
			context.Tooltip("Remove this scene")
		})
		for index := range editor.Game.Scenes {
			context.Push(sceneKey(index))
			if context.Selectable("map", editor.Game.Scenes[index].Name, index == editor.scene) {
				editor.chooseScene(index)
				view.tab = 0
			}
			context.Pop()
		}
		context.Separator()
		context.Row(func() {
			context.Heading("Kinds")
			context.Spacer()
			if context.FlatIconButton("plus", "") {
				editor.AddKind()
				view.tab = 1
			}
			context.Tooltip("Add a kind of thing")
		})
		if context.Selectable("eraser", "Eraser", editor.brush == "") {
			editor.ChooseBrush("")
		}
		for index := range editor.Game.Kinds {
			kind := &editor.Game.Kinds[index]
			context.Push(kind.Name)
			context.Row(func() {
				swatch(context, kind.Colour, 18)
				if context.Selectable(kind.Icon, kind.Name, editor.brush == kind.Name) {
					editor.ChooseBrush(kind.Name)
					view.tab = 1
				}
			})
			context.Pop()
		}
	}
}

// sceneKey scopes a scene's row by its place, so two scenes with one name stay apart.
func sceneKey(index int) string { return "scene " + strconv.Itoa(index) }

// swatch paints a square of a colour.
func swatch(context *ui.Context, colour string, size float32) {
	context.Paint(ui.Layout{Width: ui.Fixed(size), Height: ui.Fixed(size)}, func(area draw2d.Rect, list *draw2d.List) {
		list.Rectangle(area, 4, draw2d.Hex(colour), context.Theme.ControlLine, 1)
	})
}

// middle shows the scene: painted on while editing, played while playing.
func (view *View) middle(context *ui.Context) {
	editor := view.Editor
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		if session, playing := editor.Playing(); playing {
			view.canvasShown = false
			session.Bar(context)
			context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
				session.Draw(list, area.Inset(16))
			})
			return
		}
		view.followPointer(context)
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
			view.canvas = game.DrawScene(list, area.Inset(16), editor.Game, editor.Scene(), true)
			view.canvasShown = true
			if view.hoverColumn >= 0 {
				list.Rectangle(view.canvas.CellArea(view.hoverColumn, view.hoverRow), 2, draw2d.Color{}, context.Theme.Primary, 2)
			}
		})
	})
}

// followPointer paints the cells the pointer drags over, when the press began on the scene, and
// remembers the cell under it to outline.
func (view *View) followPointer(context *ui.Context) {
	x, y := context.Pointer()
	column, row, onScene := view.canvas.CellAt(x, y)
	onScene = onScene && view.canvasShown
	pressing := context.Pressing()
	if pressing && !view.wasPressing {
		view.paintingDrag = onScene
	}
	if !pressing {
		view.paintingDrag = false
	}
	view.wasPressing = pressing
	view.hoverColumn, view.hoverRow = -1, -1
	if !onScene {
		return
	}
	view.hoverColumn, view.hoverRow = column, row
	if view.paintingDrag {
		view.Editor.Paint(column, row)
	}
}

func (view *View) inspector(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(290), Height: ui.Grow()}
	bar(context, layout, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Pad(12)}, func() {
			context.Tabs("Inspector", &view.tab, "Scene", "Kind")
		})
		inside := ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Padding{Left: 12, Right: 12, Bottom: 12}, Gap: 8}
		context.ScrollArea("Inspector", inside, func() {
			if view.tab == 0 {
				view.sceneInspector(context)
				return
			}
			view.kindInspector(context)
		})
	})
}

func (view *View) sceneInspector(context *ui.Context) {
	editor := view.Editor
	scene := editor.Scene()
	context.Caption("Name")
	if context.TextField("Scene name", &editor.sceneName) && editor.sceneName != "" {
		scene.Name = editor.sceneName
		editor.changed()
	}
	width, height := float32(scene.Width), float32(scene.Height)
	context.Caption("Size, in cells")
	context.Row(func() {
		widthChanged := context.NumberField("Width", &width, game.MinimumSize, game.MaximumSize, 1, "%.0f wide")
		heightChanged := context.NumberField("Height", &height, game.MinimumSize, game.MaximumSize, 1, "%.0f high")
		if widthChanged || heightChanged {
			scene.Resize(int(width), int(height))
			editor.changed()
		}
	})
	context.Caption("Background")
	if colourPicker(context, "Background", &scene.Background) {
		editor.changed()
	}
}

func (view *View) kindInspector(context *ui.Context) {
	editor := view.Editor
	kind, found := editor.Kind()
	if !found {
		context.Caption("The game has no kinds yet: add one under Kinds.")
		return
	}
	context.Caption("Name")
	if context.TextField("Kind name", &editor.kindName) {
		view.nameTaken = !editor.RenameKind(editor.kindName) && editor.kindName != kind.Name
	}
	if view.nameTaken {
		context.Text("Another kind has that name", ui.TextStyle{Color: context.Theme.Warning})
	}
	context.Caption("Colour")
	if colourPicker(context, "Colour", &kind.Colour) {
		editor.changed()
	}
	context.Caption("Shape")
	shape := shapeIndex(kind.Shape)
	if context.Tabs("Shape", &shape, "Square", "Rounded", "Circle", "Small") {
		kind.Shape = game.Shapes[shape]
		editor.changed()
	}
	context.Caption("Icon")
	context.Grid(8, func() {
		for _, icon := range Icons {
			context.Push("icon " + icon)
			shown := icon
			if shown == "" {
				shown = "ban"
			}
			if context.Selectable(shown, "", kind.Icon == icon) {
				kind.Icon = icon
				editor.changed()
			}
			context.Pop()
		}
	})
	context.Caption("What it does")
	behaviours := []struct {
		label string
		value *bool
	}{
		{"The player: walks with the arrows", &kind.Player},
		{"Solid: nothing walks into it", &kind.Solid},
		{"Pushable: the player pushes it", &kind.Pushable},
		{"Collectable: taken when stepped on", &kind.Collectable},
		{"Goal: ends the scene", &kind.Goal},
		{"Hazard: starts the scene again", &kind.Hazard},
	}
	for _, behaviour := range behaviours {
		if context.Checkbox(behaviour.label, behaviour.value) {
			editor.changed()
		}
	}
	context.Separator()
	if context.FlatIconButton("trash-2", "Remove this kind") {
		editor.RemoveKind()
	}
}

func shapeIndex(shape game.Shape) int {
	for index, each := range game.Shapes {
		if each == shape {
			return index
		}
	}
	return 0
}

// colourPicker offers the swatches and a field to type a colour in; it tells whether the colour
// changed.
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
			looks := ui.Looks{Normal: look, Hovered: look, Pressed: look}
			if context.Clickable("swatch", ui.Layout{Width: ui.Grow(), Height: ui.Fixed(24)}, looks, nil) {
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
	bar(context, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(30), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}, func() {
		style := ui.TextStyle{Color: context.Theme.InkSoft}
		if editor.Failed {
			style.Color = context.Theme.Danger
		}
		context.Text(editor.Message, style)
		context.Spacer()
		if editor.Saved {
			context.Text("Saved", ui.TextStyle{Color: context.Theme.Positive})
		} else {
			context.Text("Unsaved changes", ui.TextStyle{Color: context.Theme.Warning})
		}
	})
}

// startExport builds the game's program in the background, in an exports folder next to the
// game's file; the status line says how it went.
func (view *View) startExport(target export.Target) {
	if view.exportRunning {
		return
	}
	editor := view.Editor
	if err := editor.Game.Check(); err != nil {
		editor.say(err.Error(), true)
		return
	}
	data, err := editor.Game.Marshal()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	copied, _ := game.Read(data) // the export works on a copy, so editing goes on meanwhile
	folder := filepath.Join(filepath.Dir(editor.Path), "exports")
	view.exportRunning = true
	editor.say("Exporting for "+string(target)+"…", false)
	go func() {
		result, err := export.Export(copied, target, folder)
		view.exports <- exportDone{result: result, err: err}
		window.Wake()
	}()
}

// collectExport takes the result of an export that finished.
func (view *View) collectExport() {
	select {
	case done := <-view.exports:
		view.exportRunning = false
		if done.err != nil {
			view.Editor.say("Export failed: "+done.err.Error(), true)
			return
		}
		view.Editor.say("Exported: "+done.result.Program, false)
	default:
	}
}
