package editor

import (
	"path"
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// The editor's two bars. Along the top: the project, the scene open, saving it, and the look.
// Along the bottom: what the editor last said, the workspaces, and whether the scene is saved.
// Everything else belongs to a workspace.

func (view *View) toolbar(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(46), Padding: ui.PadXY(10, 0), Gap: 6, AlignY: ui.Center}
	bar(context, layout, func() {
		context.IconSized(iconProject, 20, context.Theme.Primary)
		context.Heading(view.Editor.Project.Settings.Name)
		context.Box(ui.Layout{Width: ui.Fixed(8)}, nil)
		view.sceneMenu(context)
		view.saveButton(context)
		view.historyButtons(context)
		context.Spacer()
		context.Text("Go Engine "+engine.Version(), ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
		context.Tooltip("The Go Engine this editor was built with: its release and commit")
		view.themeButton(context)
	})
}

// sceneMenu chooses the scene shown, or makes a new one.
func (view *View) sceneMenu(context *ui.Context) {
	editor := view.Editor
	context.MenuButton(iconScene, sceneName(editor.SceneFile), func() {
		for _, file := range editor.Project.Scenes() {
			if context.MenuItem(iconScene, sceneName(file), "") && file != editor.SceneFile {
				view.openScene(file)
			}
		}
		if context.MenuItem(iconScene3D, "New 3D scene", "") {
			view.newScene(scenefile.Mode3D)
		}
		if context.MenuItem(iconScene2D, "New 2D scene", "") {
			view.newScene(scenefile.Mode2D)
		}
	})
	context.Tooltip("The scene open; choose another, or make one")
}

// sceneName is a scene file's name without its folder and suffix: "main" for
// "scenes/main.scene.json".
func sceneName(file string) string { return strings.TrimSuffix(path.Base(file), ".scene.json") }

func (view *View) saveButton(context *ui.Context) {
	if context.IconButton(iconSave, "Save") {
		view.Editor.Save()
	}
	context.Tooltip("Save the scene (Ctrl+S)")
}

// historyButtons undo and redo the changes of the workspace shown: the controls' in Controls, the
// scene's elsewhere.
func (view *View) historyButtons(context *ui.Context) {
	if context.FlatIconButton(iconUndo, "") {
		view.undo()
		view.shownFor = nil
	}
	context.Tooltip("Undo (Ctrl+Z)")
	if context.FlatIconButton(iconRedo, "") {
		view.redo()
		view.shownFor = nil
	}
	context.Tooltip("Redo (Ctrl+Shift+Z)")
}

// themeButton switches between the dark and the light mode.
func (view *View) themeButton(context *ui.Context) {
	icon, tip := iconDark, "Dark mode"
	if view.dark {
		icon, tip = iconLight, "Light mode"
	}
	if context.FlatIconButton(icon, "") {
		view.setDark(context, !view.dark)
	}
	context.Tooltip(tip)
}

// workspaceBar is the bottom bar: the editor's last word on the left, the workspaces in the
// middle, the scene's saved state on the right.
func (view *View) workspaceBar(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(workspaceBarHeight), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}
	bar(context, layout, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() {}) // balances the saved state, to centre the workspaces
		for _, workspace := range Workspaces {
			view.workspaceButton(context, workspace)
		}
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.End, AlignY: ui.Center}, func() { view.savedState(context) })
	})
}

// workspaceButton is a workspace's icon over its word; the one shown stands out.
func (view *View) workspaceButton(context *ui.Context, workspace Workspace) {
	shown := view.Editor.ShownWorkspace() == workspace
	ink, look := context.Theme.InkSoft, ui.Look{Radius: 8}
	if shown {
		ink, look.Fill = context.Theme.Primary, context.Theme.Sunken
	}
	hovered := look
	hovered.Fill = context.Theme.Sunken
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(76), Padding: ui.PadXY(4, 4), Gap: 2, AlignX: ui.Center}
	if context.Clickable(workspace.Name(), layout, ui.Looks{Normal: look, Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized(workspace.Icon(), 20, ink)
		context.Text(workspace.Name(), ui.TextStyle{Color: ink, Size: context.Theme.SmallTextSize, Weight: weightFor(shown)})
	}) {
		view.showWorkspace(workspace)
	}
	context.Tooltip(workspace.Name() + " (" + strconv.Itoa(int(workspace)+1) + ")")
}

func weightFor(shown bool) draw2d.Weight {
	if shown {
		return draw2d.Bold
	}
	return draw2d.Regular
}

// showWorkspace changes the workspace, and forgets what was being typed in the one left.
func (view *View) showWorkspace(workspace Workspace) {
	view.Editor.ShowWorkspace(workspace)
	view.shownFor = nil
}

// savedState says whether the scene has changes not saved.
func (view *View) savedState(context *ui.Context) {
	if view.Editor.Saved {
		context.Text("Saved", ui.TextStyle{Color: context.Theme.Positive, Size: context.Theme.SmallTextSize})
		return
	}
	context.Text("Unsaved changes", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
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

// openScene opens a scene of the project, and tells whether it is the one shown now; the scene
// open already stays as it is, unsaved changes and all.
func (view *View) openScene(file string) bool {
	if file == view.Editor.SceneFile {
		return true
	}
	if err := view.Editor.OpenScene(file); err != nil {
		view.Editor.say(err.Error(), true)
		return false
	}
	view.shownFor = nil
	return true
}

func (view *View) newScene(mode scenefile.Mode) {
	editor := view.Editor
	name := "Scene " + FieldText(float64(len(editor.Project.Scenes())+1))
	file, err := editor.Project.NewScene(name, mode)
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	view.openScene(file)
}
