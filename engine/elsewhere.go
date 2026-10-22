package engine

import (
	"io/fs"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// A scene played by someone other than Run (the editor's Play) shows the game's own 2D and
// interface as Run does, through these.

// DrawOverlay draws the game's 2D over a 3D scene (SetOverlay) into a list, in an area of a width
// and a height starting at 0, 0; it tells whether the game has any.
func DrawOverlay(list *draw2d.List, width, height float32) bool {
	if overlayPainter == nil {
		return false
	}
	overlayPainter(list, width, height)
	return true
}

// HasInterface tells whether the game shows a go-ui interface of its own (SetInterface).
func HasInterface() bool {
	_, found := surfaceNamed(InterfaceSurface)
	return found
}

// ShowSurfaces builds the game's surfaces (SetSurface, SetInterface) for a scene played, offering
// some scenes (ScenesOf), where the target UseSurfaces chose places them: those on the main screen
// in an area of the window, those on the second screen in another (an empty one shows none).
// Playing another scene asks the scene to be replaced (Scene.TakeReplacement).
func ShowSurfaces(context *ui.Context, played *Scene, scenes []MenuScene, main, second draw2d.Rectangle) {
	menu := Menu{&runner{played: played, menu: scenes}}
	paintSurfaces(context, menu, main, true)
	if second.Width == 0 {
		return
	}
	whole := ui.LayerOptions{Width: ui.Fixed(second.Width), Height: ui.Fixed(second.Height)}
	context.Within(second, func() {
		context.Overlay("Second screen", whole, func() { paintSecondScreen(context, menu) })
	})
}

// ScenesOf are the scenes in a project's files, in the order of their files, as the game's menu
// offers them.
func ScenesOf(files fs.FS) []MenuScene { return scenesOf(files) }
