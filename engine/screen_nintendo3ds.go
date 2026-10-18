//go:build nintendo3ds

package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On the console a 2D scene is drawn on the top screen, as the scene's own drawing, which go-ui's
// window paints over the (empty) 3D scene; the interface is the bottom screen's.

// drawOver draws a 2D scene over its empty world, on the top screen.
func drawOver(played *Scene, _ *input.State, flat *draw2d.List) {
	if !played.Is2D() {
		return
	}
	flat.Reset()
	played.Draw2D(flat, draw2d.Rect{Width: 400, Height: 240}, played.View2D())
	played.World().Overlay.Drawing = flat
}

// flatNeedsScene is true on the console: a 2D scene is drawn over the top screen's scene.
const flatNeedsScene = true

// interfaceFor is the bottom screen's interface: nothing yet.
func interfaceFor(func() *Scene) func(context *ui.Context) { return func(*ui.Context) {} }

// consoleKeys is the game's keys on the console: the cross as the arrow keys and A as Space, read
// from the console itself, whatever the bottom screen's interface did with them (consoleInput).
func consoleKeys(console *consoleInput, window *input.State) *input.State {
	return console.keysFrom(window, ctr.ReadControls().Held)
}
