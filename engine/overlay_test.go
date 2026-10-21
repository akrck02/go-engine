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

func TestTheSecondScreensScenePlaysAndIsSeenFromItsCamera(test *testing.T) {
	game := &runner{played: load(test)}
	below := load(test)
	SetSecondScreen(below)
	defer SetSecondScreen(nil)
	spinner, _ := below.Find("Spinner")
	before := spinner.Transform.Rotation
	state := input.NewState()
	view := game.step(state, 1.0/60)
	eye, _ := below.Find("Eye")
	if view.Below == nil || view.Below.World != below.World() || view.Below.Camera != eye.View() {
		test.Fatalf("the bottom screen shows %+v", view.Below)
	}
	game.step(state, 1.0/60)
	if spinner.Transform.Rotation == before {
		test.Fatal("the bottom scene's behaviours do not run")
	}
}

func TestAGameShownElsewhereDrawsItsOverlayAndInterface(test *testing.T) {
	var list draw2d.List
	if DrawOverlay(&list, 100, 50) || HasInterface() {
		test.Fatal("a game without an overlay or an interface says it has one")
	}
	SetOverlay(func(list *draw2d.List, width, height float32) {
		list.Rectangle(draw2d.Rect{Width: width, Height: height}, 0, draw2d.Hex("#000000"), draw2d.Transparent, 0)
	})
	defer SetOverlay(nil)
	if !DrawOverlay(&list, 100, 50) || list.Commands[0].Area.Width != 100 {
		test.Fatalf("the overlay drew %v", list.Commands)
	}
}

func TestTheSecondScreensSceneIsPlacedInACornerOfTheWindow(test *testing.T) {
	game := &runner{played: load(test)}
	SetSecondScreen(load(test))
	defer SetSecondScreen(nil)
	corner := draw2d.Rect{X: 980, Y: 520, Width: 280, Height: 180}
	PlaceSecondScreen(&corner)
	defer PlaceSecondScreen(nil)
	view := game.step(input.NewState(), 1.0/60)
	if view.Below == nil || view.Below.Area == nil || *view.Below.Area != corner {
		test.Fatalf("the second screen's scene is placed at %+v", view.Below)
	}
	PlaceSecondScreen(nil)
	if view = game.step(input.NewState(), 1.0/60); view.Below.Area != nil {
		test.Fatal("a second screen placed nowhere still has an area")
	}
}
