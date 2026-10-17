//go:build !nintendo3ds

package engine

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On the desktop a 2D scene is painted by the window's interface, which go-ui paints over the
// (empty) 3D scene: the whole window is its area.

// drawOver draws nothing over the scene on the desktop: the interface does it (interfaceFor).
func drawOver(*Scene, *input.State, *draw2d.List) {}

// interfaceFor is what the window's interface shows: a 2D scene, filling it.
func interfaceFor(played *Scene) func(context *ui.Context) {
	return func(context *ui.Context) {
		if !played.Is2D() {
			return
		}
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
			played.Draw2D(list, area, played.View2D())
		})
	}
}

// consoleKeys does nothing on the desktop, where the keyboard is the game's.
func consoleKeys(*input.State) {}
