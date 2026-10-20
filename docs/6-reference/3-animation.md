# animation

```go
import "github.com/akrck02/go-engine/animation"
```

Package anim moves values over time: a position gliding to a point, a door
turning open, a colour fading. A Player runs tweens; each frame the game calls
Update with the seconds passed, and every tween writes its value where it was
told to.

### Easing

Easing shapes a tween's pace: it takes the share of the time gone (0 to 1)
and gives the share of the way gone.

```go
type Easing func(share float64) float64
```

```go
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
```

Easings.

### Player

Player runs tweens. The zero value is ready.

```go
type Player struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (player *Player) Float(value *float64, target, seconds float64, easing Easing) *Tween` | Float moves a number to a target in some seconds. |
| `func (player *Player) Rotation(value *space.Rotation, target space.Rotation, seconds float64, easing Easing) *Tween` | Rotation turns a rotation to a target in some seconds, the shorter way round. |
| `func (player *Player) Running() int` | Running is how many tweens are playing or waiting. |
| `func (player *Player) StopAll()` | StopAll stops every tween. |
| `func (player *Player) Update(seconds float64)` | Update moves every tween on by some seconds, and forgets the ones that ended. |
| `func (player *Player) Vector(value *space.Vector, target space.Vector, seconds float64, easing Easing) *Tween` | Vector moves a point (a position, a scale) to a target in some seconds. |
| `func (player *Player) Wait(seconds float64) *Tween` | Wait does nothing for some seconds, then calls what Then gives it: a timer. |

### Tween

Tween moves one value from where it is when the tween starts to a target.

```go
type Tween struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (tween *Tween) Delay(seconds float64) *Tween` | Delay waits some seconds before the tween starts. |
| `func (tween *Tween) Done() bool` | Done tells whether the tween has ended. |
| `func (tween *Tween) Loop(times int) *Tween` | Loop plays the tween again some more times (-1 forever), from the same start. |
| `func (tween *Tween) PingPong() *Tween` | PingPong plays it there and back, forever: a light breathing, a platform going to and fro. |
| `func (tween *Tween) Stop()` | Stop ends the tween where it is, without calling what Then gave it. |
| `func (tween *Tween) Then(done func()) *Tween` | Then calls a function when the tween ends. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
