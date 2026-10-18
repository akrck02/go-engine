package editor

import ui "github.com/akrck02/go-ui"

// The Project workspace: the project's assets and scenes.

func (view *View) projectWorkspace(context *ui.Context) {
	panel(context, "Scenes", ui.Fixed(270), func() {
		context.Heading("Scenes")
		for _, file := range view.Editor.Project.Scenes() {
			if context.Selectable(iconScene, sceneName(file), file == view.Editor.SceneFile) {
				view.openScene(file)
				view.showWorkspace(LevelWorkspace)
			}
		}
	})
	context.Separator()
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() { view.assets(context) })
}
