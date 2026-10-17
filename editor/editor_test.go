package editor

import (
	"path/filepath"
	"testing"

	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/space"
)

func open(t *testing.T) *Editor {
	created, err := project.Create(filepath.Join(t.TempDir(), "game"), "Game")
	if err != nil {
		t.Fatal(err)
	}
	editor, err := Open(created)
	if err != nil {
		t.Fatal(err)
	}
	return editor
}

func TestNodesAreAddedAfterTheSelectedOneWithNamesOfTheirOwn(t *testing.T) {
	editor := open(t)
	first := editor.Add("Cube")
	second := editor.Add("Cube")
	if first.Name != "Cube 2" || second.Name != "Cube 3" || second.Index() != first.Index()+1 {
		t.Fatalf("added %q at %d and %q at %d", first.Name, first.Index(), second.Name, second.Index())
	}
	if first.Transform.Position.Y != 0.5 {
		t.Fatal("a new cube is half in the ground")
	}
	camera := editor.Add("Camera")
	if _, isCamera := camera.Camera(); !isCamera {
		t.Fatal("the camera is not one")
	}
}

func TestDuplicatingAndDeletingKeepASelection(t *testing.T) {
	editor := open(t)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	editor.Duplicate()
	if editor.Selected.Name != "Cube 2" || len(editor.Scene.Root.Children()) != 4 {
		t.Fatalf("duplicated as %q", editor.Selected.Name)
	}
	editor.Delete()
	if len(editor.Scene.Root.Children()) != 3 || editor.Selected == nil {
		t.Fatal("deleting left no selection")
	}
}

func TestNestingKeepsWhereANodeIsInTheWorld(t *testing.T) {
	editor := open(t)
	cube, _ := editor.Scene.Find("Cube")
	cube.Transform = space.Transform{Position: space.V(4, 1, 0), Rotation: space.AxisAngle(space.Up, 90), Scale: space.V(2, 2, 2)}
	editor.Select(cube)
	editor.Add("Sphere")
	sphere := editor.Selected
	sphere.Transform.Position = space.V(4, 3, 0)
	editor.Nest()
	if sphere.Parent() != cube || !sphere.World().Position.Near(space.V(4, 3, 0), 1e-9) {
		t.Fatalf("nested, the sphere is under %q at %v", sphere.Parent().Name, sphere.World().Position)
	}
	if !sphere.Transform.Scale.Near(space.V(0.5, 0.5, 0.5), 1e-9) {
		t.Fatalf("nested, the sphere's own scale is %v", sphere.Transform.Scale)
	}
	editor.Unnest()
	if sphere.Parent() != editor.Scene.Root || !sphere.World().Position.Near(space.V(4, 3, 0), 1e-9) || sphere.Index() != cube.Index()+1 {
		t.Fatal("unnested, the sphere moved")
	}
}

func TestAClickPicksTheNearestNode(t *testing.T) {
	editor := open(t)
	from := space.V(0, 0.5, 10)
	picked, found := editor.Pick(from, space.Forward)
	if !found || picked.Name != "Cube" {
		t.Fatalf("picked %v", picked)
	}
	if picked, _ := editor.Pick(space.V(5, 5, 10), space.V(0, -0.3, -1).Normalized()); picked == nil || picked.Name != "Ground" {
		t.Fatalf("looking down at the ground picked %v", picked)
	}
}

func TestASceneIsSavedAndOpenedAgain(t *testing.T) {
	editor := open(t)
	editor.Add("Sphere")
	editor.MoveTo(editor.Selected, space.V(2, 1, -3))
	editor.AttachBehaviour("Spin")
	editor.Selected.Behaviours[0].Fields["speed"] = FieldValue("120")
	if editor.Saved || editor.Save() != nil {
		t.Fatal("the scene was not saved")
	}
	if err := editor.OpenScene(editor.SceneFile); err != nil {
		t.Fatal(err)
	}
	sphere, found := editor.Scene.Find("Sphere")
	if !found || !sphere.Transform.Position.Near(space.V(2, 1, -3), 1e-9) || sphere.Behaviours[0].Fields["speed"] != 120.0 {
		t.Fatalf("the sphere came back as %+v", sphere)
	}
}

func TestTheProjectsBehavioursAreOffered(t *testing.T) {
	editor := open(t)
	if info, found := editor.BehaviourInfo("Spin"); !found || info.Fields[0].Name != "speed" {
		t.Fatalf("Spin: %+v", info)
	}
}

func TestRenamingRefusesASiblingsName(t *testing.T) {
	editor := open(t)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	if editor.Rename("Ground") || editor.Rename("a/b") || !editor.Rename("Box") {
		t.Fatal("renaming went wrong")
	}
}

func TestFieldsAreReadAsTyped(t *testing.T) {
	if FieldValue("2.5") != 2.5 || FieldValue("true") != true || FieldValue("north") != "north" {
		t.Fatal("fields read wrong")
	}
	if FieldText(2.5) != "2.5" || FieldText(nil) != "" {
		t.Fatal("fields written wrong")
	}
}
