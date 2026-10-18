package physics

import (
	"testing"

	"github.com/akrck02/go-engine/space"
)

// triggerAt is a trigger half a metre a side at a position.
func triggerAt(x float64) *Body {
	return &Body{Kind: Trigger, Position: space.NewVector(x, 0.5, 0), Size: space.NewVector(0.5, 0.5, 0)}
}

// walker moves right along the ground at a metre a second.
func walker() *Body {
	return &Body{Kind: Moving, Position: space.NewVector(0, 0.5, 0), Size: space.NewVector(1, 1, 0),
		Velocity: space.NewVector(1, 0, 0)}
}

func touchesOf(world *World, trigger *Body) int {
	count := 0
	for _, touch := range world.Begun() {
		if touch.Trigger == trigger {
			count++
		}
	}
	return count
}

func TestTouchingATriggerIsReportedOnce(test *testing.T) {
	world := NewWorld(space.Zero)
	trigger := triggerAt(1.5)
	mover := walker()
	world.Add(trigger)
	world.Add(mover)
	reported := 0
	for elapsed := 0.0; elapsed < 1.5; elapsed += frame {
		world.Step(frame)
		for _, touch := range world.Begun() {
			if touch.Trigger != trigger || touch.Other != mover {
				test.Fatalf("a touch of %v by %v", touch.Trigger, touch.Other)
			}
			reported++
		}
	}
	if reported != 1 {
		test.Fatalf("walking into the trigger and staying on it reported %d touches, not one", reported)
	}
}

func TestLeavingATriggerAndComingBackIsReportedAgain(test *testing.T) {
	world := NewWorld(space.Zero)
	zone := triggerAt(1)
	mover := walker()
	world.Add(zone)
	world.Add(mover)
	reported := 0
	// in the zone after a second, out of it a second and a half later, in it again after as long
	legs := []struct{ velocity, seconds float64 }{{1, 1}, {-1, 1.5}, {1, 1.5}}
	for _, leg := range legs {
		mover.Velocity.X = leg.velocity
		for elapsed := 0.0; elapsed < leg.seconds; elapsed += frame {
			world.Step(frame)
			reported += touchesOf(world, zone)
		}
	}
	if reported != 2 {
		test.Fatalf("going in twice reported %d touches", reported)
	}
}

func TestTriggersStopNothing(test *testing.T) {
	world := NewWorld(space.Zero)
	world.Add(triggerAt(1.5))
	mover := walker()
	world.Add(mover)
	run(world, 3)
	if mover.Position.X < 2.9 {
		test.Fatalf("the trigger stopped the body at %v", mover.Position.X)
	}
}

func TestTouchesAllocateNothingOnceWarm(test *testing.T) {
	world := NewWorld(space.Zero)
	for index := 0; index < 5; index++ {
		world.Add(triggerAt(float64(index)))
	}
	mover := walker()
	world.Add(mover)
	run(world, 1) // the lists grow once, then are reused
	allocations := testing.AllocsPerRun(100, func() {
		mover.Position.X = 0
		world.Step(frame)
		world.Begun()
	})
	if allocations != 0 {
		test.Fatalf("a step with triggers allocates %v times", allocations)
	}
}
