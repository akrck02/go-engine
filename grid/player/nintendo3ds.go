//go:build nintendo3ds

package player

import (
	"flag"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/grid/game"
)

// The top screen's size: the scene is drawn there, over an empty 3D scene, as go-ui's window
// shows a program's drawing over its scene.
const topWidth, topHeight = 400, 240

// RegisterFlags does nothing on the console, which has no command line.
func RegisterFlags(flags *flag.FlagSet) {}

// Run plays a game on the console until START ends it: the scene on the top screen, the bar on
// the bottom one.
func Run(played *game.Game) error {
	session := NewSession(played)
	var drawing draw2d.List
	empty := scene.New()
	empty.Overlay.Drawing = &drawing
	view := &window.SceneView{World: empty}
	view.Camera.Position, view.Camera.Target, view.Camera.Up = [4]float32{0, 0, 1, 1}, [4]float32{0, 0, 0, 1}, [4]float32{0, 1, 0, 0}
	options := window.Options{Title: played.Title}
	options.Scene = func(state *input.State, seconds float64) *window.SceneView {
		session.Step(state, seconds)
		drawing.Reset()
		drawing.Rectangle(draw2d.Rect{Width: topWidth, Height: topHeight}, 0, draw2d.Color{0, 0, 0, 1}, draw2d.Color{}, 0)
		session.Draw(&drawing, draw2d.Rect{X: 4, Y: 4, Width: topWidth - 8, Height: topHeight - 8})
		return view
	}
	return window.Run(options, func(context *ui.Context) {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			session.Bar(context)
			context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(12), Direction: ui.Vertical, Gap: 4}, func() {
				context.Caption(played.Title)
				context.Caption("Cross: walk   START: quit")
			})
		})
	})
}
