package editor

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/project"
)

// The Code workspace: the bridge between the project's Go code and its scenes. The behaviours on
// the left (and those the scenes use that the code lacks); the one chosen in the middle, with its
// comment, its methods and its fields; the nodes that use it on the right; building the code
// again and opening it in the code editor, with what building said, at the bottom.

// The Code workspace's measures, in pixels.
const (
	codeListWidth  = 260
	codeUsesWidth  = 300
	codeOutputTall = 170
)

// codeView is what the Code workspace remembers between frames.
type codeView struct {
	chosen string // the behaviour chosen, by its name
}

func (view *View) codeWorkspace(context *ui.Context) {
	if view.code.chosen == "" {
		view.code.chosen = view.firstBehaviour()
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		view.codeBar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			panel(context, "Behaviour list", ui.Fixed(codeListWidth), func() { view.behaviourList(context) })
			context.Separator()
			panel(context, "Behaviour", ui.Grow(), func() { view.behaviourDetail(context) })
			context.Separator()
			panel(context, "Behaviour uses", ui.Fixed(codeUsesWidth), func() { view.behaviourUses(context) })
		})
		context.Separator()
		view.outputPanel(context)
	})
}

// firstBehaviour is the one to show first: the chosen node's first, or the project's first.
func (view *View) firstBehaviour() string {
	editor := view.Editor
	if node := editor.Selected; node != nil && len(node.Behaviours) > 0 {
		return node.Behaviours[0].Name
	}
	if definitions := editor.Behaviours(); len(definitions) > 0 {
		return definitions[0].Name
	}
	return ""
}

// codeBar is the strip along the top: opening the code, and building it again.
func (view *View) codeBar(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(transportTall), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}
	bar(context, layout, func() {
		context.IconSized(iconCode, 18, context.Theme.Primary)
		context.Heading("Behaviours of the project")
		context.Spacer()
		if view.Editor.CodeChanged() {
			context.Text("The code changed since the editor was built", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
		}
		view.openCodeButton(context)
		view.rebuildButton(context)
	})
}

// behaviourList is the project's behaviours, and those the scenes use that its code lacks.
func (view *View) behaviourList(context *ui.Context) {
	editor := view.Editor
	context.Heading("In the code")
	for _, definition := range editor.Behaviours() {
		if context.Selectable(iconCode, definition.Name, definition.Name == view.code.chosen) {
			view.code.chosen = definition.Name
		}
	}
	if len(editor.Behaviours()) == 0 {
		note(context, "No behaviours yet: write a type with an Update, Start or Touched method in the project's Go code.")
	}
	missing := editor.MissingBehaviours()
	if len(missing) == 0 {
		return
	}
	context.Separator()
	context.Heading("Used but not in the code")
	for _, each := range missing {
		context.Text(each.Name+" · "+FieldText(float64(len(each.Uses)))+" nodes", ui.TextStyle{Color: context.Theme.Warning})
	}
	note(context, "A typo in a scene, or code not written yet: these nodes play without them.")
}

// behaviourDetail is the behaviour chosen: its file, comment, methods and fields.
func (view *View) behaviourDetail(context *ui.Context) {
	definition, found := view.Editor.FindBehaviour(view.code.chosen)
	if !found {
		note(context, "Choose a behaviour to see its fields, its methods and where it is used.")
		return
	}
	context.Row(func() {
		context.IconSized(iconCode, 20, context.Theme.Primary)
		context.Heading(definition.Name)
	})
	context.Caption(definition.File)
	if definition.Doc != "" {
		note(context, definition.Doc)
	}
	context.Caption("Methods the engine calls: " + strings.Join(definition.Methods, ", "))
	context.Separator()
	view.fieldTable(context, definition)
}

// fieldTable lists a behaviour's fields: the name a scene gives them, their Go type, their comment.
func (view *View) fieldTable(context *ui.Context, definition project.Behaviour) {
	context.Heading("Fields")
	if len(definition.Fields) == 0 {
		note(context, "No fields a scene can set.")
		return
	}
	context.Grid(3, func() {
		for _, header := range []string{"In the scene", "Go type", "Comment"} {
			context.Text(header, ui.TextStyle{Color: context.Theme.InkSoft, Weight: draw2d.Bold, Size: context.Theme.SmallTextSize})
		}
		for _, field := range definition.Fields {
			context.Label(field.Name)
			context.Text(field.Type, ui.TextStyle{Color: context.Theme.Primary})
			context.Text(field.Comment, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
		}
	})
}

// behaviourUses are the nodes that use the behaviour chosen; a click opens one in Level.
func (view *View) behaviourUses(context *ui.Context) {
	context.Heading("Used in")
	if view.code.chosen == "" {
		return
	}
	uses := view.Editor.BehaviourUses(view.code.chosen)
	if len(uses) == 0 {
		note(context, "No node of a saved scene has it yet.")
	}
	for _, use := range uses {
		if context.Selectable(iconScene, sceneName(use.File)+" · "+use.Node, false) {
			view.openUse(use)
		}
	}
}

// openUse opens the scene of a use in Level, with its node chosen.
func (view *View) openUse(use project.Use) {
	editor := view.Editor
	if !view.openScene(use.File) {
		return
	}
	if node, found := editor.Scene.Find(use.Node); found {
		editor.Select(node)
	}
	view.showWorkspace(LevelWorkspace)
}

// outputPanel is what building the code or the game last said.
func (view *View) outputPanel(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(codeOutputTall), Padding: ui.PadXY(10, 6), Gap: 2}
	bar(context, layout, func() {
		context.Row(func() {
			context.IconSized(iconOutput, 16, context.Theme.Primary)
			context.Text("Output", ui.TextStyle{Weight: draw2d.Bold})
		})
		inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 1}
		context.ScrollArea("Output", inside, func() {
			lines := view.outputLines()
			if len(lines) == 0 {
				note(context, "Nothing built yet in this session.")
			}
			for _, line := range lines {
				context.Text(line, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
			}
		})
	})
}

// outputLines are what rebuilding the code said when it failed, then what building or playing the
// game printed.
func (view *View) outputLines() []string {
	lines := view.Editor.buildOutput.Lines()
	if view.Editor.output != nil {
		lines = append(lines, view.Editor.output.Lines()...)
	}
	return lines
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
	context.Tooltip("Open the project's folder in the code editor its settings name")
}

// openProject opens the project in the code editor its settings name, or the one found.
func (view *View) openProject() {
	opened, err := view.Editor.Project.OpenCode()
	if err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("Opened the project in "+opened, false)
}
