package editor

import (
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// left is the scene's nodes, as a tree, over the project's models.
func (view *View) left(context *ui.Context) {
	editor := view.Editor
	bar(context, ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(270), Height: ui.Grow()}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 10, Top: 8, Right: 6}, AlignY: ui.Center}, func() {
			context.Heading("Scene")
			context.Spacer()
			context.MenuButton("plus", "", func() {
				for _, kind := range editor.KindsHere() {
					if context.MenuItem(kindIcon(kind), kind, "") {
						editor.Add(kind)
					}
				}
			})
			context.Tooltip("Add a node after the chosen one")
			view.treeButton(context, "copy", "Duplicate (Ctrl+D)", editor.Duplicate)
			view.treeButton(context, "trash-2", "Delete (Delete)", editor.Delete)
		})
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 10, Right: 6}, Gap: 2}, func() {
			view.treeButton(context, "arrow-up", "Earlier among its siblings", editor.MoveUp)
			view.treeButton(context, "arrow-down", "Later among its siblings", editor.MoveDown)
			view.treeButton(context, "indent-increase", "Hang it from the node above", editor.Nest)
			view.treeButton(context, "indent-decrease", "Hang it from its parent's parent", editor.Unnest)
		})
		tree := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.GrowBy(3), Padding: ui.PadXY(8, 4), Gap: 1}
		context.ScrollArea("Tree", tree, func() {
			for _, child := range editor.Scene.Root.Children() {
				view.treeNode(context, child, 0)
			}
		})
		context.Separator()
		view.assets(context)
	})
}

func (view *View) treeButton(context *ui.Context, icon, tip string, action func()) {
	if context.FlatIconButton(icon, "") {
		action()
	}
	context.Tooltip(tip)
}

func (view *View) treeNode(context *ui.Context, node *engine.Node, depth int) {
	context.Push(node.Name)
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 16}}, func() {
		if context.Selectable(nodeIcon(node), node.Name, node == view.Editor.Selected) {
			view.Editor.Select(node)
		}
	})
	for _, child := range node.Children() {
		view.treeNode(context, child, depth+1)
	}
	context.Pop()
}

func kindIcon(kind string) string {
	switch kind {
	case "Cube":
		return "box"
	case "Sphere":
		return "circle"
	case "Plane":
		return "square"
	case "Cylinder":
		return "cylinder"
	case "Camera":
		return "video"
	case "Square":
		return "square"
	case "Circle":
		return "circle"
	}
	return "crosshair"
}

func nodeIcon(node *engine.Node) string {
	if _, isCamera := node.Camera(); isCamera {
		return "video"
	}
	if node.Model() != "" {
		return "package"
	}
	if node.Sprite() != "" {
		return "image"
	}
	if node.Shape() != "" {
		return kindIcon(strings.ToUpper(node.Shape()[:1]) + node.Shape()[1:])
	}
	return "crosshair"
}
