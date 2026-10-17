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

// interfaceFor is the bottom screen's interface: nothing yet.
func interfaceFor(*Scene) func(context *ui.Context) { return func(*ui.Context) {} }

// consoleKeys gives the game the cross as the arrow keys, read from the console itself: go-ui's
// window lets its interface claim the arrows (to move between its controls), and a game's player
// walks with them whatever the bottom screen shows.
func consoleKeys(state *input.State) {
	held := ctr.ReadControls().Held
	for _, mapping := range []struct {
		button ctr.Buttons
		key    input.Key
	}{{ctr.ButtonUp, input.KeyUp}, {ctr.ButtonDown, input.KeyDown}, {ctr.ButtonLeft, input.KeyLeft}, {ctr.ButtonRight, input.KeyRight}} {
		if down := held&mapping.button != 0; down != state.Down(mapping.key) {
			state.SetKey(mapping.key, down)
		}
	}
}
