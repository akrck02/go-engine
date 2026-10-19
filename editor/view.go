package editor

import (
	"time"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
)

// The editor's window: a bar along the top (the project, the scene, saving), the workspace shown
// (workspaces.go), and a bar along the bottom to change workspace.

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

	emptyWorld *scene.Scene // what is drawn behind the workspaces that cover the window
	shift      bool         // Shift was held on the last frame: a click joins the choice

	art     artView      // the Art workspace's choices
	project projectView  // the Project workspace's choices
	code    codeView     // the Code workspace's choices
	build   buildView    // the Build workspace's choices
	picking *filePicking // the file asked for, while the file picker is shown

	assetTree    project.AssetEntry
	assetsReadAt time.Time
	openFolders  map[string]bool
}

// NewView shows an editor.
func NewView(editor *Editor) *View {
	editor.Wake = window.Wake
	return &View{Editor: editor, viewport: newViewport(), fieldTexts: map[string]string{}, openFolders: map[string]bool{"assets": true},
		emptyWorld: scene.New()}
}

// Options are the window's, with the step that shows the scene and reads what the interface
// leaves of the mouse and the keys.
func (view *View) Options(options window.Options) window.Options {
	options.Title = "go-engine " + engine.Version() + " — " + view.Editor.Project.Settings.Name
	options.Scene = func(state *input.State, seconds float64) *window.SceneView {
		view.shift = state.Down(input.KeyShift)
		view.keys(state)
		shown := view.Editor.ShownWorkspace()
		if shown != LevelWorkspace && shown != PlayWorkspace {
			return &window.SceneView{World: view.emptyWorld} // the other workspaces cover the window
		}
		world, camera := view.viewport.step(view.Editor, state, seconds)
		return &window.SceneView{World: world, Camera: camera}
	}
	return options
}

func (view *View) keys(state *input.State) {
	editor := view.Editor
	if view.workspaceKeys(state) || editor.ShownWorkspace() == PlayWorkspace {
		return // while playing, the game has the other keys
	}
	if state.Down(input.KeyControl) && view.controlKeys(state) {
		return
	}
	if editor.ShownWorkspace() == LevelWorkspace {
		view.levelKeys(state)
	}
}

// controlKeys save, undo and redo with Control held; they tell whether they took a key.
func (view *View) controlKeys(state *input.State) bool {
	editor := view.Editor
	switch {
	case state.Pressed(input.KeyS):
		editor.Save()
	case state.Pressed(input.KeyZ) && state.Down(input.KeyShift), state.Pressed(input.KeyY):
		editor.Redo()
	case state.Pressed(input.KeyZ):
		editor.Undo()
	default:
		return false
	}
	view.shownFor = nil // the inspector's fields show the scene as it is now
	return true
}

// workspaceKeys change workspace with 1 to 6, and leave Play with Escape; they tell whether they
// took a key.
func (view *View) workspaceKeys(state *input.State) bool {
	editor := view.Editor
	if editor.ShownWorkspace() == PlayWorkspace && state.Pressed(input.KeyEscape) {
		editor.LeavePlay()
		return true
	}
	for _, workspace := range Workspaces {
		if state.Pressed(workspace.Key()) {
			view.showWorkspace(workspace)
			return true
		}
	}
	return false
}

// levelKeys take a tool in 2D, and act on the chosen node: duplicate, delete, frame, let go.
func (view *View) levelKeys(state *input.State) {
	editor := view.Editor
	control := state.Down(input.KeyControl)
	if !control {
		for _, tool := range ToolsFor(editor.Scene.Is2D()) {
			if state.Pressed(tool.Key()) {
				editor.Tool = tool
				return
			}
		}
	}
	switch {
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
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() { view.workspaceBody(context) })
		context.Separator()
		view.workspaceBar(context)
	})
	view.filePicker(context)
}

func bar(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}
