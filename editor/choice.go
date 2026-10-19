package editor

import (
	"slices"

	"github.com/akrck02/go-engine/engine"
)

// Several nodes can be chosen at once: Selected is the one the inspector shows, Chosen the others
// taken with Shift. Dragging moves them together; Delete and Duplicate act on all of them.

// AllChosen are the nodes chosen, the inspector's first.
func (editor *Editor) AllChosen() []*engine.Node {
	if editor.Selected == nil {
		return nil
	}
	return append([]*engine.Node{editor.Selected}, editor.Chosen...)
}

// IsChosen tells whether a node is among those chosen.
func (editor *Editor) IsChosen(node *engine.Node) bool {
	return node != nil && (node == editor.Selected || slices.Contains(editor.Chosen, node))
}

// ToggleChoice adds a node to those chosen, or takes it out if it was; taking out the inspector's
// node gives the inspector the next one.
func (editor *Editor) ToggleChoice(node *engine.Node) {
	switch {
	case node == nil:
	case editor.Selected == nil:
		editor.Selected = node
	case node == editor.Selected:
		editor.Selected = nil
		if len(editor.Chosen) > 0 {
			editor.Selected, editor.Chosen = editor.Chosen[0], editor.Chosen[1:]
		}
	case slices.Contains(editor.Chosen, node):
		editor.Chosen = slices.DeleteFunc(editor.Chosen, func(each *engine.Node) bool { return each == node })
	default:
		editor.Chosen = append(editor.Chosen, node)
	}
}

// outermostChosen are the nodes chosen that no other chosen node holds: acting on a node acts on
// what hangs from it, so its chosen children would be acted on twice.
func (editor *Editor) outermostChosen() []*engine.Node {
	all := editor.AllChosen()
	var outermost []*engine.Node
	for _, node := range all {
		if !heldByAny(node, all) {
			outermost = append(outermost, node)
		}
	}
	return outermost
}

func heldByAny(node *engine.Node, others []*engine.Node) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if slices.Contains(others, parent) {
			return true
		}
	}
	return false
}
