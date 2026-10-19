package physics

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

// zone is a trigger ten metres a side round a point, pulling with a gravity.
func zone(centre, gravity space.Vector) *Body {
	return &Body{Kind: Trigger, Position: centre, Size: space.NewVector(10, 10, 0), ZonePull: &gravity}
}

func TestABodyFallsTheWayItsZonePulls(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	ceiling := &Body{Kind: Solid, Position: space.NewVector(0, 6.5, 0), Size: space.NewVector(20, 1, 0)}
	world.Add(ceiling)
	world.Add(zone(space.NewVector(0, 3, 0), space.NewVector(0, 9.81, 0)))
	lifted := crate(1)
	world.Add(lifted)
	run(world, 2)
	if top := lifted.Position.Y + 0.5; math.Abs(top-6) > 1e-6 {
		test.Fatalf("the crate's top is at %v: pulled up, it should rest against the ceiling at 6", top)
	}
	if !lifted.OnFloor() || lifted.Pull().Y <= 0 {
		test.Fatalf("against the ceiling it is on its floor: %v, pulled %v", lifted.OnFloor(), lifted.Pull())
	}
}

func TestABodyStandsOnTheFloorItsGravityPullsItTo(test *testing.T) {
	world := NewWorld(earthGravity)
	wall := &Body{Kind: Solid, Position: space.NewVector(4.5, 0, 0), Size: space.NewVector(1, 20, 0)}
	world.Add(wall)
	world.Add(zone(space.Zero, space.NewVector(9.81, 0, 0)))
	sideways := crate(0)
	world.Add(sideways)
	run(world, 2)
	if right := sideways.Position.X + 0.5; math.Abs(right-4) > 1e-6 || !sideways.OnFloor() {
		test.Fatalf("the crate's right side is at %v (on its floor: %v): pulled right, it should stand on the wall at 4", right, sideways.OnFloor())
	}
	if math.Abs(sideways.Position.Y) > 1e-6 {
		test.Fatalf("pulled right, the crate still fell to %v", sideways.Position.Y)
	}
}

func TestABodyOutOfTheZoneFallsAsBefore(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	world.Add(zone(space.NewVector(30, 3, 0), space.NewVector(0, 9.81, 0)))
	falling := crate(4)
	world.Add(falling)
	run(world, 2)
	if bottom := falling.Position.Y - 0.5; math.Abs(bottom) > 1e-6 || !falling.OnFloor() || falling.Pull().Y >= 0 {
		test.Fatalf("away from the zone the crate's bottom is at %v, pulled %v: it should fall to the ground", bottom, falling.Pull())
	}
}

func TestABodyThatIgnoresZonesKeepsTheWorldsGravity(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	world.Add(zone(space.NewVector(0, 3, 0), space.NewVector(0, 9.81, 0)))
	walker := crate(1)
	walker.IgnoresZones = true
	world.Add(walker)
	run(world, 1)
	if bottom := walker.Position.Y - 0.5; math.Abs(bottom) > 1e-6 || walker.Pull().Y >= 0 {
		test.Fatalf("a body that ignores zones is at %v, pulled %v: it should stand on the ground", bottom, walker.Pull())
	}
}
