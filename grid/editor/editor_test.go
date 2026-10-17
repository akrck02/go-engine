package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akrck02/go-engine/grid/game"
)

func TestPaintingPlacesTheBrushAndTheEraserEmpties(t *testing.T) {
	editor := New(game.Sample(), "")
	editor.ChooseBrush("Crate")
	editor.Paint(5, 5)
	scene := editor.Scene()
	if index, found := scene.ThingAt(5, 5); !found || scene.Things[index].Kind != "Crate" || editor.Saved {
		t.Fatal("the crate was not painted")
	}
	editor.ChooseBrush("")
	editor.Paint(5, 5)
	if _, found := scene.ThingAt(5, 5); found {
		t.Fatal("the eraser left the crate")
	}
}

func TestScenesAreAddedMovedAndRemovedButOneStays(t *testing.T) {
	editor := New(game.Sample(), "")
	editor.AddScene()
	if editor.SceneIndex() != 1 || len(editor.Game.Scenes) != 3 || editor.Scene().Width != 12 {
		t.Fatalf("the new scene is %d of %d", editor.SceneIndex(), len(editor.Game.Scenes))
	}
	editor.MoveScene(1)
	if editor.SceneIndex() != 2 || editor.Game.Scenes[2].Name != "Scene 3" {
		t.Fatal("the scene did not move")
	}
	for range 5 {
		editor.RemoveScene()
	}
	if len(editor.Game.Scenes) != 1 || !editor.Failed {
		t.Fatalf("%d scenes left", len(editor.Game.Scenes))
	}
}

func TestKindsAreAddedRenamedAndRemoved(t *testing.T) {
	editor := New(game.Sample(), "")
	editor.AddKind()
	if editor.Brush() != "Kind 7" {
		t.Fatalf("painting with %q", editor.Brush())
	}
	editor.Paint(1, 2)
	if !editor.RenameKind("Bush") || editor.Brush() != "Bush" || editor.RenameKind("Wall") {
		t.Fatal("renaming went wrong")
	}
	editor.RemoveKind()
	if _, found := editor.Game.Kind("Bush"); found || editor.Brush() != "" {
		t.Fatal("the kind stayed")
	}
	if _, found := editor.Scene().ThingAt(1, 2); found {
		t.Fatal("its thing stayed")
	}
}

func TestAGameIsSavedAndOpenedAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.json")
	editor, err := Open(path)
	if err != nil || editor.Saved {
		t.Fatalf("a new game: %v", err)
	}
	editor.ChooseBrush("Oil")
	editor.Paint(10, 1)
	if err := editor.Save(); err != nil {
		t.Fatal(err)
	}
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if index, found := opened.Scene().ThingAt(10, 1); !found || opened.Scene().Things[index].Kind != "Oil" || !opened.Saved {
		t.Fatal("the game came back different")
	}
	os.WriteFile(path, []byte("{"), 0o644)
	if _, err := Open(path); err == nil {
		t.Fatal("a broken file opened")
	}
}

func TestPlayingStartsFromTheSceneShown(t *testing.T) {
	editor := New(game.Sample(), "")
	editor.chooseScene(1)
	editor.Play()
	session, playing := editor.Playing()
	if !playing || session.World.SceneNumber() != 1 {
		t.Fatal("play did not start from the scene shown")
	}
	editor.Stop()
	if _, playing := editor.Playing(); playing {
		t.Fatal("still playing")
	}
}
