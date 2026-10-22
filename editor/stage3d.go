package editor

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

// A 3D game played in the view is drawn behind the interface, in the stage at its target's shape:
// a computer's or a phone's screen, or the 3DS's top screen with its bottom one under it, where the
// game's second screen's scene (engine.SetSecondScreen) plays. Over it, the game's own 2D
// (engine.SetOverlay) and its surfaces (engine.SetSurface, engine.SetInterface) are shown where
// that target shows them.

// stagedScreens are where the stage showed a 3D game last frame, in window pixels: the game's
// screen, and the 3DS's bottom one (empty for a computer's shape).
type stagedScreens struct{ top, bottom draw2d.Rectangle }

// stage3D shows a 3D game at its target's shape, with the game's 2D and interface over it.
func (view *View) stage3D(context *ui.Context, played *engine.Scene) {
	console := view.Editor.Target == Nintendo3DSScreens
	_, hasBelow := engine.SecondScreen()
	ink := context.Theme.InkSoft
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rectangle, list *draw2d.List) {
		view.staged = stagedScreens{top: fitted(area, view.Editor.Target.Aspect())}
		if console {
			view.staged.top, view.staged.bottom = consoleScreens(area)
		}
		view.overlay.Reset()
		if engine.DrawOverlay(&view.overlay, view.staged.top.Width, view.staged.top.Height) {
			appendMoved(list, &view.overlay, view.staged.top)
		}
		if console && !hasBelow {
			list.Rectangle(view.staged.bottom, 0, bottomScreenColor, ink, 1)
		}
	})
	view.gameSurfaces(context, played)
}

// gameSurfaces shows the game's surfaces (its interface among them) where the target shown places
// them: over its screen, in windows and sheets there, or on the 3DS's second screen. They are laid
// out in those screens as in the game's own window, placed where the stage was last frame (the
// layout places it after this).
func (view *View) gameSurfaces(context *ui.Context, played *engine.Scene) {
	if view.staged.top.Width == 0 {
		return // the stage is not laid out yet
	}
	engine.ShowSurfaces(context, played, view.scenesOffered(), view.staged.top, view.staged.bottom)
}

// playedView is how the window draws a 3D game played in the view: its scene in the stage's screen
// and, with the 3DS's shape, its bottom scene under it.
func (view *View) playedView(world *scene.Scene, camera camera.Camera) *window.SceneView {
	staged := &view.staged
	shown := &window.SceneView{World: world, Camera: camera}
	if staged.top.Width == 0 {
		return shown // the stage is not laid out yet: the whole window, for a frame
	}
	shown.Area = &staged.top
	below, found := engine.SecondScreen()
	if !found || staged.bottom.Width == 0 {
		return shown
	}
	view.below = window.BelowView{World: below.World(), Camera: below.Camera(), Area: &staged.bottom}
	shown.Below = &view.below
	return shown
}

// gameLayer is the layer of the game's own 2D: under every layer of the interface, the game's
// own interface included, as Go UI paints it under the interface in the game's window.
const gameLayer = -1

// appendMoved adds a list's commands to another, moved into an area, kept inside it, and under the
// interface.
func appendMoved(list, moved *draw2d.List, area draw2d.Rectangle) {
	list.PushClip(area)
	for _, command := range moved.Commands {
		command.Area.X, command.Area.Y = command.Area.X+area.X, command.Area.Y+area.Y
		command.Clip.X, command.Clip.Y = command.Clip.X+area.X, command.Clip.Y+area.Y
		command.Baseline += area.Y
		command.Layer = gameLayer
		if !command.HasClip {
			command.Clip, command.HasClip = area, true
		}
		list.Commands = append(list.Commands, command)
	}
	list.PopClip()
}

// scenesOffered are the project's scenes, for the game's interface to offer, read the first time
// it asks.
func (view *View) scenesOffered() []engine.MenuScene {
	if view.playedScenes == nil {
		view.playedScenes = engine.ScenesOf(view.Editor.Project.Files())
	}
	return view.playedScenes
}
