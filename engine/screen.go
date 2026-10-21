//go:build !nintendo3ds && !android

package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On the desktop a 2D scene is painted by the window's interface, which go-ui paints over the
// (empty) 3D scene: the whole window is its area.

// flatNeedsScene is false on the desktop: a 2D scene is painted by the interface alone, and a
// window given no scene draws only its interface.
const flatNeedsScene = false

// drawOver draws nothing over the scene on the desktop: the interface does it (interfaceFor).
func drawOver(*Scene, *input.State, *draw2d.List) {}

// interfaceFor is what the window's interface shows: a 2D scene, filling it.
func interfaceFor(game *runner) func(context *ui.Context) {
	return func(context *ui.Context) {
		played := game.played
		if !played.Is2D() {
			if gameInterface != nil {
				gameInterface(context, Menu{game}) // over the 3D scene, and over the game's overlay
			}
			return
		}
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
			played.Draw2D(list, area, played.View2D())
		})
	}
}

// consoleKeys is the window's own keys on the desktop, where the keyboard is the game's.
func consoleKeys(_ *consoleInput, window *input.State) *input.State { return window }

// readConsole is an empty reading: there is no 3DS here.
func readConsole() ctr.Reading { return ctr.Reading{} }

// startConsole does nothing: there is no 3DS here.
func startConsole() {}

// stopConsole does nothing: there is no 3DS here.
func stopConsole() {}
