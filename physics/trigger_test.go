package physics

import (
	"testing"

	"github.com/akrck02/go-engine/space"
)

// coinAt is a trigger half a metre a side at a position.
func coinAt(x float64) *Body {
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
	coin := coinAt(1.5)
	player := walker()
	world.Add(coin)
	world.Add(player)
	reported := 0
	for elapsed := 0.0; elapsed < 1.5; elapsed += frame {
		world.Step(frame)
		for _, touch := range world.Begun() {
			if touch.Trigger != coin || touch.Other != player {
				test.Fatalf("a touch of %v by %v", touch.Trigger, touch.Other)
			}
			reported++
		}
	}
	if reported != 1 {
		test.Fatalf("walking into the coin and staying on it reported %d touches, not one", reported)
	}
}

func TestLeavingATriggerAndComingBackIsReportedAgain(test *testing.T) {
	world := NewWorld(space.Zero)
	spikes := coinAt(1)
	player := walker()
	world.Add(spikes)
	world.Add(player)
	reported := 0
	// on the spikes after a second, off them a second and a half later, on them again after as long
	legs := []struct{ velocity, seconds float64 }{{1, 1}, {-1, 1.5}, {1, 1.5}}
	for _, leg := range legs {
		player.Velocity.X = leg.velocity
		for elapsed := 0.0; elapsed < leg.seconds; elapsed += frame {
			world.Step(frame)
			reported += touchesOf(world, spikes)
		}
	}
	if reported != 2 {
		test.Fatalf("going in twice reported %d touches", reported)
	}
}

func TestTriggersStopNothing(test *testing.T) {
	world := NewWorld(space.Zero)
	world.Add(coinAt(1.5))
	player := walker()
	world.Add(player)
	run(world, 3)
	if player.Position.X < 2.9 {
		test.Fatalf("the trigger stopped the body at %v", player.Position.X)
	}
}

func TestTouchesAllocateNothingOnceWarm(test *testing.T) {
	world := NewWorld(space.Zero)
	for index := 0; index < 5; index++ {
		world.Add(coinAt(float64(index)))
	}
	player := walker()
	world.Add(player)
	run(world, 1) // the lists grow once, then are reused
	allocations := testing.AllocsPerRun(100, func() {
		player.Position.X = 0
		world.Step(frame)
		world.Begun()
	})
	if allocations != 0 {
		test.Fatalf("a step with triggers allocates %v times", allocations)
	}
}
