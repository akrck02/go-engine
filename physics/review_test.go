package physics

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

func TestBoxesPlacedInsideOneAnotherDoNotPushForever(test *testing.T) {
	world := NewWorld(space.Zero)
	first, second, third := box(0), box(0.1), box(0.2)
	first.Velocity = space.NewVector(2, 0, 0)
	world.Add(first)
	world.Add(second)
	world.Add(third)
	run(world, 0.5) // a ring of pushes would overflow the stack here
}

func TestAPlatformDoesNotCarryABodyThroughACeiling(test *testing.T) {
	world := NewWorld(earthGravity)
	ceiling := &Body{Kind: Solid, Position: space.NewVector(0, 4.5, 0), Size: space.NewVector(20, 1, 0)}
	world.Add(ceiling)
	carrier := lift(2, space.NewVector(0, 1, 0))
	world.Add(carrier)
	rider := standingOn(2)
	world.Add(rider)
	run(world, 3)
	if top := rider.Position.Y + 0.5; top > 4+1e-6 {
		test.Fatalf("the rider's top went to %v, through the ceiling at 4", top)
	}
	if overlap := rider.overlaps(carrier) || rider.overlaps(ceiling); overlap {
		test.Fatalf("the rider was left inside the lift or the ceiling: rider %v, lift %v", rider.Position, carrier.Position)
	}
}

func TestAPlatformStopsAgainstABodyItCannotPush(test *testing.T) {
	world := NewWorld(space.Zero)
	wall := &Body{Kind: Solid, Position: space.NewVector(5, 0, 0), Size: space.NewVector(1, 4, 0)}
	world.Add(wall)
	carrier := &Body{Kind: Platform, Position: space.Zero, Size: space.NewVector(1, 1, 0), Velocity: space.NewVector(2, 0, 0)}
	world.Add(carrier)
	squeezed := &Body{Kind: Moving, Position: space.NewVector(2, 0, 0), Size: space.NewVector(1, 1, 0)}
	world.Add(squeezed)
	run(world, 3)
	if squeezed.overlaps(carrier) || squeezed.overlaps(wall) {
		test.Fatalf("the body was crushed into the platform or the wall: body %v, platform %v", squeezed.Position, carrier.Position)
	}
}

func TestAPlatformCarriesWhatItsGravityHoldsAgainstIt(test *testing.T) {
	world := NewWorld(space.NewVector(0, 9.81, 0)) // everything falls up
	carrier := &Body{Kind: Platform, Position: space.NewVector(0, 2.25, 0), Size: space.NewVector(4, 0.5, 0), Velocity: space.NewVector(1, 0, 0)}
	world.Add(carrier)
	hanging := &Body{Kind: Moving, Position: space.NewVector(0, 1.5, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
	world.Add(hanging)
	run(world, 1)
	if math.Abs(hanging.Position.X-carrier.Position.X) > 1e-6 {
		test.Fatalf("the body under the platform is at %v and the platform at %v: pulled up against it, it was not carried", hanging.Position.X, carrier.Position.X)
	}
}

func TestAPlatformCarriesAStack(test *testing.T) {
	world := NewWorld(earthGravity)
	carrier := lift(2, space.NewVector(0, 1, 0))
	world.Add(carrier)
	lower := box(0)
	lower.Position.Y = 2.5
	world.Add(lower)
	upper := standingOn(3)
	world.Add(upper)
	run(world, 1)
	if lowerBottom, top := lower.Position.Y-0.5, carrier.Position.Y+0.25; math.Abs(lowerBottom-top) > 1e-6 || top < 2.9 {
		test.Fatalf("the box's bottom is at %v and the lift's top at %v: the box was not carried", lowerBottom, top)
	}
	if upperBottom := upper.Position.Y - 0.5; math.Abs(upperBottom-(lower.Position.Y+0.5)) > 1e-6 {
		test.Fatalf("the body on the box is at %v, the box's top at %v: the stack came apart", upperBottom, lower.Position.Y+0.5)
	}
}

func TestABodyFallingSidewaysLandsOnABox(test *testing.T) {
	world := NewWorld(space.NewVector(-9.81, 0, 0)) // everything falls left
	wall := &Body{Kind: Solid, Position: space.NewVector(-0.5, 0, 0), Size: space.NewVector(1, 20, 0)}
	world.Add(wall)
	landing := &Body{Kind: Pushable, Position: space.NewVector(0.5, 0, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
	world.Add(landing)
	falling := &Body{Kind: Moving, Position: space.NewVector(4, 0, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
	world.Add(falling)
	run(world, 2)
	if math.Abs(landing.Position.X-0.5) > 1e-6 || math.Abs(falling.Position.X-1.5) > 1e-6 || !falling.OnFloor() {
		test.Fatalf("the box is at %v and the body at %v: falling onto the box, it should stand on it, not push it", landing.Position.X, falling.Position.X)
	}
}

func TestWhereZonesOverlapTheFirstAddedPulls(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(zone(space.Zero, space.NewVector(0, 9.81, 0)))
	world.Add(zone(space.Zero, space.NewVector(9.81, 0, 0)))
	inside := crate(0)
	world.Add(inside)
	world.Step(frame)
	if inside.Pull() != space.NewVector(0, 9.81, 0) {
		test.Fatalf("in two zones the body is pulled %v, not by the first one added", inside.Pull())
	}
}

func TestMovingBodiesPassThroughEachOther(test *testing.T) {
	world := NewWorld(space.Zero)
	runner := &Body{Kind: Moving, Position: space.Zero, Size: space.NewVector(1, 1, 0), Velocity: space.NewVector(3, 0, 0)}
	standing := &Body{Kind: Moving, Position: space.NewVector(2, 0, 0), Size: space.NewVector(1, 1, 0)}
	world.Add(runner)
	world.Add(standing)
	run(world, 1)
	if runner.Position.X < 2.9 || standing.Position.X != 2 {
		test.Fatalf("the runner is at %v and the other at %v: moving bodies should pass through each other", runner.Position.X, standing.Position.X)
	}
}

func TestAStepWithBoxesPlatformsAndZonesAllocatesNothing(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	world.Add(zone(space.NewVector(-6, 3, 0), space.NewVector(0, 9.81, 0)))
	world.Add(lift(2, space.NewVector(0.5, 0, 0)))
	world.Add(standingOn(2))
	for index := 0; index < 4; index++ {
		world.Add(box(float64(index) + 3))
	}
	pusher := stroller(1.5, 2)
	world.Add(pusher)
	run(world, 0.2) // the lists the step reuses grow to their size here
	allocations := testing.AllocsPerRun(100, func() {
		pusher.Velocity.X = 2
		world.Step(frame)
	})
	if allocations != 0 {
		test.Fatalf("a step with boxes, a platform and a zone allocates %v times", allocations)
	}
}
