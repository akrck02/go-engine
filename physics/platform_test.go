package physics

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

// lift is a platform four metres wide and half a metre tall whose top is at a height, moving at a
// velocity.
func lift(height float64, velocity space.Vector) *Body {
	return &Body{Kind: Platform, Position: space.NewVector(0, height-0.25, 0), Size: space.NewVector(4, 0.5, 0), Velocity: velocity}
}

// standingOn is a moving body a metre a side, standing on a top at a height.
func standingOn(height float64) *Body {
	return &Body{Kind: Moving, Position: space.NewVector(0, height+0.5, 0), Size: space.NewVector(1, 1, 0), GravityScale: 1}
}

func TestAPlatformCarriesWhatStandsOnIt(test *testing.T) {
	world := NewWorld(earthGravity)
	carrier := lift(2, space.NewVector(1, 0, 0))
	world.Add(carrier)
	rider := standingOn(2)
	world.Add(rider)
	run(world, 1)
	if math.Abs(rider.Position.X-carrier.Position.X) > 1e-6 || !rider.OnFloor() {
		test.Fatalf("the rider is at %v and the platform at %v (on the floor: %v): it was not carried", rider.Position.X, carrier.Position.X, rider.OnFloor())
	}
	if math.Abs(carrier.Position.X-1) > 0.05 {
		test.Fatalf("the platform went to %v, not a metre in a second", carrier.Position.X)
	}
}

func TestAPlatformGoingUpLiftsWhatStandsOnIt(test *testing.T) {
	world := NewWorld(earthGravity)
	carrier := lift(2, space.NewVector(0, 1, 0))
	world.Add(carrier)
	rider := standingOn(2)
	world.Add(rider)
	run(world, 1)
	top := carrier.Position.Y + 0.25
	if bottom := rider.Position.Y - 0.5; math.Abs(bottom-top) > 1e-6 || top < 2.9 {
		test.Fatalf("the rider's bottom is at %v and the platform's top at %v: it was not lifted", bottom, top)
	}
}

func TestAPlatformGoingDownKeepsWhatStandsOnIt(test *testing.T) {
	world := NewWorld(earthGravity)
	carrier := lift(2, space.NewVector(0, -1, 0))
	world.Add(carrier)
	rider := standingOn(2)
	world.Add(rider)
	run(world, 1)
	if bottom, top := rider.Position.Y-0.5, carrier.Position.Y+0.25; math.Abs(bottom-top) > 1e-6 || !rider.OnFloor() {
		test.Fatalf("the rider's bottom is at %v and the platform's top at %v: it did not ride down", bottom, top)
	}
}

func TestAPlatformPushesABodyInItsWay(test *testing.T) {
	world := NewWorld(space.Zero)
	carrier := &Body{Kind: Platform, Position: space.NewVector(0, 0, 0), Size: space.NewVector(1, 1, 0), Velocity: space.NewVector(2, 0, 0)}
	world.Add(carrier)
	inTheWay := &Body{Kind: Moving, Position: space.NewVector(2, 0, 0), Size: space.NewVector(1, 1, 0)}
	world.Add(inTheWay)
	run(world, 2)
	if gap := inTheWay.Position.X - carrier.Position.X; math.Abs(gap-1) > 1e-6 || carrier.Position.X < 3.9 {
		test.Fatalf("the platform is at %v and the body at %v: it was not pushed ahead", carrier.Position.X, inTheWay.Position.X)
	}
}

func TestAPlatformIgnoresGravityAndSolids(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(ground())
	carrier := lift(0.5, space.NewVector(0, -1, 0))
	world.Add(carrier)
	run(world, 1)
	if top := carrier.Position.Y + 0.25; math.Abs(top-(-0.5)) > 0.05 {
		test.Fatalf("the platform's top is at %v: it should go down through the ground at a metre a second", top)
	}
}

func TestABodyLandsOnAPlatform(test *testing.T) {
	world := NewWorld(earthGravity)
	world.Add(lift(1, space.Zero))
	falling := crate(4)
	world.Add(falling)
	run(world, 2)
	if bottom := falling.Position.Y - 0.5; math.Abs(bottom-1) > 1e-6 || !falling.OnFloor() {
		test.Fatalf("the crate's bottom is at %v: it should stand on the platform at 1", bottom)
	}
}
