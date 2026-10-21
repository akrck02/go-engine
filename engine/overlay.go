package engine

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// A 3D game draws its own 2D over the scene each frame (a life bar, a sight, a pause screen): over
// the window on the desktop and phones, under the game's interface (SetInterface); on the 3DS's
// top screen, in each eye when the slider shows depth. 2D scenes are drawn whole by Draw2D and do
// not use it.

// overlayPainter draws the game's 2D over a 3D scene, when the game sets one.
var overlayPainter func(list *draw2d.List, width, height float32)

// SetOverlay sets what the game draws over a 3D scene each frame, in the pixels of the screen it
// is on (the window's, or the 3DS's top screen, 400 × 240): the list is empty when it is called.
// nil removes it.
func SetOverlay(paint func(list *draw2d.List, width, height float32)) { overlayPainter = paint }

// paintOverlay draws the game's 2D for this frame into the list the runner keeps.
func (game *runner) paintOverlay(state *input.State) *draw2d.List {
	game.overlay.Reset()
	overlayPainter(&game.overlay, float32(state.WindowWidth), float32(state.WindowHeight))
	return &game.overlay
}
