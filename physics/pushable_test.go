package physics

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

// box is a pushable box a metre a side, standing on the ground at a place across.
func box(across float64) *Body {
	return &Body{Kind: Pushable, Position: space.NewVector(across, 0.5, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
}

// stroller is a moving body a metre a side on the ground, walking at a speed.
func stroller(across, speed float64) *Body {
	return &Body{Kind: Moving, Position: space.NewVector(across, 0.5, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1,
		Velocity: space.NewVector(speed, 0, 0)}
}

func TestABodyPushesABoxUntilAWallStopsBoth(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	wall := &Body{Kind: Solid, Position: space.NewVector(6.5, 2, 0), Size: space.NewVector(1, 4, 0)}
	world.Add(wall)
	pushed := box(2)
	world.Add(pushed)
	pusher := stroller(0, 2)
	world.Add(pusher)
	for elapsed := 0.0; elapsed < 0.5; elapsed += frame {
		pusher.Velocity.X = 2
		world.Step(frame)
	}
	if pushed.Position.X <= 2 {
		test.Fatalf("the box was not pushed: it is at %v", pushed.Position.X)
	}
	if gap := pushed.Position.X - pusher.Position.X; math.Abs(gap-1) > 1e-6 {
		test.Fatalf("the walker and the box are %v apart, not side by side", gap)
	}
	for elapsed := 0.0; elapsed < 5; elapsed += frame {
		pusher.Velocity.X = 2
		world.Step(frame)
	}
	if right := pushed.Position.X + 0.5; math.Abs(right-6) > 1e-6 {
		test.Fatalf("the box's right side is at %v, not against the wall at 6", right)
	}
	if right := pusher.Position.X + 0.5; math.Abs(right-5) > 1e-6 {
		test.Fatalf("the walker's right side is at %v, not against the box at 5", right)
	}
}

func TestABodyStandsOnABox(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	world.Add(box(0))
	falling := crate(4)
	world.Add(falling)
	run(world, 2)
	if bottom := falling.Position.Y - 0.5; math.Abs(bottom-1) > 1e-6 || !falling.OnFloor() {
		test.Fatalf("the crate's bottom is at %v (on the floor: %v), not standing on the box at 1", bottom, falling.OnFloor())
	}
}

func TestABoxFallsAndLandsOnAnother(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	world.Add(box(0))
	upper := box(0)
	upper.Position.Y = 5
	world.Add(upper)
	run(world, 2)
	if bottom := upper.Position.Y - 0.5; math.Abs(bottom-1) > 1e-6 {
		test.Fatalf("the upper box's bottom is at %v, not on the lower box at 1", bottom)
	}
}

func TestABodyPushesARowOfBoxes(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	first, second := box(1), box(2)
	world.Add(first)
	world.Add(second)
	pusher := stroller(0, 0)
	world.Add(pusher)
	for elapsed := 0.0; elapsed < 0.5; elapsed += frame {
		pusher.Velocity.X = 2
		world.Step(frame)
	}
	if first.Position.X <= 1 || second.Position.X <= 2 || math.Abs(second.Position.X-first.Position.X-1) > 1e-6 {
		test.Fatalf("the boxes are at %v and %v: both should move, side by side", first.Position.X, second.Position.X)
	}
}

func TestATriggerNoticesABoxPushedOntoIt(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	button := &Body{Kind: Trigger, Position: space.NewVector(3, 0.25, 0), Size: space.NewVector(1, 0.5, 0)}
	world.Add(button)
	pushed := box(1)
	world.Add(pushed)
	pusher := stroller(-0.5, 0)
	world.Add(pusher)
	noticed := false
	for elapsed := 0.0; elapsed < 2 && !noticed; elapsed += frame {
		pusher.Velocity.X = 2
		world.Step(frame)
		for _, touch := range world.Begun() {
			noticed = noticed || touch.Trigger == button && touch.Other == pushed
		}
	}
	if !noticed {
		test.Fatal("a box pushed onto a trigger is not noticed")
	}
}
