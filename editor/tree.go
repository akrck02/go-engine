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
			context.MenuButton(iconAdd, "", func() {
				for _, kind := range editor.KindsHere() {
					if context.MenuItem(kindIcon(kind), kind, "") {
						editor.Add(kind)
					}
				}
			})
			context.Tooltip("Add a node after the chosen one")
			view.treeButton(context, iconDuplicate, "Duplicate (Ctrl+D)", editor.Duplicate)
			view.treeButton(context, iconDelete, "Delete (Delete)", editor.Delete)
		})
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 10, Right: 6}, Gap: 2}, func() {
			view.treeButton(context, iconEarlier, "Earlier among its siblings", editor.MoveUp)
			view.treeButton(context, iconLater, "Later among its siblings", editor.MoveDown)
			view.treeButton(context, iconMakeChild, "Make it a child of the node above", editor.Nest)
			view.treeButton(context, iconMoveOut, "Move it out, beside its parent", editor.Unnest)
		})
		tree := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.PadXY(8, 4), Gap: 1}
		context.ScrollArea("Tree", tree, func() {
			for _, child := range editor.Scene.Root.Children() {
				view.treeNode(context, child, 0)
			}
		})
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
		if context.Selectable(nodeIcon(node), node.Name, view.Editor.IsChosen(node)) {
			chooseWith(view.Editor, node, view.shift)
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
		return iconCube
	case "Sphere":
		return iconRound
	case "Plane":
		return iconFlat
	case "Cylinder":
		return iconCylinder
	case "Camera":
		return iconCamera
	case "Square":
		return iconFlat
	case "Circle":
		return iconRound
	case "Tile map":
		return iconTileMap
	case "Light":
		return iconLamp
	}
	return iconEmpty
}

func nodeIcon(node *engine.Node) string {
	if _, isCamera := node.Camera(); isCamera {
		return iconCamera
	}
	if node.Model() != "" {
		return iconModel
	}
	if node.Sprite() != "" {
		return iconSprite
	}
	if node.Shape() != "" {
		return kindIcon(strings.ToUpper(node.Shape()[:1]) + node.Shape()[1:])
	}
	if _, isTileMap := node.Tiles(); isTileMap {
		return iconTileMap
	}
	if _, isLight := node.Light(); isLight {
		return iconLamp
	}
	return iconEmpty
}
