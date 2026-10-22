package engine

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// tilesetPicture is a tileset of two 16-pixel tiles side by side: 0 is grass, 1 is dirt.
func tilesetPicture() []byte {
	picture := image.NewNRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			picture.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 8), G: 200, B: 0, A: 255})
		}
	}
	var encoded bytes.Buffer
	png.Encode(&encoded, picture)
	return encoded.Bytes()
}

// tiledLevel is a 2D scene at 16 pixels a metre (a tile is a metre): a map 6 tiles wide and 3 high
// whose top-left corner is at (-3, 1), with a solid floor along its bottom row (two runs, a gap
// between them) and a crate falling onto it.
func tiledLevel(test *testing.T) *Scene {
	empty := scenefile.EmptyTile
	description := &scenefile.Scene{Name: "Tiled", Mode: scenefile.Mode2D, PixelsPerMetre: 16, Nodes: []scenefile.Node{
		{Name: "Ground", Position: [3]float64{-3, 1, 0}, Scale: [3]float64{1, 1, 1}, Tiles: &scenefile.TileMap{
			Tileset: "assets/tiles.png", TileSize: 16, Columns: 6, Rows: 3, Solid: []int{1},
			Tiles: []int{
				empty, empty, empty, empty, empty, empty,
				empty, empty, 0, 0, empty, empty,
				1, 1, 1, empty, 1, 1,
			}}},
		{Name: "Crate", Shape: "cube", Position: [3]float64{-1.5, 4, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}},
	}}
	built, err := Build(fstest.MapFS{"assets/tiles.png": {Data: tilesetPicture()}}, description)
	if err != nil {
		test.Fatal(err)
	}
	return built
}

func TestSolidTilesStopABody(test *testing.T) {
	level := tiledLevel(test)
	level.Start()
	playFor(level, 3)
	crate, _ := level.Find("Crate")
	// the bottom row's top is at 1 - 2 = -1, so the crate's centre rests half a metre above it
	if math.Abs(crate.Transform.Position.Y-(-0.5)) > 1e-6 {
		test.Fatalf("the crate rests at %v, not on the solid tiles", crate.Transform.Position.Y)
	}
}

func TestARowOfSolidTilesIsOneBox(test *testing.T) {
	level := tiledLevel(test)
	level.Start()
	ground, _ := level.Find("Ground")
	if len(ground.tileBodies) != 2 {
		test.Fatalf("the solid row made %d bodies, not one per run of tiles", len(ground.tileBodies))
	}
	first := ground.tileBodies[0]
	if first.Size.X != 3 || first.Size.Y != 1 || first.Position.X != -1.5 || first.Position.Y != -1.5 {
		test.Fatalf("the first run is %+v", first)
	}
}

func TestOnlyTheTilesInSightAreDrawn(test *testing.T) {
	level := tiledLevel(test)
	ground, _ := level.Find("Ground")
	if size := ground.Bounds().Size(); size.X != 6 || size.Y != 3 {
		test.Fatalf("the map's box is %v", size)
	}
	var whole, narrow draw2d.List
	area := draw2d.Rectangle{Width: 160, Height: 160} // 16 pixels a metre on screen: ten metres
	level.DrawContent2D(&whole, area, View2D{Height: 10})
	// two metres around a point over the floor's left end: two of its tiles
	level.DrawContent2D(&narrow, area, View2D{Centre: space.NewVector(-2, -1.5, 0), Height: 2})
	if images(&whole) != 7 {
		test.Fatalf("the whole map drew %d tiles, not its 7", images(&whole))
	}
	if images(&narrow) != 2 {
		test.Fatalf("a view of its corner drew %d tiles", images(&narrow))
	}
}

func images(list *draw2d.List) int {
	count := 0
	for _, command := range list.Commands {
		if command.Picture != nil {
			count++
		}
	}
	return count
}

func TestDrawingA2DFrameAllocatesNothing(test *testing.T) {
	level := tiledLevel(test)
	var list draw2d.List
	area := draw2d.Rectangle{Width: 160, Height: 160}
	level.DrawContent2D(&list, area, View2D{Height: 10}) // the tileset is cut once, the list grows once
	allocations := testing.AllocsPerRun(100, func() {
		list.Reset()
		level.DrawContent2D(&list, area, View2D{Height: 10})
	})
	if allocations != 0 {
		test.Fatalf("drawing a frame allocates %v times", allocations)
	}
}

func TestATileMapTakesItsSolidTilesFromItsTileset(test *testing.T) {
	tileset := scenefile.Tileset{Picture: "assets/tiles.png", TileSize: 16, Solid: []int{1}}
	data, _ := tileset.Marshal()
	empty := scenefile.EmptyTile
	description := &scenefile.Scene{Name: "Tiled", Mode: scenefile.Mode2D, PixelsPerMetre: 16, Nodes: []scenefile.Node{
		{Name: "Ground", Position: [3]float64{-3, 1, 0}, Scale: [3]float64{1, 1, 1}, Tiles: &scenefile.TileMap{
			Tileset: "assets/tiles.tileset.json", Columns: 3, Rows: 2, Tiles: []int{empty, empty, empty, 1, 1, 1}}},
		{Name: "Crate", Shape: "cube", Position: [3]float64{-1.5, 4, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}},
	}}
	files := fstest.MapFS{"assets/tiles.png": {Data: tilesetPicture()}, "assets/tiles.tileset.json": {Data: data}}
	built, err := Build(files, description)
	if err != nil {
		test.Fatal(err)
	}
	built.Start()
	playFor(built, 2)
	crate, _ := built.Find("Crate")
	if body, _ := crate.Body(); !body.OnFloor() || math.Abs(crate.Transform.Position.Y-0.5) > 1e-6 {
		test.Fatalf("the crate is at %v, on the floor %v", crate.Transform.Position, body.OnFloor())
	}
}

func TestAMissingPictureIsTriedOnceAndCostsNothingAfter(test *testing.T) {
	description := &scenefile.Scene{Name: "Missing", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "Ghost", Sprite: "assets/nowhere.png", Scale: [3]float64{1, 1, 1}},
	}}
	built, _ := Build(fstest.MapFS{}, description)
	var list draw2d.List
	area := draw2d.Rectangle{Width: 100, Height: 100}
	built.Draw2D(&list, area, built.View2D())
	allocations := testing.AllocsPerRun(50, func() {
		list.Reset()
		built.Draw2D(&list, area, built.View2D())
	})
	if allocations != 0 {
		test.Fatalf("drawing a missing picture allocates %v times a frame", allocations)
	}
}

func TestATilesetHasTheColumnsOfItsPicture(test *testing.T) {
	level := tiledLevel(test)
	ground, _ := level.Find("Ground")
	if columns := level.TilesetColumns(ground); columns != 2 {
		test.Fatalf("a 32-pixel picture of 16-pixel tiles has %d columns", columns)
	}
}
