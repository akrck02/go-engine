package main

import (
	"os"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/space"
)

// frame is a sixtieth of a second.
const frame = 1.0 / 60

func TestMain(main *testing.M) {
	register()
	os.Exit(main.Run())
}

// level loads a level of the demo and starts it, failing on any problem.
func level(test *testing.T, file string) *engine.Scene {
	loaded, err := engine.Load(os.DirFS("."), file)
	if err != nil {
		test.Fatal(err)
	}
	if problems := loaded.Start(); len(problems) > 0 {
		test.Fatal(problems)
	}
	return loaded
}

// play moves a level on for some seconds, with the keys held.
func play(level *engine.Scene, seconds float64, held ...input.Key) {
	keys := input.NewState()
	for _, key := range held {
		keys.SetKey(key, true)
	}
	for played := 0.0; played < seconds; played += frame {
		level.Update(frame, keys)
		keys.EndFrame()
	}
}

// find finds a node of a level, failing when it is not there.
func find(test *testing.T, level *engine.Scene, path string) *engine.Node {
	node, found := level.Find(path)
	if !found {
		test.Fatalf("%s is not in %s", path, level.Name)
	}
	return node
}

func TestEveryLevelLoadsAndItsSignLeadsToAnother(test *testing.T) {
	for _, file := range []string{"scenes/forest-1.scene.json", "scenes/forest-2.scene.json", "scenes/forest-3.scene.json"} {
		sign := find(test, level(test, file), "Sign")
		flag, found := engine.Behaviour[Flag](sign)
		if !found {
			test.Fatalf("the sign of %s has no Flag", file)
		}
		if _, err := os.Stat(flag.Next); err != nil || flag.Next == file {
			test.Errorf("the sign of %s leads to %q", file, flag.Next)
		}
	}
}

func TestThePlayerLandsOnTheGroundAndStaysThere(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	play(forest, 2)
	player := find(test, forest, "Player")
	body, _ := player.Body()
	resting := player.Transform.Position
	play(forest, 1)
	if !body.OnFloor() || !player.Transform.Position.Near(resting, 1e-6) {
		test.Fatalf("the player is at %v, on the floor %v", player.Transform.Position, body.OnFloor())
	}
}

func TestThePlayerWalksAndFacesTheWayItWalks(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	play(forest, 1)
	player := find(test, forest, "Player")
	sprite := find(test, forest, "Player/Sprite")
	from := player.Transform.Position.X
	play(forest, 0.5, input.KeyLeft)
	if player.Transform.Position.X >= from-0.5 || sprite.Animation() != "walk" || sprite.FacingLeft() {
		test.Fatalf("walking left, the player went from %v to %v, playing %q, mirrored %v",
			from, player.Transform.Position.X, sprite.Animation(), sprite.FacingLeft())
	}
	play(forest, 0.1, input.KeyRight)
	if !sprite.FacingLeft() {
		test.Fatal("walking right, the slime (drawn facing left) is not mirrored")
	}
	play(forest, 0.1)
	if sprite.Animation() != "idle" {
		test.Fatalf("standing, the player plays %q", sprite.Animation())
	}
}

func TestThePlayerJumpsFromTheGroundOnly(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	play(forest, 1)
	player := find(test, forest, "Player")
	ground := player.Transform.Position.Y
	play(forest, 0.3, input.KeySpace)
	if rise := player.Transform.Position.Y - ground; rise < 2 {
		test.Fatalf("a jump rose %v metres", rise)
	}
	body, _ := player.Body()
	before := body.Velocity.Y
	play(forest, frame, input.KeySpace) // held, not pressed again, and in the air
	if body.Velocity.Y > before {
		test.Fatal("the player jumped again in the air")
	}
}

func TestTouchingACoinCollectsIt(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	player := find(test, forest, "Player")
	coin := find(test, forest, "Coin 1")
	placeOn(player, coin)
	play(forest, frame)
	if _, found := forest.Find("Coin 1"); found {
		test.Fatal("the coin is still there")
	}
	if collector, _ := engine.Behaviour[Player](player); collector.Coins != 1 {
		test.Fatalf("the player has %d coins", collector.Coins)
	}
}

func TestSpikesStartTheLevelAgain(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	placeOn(find(test, forest, "Player"), find(test, forest, "Spikes 1"))
	play(forest, frame)
	again, replaced, err := forest.Replacement()
	if err != nil || !replaced || again.Name != "Forest 1" {
		test.Fatalf("after the spikes: %v, %v, %v", again, replaced, err)
	}
}

func TestTheSignGoesToTheNextLevel(test *testing.T) {
	forest := level(test, "scenes/forest-1.scene.json")
	placeOn(find(test, forest, "Player"), find(test, forest, "Sign"))
	play(forest, frame)
	next, replaced, err := forest.Replacement()
	if err != nil || !replaced || next.Name != "Forest 2" {
		test.Fatalf("after the sign: %v, %v, %v", next, replaced, err)
	}
}

// placeOn puts a node's body where another node is, as if it had walked there.
func placeOn(walker, target *engine.Node) {
	walker.Transform.Position = space.NewVector(target.Transform.Position.X, target.Transform.Position.Y, walker.Transform.Position.Z)
	if body, found := walker.Body(); found {
		body.Position = walker.World().Position
	}
}
