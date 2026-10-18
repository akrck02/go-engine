package editor

import (
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/project"
)

// The Art workspace: the project's pictures, to make into sprites and tilesets.

func (view *View) artWorkspace(context *ui.Context) {
	view.readAssetsNow()
	panel(context, "Art", ui.Grow(), func() {
		context.Heading("Pictures")
		for _, picture := range pictures(view.assetTree) {
			context.Row(func() {
				context.IconSized(iconPicture, 16, context.Theme.InkSoft)
				context.Label(picture)
			})
		}
		if len(pictures(view.assetTree)) == 0 {
			note(context, "No pictures yet: copy them into "+view.Editor.Project.Path(project.AssetsFolder))
		}
	})
}
