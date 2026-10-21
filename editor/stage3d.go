package editor

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// A 3D game played in the view is drawn behind the interface, filling the window; over the part
// of it the stage leaves open, the game's own 2D (engine.SetOverlay) and its interface
// (engine.SetInterface) are shown as the game shows them over its window.

// stage3D leaves the stage open on the 3D game, with the game's 2D and interface over it.
func (view *View) stage3D(context *ui.Context, played *engine.Scene) {
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
		view.playedArea = area
		view.overlay.Reset()
		if engine.DrawOverlay(&view.overlay, area.Width, area.Height) {
			appendMoved(list, &view.overlay, area)
		}
	})
	if !engine.HasInterface() {
		return
	}
	area := view.playedArea // where the stage was last frame: the layout places it after this
	options := ui.LayerOptions{Padding: ui.Padding{Left: area.X, Top: area.Y}, Width: ui.Fixed(area.Width)}
	context.Overlay("Game interface", options, func() {
		engine.ShowInterface(context, played, view.scenesOffered())
	})
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
