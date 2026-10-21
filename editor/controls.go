package editor

import (
	"slices"
	"strings"

	"github.com/akrck02/go-engine/engine"
	ui "github.com/akrck02/go-ui"
)

// The project's controls, in its settings: each action the game reads (engine.json's actions),
// and for each target the inputs that press it, typed as words separated by commas ("Space, Pad
// A"). A word the target does not know is shown in red under its field.

// actionsTyped are the project's actions as typed: for each action and target, its words.
type actionsTyped map[string]map[string]string

// typedFrom is a project's actions ready to be typed over.
func typedFrom(actions map[string]map[string][]string) actionsTyped {
	typed := actionsTyped{}
	for name, targets := range actions {
		typed[name] = map[string]string{}
		for target, words := range targets {
			typed[name][target] = strings.Join(words, ", ")
		}
	}
	return typed
}

// bindings are the actions as engine.json keeps them, with the words typed split and trimmed;
// a target with none is left out.
func (typed actionsTyped) bindings() map[string]map[string][]string {
	actions := map[string]map[string][]string{}
	for name, targets := range typed {
		actions[name] = map[string][]string{}
		for target, text := range targets {
			if words := wordsOf(text); len(words) > 0 {
				actions[name][target] = words
			}
		}
	}
	return actions
}

// wordsOf are the words typed in a field, split at commas.
func wordsOf(text string) []string {
	var words []string
	for _, word := range strings.Split(text, ",") {
		if word = strings.TrimSpace(word); word != "" {
			words = append(words, word)
		}
	}
	return words
}

// controlsSection shows the actions, a field for each target's inputs, and a way to add one.
func (view *View) controlsSection(context *ui.Context) {
	context.Separator()
	context.Heading("Controls")
	note(context, "Each action the game reads (Input.Action), and the inputs that press it on each target, separated by commas: Space, Mouse Left, Pad A; on a 3DS, its buttons: A, R, Start.")
	typed := view.project.actions
	names := make([]string, 0, len(typed))
	for name := range typed {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		view.actionRows(context, name)
	}
	context.Row(func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.TextField("New action", &view.project.newAction) })
		if context.IconButton(iconAdd, "Add action") {
			view.addAction()
		}
	})
}

// actionRows are one action: its name, a button that removes it, and a field for each target.
func (view *View) actionRows(context *ui.Context, name string) {
	context.Push(name)
	defer context.Pop()
	context.Row(func() {
		context.Text(name, ui.TextStyle{Weight: weightFor(true)})
		if context.FlatIconButton(iconDelete, "") {
			delete(view.project.actions, name)
		}
		context.Tooltip("Remove the action " + name)
	})
	for _, target := range engine.InputTargets {
		text := view.project.actions[name][target]
		context.Caption(target)
		context.TextField(target, &text)
		view.project.actions[name][target] = text
		for _, problem := range engine.CheckInputs(wordsOf(text), target) {
			wrapped(context, explained(problem), ui.TextStyle{Color: context.Theme.Danger, Size: context.Theme.SmallTextSize})
		}
	}
}

// addAction adds the action named in the new action's field, with no inputs yet.
func (view *View) addAction() {
	name := strings.TrimSpace(view.project.newAction)
	if name == "" || view.project.actions[name] != nil {
		return
	}
	view.project.actions[name] = map[string]string{}
	view.project.newAction = ""
}
