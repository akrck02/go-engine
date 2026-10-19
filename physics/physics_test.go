package physics

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

// frame is one frame at sixty frames a second.
const frame = 1.0 / 60

// earthGravity pulls down at 9.81 metres a second, each second.
var earthGravity = space.NewVector(0, -9.81, 0)

// ground is a solid floor 20 metres wide whose top is at height 0.
func ground() *Body {
	return &Body{Kind: Solid, Position: space.NewVector(0, -0.5, 0), Size: space.NewVector(20, 1, 0)}
}

// crate is a moving box a metre a side, its centre at a height.
func crate(height float64) *Body {
	return &Body{Kind: Moving, Position: space.NewVector(0, height, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
}

func run(world *World, seconds float64) {
	for elapsed := 0.0; elapsed < seconds; elapsed += frame {
		world.Step(frame)
	}
}

func TestAFallingBodyLandsOnTheGroundAndStaysThere(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	falling := crate(5)
	world.Add(falling)
	run(world, 3)
	if bottom := falling.Position.Y - falling.Size.Y/2; math.Abs(bottom) > 1e-9 {
		test.Fatalf("the crate's bottom is at %v, not on the ground at 0", bottom)
	}
	if falling.Velocity.Y != 0 {
		test.Fatalf("the crate on the ground still falls at %v", falling.Velocity.Y)
	}
}

func TestAFastBodyDoesNotPassThroughAThinWall(test *testing.T) {
	world := NewWorld(space.Zero)
	wall := &Body{Kind: Solid, Position: space.NewVector(5, 0, 0), Size: space.NewVector(0.1, 4, 0)}
	world.Add(wall)
	bullet := &Body{Kind: Moving, Position: space.NewVector(0, 0, 0), Size: space.NewVector(0.2, 0.2, 0),
		Velocity: space.NewVector(600, 0, 0)} // ten metres in one frame
	world.Add(bullet)
	world.Step(frame)
	if right := bullet.Position.X + bullet.Size.X/2; math.Abs(right-(5-0.05)) > 1e-9 {
		test.Fatalf("the bullet's front is at %v: it should stop against the wall at 4.95", right)
	}
}

func TestABodyPushedIntoAWallSlidesUpIt(test *testing.T) {
	world := NewWorld(space.Zero)
	world.Add(&Body{Kind: Solid, Position: space.NewVector(2, 0, 0), Size: space.NewVector(1, 10, 0)})
	climber := &Body{Kind: Moving, Position: space.NewVector(0, 0, 0), Size: space.NewVector(1, 1, 0),
		Velocity: space.NewVector(3, 2, 0)}
	world.Add(climber)
	run(world, 1)
	if climber.Position.X != 1 {
		test.Fatalf("the body went into the wall: its centre is at x %v", climber.Position.X)
	}
	if climber.Position.Y < 1.9 {
		test.Fatalf("the body stopped instead of sliding up: y %v", climber.Position.Y)
	}
}

func TestABodyOnTheGroundIsOnTheFloor(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	standing := crate(0.5)
	world.Add(standing)
	world.Step(frame)
	if !standing.OnFloor() {
		test.Fatal("a body standing on the ground is not on the floor")
	}
	standing.Velocity.Y = 5 // a jump
	world.Step(frame)
	if standing.OnFloor() {
		test.Fatal("a body jumping off the ground is still on the floor")
	}
}

func TestASolidNeverMoves(test *testing.T) {
	world := NewWorld(earthGravity)
	floor := ground()
	floor.Velocity = space.NewVector(1, 1, 0)
	world.Add(floor)
	run(world, 1)
	if !floor.Position.Near(space.NewVector(0, -0.5, 0), 0) {
		test.Fatalf("the solid moved to %v", floor.Position)
	}
}

func TestAStepAllocatesNothing(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	for index := 0; index < 10; index++ {
		world.Add(crate(float64(index) * 2))
	}
	allocations := testing.AllocsPerRun(100, func() { world.Step(frame) })
	if allocations != 0 {
		test.Fatalf("a step allocates %v times", allocations)
	}
}

func TestGravityCanChange(test *testing.T) {
	world := NewWorld(earthGravity)
	falling := crate(5)
	world.Add(falling)
	world.SetGravity(space.NewVector(0, 9.81, 0)) // upside down
	run(world, 1)
	if falling.Position.Y <= 5 || world.Gravity() != space.NewVector(0, 9.81, 0) {
		test.Fatalf("with gravity up the crate is at %v", falling.Position.Y)
	}
}

func TestABodyWithNoSizeDoesNotPassThroughAThinWall(test *testing.T) {
	world := NewWorld(space.Vector{})
	wall := &Body{Kind: Solid, Position: space.NewVector(5, 0, 0), Size: space.NewVector(0.1, 10, 0)}
	point := &Body{Kind: Moving, Velocity: space.NewVector(600, 0, 0)}
	world.Add(wall)
	world.Add(point)
	world.Step(1.0 / 60) // ten metres in one step
	if point.Position.X > 5 {
		test.Fatalf("a body with no size went through the wall to %v", point.Position.X)
	}
}
