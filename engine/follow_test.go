package engine

import (
	"math"
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// followedLevel is a 2D scene at 16 pixels a metre (a tile a metre): a map 20 tiles wide and 6 high
// from (0, 6) to (20, 0), a runner, and a camera seeing 4 metres that follows it with a smoothing.
func followedLevel(test *testing.T, smoothing float64) *Scene {
	return followedLevelWith(test, &scenefile.Camera{Projection: scenefile.Orthographic, Height: 4, Follow: "Runner", SmoothingSeconds: smoothing})
}

// followedLevelWith is followedLevel with a camera of its own.
func followedLevelWith(test *testing.T, camera *scenefile.Camera) *Scene {
	description := &scenefile.Scene{Name: "Followed", Mode: scenefile.Mode2D, PixelsPerMetre: 16, Nodes: []scenefile.Node{
		{Name: "Ground", Position: [3]float64{0, 6, 0}, Scale: [3]float64{1, 1, 1}, Tiles: &scenefile.TileMap{
			Tileset: "assets/tiles.png", TileSize: 16, Columns: 20, Rows: 6, Tiles: make([]int, 120)}},
		{Name: "Runner", Position: [3]float64{5, 2, 0}, Scale: [3]float64{1, 1, 1}},
		{Name: "Camera", Position: [3]float64{5, 2, 0}, Scale: [3]float64{1, 1, 1},
			Camera: camera},
	}}
	built, err := Build(fstest.MapFS{"assets/tiles.png": {Data: tilesetPicture()}}, description)
	if err != nil {
		test.Fatal(err)
	}
	built.Start()
	var list draw2d.List
	built.DrawContent2D(&list, draw2d.Rectangle{Width: 100, Height: 100}, built.View2D()) // a square screen
	return built
}

func TestTheCameraFollowsItsTargetAndStopsAtTheLevelsEdge(test *testing.T) {
	level := followedLevel(test, 0)
	runner, _ := level.Find("Runner")
	camera, _ := level.Find("Camera")
	runner.Transform.Position = space.NewVector(9, 3, 0)
	playFor(level, 0.1)
	if !camera.Transform.Position.Near(space.NewVector(9, 3, 0), 1e-9) {
		test.Fatalf("the camera is at %v, not on the runner", camera.Transform.Position)
	}
	runner.Transform.Position = space.NewVector(19.5, 0.5, 0) // the level's bottom-right corner
	playFor(level, 0.1)
	// the view is 4 metres high and, on a square screen, 4 wide: its centre stays 2 from each edge
	if !camera.Transform.Position.Near(space.NewVector(18, 2, 0), 1e-9) {
		test.Fatalf("at the level's corner the camera is at %v, not at (18, 2)", camera.Transform.Position)
	}
}

func TestTheSkyAboveTheTileMapsStaysOpen(test *testing.T) {
	level := followedLevel(test, 0)
	runner, _ := level.Find("Runner")
	camera, _ := level.Find("Camera")
	runner.Transform.Position = space.NewVector(9, 20, 0) // far above the map, whose top is at 6
	playFor(level, 0.1)
	if !camera.Transform.Position.Near(space.NewVector(9, 20, 0), 1e-9) {
		test.Fatalf("above the tile maps the camera is at %v, not on the runner", camera.Transform.Position)
	}
}

func TestACeilingKeepsTheViewUnderTheTileMapsTop(test *testing.T) {
	level := followedLevelWith(test, &scenefile.Camera{Projection: scenefile.Orthographic, Height: 4, Follow: "Runner", Ceiling: true})
	runner, _ := level.Find("Runner")
	camera, _ := level.Find("Camera")
	runner.Transform.Position = space.NewVector(9, 20, 0)
	playFor(level, 0.1)
	// the map's top is at 6 and the view 4 high: its centre stays 2 under it
	if !camera.Transform.Position.Near(space.NewVector(9, 4, 0), 1e-9) {
		test.Fatalf("under a ceiling the camera is at %v, not at (9, 4)", camera.Transform.Position)
	}
}

func TestASmoothCameraCatchesUpOverTime(test *testing.T) {
	level := followedLevel(test, 0.5)
	runner, _ := level.Find("Runner")
	camera, _ := level.Find("Camera")
	runner.Transform.Position = space.NewVector(9, 2, 0)
	playFor(level, 0.1)
	if x := camera.Transform.Position.X; x <= 5 || x >= 8 {
		test.Fatalf("after a tenth of a second a smooth camera is at x %v", x)
	}
	playFor(level, 3)
	if math.Abs(camera.Transform.Position.X-9) > 0.01 {
		test.Fatalf("after three seconds the camera is at x %v, not on the runner", camera.Transform.Position.X)
	}
}

func TestAStartedSceneSeesThroughItsFollowingCamera(test *testing.T) {
	level := followedLevel(test, 0)
	if view := level.View2D(); view.Height != 4 || !view.Centre.Near(space.NewVector(5, 2, 0), 1e-9) {
		test.Fatalf("the scene sees %+v", view)
	}
}

func TestFollowingAllocatesNothing(test *testing.T) {
	level := followedLevel(test, 0.5)
	playFor(level, 1)
	allocations := testing.AllocsPerRun(100, func() { level.Update(1.0/60, nil) })
	if allocations != 0 {
		test.Fatalf("a frame with a following camera allocates %v times", allocations)
	}
}

func TestACamerasBoundsAreOpenAboveUnlessItHasACeiling(test *testing.T) {
	level := followedLevel(test, 0)
	camera, _ := level.Find("Camera")
	bounds, found := level.CameraBounds(camera)
	if !found || bounds.Minimum.Y != 0 || !math.IsInf(bounds.Maximum.Y, 1) || bounds.Maximum.X != 20 {
		test.Fatalf("the camera's bounds are %v", bounds)
	}
	settings, _ := camera.Camera()
	settings.Ceiling = true
	if bounds, _ := level.CameraBounds(camera); bounds.Maximum.Y != 6 {
		test.Fatalf("with a ceiling the bounds are %v", bounds)
	}
}

func TestANarrowScreenSeesMoreHeightToKeepTheCamerasWidth(test *testing.T) {
	level := followedLevelWith(test, &scenefile.Camera{Projection: scenefile.Orthographic, Height: 4, Width: 4})
	var list draw2d.List
	level.DrawContent2D(&list, draw2d.Rectangle{Width: 50, Height: 100}, level.View2D()) // twice as tall as wide
	if view := level.View2D(); math.Abs(view.Height-8) > 1e-9 {
		test.Fatalf("a screen half as wide as tall sees %v metres high, not 8", view.Height)
	}
	level.DrawContent2D(&list, draw2d.Rectangle{Width: 200, Height: 100}, level.View2D())
	if view := level.View2D(); view.Height != 4 {
		test.Fatalf("a wide screen sees %v metres high, not the camera's 4", view.Height)
	}
}
