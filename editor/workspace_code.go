package editor

import ui "github.com/akrck02/go-ui"

// The Code workspace: the behaviours the project's Go code defines, building that code again,
// and opening it in the code editor.

func (view *View) codeWorkspace(context *ui.Context) {
	panel(context, "Code", ui.Grow(), func() {
		context.Row(func() {
			context.Heading("Behaviours of the project")
			context.Spacer()
			view.openCodeButton(context)
			view.rebuildButton(context)
		})
		for _, definition := range view.Editor.Behaviours() {
			context.Row(func() {
				context.IconSized(iconCode, 16, context.Theme.Primary)
				context.Label(definition.Name)
				context.Caption(definition.File)
			})
			if definition.Doc != "" {
				note(context, definition.Doc)
			}
		}
	})
}

// rebuildButton builds the project's code again; it stands out while the code is newer than the
// game the editor runs in.
func (view *View) rebuildButton(context *ui.Context) {
	editor := view.Editor
	button := context.IconButton
	if editor.CodeChanged() {
		button = context.PrimaryIconButton
	}
	if button(iconBuild, "Rebuild") {
		editor.Reload()
	}
	context.Tooltip("Build the project's Go code again and reopen the editor with it")
}

// openCodeButton opens the project in its code editor.
func (view *View) openCodeButton(context *ui.Context) {
	if context.IconButton(iconOpenCode, "Open in code editor") {
		view.openProject()
	}
	context.Tooltip("Open the project's folder in the code editor")
}

func (view *View) openProject() {
	if err := view.Editor.Project.OpenInZed(); err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("Opened the project in Zed", false)
}
