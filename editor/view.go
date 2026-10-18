package editor

import (
	"time"

	"github.com/akrck02/go-renderer/input"
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
	if editor.ShownWorkspace() == PlayWorkspace {
		if state.Pressed(input.KeyEscape) {
			editor.LeavePlay()
		}
		return // the game has the other keys
	}
	for _, workspace := range Workspaces {
		if state.Pressed(workspace.Key()) {
			view.showWorkspace(workspace)
			return
		}
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
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() { view.workspaceBody(context) })
		context.Separator()
		view.workspaceBar(context)
	})
}

func bar(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}
