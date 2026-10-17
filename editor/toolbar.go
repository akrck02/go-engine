package editor

import (
	"path"
	"strings"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/scenefile"
)

func (view *View) toolbar(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(46), Padding: ui.PadXY(10, 0), Gap: 6, AlignY: ui.Center}
	bar(context, layout, func() {
		context.IconSized("gamepad-2", 20, context.Theme.Primary)
		context.Heading(view.Editor.Project.Settings.Name)
		context.Box(ui.Layout{Width: ui.Fixed(8)}, nil)
		view.sceneMenu(context)
		view.saveButton(context)
		context.Spacer()
		view.playButton(context)
		view.windowButton(context)
		view.reloadButton(context)
		view.openProjectButton(context)
		view.buildMenu(context)
		view.themeButton(context)
	})
}

// sceneMenu chooses the scene shown, or makes a new one.
func (view *View) sceneMenu(context *ui.Context) {
	editor := view.Editor
	context.MenuButton("map", sceneName(editor.SceneFile), func() {
		for _, file := range editor.Project.Scenes() {
			if context.MenuItem("map", sceneName(file), "") && file != editor.SceneFile {
				view.openScene(file)
			}
		}
		if context.MenuItem("box", "New 3D scene", "") {
			view.newScene(scenefile.Mode3D)
		}
		if context.MenuItem("square", "New 2D scene", "") {
			view.newScene(scenefile.Mode2D)
		}
	})
}

// sceneName is a scene file's name without its folder and suffix: "main" for
// "scenes/main.scene.json".
func sceneName(file string) string { return strings.TrimSuffix(path.Base(file), ".scene.json") }

func (view *View) saveButton(context *ui.Context) {
	if context.IconButton("save", "Save") {
		view.Editor.Save()
	}
	context.Tooltip("Save the scene (Ctrl+S)")
}

// playButton plays the scene in the view, or stops it.
func (view *View) playButton(context *ui.Context) {
	editor := view.Editor
	if _, playing := editor.PlayingHere(); playing {
		if context.PrimaryIconButton("square", "Stop") {
			editor.StopHere()
		}
		context.Tooltip("Back to editing (Escape)")
		return
	}
	if context.PrimaryIconButton("play", "Play") {
		editor.PlayHere()
	}
	context.Tooltip("Play the scene in the view, with the project's code")
}

// windowButton plays the game in a window of its own, or closes it.
func (view *View) windowButton(context *ui.Context) {
	editor := view.Editor
	if editor.Running() && editor.playing != nil {
		if context.IconButton("square", "Window") {
			editor.Stop()
		}
		context.Tooltip("Close the game's window")
		return
	}
	if context.FlatIconButton("external-link", "") {
		editor.PlayInWindow()
	}
	context.Tooltip("Build the game and play it in a window of its own")
}

// reloadButton builds the project's code again; it stands out while the code is newer than the
// game the editor runs in.
func (view *View) reloadButton(context *ui.Context) {
	editor := view.Editor
	changed := editor.CodeChanged()
	button := context.FlatIconButton
	if changed {
		button = context.IconButton
	}
	if button("refresh-cw", "Reload code") {
		editor.Reload()
	}
	if changed {
		context.Tooltip("The project's Go code changed: build it again and reopen the editor with it")
		return
	}
	context.Tooltip("Build the project's Go code again and reopen the editor with it")
}

func (view *View) openProjectButton(context *ui.Context) {
	if context.IconButton("folder-open", "Open project") {
		view.openProject()
	}
	context.Tooltip("Open the project's folder in Zed")
}

func (view *View) openProject() {
	if err := view.Editor.Project.OpenInZed(); err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("Opened the project in Zed", false)
}

// buildMenu builds the game for this computer or for the 3DS, in the background.
func (view *View) buildMenu(context *ui.Context) {
	editor := view.Editor
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
}

// themeButton switches between the dark and the light mode.
func (view *View) themeButton(context *ui.Context) {
	icon, tip := "moon", "Dark mode"
	if view.dark {
		icon, tip = "sun", "Light mode"
	}
	if context.FlatIconButton(icon, "") {
		view.setDark(context, !view.dark)
	}
	context.Tooltip(tip)
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
