package animation

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/space"
)

func run(player *Player, seconds float64) {
	for step := 0.0; step < seconds-1e-9; step += 0.01 {
		player.Update(0.01)
	}
}

func TestAValueGetsThereOnTimeAndThenIsCalled(test *testing.T) {
	var player Player
	value, ended := 2.0, false
	player.Float(&value, 10, 1, EaseInOut).Then(func() { ended = true })
	run(&player, 0.5)
	if math.Abs(value-6) > 0.01 || ended {
		test.Fatalf("half way the value is %v", value)
	}
	run(&player, 0.6)
	if value != 10 || !ended || player.Running() != 0 {
		test.Fatalf("at the end the value is %v, ended %v", value, ended)
	}
}

func TestATweenStartsFromWhereTheValueIsWhenItStarts(test *testing.T) {
	var player Player
	position := space.NewVector(0, 0, 0)
	player.Vector(&position, space.NewVector(0, 4, 0), 1, Linear).Delay(1)
	position = space.NewVector(0, 2, 0) // moved meanwhile
	run(&player, 1.5)
	if !position.Near(space.NewVector(0, 3, 0), 0.02) {
		test.Fatalf("the tween did not start from where the value was: %v", position)
	}
}

func TestPingPongGoesThereAndBack(test *testing.T) {
	var player Player
	value := 0.0
	player.Float(&value, 1, 1, Linear).PingPong()
	run(&player, 1.5)
	if math.Abs(value-0.5) > 0.02 {
		test.Fatalf("on the way back the value is %v", value)
	}
	run(&player, 0.5)
	if value > 0.02 || player.Running() != 1 {
		test.Fatalf("back at the start: %v, %d running", value, player.Running())
	}
}

func TestTweensAddedFromThenRun(test *testing.T) {
	var player Player
	value := 0.0
	player.Wait(0.5).Then(func() { player.Float(&value, 1, 0.5, Linear) })
	run(&player, 1.1)
	if value != 1 {
		test.Fatalf("the chained tween left the value at %v", value)
	}
}

func TestEasingsStartAtZeroAndEndAtOne(test *testing.T) {
	for name, easing := range map[string]Easing{"linear": Linear, "in": EaseIn, "out": EaseOut, "in-out": EaseInOut, "back": Back, "bounce": Bounce} {
		if math.Abs(easing(0)) > 1e-9 || math.Abs(easing(1)-1) > 1e-9 {
			test.Errorf("%s: %v at 0, %v at 1", name, easing(0), easing(1))
		}
	}
}

func TestARotationTurnsTheShortWay(test *testing.T) {
	var player Player
	rotation := space.RotationAround(space.Up, 170)
	player.Rotation(&rotation, space.RotationAround(space.Up, -170), 1, Linear)
	run(&player, 0.5)
	if yaw := rotation.Angles().Y; math.Abs(math.Abs(yaw)-180) > 0.5 {
		test.Fatalf("half way the yaw is %v, not through 180", yaw)
	}
}
