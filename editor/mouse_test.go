package editor

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// viewOver is a 2D view 400 pixels square, ten metres high, centred on a point.
func viewOver(centre space.Vector) *flatView {
	flat := newFlatView()
	flat.area = draw2d.Rectangle{Width: 400, Height: 400}
	flat.view.Centre, flat.view.Height = centre, 10
	return flat
}

// press holds a button over a point of the view for a frame, with keys held too.
func press(flat *flatView, editor *Editor, state *input.State, button input.MouseButton, point space.Vector, held ...input.Key) {
	for _, key := range held {
		state.Feed().SetKey(key, true)
	}
	x, y := flat.view.ToScreen(flat.area, point)
	state.Feed().MoveCursor(float64(x), float64(y))
	state.Feed().SetButton(button, true)
	flat.followMouse(editor, state)
	state.Feed().EndFrame()
}

// release lets a button go where the pointer is.
func release(flat *flatView, editor *Editor, state *input.State, button input.MouseButton) {
	state.Feed().SetButton(button, false)
	flat.followMouse(editor, state)
	state.Feed().EndFrame()
}

func TestTheRightButtonErasesWithATileTool(test *testing.T) {
	editor := openTileMap(test)
	tiles, _ := editor.Selected.Tiles()
	editor.PaintTileAt(space.NewVector(2.5, -1.5, 0), 4)
	flat, state := viewOver(space.NewVector(4, -4, 0)), input.NewState()
	editor.Tool = PaintTool
	press(flat, editor, state, input.MouseRight, space.NewVector(2.5, -1.5, 0))
	release(flat, editor, state, input.MouseRight)
	if tiles.TileAt(2, 1) != scenefile.EmptyTile {
		test.Fatal("the right button did not erase")
	}
}

func TestTheHandAndTheMiddleButtonMoveTheView(test *testing.T) {
	editor := openFlat(test)
	for _, way := range []struct {
		tool   Tool
		button input.MouseButton
	}{{HandTool, input.MouseLeft}, {SelectTool, input.MouseMiddle}} {
		editor.Tool = way.tool
		flat, state := viewOver(space.Zero), input.NewState()
		press(flat, editor, state, way.button, space.NewVector(0, 0, 0))
		press(flat, editor, state, way.button, space.NewVector(2, 1, 0))
		release(flat, editor, state, way.button)
		if !flat.view.Centre.Near(space.NewVector(-2, -1, 0), 1e-6) {
			test.Errorf("%s with button %d left the view at %v", way.tool.Name(), way.button, flat.view.Centre)
		}
	}
}

func TestFillIsAClick(test *testing.T) {
	editor := openTileMap(test)
	editor.ResizeTileMap(3, 2)
	tiles, _ := editor.Selected.Tiles()
	flat, state := viewOver(space.NewVector(1.5, -1, 0)), input.NewState()
	editor.Tool, editor.Brush = FillTool, 6
	press(flat, editor, state, input.MouseLeft, space.NewVector(0.5, -0.5, 0))
	release(flat, editor, state, input.MouseLeft)
	for _, tile := range tiles.Tiles {
		if tile != 6 {
			test.Fatalf("one click filled %v", tiles.Tiles)
		}
	}
}

func TestShiftClickJoinsTheChoiceAndABackgroundClickLetsGo(test *testing.T) {
	editor := openFlat(test)
	square := editor.Selected
	square.Transform.Position = space.NewVector(0, 0, 1)
	other := editor.Add("Square")
	other.Transform.Position = space.NewVector(3, 0, 1)
	editor.Select(square)
	flat, state := viewOver(space.NewVector(1.5, 0, 0)), input.NewState()
	press(flat, editor, state, input.MouseLeft, space.NewVector(3, 0, 0), input.KeyShift)
	release(flat, editor, state, input.MouseLeft)
	if len(editor.AllChosen()) != 2 {
		test.Fatalf("a shift-click chose %d nodes", len(editor.AllChosen()))
	}
	state.Feed().SetKey(input.KeyShift, false)
	press(flat, editor, state, input.MouseLeft, space.NewVector(1.5, 4, 0))
	release(flat, editor, state, input.MouseLeft)
	if editor.Selected != nil && editor.Selected.Name != "Camera" {
		test.Fatalf("a click on the background chose %v", editor.Selected)
	}
}
