package game

import (
	"fmt"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// room is a game of one scene laid out from rows of letters, with the sample's kinds.
func room(rows ...string) *Game {
	sample := Sample()
	sample.Scenes = []Scene{sceneFrom("Room", "#ffffff", rows)}
	return sample
}

func walk(world *World, directions ...Direction) {
	for _, direction := range directions {
		world.Move(direction)
	}
}

func playerAt(world *World) (int, int) {
	player := world.things[world.player]
	return player.x, player.y
}

func TestThePlayerWalksButNotIntoWallsOrOutOfTheScene(t *testing.T) {
	world := NewWorld(room(
		"#####",
		"#K..#",
		"#####",
	))
	walk(world, Up, Left, Right, Right, Right)
	if x, y := playerAt(world); x != 3 || y != 1 || world.Steps != 2 {
		t.Fatalf("the player is at %d,%d after %d steps", x, y, world.Steps)
	}
	open := NewWorld(room("K..", "...", "..."))
	walk(open, Left, Up)
	if x, y := playerAt(open); x != 0 || y != 0 || open.Steps != 0 {
		t.Fatalf("the player left the scene: %d,%d", x, y)
	}
}

func TestThePlayerPushesACrateUnlessSomethingIsBehindIt(t *testing.T) {
	world := NewWorld(room(
		"######",
		"#KC..#",
		"#KCC.#",
		"######",
	))
	walk(world, Right)
	if x, _ := playerAt(world); x != 2 || world.pushableAt(3, 1) < 0 {
		t.Fatalf("the crate was not pushed: player at %d", x)
	}
	walk(world, Right, Right)
	if x, _ := playerAt(world); x != 3 || world.pushableAt(4, 1) < 0 {
		t.Fatalf("the player pushed a crate into the wall: player at %d", x)
	}
}

func TestTheGoalOpensOnceEverythingIsCollectedAndTheLastOneWins(t *testing.T) {
	game := room("KLO..")
	game.Scenes = append(game.Scenes, sceneFrom("Second", "#ffffff", []string{"K.L"}))
	world := NewWorld(game)
	walk(world, Right)
	if world.SceneNumber() != 0 || world.Notice == "" {
		t.Fatal("the goal opened before the oil was collected")
	}
	walk(world, Right, Left)
	if world.SceneNumber() != 1 || world.Steps != 0 {
		t.Fatalf("the goal did not lead on: scene %d", world.SceneNumber())
	}
	walk(world, Right, Right)
	if world.State != Won {
		t.Fatal("the last goal did not win")
	}
	walk(world, Left)
	if x, _ := playerAt(world); x != 2 {
		t.Fatal("the player walked after winning")
	}
}

func TestAHazardStartsTheSceneAgain(t *testing.T) {
	world := NewWorld(room("KOR.."))
	walk(world, Right, Right)
	if x, _ := playerAt(world); x != 0 || world.Collected != 0 || world.Steps != 0 {
		t.Fatalf("the scene did not start again: player at %d, %d collected", x, world.Collected)
	}
}

func TestAGameComesBackAsWritten(t *testing.T) {
	data, err := Sample().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if read.Title != Sample().Title || len(read.Scenes) != 2 || len(read.Scenes[1].Things) != len(Sample().Scenes[1].Things) {
		t.Fatal("the game came back different")
	}
	if _, err := Read([]byte(`{"scenes": []}`)); err == nil {
		t.Fatal("a game with no scenes was read")
	}
	if _, err := Read([]byte(`{"scenes": [{"name": "A", "width": 3, "height": 3, "things": [{"kind": "Ghost"}]}]}`)); err == nil {
		t.Fatal("a thing of an unknown kind was read")
	}
}

func TestRenamingAndRemovingAKindReachesItsThings(t *testing.T) {
	game := room("KCC")
	if !game.RenameKind("Crate", "Barrel") || game.Scenes[0].Things[1].Kind != "Barrel" {
		t.Fatal("the things kept the old name")
	}
	if game.RenameKind("Barrel", "Wall") {
		t.Fatal("two kinds got one name")
	}
	game.RemoveKind("Barrel")
	if len(game.Scenes[0].Things) != 1 {
		t.Fatalf("%d things left", len(game.Scenes[0].Things))
	}
}

func TestPlacingFillsACellOnceAndResizingForgetsWhatNoLongerFits(t *testing.T) {
	scene := Scene{Width: 5, Height: 5}
	scene.Place("Wall", 4, 4)
	scene.Place("Crate", 4, 4)
	scene.Place("Wall", 9, 9)
	if len(scene.Things) != 1 || scene.Things[0].Kind != "Crate" {
		t.Fatalf("the cell holds %v", scene.Things)
	}
	scene.Resize(3, 100)
	if len(scene.Things) != 0 || scene.Height != MaximumSize {
		t.Fatalf("after resizing: %v, height %d", scene.Things, scene.Height)
	}
}

func TestTheGridFitsInWholePixelsAndFindsCells(t *testing.T) {
	grid := Fit(draw2d.Rect{X: 10, Y: 0, Width: 400, Height: 240}, &Scene{Width: 12, Height: 8})
	if grid.Cell != 30 || grid.X != 30 || grid.Y != 0 {
		t.Fatalf("grid %+v", grid)
	}
	if column, row, found := grid.CellAt(31, 239); !found || column != 0 || row != 7 {
		t.Fatalf("cell %d,%d", column, row)
	}
	if _, _, found := grid.CellAt(29, 10); found {
		t.Fatal("a point left of the grid was on it")
	}
}

func TestHoldingAnArrowWalks(t *testing.T) {
	var walker Walker
	steps := 0
	for frame := 0; frame < 60; frame++ { // a second at 60 frames
		if _, step := walker.Step(Right, 1.0/60); step {
			steps++
		}
	}
	if steps < 5 || steps > 6 { // at once, then at 0.3, 0.45, 0.6, 0.75 and 0.9 seconds

		t.Fatalf("%d steps in a second", steps)
	}
	if _, step := walker.Step(Direction{}, 1.0/60); step {
		t.Fatal("a step with no arrow held")
	}
	if _, step := walker.Step(Right, 1.0/60); !step {
		t.Fatal("pressing again did not step at once")
	}
	// a window drawing only on events: the frame after the press may come late
	if _, step := walker.Step(Right, 2); step {
		t.Fatal("a late frame took a second step")
	}
}

func TestTheSampleCanBeWon(t *testing.T) {
	game := Sample()
	for scene := range game.Scenes {
		if !solvable(game, scene) {
			t.Fatalf("scene %q cannot be finished", game.Scenes[scene].Name)
		}
	}
}

// solvable searches every way of playing a scene, breadth first, for one that finishes it.
func solvable(game *Game, scene int) bool {
	start := NewWorld(game)
	start.Start(scene)
	seen := map[string]bool{}
	queue := [][]Direction{nil}
	for len(queue) > 0 && len(seen) < 100000 {
		path := queue[0]
		queue = queue[1:]
		for _, direction := range []Direction{Up, Down, Left, Right} {
			world := NewWorld(game)
			world.Start(scene)
			walk(world, append(append([]Direction{}, path...), direction)...)
			if world.State == Won || world.SceneNumber() != scene {
				return true
			}
			key := fmt.Sprint(world.things)
			if seen[key] {
				continue
			}
			seen[key] = true
			queue = append(queue, append(append([]Direction{}, path...), direction))
		}
	}
	return false
}
