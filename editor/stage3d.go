package editor

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

// A 3D game played in the view is drawn behind the interface, in the stage at its target's shape:
// a computer's screen, or the 3DS's top screen with its bottom one under it, where the game's
// second screen's scene (engine.SetSecondScreen) plays. Over it, the game's own 2D (engine.SetOverlay) and
// its interface (engine.SetInterface) are shown as the game shows them: over its window, or on the
// 3DS's bottom screen.

// stagedScreens are where the stage showed a 3D game last frame, in window pixels: the game's
// screen, and the 3DS's bottom one (empty for a computer's shape).
type stagedScreens struct{ top, bottom draw2d.Rect }

// stage3D shows a 3D game at its target's shape, with the game's 2D and interface over it.
func (view *View) stage3D(context *ui.Context, played *engine.Scene) {
	console := view.Editor.Target == ConsoleScreen
	_, hasBelow := engine.SecondScreen()
	ink := context.Theme.InkSoft
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
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
	if engine.HasInterface() {
		view.gameInterface(context, played, console)
	}
}

// gameInterface shows the game's interface where the game shows it: over its window's corner, or
// filling the 3DS's bottom screen. It is placed where the stage was last frame, as the layout
// places the stage after this.
func (view *View) gameInterface(context *ui.Context, played *engine.Scene, console bool) {
	area := view.staged.top
	options := ui.LayerOptions{Padding: ui.Padding{Left: area.X, Top: area.Y}, Width: ui.Fixed(area.Width)}
	if console {
		area = view.staged.bottom
		options = ui.LayerOptions{Padding: ui.Padding{Left: area.X, Top: area.Y}, Width: ui.Fixed(area.Width), Height: ui.Fixed(area.Height)}
	}
	context.Overlay("Game interface", options, func() {
		engine.ShowInterface(context, played, view.scenesOffered())
	})
}

// playedView is how the window draws a 3D game played in the view: its scene in the stage's screen
// and, with the 3DS's shape, its bottom scene under it.
func (view *View) playedView(world *scene.Scene, camera models.Camera) *window.SceneView {
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
func appendMoved(list, moved *draw2d.List, area draw2d.Rect) {
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
