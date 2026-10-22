package editor

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAGamesOwn2DIsMovedOntoTheStageUnderTheInterface(test *testing.T) {
	var game, list draw2d.List
	game.Rectangle(draw2d.Rectangle{X: 10, Y: 20, Width: 30, Height: 5}, 0, draw2d.MustHex("#ffffff"), draw2d.Transparent, 0)
	game.Text("score", 10, 40, 12, draw2d.Regular, draw2d.MustHex("#ffffff"))
	stage := draw2d.Rectangle{X: 200, Y: 100, Width: 600, Height: 400}
	appendMoved(&list, &game, stage)
	box, text := list.Commands[0], list.Commands[1]
	if box.Area.X != 210 || box.Area.Y != 120 || text.Baseline != 140 {
		test.Fatalf("the game's 2D is at %v and its text on %v, not moved onto the stage", box.Area, text.Baseline)
	}
	if box.Layer != gameLayer || !box.HasClip || box.Clip != stage {
		test.Fatalf("the game's 2D is on layer %d, clipped %v to %v", box.Layer, box.HasClip, box.Clip)
	}
}
