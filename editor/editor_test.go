package editor

import (
	"path/filepath"
	"testing"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

func open(test *testing.T) *Editor {
	created, err := project.Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	editor, err := Open(created)
	if err != nil {
		test.Fatal(err)
	}
	return editor
}

func TestNodesAreAddedAfterTheSelectedOneWithNamesOfTheirOwn(test *testing.T) {
	editor := open(test)
	first := editor.Add("Cube")
	second := editor.Add("Cube")
	if first.Name != "Cube 2" || second.Name != "Cube 3" || second.Index() != first.Index()+1 {
		test.Fatalf("added %q at %d and %q at %d", first.Name, first.Index(), second.Name, second.Index())
	}
	if first.Transform.Position.Y != 0.5 {
		test.Fatal("a new cube is half in the ground")
	}
	camera := editor.Add("Camera")
	if _, isCamera := camera.Camera(); !isCamera {
		test.Fatal("the camera is not one")
	}
}

func TestDuplicatingAndDeletingKeepASelection(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	editor.Duplicate()
	if editor.Selected.Name != "Cube 2" || len(editor.Scene.Root.Children()) != 4 {
		test.Fatalf("duplicated as %q", editor.Selected.Name)
	}
	editor.Delete()
	if len(editor.Scene.Root.Children()) != 3 || editor.Selected == nil {
		test.Fatal("deleting left no selection")
	}
}

func TestNestingKeepsWhereANodeIsInTheWorld(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	cube.Transform = space.Transform{Position: space.NewVector(4, 1, 0), Rotation: space.RotationAround(space.Up, 90), Scale: space.NewVector(2, 2, 2)}
	editor.Select(cube)
	editor.Add("Sphere")
	sphere := editor.Selected
	sphere.Transform.Position = space.NewVector(4, 3, 0)
	editor.Nest()
	if sphere.Parent() != cube || !sphere.World().Position.Near(space.NewVector(4, 3, 0), 1e-9) {
		test.Fatalf("nested, the sphere is under %q at %v", sphere.Parent().Name, sphere.World().Position)
	}
	if !sphere.Transform.Scale.Near(space.NewVector(0.5, 0.5, 0.5), 1e-9) {
		test.Fatalf("nested, the sphere's own scale is %v", sphere.Transform.Scale)
	}
	editor.Unnest()
	if sphere.Parent() != editor.Scene.Root || !sphere.World().Position.Near(space.NewVector(4, 3, 0), 1e-9) || sphere.Index() != cube.Index()+1 {
		test.Fatal("unnested, the sphere moved")
	}
}

func TestAClickPicksTheNearestNode(test *testing.T) {
	editor := open(test)
	from := space.NewVector(0, 0.5, 10)
	picked, found := editor.Pick(from, space.Forward)
	if !found || picked.Name != "Cube" {
		test.Fatalf("picked %v", picked)
	}
	if picked, _ := editor.Pick(space.NewVector(5, 5, 10), space.NewVector(0, -0.3, -1).Normalized()); picked == nil || picked.Name != "Ground" {
		test.Fatalf("looking down at the ground picked %v", picked)
	}
}

func TestASceneIsSavedAndOpenedAgain(test *testing.T) {
	editor := open(test)
	editor.Add("Sphere")
	editor.MoveTo(editor.Selected, space.NewVector(2, 1, -3))
	editor.AttachBehaviour("Spin")
	editor.Selected.Behaviours[0].Fields["speed"] = FieldValue("120")
	if editor.Saved || editor.Save() != nil {
		test.Fatal("the scene was not saved")
	}
	if err := editor.OpenScene(editor.SceneFile); err != nil {
		test.Fatal(err)
	}
	sphere, found := editor.Scene.Find("Sphere")
	if !found || !sphere.Transform.Position.Near(space.NewVector(2, 1, -3), 1e-9) || sphere.Behaviours[0].Fields["speed"] != 120.0 {
		test.Fatalf("the sphere came back as %+v", sphere)
	}
}

func TestTheProjectsBehavioursAreOffered(test *testing.T) {
	editor := open(test)
	if definition, found := editor.FindBehaviour("Spin"); !found || definition.Fields[0].Name != "speed" {
		test.Fatalf("Spin: %+v", definition)
	}
}

func TestRenamingRefusesASiblingsName(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	if editor.Rename("Ground") || editor.Rename("a/b") || !editor.Rename("Box") {
		test.Fatal("renaming went wrong")
	}
}

func TestFieldsAreReadAsTyped(test *testing.T) {
	if FieldValue("2.5") != 2.5 || FieldValue("true") != true || FieldValue("north") != "north" {
		test.Fatal("fields read wrong")
	}
	if FieldText(2.5) != "2.5" || FieldText(nil) != "" {
		test.Fatal("fields written wrong")
	}
}

func TestPlayingInTheViewLeavesTheSceneAsItWas(test *testing.T) {
	editor := open(test)
	editor.PlayHere()
	played, playing := editor.PlayingHere()
	if !playing || played == editor.Scene {
		test.Fatal("the view does not play a copy of the scene")
	}
	cube, _ := played.Find("Cube")
	cube.Transform.Position = space.NewVector(9, 9, 9)
	editor.StopHere()
	if _, playing := editor.PlayingHere(); playing {
		test.Fatal("still playing")
	}
	edited, _ := editor.Scene.Find("Cube")
	if edited.Transform.Position.Near(space.NewVector(9, 9, 9), 0) {
		test.Fatal("playing moved the scene edited")
	}
}

func TestA2DSceneTakesSquaresAndSpritesAndIsPickedFromTheFront(test *testing.T) {
	editor := open(test)
	file, err := editor.Project.NewScene("Flat", scenefile.Mode2D)
	if err != nil || editor.OpenScene(file) != nil {
		test.Fatal(err)
	}
	if editor.KindsHere()[1] != "Square" {
		test.Fatal("a 2D scene offers 3D shapes")
	}
	front := editor.Add("Square")
	front.Transform.Position = space.NewVector(0, -2.5, 1) // over the square of the starter scene
	if picked, found := editor.PickFlat(space.NewVector(0.2, -2.4, 0), 0.3); !found || picked != front {
		test.Fatalf("picked %v", picked)
	}
	if picked, _ := editor.PickFlat(space.NewVector(0, 0, 0), 0.3); picked == nil || picked.Name != "Camera" {
		test.Fatalf("the camera at the origin was not picked: %v", picked)
	}
}

func TestASceneTurnsFrom3DTo2DAndBack(test *testing.T) {
	editor := open(test)
	editor.SetMode(scenefile.Mode2D)
	cube, _ := editor.Scene.Find("Cube")
	if !editor.Scene.Is2D() || editor.Saved || cube.Bounds().Size().Z > 0.1 {
		test.Fatal("the scene did not turn 2D")
	}
	editor.SetMode(scenefile.Mode3D)
	if editor.Scene.Is2D() || cube.Bounds().Size().Z < 0.9 {
		test.Fatal("the scene did not turn 3D again")
	}
}

// Restarter starts its scene again on its first frame.
type Restarter struct{ done bool }

func (restarter *Restarter) Update(context *engine.Context) {
	if restarter.done {
		return
	}
	restarter.done = true
	context.RestartScene()
}

func TestPlayingInTheViewFollowsTheGamesSceneChanges(test *testing.T) {
	engine.RegisterAs("Restarter", Restarter{})
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	editor.AttachBehaviour("Restarter")
	editor.PlayHere()
	first, _ := editor.PlayingHere()
	editor.AdvancePlay(1.0/60, nil)
	again, playing := editor.PlayingHere()
	if !playing || again == first {
		test.Fatal("the view did not go on to the restarted scene")
	}
}
