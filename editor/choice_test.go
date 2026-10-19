package editor

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/space"
)

func TestShiftAddsANodeToTheChoiceAndTakesItOut(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	ground, _ := editor.Scene.Find("Ground")
	editor.Select(cube)
	editor.ToggleChoice(ground)
	if all := editor.AllChosen(); len(all) != 2 || all[0] != cube || all[1] != ground {
		test.Fatalf("chosen: %v", all)
	}
	editor.ToggleChoice(cube)
	if editor.Selected != ground || len(editor.Chosen) != 0 {
		test.Fatalf("taking the first out left %v and %v", editor.Selected, editor.Chosen)
	}
	editor.ToggleChoice(cube)
	editor.Select(ground)
	if len(editor.AllChosen()) != 1 {
		test.Fatal("choosing one node did not let the others go")
	}
}

func TestDeletingAndDuplicatingActOnEveryNodeChosen(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	camera, _ := editor.Scene.Find("Camera")
	before := len(editor.Scene.Root.Children())
	editor.Select(cube)
	editor.ToggleChoice(camera)
	editor.Duplicate()
	if count := len(editor.Scene.Root.Children()); count != before+2 || len(editor.AllChosen()) != 2 {
		test.Fatalf("duplicating two made %d nodes, %d chosen", count-before, len(editor.AllChosen()))
	}
	editor.Delete()
	if count := len(editor.Scene.Root.Children()); count != before {
		test.Fatalf("deleting the two copies left %d nodes, not %d", count, before)
	}
}

func TestDraggingMovesEveryNodeChosen(test *testing.T) {
	editor := openFlat(test)
	square := editor.Selected
	square.Transform.Position = space.NewVector(0, 0, 1)
	other := editor.Add("Square")
	other.Transform.Position = space.NewVector(3, 0, 1)
	editor.Select(square)
	editor.ToggleChoice(other)
	flat := newFlatView()
	flat.area = draw2d.Rect{Width: 400, Height: 400}
	flat.view.Centre, flat.view.Height = space.NewVector(1.5, 0, 0), 10
	dragFrom(flat, editor, space.NewVector(0, 0, 0), space.NewVector(1, 2, 0))
	if !square.Transform.Position.Near(space.NewVector(1, 2, 1), 1e-6) || !other.Transform.Position.Near(space.NewVector(4, 2, 1), 1e-6) {
		test.Fatalf("dragged, the squares are at %v and %v", square.Transform.Position, other.Transform.Position)
	}
}

// dragFrom presses the left button over a point of the 2D view, moves to another and lets go.
func dragFrom(flat *flatView, editor *Editor, from, to space.Vector) {
	state := input.NewState()
	x, y := flat.view.ToScreen(flat.area, from)
	state.MoveCursor(float64(x), float64(y))
	state.SetButton(input.MouseLeft, true)
	flat.followMouse(editor, state)
	state.EndFrame()
	x, y = flat.view.ToScreen(flat.area, to)
	state.MoveCursor(float64(x), float64(y))
	flat.followMouse(editor, state)
	state.EndFrame()
	state.SetButton(input.MouseLeft, false)
	flat.followMouse(editor, state)
}
