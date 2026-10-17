// Package anim moves values over time: a position gliding to a point, a door turning open, a
// colour fading. A Player runs tweens; each frame the game calls Update with the seconds passed,
// and every tween writes its value where it was told to.
//
//	player.Vec3(&node.Transform.Position, space.V(0, 2, 0), 0.5, anim.EaseOut).
//		Then(func() { door.Open = true })
package anim

import (
	"math"

	"github.com/akrck02/go-engine/space"
)

// Easing shapes a tween's pace: it takes the share of the time gone (0 to 1) and gives the share
// of the way gone.
type Easing func(share float64) float64

// Easings.
var (
	Linear    Easing = func(share float64) float64 { return share }
	EaseIn    Easing = func(share float64) float64 { return share * share * share }
	EaseOut   Easing = func(share float64) float64 { rest := 1 - share; return 1 - rest*rest*rest }
	EaseInOut Easing = func(share float64) float64 { return (1 - math.Cos(share*math.Pi)) / 2 }
	// Back overshoots the end a little and settles, as something thrown into place.
	Back Easing = func(share float64) float64 {
		const overshoot = 1.70158
		rest := share - 1
		return 1 + (overshoot+1)*rest*rest*rest + overshoot*rest*rest
	}
	// Bounce lands like a ball.
	Bounce Easing = bounce
)

func bounce(share float64) float64 {
	const strength, width = 7.5625, 2.75
	switch {
	case share < 1/width:
		return strength * share * share
	case share < 2/width:
		share -= 1.5 / width
		return strength*share*share + 0.75
	case share < 2.5/width:
		share -= 2.25 / width
		return strength*share*share + 0.9375
	}
	share -= 2.625 / width
	return strength*share*share + 0.984375
}

// Tween moves one value from where it is when the tween starts to a target.
type Tween struct {
	duration, delay, elapsed float64
	easing                   Easing
	start                    func()              // reads the value it starts from
	apply                    func(eased float64) // writes the value a share of the way
	started, done, stopped   bool
	loops                    int // more times to play after this one; -1 forever
	pingPong, backwards      bool
	then                     []func()
}

// Delay waits some seconds before the tween starts.
func (tween *Tween) Delay(seconds float64) *Tween { tween.delay = seconds; return tween }

// Loop plays the tween again some more times (-1 forever), from the same start.
func (tween *Tween) Loop(times int) *Tween { tween.loops = times; return tween }

// PingPong plays it there and back, forever: a light breathing, a platform going to and fro.
func (tween *Tween) PingPong() *Tween { tween.pingPong, tween.loops = true, -1; return tween }

// Then calls a function when the tween ends.
func (tween *Tween) Then(done func()) *Tween { tween.then = append(tween.then, done); return tween }

// Stop ends the tween where it is, without calling what Then gave it.
func (tween *Tween) Stop() { tween.stopped = true }

// Done tells whether the tween has ended.
func (tween *Tween) Done() bool { return tween.done || tween.stopped }

// advance moves the tween on; it tells whether it is still running.
func (tween *Tween) advance(seconds float64) bool {
	if tween.Done() {
		return false
	}
	if tween.delay > 0 {
		tween.delay -= seconds
		if tween.delay > 0 {
			return true
		}
		seconds = -tween.delay
	}
	if !tween.started {
		tween.started = true
		tween.start()
	}
	tween.elapsed += seconds
	share := 1.0
	if tween.duration > 0 {
		share = min(tween.elapsed/tween.duration, 1)
	}
	shown := share
	if tween.backwards {
		shown = 1 - share
	}
	tween.apply(tween.easing(shown))
	if share < 1 {
		return true
	}
	if tween.loops != 0 {
		if tween.loops > 0 {
			tween.loops--
		}
		tween.elapsed -= tween.duration
		if tween.pingPong {
			tween.backwards = !tween.backwards
		}
		return true
	}
	tween.done = true
	for _, done := range tween.then {
		done()
	}
	return false
}

// Player runs tweens. The zero value is ready.
type Player struct {
	tweens   []*Tween
	updating bool
	added    []*Tween // tweens added while Update walks the list (from a Then)
}

// Update moves every tween on by some seconds, and forgets the ones that ended.
func (player *Player) Update(seconds float64) {
	player.updating = true
	running := player.tweens[:0]
	for _, tween := range player.tweens {
		if tween.advance(seconds) {
			running = append(running, tween)
		}
	}
	player.updating = false
	clear(player.tweens[len(running):])
	player.tweens = append(running, player.added...)
	clear(player.added)
	player.added = player.added[:0]
}

// Running is how many tweens are playing or waiting.
func (player *Player) Running() int { return len(player.tweens) }

// StopAll stops every tween.
func (player *Player) StopAll() {
	for _, tween := range player.tweens {
		tween.Stop()
	}
	player.tweens = player.tweens[:0]
}

func (player *Player) add(tween *Tween, easing Easing, seconds float64) *Tween {
	if easing == nil {
		easing = Linear
	}
	tween.easing, tween.duration = easing, seconds
	if player.updating {
		player.added = append(player.added, tween)
		return tween
	}
	player.tweens = append(player.tweens, tween)
	return tween
}

// Float moves a number to a target in some seconds.
func (player *Player) Float(value *float64, target, seconds float64, easing Easing) *Tween {
	var from float64
	tween := &Tween{
		start: func() { from = *value },
		apply: func(eased float64) { *value = from + (target-from)*eased },
	}
	return player.add(tween, easing, seconds)
}

// Vec3 moves a point (a position, a scale) to a target in some seconds.
func (player *Player) Vec3(value *space.Vec3, target space.Vec3, seconds float64, easing Easing) *Tween {
	var from space.Vec3
	tween := &Tween{
		start: func() { from = *value },
		apply: func(eased float64) { *value = from.Lerp(target, eased) },
	}
	return player.add(tween, easing, seconds)
}

// Rotation turns a rotation to a target in some seconds, the shorter way round.
func (player *Player) Rotation(value *space.Quat, target space.Quat, seconds float64, easing Easing) *Tween {
	var from space.Quat
	tween := &Tween{
		start: func() { from = *value },
		apply: func(eased float64) { *value = from.Slerp(target, eased) },
	}
	return player.add(tween, easing, seconds)
}

// Wait does nothing for some seconds, then calls what Then gives it: a timer.
func (player *Player) Wait(seconds float64) *Tween {
	return player.add(&Tween{start: func() {}, apply: func(float64) {}}, Linear, seconds)
}
