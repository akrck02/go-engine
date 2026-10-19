package scenefile

import (
	"errors"
	"testing"
)

func TestASpriteComesBackAsWritten(test *testing.T) {
	sprite := &Sprite{Picture: "assets/sprites/lamp.png", SpriteSheet: SpriteSheet{Columns: 4, Rows: 1, Start: "flicker",
		Animations: []Animation{{Name: "flicker", Frames: []int{0, 1, 2, 3}, FramesPerSecond: 7, Loop: true}}}}
	data, err := sprite.Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := ReadSprite(data)
	if err != nil || read.Picture != sprite.Picture || read.Columns != 4 || read.Start != "flicker" || len(read.Animations[0].Frames) != 4 {
		test.Fatalf("the sprite came back as %+v, %v", read, err)
	}
}

func TestASpriteIsChecked(test *testing.T) {
	cases := []struct {
		data string
		code error
	}{
		{`{`, Unreadable},
		{`{"columns": 1, "rows": 1}`, NoPicture},
		{`{"picture": "a.png", "columns": 0, "rows": 1}`, SheetSizeWrong},
		{`{"picture": "a.png", "columns": 2, "rows": 1, "animations": [{"name": "a", "frames": [5]}]}`, FrameOutsideSheet},
	}
	for _, each := range cases {
		if _, err := ReadSprite([]byte(each.data)); !errors.Is(err, each.code) {
			test.Errorf("%s: %v, not %v", each.data, err, each.code)
		}
	}
}

func TestATilesetComesBackAsWritten(test *testing.T) {
	tileset := &Tileset{Picture: "assets/sprites/coast.png", TileSize: 16, Solid: []int{14, 15}}
	data, err := tileset.Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := ReadTileset(data)
	if err != nil || read.Picture != tileset.Picture || read.TileSize != 16 || !read.IsSolid(15) || read.IsSolid(0) {
		test.Fatalf("the tileset came back as %+v, %v", read, err)
	}
	if _, err := ReadTileset([]byte(`{"picture": "a.png"}`)); !errors.Is(err, TileSizeWrong) {
		test.Fatalf("a tileset without a tile size: %v", err)
	}
}

func TestFilesAreToldBySuffix(test *testing.T) {
	if !IsSpriteFile("assets/lamp.sprite.json") || IsSpriteFile("assets/lamp.png") || !IsTilesetFile("a/coast.tileset.json") {
		test.Fatal("sprite and tileset files are not told apart")
	}
}

func TestATileMapOfATilesetFileNeedsNoTileSize(test *testing.T) {
	node := NewNode("Ground")
	node.Tiles = &TileMap{Tileset: "assets/coast.tileset.json", Columns: 1, Rows: 1, Tiles: []int{0}}
	if err := (&Scene{Name: "Level", Mode: Mode2D, Nodes: []Node{node}}).Check(); err != nil {
		test.Fatal(err)
	}
}
