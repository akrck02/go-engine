package editor

import (
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/project"
)

// assets shows the project's assets folder as a tree: whatever is copied there, from the file
// manager or the code editor, shows here. A model clicked is placed in the scene.
func (view *View) assets(context *ui.Context) {
	view.readAssetsNow()
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.GrowBy(2), Padding: ui.Pad(10), Gap: 1}
	context.ScrollArea("Assets", layout, func() {
		context.Row(func() {
			context.Heading("Assets")
			context.Spacer()
			if context.FlatIconButton(iconReadAgain, "") {
				view.assetsReadAt = time.Time{}
			}
			context.Tooltip("Read the assets folder again")
		})
		if len(view.assetTree.Children) == 0 {
			note(context, "The assets folder is empty: copy models (.glb) and pictures (.png) into "+view.Editor.Project.Path("assets"))
		}
		for _, child := range view.assetTree.Children {
			view.assetNode(context, child, 0)
		}
	})
}

func (view *View) assetNode(context *ui.Context, entry project.AssetEntry, depth int) {
	editor := view.Editor
	context.Push(entry.Path)
	defer context.Pop()
	indent := ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 14}}
	if entry.Folder {
		open := view.openFolders[entry.Path]
		icon := iconFolder
		if open {
			icon = iconOpenFolder
		}
		context.Box(indent, func() {
			if context.Selectable(icon, entry.Name, false) {
				view.openFolders[entry.Path] = !open
			}
		})
		if open {
			for _, child := range entry.Children {
				view.assetNode(context, child, depth+1)
			}
		}
		return
	}
	kind := project.KindOf(entry.Path)
	context.Box(indent, func() {
		if context.Selectable(assetIcon(kind), entry.Name, false) {
			switch {
			case kind == project.ModelAsset && !editor.Scene.Is2D():
				editor.AddModel(entry.Path)
			case kind == project.ImageAsset && editor.Scene.Is2D():
				editor.AddSprite(entry.Path)
			case kind == project.ModelAsset:
				editor.say(entry.Path+": a model goes in a 3D scene", false)
			case kind == project.ImageAsset:
				editor.say(entry.Path+": a picture goes in a 2D scene, as a sprite", false)
			default:
				editor.say(entry.Path+": not something a scene shows yet", false)
			}
		}
	})
	switch kind {
	case project.ModelAsset, project.ImageAsset:
		context.Tooltip("Place " + entry.Path + " in the scene")
	default:
		context.Tooltip(entry.Path)
	}
}

func assetIcon(kind project.AssetKind) string {
	switch kind {
	case project.ModelAsset:
		return iconModel
	case project.ImageAsset:
		return iconPicture
	case project.SoundAsset:
		return iconSound
	}
	return iconFile
}
