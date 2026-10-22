package editor

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"

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
	editor.Tool = PaintTool
	if !editor.PaintingHere() {
		test.Fatal("a chosen tile map is not painted")
	}
	editor.Select(editor.Add("Square"))
	if editor.PaintingHere() {
		test.Fatal("a square is painted")
	}
}

func TestFillPaintsTheRegionOfTheTileClicked(test *testing.T) {
	editor := openTileMap(test)
	editor.ResizeTileMap(4, 3)
	for column := 0; column < 4; column++ { // a wall across the middle row
		editor.PaintTileAt(space.NewVector(float64(column)+0.5, -1.5, 0), 2)
	}
	if !editor.FillTileAt(space.NewVector(0.5, -0.5, 0), 5) {
		test.Fatal("filling changed nothing")
	}
	tiles, _ := editor.Selected.Tiles()
	for column := 0; column < 4; column++ {
		if tiles.TileAt(column, 0) != 5 || tiles.TileAt(column, 1) != 2 || tiles.TileAt(column, 2) != scenefile.EmptyTile {
			test.Fatalf("filled, the map is %v", tiles.Tiles)
		}
	}
}

// tilesetPicture writes a picture of three 16-pixel tiles side by side into the project.
func tilesetPicture(test *testing.T, editor *Editor, file string) {
	picture := image.NewNRGBA(image.Rect(0, 0, 48, 16))
	for x := 0; x < 48; x++ {
		for y := 0; y < 16; y++ {
			picture.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 5), G: 120, B: 60, A: 255})
		}
	}
	var encoded bytes.Buffer
	png.Encode(&encoded, picture)
	if err := os.WriteFile(editor.Project.Path(file), encoded.Bytes(), 0o644); err != nil {
		test.Fatal(err)
	}
}

func TestThePaletteKeepsTheTilesetsColumns(test *testing.T) {
	editor := openTileMap(test)
	tilesetPicture(test, editor, "assets/tiles.png")
	editor.SetTileset("assets/tiles.png")
	editor.SetTileSize(16)
	if columns := editor.PaletteColumns(); columns != 3 {
		test.Fatalf("the palette has %d columns, not the tileset's 3", columns)
	}
}

// clickAt presses and lets go the left button over a point of the 2D view.
func clickAt(flat *flatView, editor *Editor, point space.Vector) {
	x, y := flat.view.ToScreen(flat.area, point)
	state := input.NewState()
	state.Feed().MoveCursor(float64(x), float64(y))
	state.Feed().SetButton(input.MouseLeft, true)
	flat.followMouse(editor, state)
	state.Feed().EndFrame()
	state.Feed().SetButton(input.MouseLeft, false)
	flat.followMouse(editor, state)
}

func TestThePaintToolPaintsAndTheSelectToolChooses(test *testing.T) {
	editor := openTileMap(test)
	tiles, _ := editor.Selected.Tiles()
	flat := newFlatView()
	flat.area = draw2d.Rectangle{Width: 400, Height: 400}
	flat.view.Centre, flat.view.Height = space.NewVector(4, -4, 0), 10
	editor.Tool, editor.Brush = PaintTool, 1
	clickAt(flat, editor, space.NewVector(2.5, -1.5, 0))
	if tiles.TileAt(2, 1) != 1 {
		test.Fatal("the paint tool did not paint")
	}
	square := editor.Add("Square")
	square.Transform.Position = space.NewVector(6, -6, 1)
	editor.Select(nil)
	editor.Tool = SelectTool
	clickAt(flat, editor, space.NewVector(6, -6, 0))
	if editor.Selected != square {
		test.Fatalf("the select tool chose %v", editor.Selected)
	}
}
