package anim

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

func TestAValueGetsThereOnTimeAndThenIsCalled(t *testing.T) {
	var player Player
	value, ended := 2.0, false
	player.Float(&value, 10, 1, EaseInOut).Then(func() { ended = true })
	run(&player, 0.5)
	if math.Abs(value-6) > 0.01 || ended {
		t.Fatalf("half way the value is %v", value)
	}
	run(&player, 0.6)
	if value != 10 || !ended || player.Running() != 0 {
		t.Fatalf("at the end the value is %v, ended %v", value, ended)
	}
}

func TestATweenStartsFromWhereTheValueIsWhenItStarts(t *testing.T) {
	var player Player
	position := space.V(0, 0, 0)
	player.Vec3(&position, space.V(0, 4, 0), 1, Linear).Delay(1)
	position = space.V(0, 2, 0) // moved meanwhile
	run(&player, 1.5)
	if !position.Near(space.V(0, 3, 0), 0.02) {
		t.Fatalf("the tween did not start from where the value was: %v", position)
	}
}

func TestPingPongGoesThereAndBack(t *testing.T) {
	var player Player
	value := 0.0
	player.Float(&value, 1, 1, Linear).PingPong()
	run(&player, 1.5)
	if math.Abs(value-0.5) > 0.02 {
		t.Fatalf("on the way back the value is %v", value)
	}
	run(&player, 0.5)
	if value > 0.02 || player.Running() != 1 {
		t.Fatalf("back at the start: %v, %d running", value, player.Running())
	}
}

func TestTweensAddedFromThenRun(t *testing.T) {
	var player Player
	value := 0.0
	player.Wait(0.5).Then(func() { player.Float(&value, 1, 0.5, Linear) })
	run(&player, 1.1)
	if value != 1 {
		t.Fatalf("the chained tween left the value at %v", value)
	}
}

func TestEasingsStartAtZeroAndEndAtOne(t *testing.T) {
	for name, easing := range map[string]Easing{"linear": Linear, "in": EaseIn, "out": EaseOut, "in-out": EaseInOut, "back": Back, "bounce": Bounce} {
		if math.Abs(easing(0)) > 1e-9 || math.Abs(easing(1)-1) > 1e-9 {
			t.Errorf("%s: %v at 0, %v at 1", name, easing(0), easing(1))
		}
	}
}

func TestARotationTurnsTheShortWay(t *testing.T) {
	var player Player
	rotation := space.AxisAngle(space.Up, 170)
	player.Rotation(&rotation, space.AxisAngle(space.Up, -170), 1, Linear)
	run(&player, 0.5)
	if yaw := rotation.Euler().Y; math.Abs(math.Abs(yaw)-180) > 0.5 {
		t.Fatalf("half way the yaw is %v, not through 180", yaw)
	}
}
