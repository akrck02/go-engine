package editor

import (
	"time"

	ui "github.com/akrck02/go-ui"
)

// What each workspace shows between the two bars. Level is the scene's tree, its view and the
// inspector; the others are built in their own files.

// workspaceBody draws the workspace shown.
func (view *View) workspaceBody(context *ui.Context) {
	switch view.Editor.ShownWorkspace() {
	case ProjectWorkspace:
		view.projectWorkspace(context)
	case ArtWorkspace:
		view.artWorkspace(context)
	case CodeWorkspace:
		view.codeWorkspace(context)
	case ControlsWorkspace:
		view.controlsWorkspace(context)
	case PlayWorkspace:
		view.playWorkspace(context)
	case BuildWorkspace:
		view.buildWorkspace(context)
	default:
		view.levelWorkspace(context)
	}
}

// levelWorkspace is the scene's tree and assets, its view, and the inspector.
func (view *View) levelWorkspace(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		view.left(context)
		context.Separator()
		view.middle(context)
		context.Separator()
		view.inspector(context)
	})
}

// readAssetsNow reads the project's assets folder again, at most every couple of seconds.
func (view *View) readAssetsNow() {
	if time.Since(view.assetsReadAt) > assetsReadEvery {
		view.assetTree, view.assetsReadAt = view.Editor.Project.AssetTree(), time.Now()
	}
}

// assetsReadEvery is how often the assets folder is read again while it is shown.
const assetsReadEvery = 2 * time.Second

// panel is a workspace's column: a surface with a heading and room for what it holds.
func panel(context *ui.Context, name string, width ui.Sizing, content func()) {
	layout := ui.Layout{Direction: ui.Vertical, Width: width, Height: ui.Grow()}
	bar(context, layout, func() {
		inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(12), Gap: 6}
		context.ScrollArea(name, inside, content)
	})
}
