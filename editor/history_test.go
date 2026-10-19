package editor

import (
	"testing"
	"time"

	"github.com/akrck02/go-engine/space"
)

// clockAt makes the editor's clock a time that a test moves on by hand.
func clockAt(editor *Editor) *time.Time {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	editor.now = func() time.Time { return now }
	return &now
}

func TestAChangeIsUndoneAndRedone(test *testing.T) {
	editor := open(test)
	now := clockAt(editor)
	before := len(editor.Scene.Root.Children())
	editor.Add("Cube")
	*now = now.Add(time.Second)
	editor.Add("Sphere")
	editor.Undo()
	if count := len(editor.Scene.Root.Children()); count != before+1 {
		test.Fatalf("one undo left %d nodes, not %d", count, before+1)
	}
	editor.Undo()
	if count := len(editor.Scene.Root.Children()); count != before || !editor.Saved {
		test.Fatalf("two undos left %d nodes (saved %v)", count, editor.Saved)
	}
	editor.Redo()
	editor.Redo()
	if _, found := editor.Scene.Find("Sphere"); !found || editor.Saved {
		test.Fatal("redoing did not bring the sphere back")
	}
}

func TestADragIsOneStepToUndo(test *testing.T) {
	editor := open(test)
	now := clockAt(editor)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	start := cube.Transform.Position
	for step := 1; step <= 10; step++ {
		*now = now.Add(100 * time.Millisecond)
		editor.MoveTo(cube, space.NewVector(float64(step), 0.5, 0))
	}
	editor.Undo()
	again, _ := editor.Scene.Find("Cube")
	if !again.Transform.Position.Near(start, 1e-9) || editor.Selected == nil || editor.Selected.Name != "Cube" {
		test.Fatalf("one undo took the cube to %v, with %v chosen", again.Transform.Position, editor.Selected)
	}
}
