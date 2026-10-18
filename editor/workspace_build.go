package editor

import ui "github.com/akrck02/go-ui"

// The Build workspace: making the game for each place it runs.

func (view *View) buildWorkspace(context *ui.Context) {
	editor := view.Editor
	panel(context, "Build", ui.Grow(), func() {
		context.Heading("Build the game")
		if context.IconButton(iconComputer, "For this computer") {
			editor.Run("Building for this computer…", func() (string, error) {
				built, err := editor.Project.BuildDesktop()
				return "Built " + built.Program, err
			})
		}
		if context.IconButton(iconConsole, "For the Nintendo 3DS") {
			editor.Run("Building for the 3DS…", func() (string, error) {
				built, err := editor.Project.Build3DS()
				return "Built " + built.Program, err
			})
		}
	})
}
