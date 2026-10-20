# physics

```go
import "github.com/akrck02/go-engine/physics"
```

Package physics moves boxes in 2D: bodies that fall, collide with solid ground
and walls, and slide along them, as a game seen from the side needs. It knows
nothing of scenes or nodes; the engine keeps a World in step with them.

Bodies are boxes lined up with the axes, in metres, with X to the right and Y
up; their Z is not used. A moving body moves one axis at a time, in steps of at
most half its smallest side. So even a fast body meets a thin wall instead of
jumping over it.

### Body

Body is a box in the world.

```go
type Body struct {
	Kind         Kind
	Position     space.Vector // the box's centre, in metres
	Size         space.Vector // the box's width (X) and height (Y), in metres
	Velocity     space.Vector // metres a second
	GravityScale float64      // how much gravity pulls it: 1 fully, 0 not at all
	// ZonePull, on a trigger, is the gravity of the area it covers: the falling bodies inside it are
	// pulled this way instead of the world's (a magnetic field, a room upside down). Nil leaves
	// the world's.
	ZonePull *space.Vector
	// IgnoresZones keeps the world's gravity for the body, inside a zone too: a character that
	// walks through a field meant for the things it moves.
	IgnoresZones bool

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (body *Body) OnFloor() bool` | OnFloor tells whether the body stood on something solid in the last step. |
| `func (body *Body) Pull() space.Vector` | Pull is the gravity that pulled the body in the last step: the world's, or the zone's it was in. |

### Kind

Kind is what a body does in the world.

```go
type Kind int
```

```go
const (
	// Solid bodies never move: ground, walls, platforms. Moving bodies stop against them.
	Solid Kind = iota
	// Moving bodies fall with gravity, move with their velocity and are stopped by solids.
	Moving
	// Trigger bodies don't move nor stop anything; they notice the moving and pushable bodies that
	// touch them (World.Begun).
	Trigger
	// Pushable bodies fall and move as moving ones do, and are obstacles to them: a moving body
	// stands on one, and walking into one pushes it, until a solid stops both (a box, a cart).
	Pushable
	// Platform bodies move at their velocity, with no gravity and through solids; they carry the
	// bodies standing on them and push those in their way (a lift, a moving floor).
	Platform
)
```

| Function or method | |
|---|---|
| `func (kind Kind) Moves() bool` | Moves tells whether bodies of the kind move by the physics: their place is the world's to set. |

### Touch

Touch is a moving or pushable body touching a trigger: a door reached,
a line crossed, a box on a button.

```go
type Touch struct {
	Trigger *Body
	Other   *Body // the moving or pushable body that touches it
}
```

### World

World holds the bodies and moves them.

```go
type World struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func NewWorld(gravity space.Vector) *World` | NewWorld makes a world with a gravity, in metres a second each second (Earth's is 9.81 down). |
| `func (world *World) Add(body *Body)` | Add puts a body in the world. |
| `func (world *World) Begun() []Touch` | Begun are the touches that began in the last step: each is reported once, when a body first overlaps a trigger, and again only after it has left and comes back. |
| `func (world *World) Gravity() space.Vector` | Gravity is the world's gravity, in metres a second each second. |
| `func (world *World) Remove(body *Body)` | Remove takes a body out of the world. |
| `func (world *World) SetGravity(gravity space.Vector)` | SetGravity changes the world's gravity: a floatier jump, a world upside down. |
| `func (world *World) Step(seconds float64)` | Step moves the world on by some seconds: gravity pulls the moving bodies, each moves by its velocity until a solid stops it, and the touches of triggers that began are found (Begun). |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
