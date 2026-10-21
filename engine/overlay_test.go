package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

func TestTheGamesOverlayIsDrawnOverA3DSceneAtTheScreensSize(test *testing.T) {
	game := &runner{played: load(test)}
	var seenWidth, seenHeight float32
	SetOverlay(func(list *draw2d.List, width, height float32) {
		seenWidth, seenHeight = width, height
		list.Rectangle(draw2d.Rect{Width: width, Height: 8}, 0, draw2d.Hex("#00ff00"), draw2d.Transparent, 0)
	})
	defer SetOverlay(nil)
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 400, 240
	view := game.step(state, 1.0/60)
	if view == nil || view.Drawing.IsEmpty() || seenWidth != 400 || seenHeight != 240 {
		test.Fatalf("the overlay was drawn at %v × %v into %v", seenWidth, seenHeight, view)
	}
	game.step(state, 1.0/60)
	if len(view.Drawing.Commands) != 1 {
		test.Fatalf("the overlay keeps %d commands after a second frame: it is not made again", len(view.Drawing.Commands))
	}
}
