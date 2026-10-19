package editor

import (
	"errors"
	"os"
	"testing"

	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
)

// spriteForTests is a sprite of one frame.
var spriteForTests = scenefile.Sprite{Picture: "assets/lamp.png", SpriteSheet: scenefile.SpriteSheet{Columns: 1, Rows: 1}}

func TestOpeningAnotherSceneKeepsUnsavedChanges(test *testing.T) {
	editor := open(test)
	other, _ := editor.Project.NewScene("Other", scenefile.Mode3D)
	editor.Add("Cube")
	if err := editor.OpenScene(other); !errors.Is(err, UnsavedChanges) {
		test.Fatalf("opening another scene with changes: %v", err)
	}
	if editor.SceneFile == other {
		test.Fatal("the unsaved scene was left")
	}
}

func TestOpeningAnotherArtFileKeepsUnsavedWork(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	tilesetPicture(test, editor, "assets/coast.png")
	editor.MakeSprite("assets/lamp.png", 3, 1)
	editor.AddAnimation()
	if err := editor.OpenArt("assets/lamp.sprite.json"); !errors.Is(err, ArtUnsaved) {
		test.Fatalf("opening over unsaved art: %v", err)
	}
	if _, err := editor.MakeTileset("assets/coast.png", 16); !errors.Is(err, ArtUnsaved) {
		test.Fatalf("making a tileset over unsaved art: %v", err)
	}
	if err := editor.CloseArt(); !errors.Is(err, ArtUnsaved) {
		test.Fatalf("closing unsaved art: %v", err)
	}
}

func TestASecondNodesSheetDoesNotOverwriteASharedSprite(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	editor.SetSprite("assets/lamp.png")
	editor.ResizeSheet(3, 1)
	shared, _ := editor.MakeSpriteOfNode()
	before, _ := os.ReadFile(editor.Project.Path(shared))
	second := editor.AddSprite("assets/lamp.png")
	editor.Select(second)
	editor.CloseArt() // with a sprite open in Art, the sheet tools change it, not the node
	editor.ResizeSheet(1, 1)
	file, err := editor.MakeSpriteOfNode()
	after, _ := os.ReadFile(editor.Project.Path(shared))
	if err != nil || file != shared || string(after) != string(before) || second.Sprite() != shared {
		test.Fatalf("the second node made %q (%v); the shared file changed %v", file, err, string(after) != string(before))
	}
}

func TestErrorCodesComeFromTheirFailures(test *testing.T) {
	editor := openFlat(test)
	if err := editor.OpenArt("assets/nowhere.sprite.json"); !errors.Is(err, ArtNotFound) {
		test.Errorf("a missing sprite: %v", err)
	}
	if err := editor.OpenArt("scenes/main.scene.json"); !errors.Is(err, ArtWrongKind) {
		test.Errorf("a scene opened as art: %v", err)
	}
	if _, err := editor.MakeSpriteOfNode(); !errors.Is(err, NoOwnSheet) {
		test.Errorf("a node with no sheet of its own: %v", err)
	}
	editor.art = artFile{path: "assets/no/such/folder/lamp.sprite.json", sprite: &spriteForTests}
	if err := editor.SaveArt(); !errors.Is(err, ArtNotSaved) {
		test.Errorf("a sprite that cannot be written: %v", err)
	}
	os.Remove(editor.Project.Path("engine.json"))
	os.Mkdir(editor.Project.Path("engine.json"), 0o755)
	if err := editor.Project.SaveSettings(); !errors.Is(err, project.SettingsNotSaved) {
		test.Errorf("settings that cannot be written: %v", err)
	}
}
