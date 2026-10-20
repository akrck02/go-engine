package scenefile

import (
	"testing"

	"github.com/akrck02/go-engine/failure"
)

// flatScene is a 2D scene with a tile map, a mover with a body and animations, and a trigger
// that is a trigger.
func flatScene() *Scene {
	ground := NewNode("Ground")
	ground.Tiles = &TileMap{Tileset: "assets/tiles.png", TileSize: 16, Columns: 4, Rows: 2,
		Tiles: []int{-1, -1, 3, 3, 0, 1, 1, 2}, Solid: []int{0, 1, 2}}
	mover := NewNode("Keeper")
	mover.Sprite = "assets/hero.png"
	mover.Sheet = &SpriteSheet{Columns: 5, Rows: 2, Start: "idle", Animations: []Animation{
		{Name: "idle", Frames: []int{0}, FramesPerSecond: 1, Loop: true},
		{Name: "walk", Frames: []int{0, 1, 2, 3}, FramesPerSecond: 8, Loop: true},
	}}
	floating := 0.5
	mover.Body = &Body{Kind: Moving, Size: [2]float64{0.8, 0.6}, GravityScale: &floating}
	marker := NewNode("Marker")
	marker.Body = &Body{Kind: Trigger}
	return &Scene{Name: "Level", Mode: Mode2D, Nodes: []Node{ground, mover, marker}}
}

func TestASceneWithBodiesTilesAndAnimationsComesBackAsWritten(test *testing.T) {
	data, err := flatScene().Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		test.Fatal(err)
	}
	ground, mover, marker := read.Nodes[0], read.Nodes[1], read.Nodes[2]
	if ground.Tiles.TileAt(2, 0) != 3 || ground.Tiles.TileAt(0, 0) != EmptyTile || !ground.Tiles.IsSolid(1) || ground.Tiles.IsSolid(3) {
		test.Fatalf("the tile map came back as %+v", ground.Tiles)
	}
	walk, found := mover.Sheet.Find("walk")
	if !found || walk.FramesPerSecond != 8 || len(walk.Frames) != 4 || mover.Sheet.Start != "idle" {
		test.Fatalf("the sprite sheet came back as %+v", mover.Sheet)
	}
	if mover.Body.Kind != Moving || mover.Body.Gravity() != 0.5 || marker.Body.Gravity() != 1 {
		test.Fatalf("the bodies came back as %+v and %+v", mover.Body, marker.Body)
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
		{"a light of no kind", func(scene *Scene) { scene.Nodes[2].Light = &Light{Kind: "torch"} }, UnknownLightKind},
		{"a light inner beyond outer", func(scene *Scene) { scene.Nodes[2].Light = &Light{Kind: PointLight, Inner: 4, Outer: 2} }, LightRadiiWrong},
		{"a light of negative radius", func(scene *Scene) { scene.Nodes[2].Light = &Light{Kind: PointLight, Outer: -1} }, LightRadiiWrong},
	}
	for _, each := range cases {
		scene := flatScene()
		each.change(scene)
		if err := scene.Check(); !failure.Is(err, each.code) {
			test.Errorf("%s: %v, not %s", each.name, err, each.code.Name())
		}
	}
}

func TestACopyOfA2DPartIsIndependent(test *testing.T) {
	scene := flatScene()
	mover, ground := scene.Nodes[1], scene.Nodes[0]
	body, sheet, tiles := mover.Body.Copy(), mover.Sheet.Copy(), ground.Tiles.Copy()
	*body.GravityScale = 3
	sheet.Animations[1].Frames[0] = 4
	tiles.Tiles[0] = 7
	if mover.Body.Gravity() == 3 || mover.Sheet.Animations[1].Frames[0] == 4 || ground.Tiles.Tiles[0] == 7 {
		test.Fatal("changing a copy changed the original")
	}
	var none *Body
	if none.Copy() != nil {
		test.Fatal("a copy of nothing is something")
	}
}

func TestACameraThatFollowsComesBackAsWritten(test *testing.T) {
	camera := NewNode("Camera")
	camera.Camera = &Camera{Projection: Orthographic, Height: 8, Follow: "Keeper", SmoothingSeconds: 0.25,
		Ceiling: true, Bounds: &[4]float64{-10, -5, 40, 15}}
	data, err := (&Scene{Name: "Level", Mode: Mode2D, Nodes: []Node{camera}}).Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		test.Fatal(err)
	}
	again := read.Nodes[0].Camera
	if again.Follow != "Keeper" || again.SmoothingSeconds != 0.25 || !again.Ceiling || again.Bounds == nil || again.Bounds[2] != 40 {
		test.Fatalf("the camera came back as %+v", again)
	}
}

func TestACopyOfABodyKeepsAPullOfItsOwn(test *testing.T) {
	body := &Body{Kind: Trigger, Pull: &[2]float64{0, 5}}
	copied := body.Copy()
	copied.Pull[1] = -9
	if body.Pull[1] != 5 {
		test.Fatal("changing the copy's pull changed the original's")
	}
}
