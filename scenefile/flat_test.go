package scenefile

import (
	"testing"

	"github.com/akrck02/go-engine/failure"
)

// platformScene is a 2D scene with a tile map, a player with a body and animations, and a coin
// that is a trigger.
func platformScene() *Scene {
	ground := NewNode("Ground")
	ground.Tiles = &TileMap{Tileset: "assets/tiles.png", TileSize: 16, Columns: 4, Rows: 2,
		Tiles: []int{-1, -1, 3, 3, 0, 1, 1, 2}, Solid: []int{0, 1, 2}}
	player := NewNode("Player")
	player.Sprite = "assets/hero.png"
	player.Sheet = &SpriteSheet{Columns: 5, Rows: 2, Start: "idle", Animations: []Animation{
		{Name: "idle", Frames: []int{0}, FramesPerSecond: 1, Loop: true},
		{Name: "walk", Frames: []int{0, 1, 2, 3}, FramesPerSecond: 8, Loop: true},
	}}
	floating := 0.5
	player.Body = &Body{Kind: Moving, Size: [2]float64{0.8, 0.6}, GravityScale: &floating}
	coin := NewNode("Coin")
	coin.Body = &Body{Kind: Trigger}
	return &Scene{Name: "Level", Mode: Mode2D, Nodes: []Node{ground, player, coin}}
}

func TestASceneWithBodiesTilesAndAnimationsComesBackAsWritten(test *testing.T) {
	data, err := platformScene().Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		test.Fatal(err)
	}
	ground, player, coin := read.Nodes[0], read.Nodes[1], read.Nodes[2]
	if ground.Tiles.TileAt(2, 0) != 3 || ground.Tiles.TileAt(0, 0) != EmptyTile || !ground.Tiles.IsSolid(1) || ground.Tiles.IsSolid(3) {
		test.Fatalf("the tile map came back as %+v", ground.Tiles)
	}
	walk, found := player.Sheet.Find("walk")
	if !found || walk.FramesPerSecond != 8 || len(walk.Frames) != 4 || player.Sheet.Start != "idle" {
		test.Fatalf("the sprite sheet came back as %+v", player.Sheet)
	}
	if player.Body.Kind != Moving || player.Body.Gravity() != 0.5 || coin.Body.Gravity() != 1 {
		test.Fatalf("the bodies came back as %+v and %+v", player.Body, coin.Body)
	}
}

func TestImpossible2DPartsAreRefusedWithTheirCodes(test *testing.T) {
	cases := []struct {
		name   string
		change func(scene *Scene)
		code   *failure.Code
	}{
		{"a body of no kind", func(scene *Scene) { scene.Nodes[2].Body.Kind = "ghost" }, UnknownBodyKind},
		{"tiles that don't fill the map", func(scene *Scene) { scene.Nodes[0].Tiles.Tiles = []int{1} }, TileCountWrong},
		{"tiles of no size", func(scene *Scene) { scene.Nodes[0].Tiles.TileSize = 0 }, TileSizeWrong},
		{"a sheet of no frames", func(scene *Scene) { scene.Nodes[1].Sheet.Columns = 0 }, SheetSizeWrong},
		{"a frame beyond the sheet", func(scene *Scene) { scene.Nodes[1].Sheet.Animations[1].Frames[0] = 10 }, FrameOutsideSheet},
		{"two animations with one name", func(scene *Scene) { scene.Nodes[1].Sheet.Animations[1].Name = "idle" }, SameAnimationName},
		{"a start that does not exist", func(scene *Scene) { scene.Nodes[1].Sheet.Start = "fly" }, UnknownAnimation},
	}
	for _, each := range cases {
		scene := platformScene()
		each.change(scene)
		if err := scene.Check(); !failure.Is(err, each.code) {
			test.Errorf("%s: %v, not %s", each.name, err, each.code.Name())
		}
	}
}

func TestACopyOfA2DPartIsIndependent(test *testing.T) {
	scene := platformScene()
	player, ground := scene.Nodes[1], scene.Nodes[0]
	body, sheet, tiles := player.Body.Copy(), player.Sheet.Copy(), ground.Tiles.Copy()
	*body.GravityScale = 3
	sheet.Animations[1].Frames[0] = 4
	tiles.Tiles[0] = 7
	if player.Body.Gravity() == 3 || player.Sheet.Animations[1].Frames[0] == 4 || ground.Tiles.Tiles[0] == 7 {
		test.Fatal("changing a copy changed the original")
	}
	var none *Body
	if none.Copy() != nil {
		test.Fatal("a copy of nothing is something")
	}
}
