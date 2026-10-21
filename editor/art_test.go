package editor

import (
	"os"
	"slices"
	"testing"

	"github.com/akrck02/go-engine/scenefile"
)

func TestAPictureBecomesASpriteWithItsFrames(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	file, err := editor.MakeSprite("assets/lamp.png", 3, 1)
	if err != nil || file != "assets/lamp.sprite.json" {
		test.Fatalf("made %q, %v", file, err)
	}
	if path, isSprite, _ := editor.ArtOpen(); path != file || !isSprite {
		test.Fatalf("Art has %q open", path)
	}
	sheet, _ := editor.editedSheet()
	if sheet.FrameCount() != 3 {
		test.Fatalf("the sprite has %d frames", sheet.FrameCount())
	}
	if _, err := os.Stat(editor.Project.Path(file)); err != nil {
		test.Fatal("the sprite's file was not written")
	}
}

func TestClickedFramesMakeAnAnimation(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	editor.MakeSprite("assets/lamp.png", 3, 1)
	flicker := editor.AddAnimation()
	editor.ClearFrames(flicker)
	editor.AddFrame(flicker, 2)
	editor.AddFrame(flicker, 0)
	editor.AddFrame(flicker, 7) // the sheet has no frame 7
	sheet, _ := editor.editedSheet()
	if !slices.Equal(sheet.Animations[flicker].Frames, []int{2, 0}) || editor.ArtSaved() {
		test.Fatalf("the animation shows %v, saved %v", sheet.Animations[flicker].Frames, editor.ArtSaved())
	}
}

func TestAClickedTileTurnsSolidAndBack(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/coast.png")
	if _, err := editor.MakeTileset("assets/coast.png", 16); err != nil {
		test.Fatal(err)
	}
	editor.ToggleSolidTile(1)
	if _, _, isTileset := editor.ArtOpen(); !isTileset || !editor.art.tileset.IsSolid(1) {
		test.Fatal("the tile did not turn solid")
	}
	editor.ToggleSolidTile(1)
	if editor.art.tileset.IsSolid(1) {
		test.Fatal("the tile did not turn back")
	}
}

func TestAnEditedSpriteIsSavedAndReadBack(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	file, _ := editor.MakeSprite("assets/lamp.png", 3, 1)
	flicker := editor.AddAnimation()
	editor.RenameAnimation(flicker, "flicker")
	editor.AddFrame(flicker, 1)
	if err := editor.SaveArt(); err != nil || !editor.ArtSaved() {
		test.Fatal(err)
	}
	data, _ := os.ReadFile(editor.Project.Path(file))
	sprite, err := scenefile.ReadSprite(data)
	if err != nil || sprite.Start != "flicker" || !slices.Equal(sprite.Animations[0].Frames, []int{0, 1}) {
		test.Fatalf("the file holds %+v, %v", sprite, err)
	}
}

func TestANodesOwnSheetBecomesASpriteFile(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	editor.SetSprite("assets/lamp.png")
	editor.CloseArt()
	editor.ResizeSheet(3, 1)
	editor.AddAnimation()
	file, err := editor.MakeSpriteOfNode()
	if err != nil {
		test.Fatal(err)
	}
	if _, own := editor.Selected.Sheet(); own || editor.Selected.Sprite() != file || editor.Saved {
		test.Fatalf("the node shows %q, with a sheet of its own %v", editor.Selected.Sprite(), own)
	}
	if editor.Selected.Play("animation") == false {
		test.Fatal("the node does not play its sprite file's animation")
	}
}

func TestArtOpensTheChosenNodesSprite(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	file, _ := editor.MakeSprite("assets/lamp.png", 3, 1)
	editor.SetSprite(file)
	editor.CloseArt()
	editor.ShowWorkspace(ArtWorkspace)
	if open, _, _ := editor.ArtOpen(); open != file {
		test.Fatalf("Art opened %q, not the node's sprite", open)
	}
}

func TestArtOpensTheChosenMapsTileset(test *testing.T) {
	editor := openTileMap(test)
	tilesetPicture(test, editor, "assets/coast.png")
	file, _ := editor.MakeTileset("assets/coast.png", 16)
	editor.SetTileset(file)
	editor.CloseArt()
	editor.ShowWorkspace(ArtWorkspace)
	if open, _, isTileset := editor.ArtOpen(); open != file || !isTileset {
		test.Fatalf("Art opened %q", open)
	}
}

func TestAFrameIsTakenOutOfAnAnimation(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	editor.MakeSprite("assets/lamp.png", 3, 1)
	flicker := editor.AddAnimation()
	editor.ClearFrames(flicker)
	editor.AddFrame(flicker, 0)
	editor.AddFrame(flicker, 1)
	editor.AddFrame(flicker, 2)
	editor.RemoveFrameAt(flicker, 1)
	editor.RemoveFrameAt(flicker, 5) // no frame there: nothing changes
	sheet, _ := editor.editedSheet()
	if !slices.Equal(sheet.Animations[flicker].Frames, []int{0, 2}) {
		test.Fatalf("the animation shows %v after taking its middle frame out", sheet.Animations[flicker].Frames)
	}
}

func TestTheAnimationChosenToStartIsTheOneTheSceneStartsWith(test *testing.T) {
	editor := openFlat(test)
	tilesetPicture(test, editor, "assets/lamp.png")
	editor.MakeSprite("assets/lamp.png", 3, 1)
	first, second := editor.AddAnimation(), editor.AddAnimation()
	editor.RenameAnimation(first, "idle")
	editor.RenameAnimation(second, "flicker")
	editor.StartWith(second)
	if sheet, _ := editor.editedSheet(); sheet.Start != "flicker" {
		test.Fatalf("the sprite starts with %q", sheet.Start)
	}
}
