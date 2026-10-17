//go:build !nintendo3ds

package player

import (
	"flag"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/grid/game"
)

// options are the window's, which the command line may change (RegisterFlags).
var options = window.Options{Width: 960, Height: 640}

// RegisterFlags declares the window's flags: -width, -height, -capture, -dark and the rest.
func RegisterFlags(flags *flag.FlagSet) { options.RegisterFlags(flags) }

// Run plays a game in a window until it is closed.
func Run(played *game.Game) error {
	session := NewSession(played)
	var walking bool
	options := options
	options.Title = played.Title
	// The scene step gets the keys the interface did not claim: the arrows walk there.
	options.Scene = func(state *input.State, seconds float64) *window.SceneView {
		walking = session.Step(state, seconds)
		return nil
	}
	return window.Run(options, func(context *ui.Context) {
		if walking {
			context.RequestFrame(0) // a held arrow repeats, so keep drawing
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			session.Bar(context)
			context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
				session.Draw(list, area.Inset(16))
			})
		})
	})
}
