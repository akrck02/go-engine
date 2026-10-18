package editor

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
)

func (view *View) behaviourInspector(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	context.Separator()
	context.Row(func() {
		context.Heading("Behaviours")
		context.Spacer()
		context.MenuButton(iconAdd, "Add", func() {
			for _, definition := range editor.Behaviours() {
				if context.MenuItem(iconCode, definition.Name, definition.File) {
					editor.AttachBehaviour(definition.Name)
				}
			}
			if context.MenuItem(iconReadAgain, "Read the code again", "") {
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
			definition, known := editor.FindBehaviour(behaviour.Name)
			context.Row(func() {
				context.IconSized(iconCode, 16, context.Theme.Primary)
				context.Text(behaviour.Name, ui.TextStyle{Weight: draw2d.Bold})
				context.Spacer()
				if context.FlatIconButton(iconRemove, "") {
					editor.DetachBehaviour(index)
				}
			})
			if !known {
				context.Text("Not in the project's code (yet?)", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
			} else if definition.Doc != "" {
				note(context, definition.Doc)
			}
			view.behaviourFields(context, behaviour, definition.Fields)
		})
		context.Pop()
	}
}

// behaviourFields edits a behaviour's fields: the ones its code declares, then any others the
// scene holds.
func (view *View) behaviourFields(context *ui.Context, behaviour *scenefile.Behaviour, declared []projectField) {
	shown := map[string]bool{}
	for _, field := range declared {
		shown[field.Name] = true
		view.behaviourField(context, behaviour, field.Name, field.Comment) // its Go type is shown in Code
	}
	for name := range behaviour.Fields {
		if !shown[name] {
			view.behaviourField(context, behaviour, name, "not in the code")
		}
	}
}

func (view *View) behaviourField(context *ui.Context, behaviour *scenefile.Behaviour, name, hint string) {
	key := behaviour.Name + "." + name
	text, editing := view.fieldTexts[key]
	if !editing {
		text = FieldText(behaviour.Fields[name])
	}
	if hint == "" {
		context.Caption(name)
	} else {
		context.Caption(name + " · " + hint)
	}
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
type projectField = project.Field
