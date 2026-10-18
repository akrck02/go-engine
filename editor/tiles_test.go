package editor

import (
	"testing"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// openTileMap opens a new 2D scene with a new tile map chosen, its corner at the origin, at the
// scene's 32 pixels a metre and tiles of 32 pixels: a tile a metre.
func openTileMap(test *testing.T) *Editor {
	editor := openFlat(test)
	editor.Add("Tile map")
	editor.MoveTo(editor.Selected, space.NewVector(0, 0, 0))
	editor.SetTileSize(32)
	return editor
}

func TestAnAddedTileMapIsEmpty(test *testing.T) {
	editor := openTileMap(test)
	tiles, found := editor.Selected.Tiles()
	if !found || tiles.Columns == 0 || tiles.Rows == 0 || len(tiles.Tiles) != tiles.Columns*tiles.Rows {
		test.Fatalf("the tile map is %+v", tiles)
	}
	for _, tile := range tiles.Tiles {
		if tile != scenefile.EmptyTile {
			test.Fatal("a new tile map has tiles")
		}
	}
}

func TestPaintingATileAndErasingIt(test *testing.T) {
	editor := openTileMap(test)
	// a metre a tile, rows going down from the corner: (2.5, -1.5) is column 2, row 1
	if !editor.PaintTileAt(space.NewVector(2.5, -1.5, 0), 3) {
		test.Fatal("painting changed nothing")
	}
	if editor.PaintTileAt(space.NewVector(2.5, -1.5, 0), 3) {
		test.Fatal("painting the same tile again is a change")
	}
	if editor.PaintTileAt(space.NewVector(-5, 5, 0), 3) {
		test.Fatal("painting outside the map is a change")
	}
	editor.ToggleSolid(3)
	tiles := saveAndOpenAgain(test, editor, editor.Selected.Path())
	saved, _ := tiles.Tiles()
	if saved.TileAt(2, 1) != 3 || !saved.IsSolid(3) {
		test.Fatalf("the tile came back as %d, solid %v", saved.TileAt(2, 1), saved.IsSolid(3))
	}
	editor.Select(tiles)
	editor.PaintTileAt(space.NewVector(2.5, -1.5, 0), scenefile.EmptyTile)
	if saved.TileAt(2, 1) != scenefile.EmptyTile {
		test.Fatal("the tile was not erased")
	}
}

func TestAResizedTileMapKeepsItsTilesFromTheCorner(test *testing.T) {
	editor := openTileMap(test)
	editor.PaintTileAt(space.NewVector(1.5, -0.5, 0), 4) // column 1, row 0
	editor.PaintTileAt(space.NewVector(9.5, -0.5, 0), 5) // column 9, row 0: cut by a smaller map
	editor.ResizeTileMap(4, 2)
	tiles, _ := editor.Selected.Tiles()
	if tiles.Columns != 4 || tiles.Rows != 2 || len(tiles.Tiles) != 8 || tiles.TileAt(1, 0) != 4 {
		test.Fatalf("resized, the map is %+v", tiles)
	}
	editor.ResizeTileMap(6, 3)
	if tiles.TileAt(1, 0) != 4 || tiles.TileAt(5, 2) != scenefile.EmptyTile {
		test.Fatalf("grown, the map is %+v", tiles)
	}
}

func TestPaintingIsOnlyForATileMap(test *testing.T) {
	editor := openTileMap(test)
	editor.PaintingTiles = true
	if !editor.PaintingHere() {
		test.Fatal("a chosen tile map is not painted")
	}
	editor.Select(editor.Add("Square"))
	if editor.PaintingHere() {
		test.Fatal("a square is painted")
	}
}
