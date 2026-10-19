// Package physics moves boxes in 2D: bodies that fall, collide with solid ground and walls, and
// slide along them, as a game seen from the side needs. It knows nothing of scenes or nodes; the engine keeps
// a World in step with them.
//
// Bodies are boxes lined up with the axes, sized and placed in metres, with X to the right and Y
// up; their Z is not used. A moving body is moved one axis at a time, in steps no longer than half
// its smallest side, so even a fast body meets a thin wall instead of jumping over it.
//
//	world := physics.NewWorld(space.NewVector(0, -9.81, 0))
//	world.Add(&physics.Body{Kind: physics.Solid, Position: groundCentre, Size: groundSize})
//	world.Add(mover) // a Moving body
//	world.Step(seconds)
//	for _, touch := range world.Begun() { … } // triggers touched in this step
package physics

import (
	"math"

	"github.com/akrck02/go-engine/space"
)

// Kind is what a body does in the world.
type Kind int

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

// Moves tells whether bodies of the kind move by the physics: their place is the world's to set.
func (kind Kind) Moves() bool { return kind == Moving || kind == Pushable || kind == Platform }

// falls tells whether bodies of the kind fall with gravity and are stopped by obstacles.
func (kind Kind) falls() bool { return kind == Moving || kind == Pushable }

// touchTolerance is how far, in metres, two boxes may overlap and still count as only touching:
// a body placed against a wall must not be found inside it because of rounding.
const touchTolerance = 1e-9

// longestStepShare is the longest step a moving body takes, as a share of its smallest side.
const longestStepShare = 0.5

// pointStepMetres is the longest step of a body with no size, which has no side to measure by.
const pointStepMetres = 0.05

// Body is a box in the world.
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

	onFloor bool
	pull    space.Vector // the gravity that pulled it in the last step
}

// Pull is the gravity that pulled the body in the last step: the world's, or the zone's it was
// in. Its floor is the side this pulls it to.
func (body *Body) Pull() space.Vector { return body.pull }

// OnFloor tells whether the body stood on something solid in the last step.
func (body *Body) OnFloor() bool { return body.onFloor }

// overlaps tells whether two bodies' boxes overlap by more than the touch tolerance.
func (body *Body) overlaps(other *Body) bool {
	return overlapAlong(body, other, horizontal) && overlapAlong(body, other, vertical)
}

func overlapAlong(body, other *Body, axis axis) bool {
	distance := math.Abs(axis.of(body.Position) - axis.of(other.Position))
	reach := (axis.of(body.Size) + axis.of(other.Size)) / 2
	return distance < reach-touchTolerance
}

// World holds the bodies and moves them.
type World struct {
	gravity space.Vector // metres a second, each second
	bodies  []*Body

	touching    []Touch // the moving bodies overlapping a trigger now
	wasTouching []Touch // the same, a step ago
	begun       []Touch // the touches that began in the last step
	riders      []*Body // the bodies standing on the platform moving, reused
	zones       []*Body // the triggers with a gravity of their own, found again each step
	// passing are the bodies moving others now (pushing them, carrying them): no obstacle to the
	// bodies they move, and not pushed back by them, so a ring of bodies cannot push forever.
	passing []*Body
}

// longestPush is how many bodies in a row may push one another: a longer row is as heavy as a wall.
const longestPush = 8

// NewWorld makes a world with a gravity, in metres a second each second (Earth's is 9.81 down).
func NewWorld(gravity space.Vector) *World { return &World{gravity: gravity} }

// Gravity is the world's gravity, in metres a second each second.
func (world *World) Gravity() space.Vector { return world.gravity }

// SetGravity changes the world's gravity: a floatier jump, a world upside down.
func (world *World) SetGravity(gravity space.Vector) { world.gravity = gravity }

// Add puts a body in the world.
func (world *World) Add(body *Body) { world.bodies = append(world.bodies, body) }

// Remove takes a body out of the world.
func (world *World) Remove(body *Body) {
	kept := world.bodies[:0]
	for _, each := range world.bodies {
		if each != body {
			kept = append(kept, each)
		}
	}
	clear(world.bodies[len(kept):])
	world.bodies = kept
}

// Step moves the world on by some seconds: gravity pulls the moving bodies, each moves by its
// velocity until a solid stops it, and the touches of triggers that began are found (Begun).
func (world *World) Step(seconds float64) {
	world.findZones()
	for _, body := range world.bodies {
		if body.Kind.falls() {
			body.pull = world.gravityAt(body) // which side is its floor, for the platforms
		}
	}
	for _, body := range world.bodies {
		if body.Kind == Platform {
			world.movePlatform(body, seconds)
		}
	}
	for _, body := range world.bodies {
		if body.Kind.falls() {
			world.move(body, seconds)
		}
	}
	world.findTouches()
}

// move pulls a body by gravity and moves it, across first and then up or down, so that blocked on
// one axis it still slides along the other.
func (world *World) move(body *Body, seconds float64) {
	body.onFloor = false
	body.pull = world.gravityAt(body)
	body.Velocity = body.Velocity.Add(body.pull.Scale(body.GravityScale * seconds))
	world.moveAlong(body, horizontal, body.Velocity.X*seconds)
	world.moveAlong(body, vertical, body.Velocity.Y*seconds)
}

// moveAlong moves a body a distance along one axis, in short steps, and stops it against the
// first obstacle it meets; a pushable one met across its gravity is pushed first, and stops it
// only if it could not go.
func (world *World) moveAlong(body *Body, axis axis, distance float64) {
	if distance == 0 {
		return
	}
	steps := stepsFor(body, distance)
	stride := distance / float64(steps)
	for step := 0; step < steps; step++ {
		axis.set(&body.Position, axis.of(body.Position)+stride)
		blocker, found := world.obstacleTo(body)
		if found && world.canPush(body, blocker, axis) {
			world.passing = append(world.passing, body)
			world.moveAlong(blocker, axis, stride)
			world.passing = world.passing[:len(world.passing)-1]
			blocker, found = world.obstacleTo(body)
		}
		if found {
			world.stopAgainst(body, blocker, axis, stride)
			return
		}
	}
}

// canPush tells whether a body moving along an axis pushes another it ran into: a pushable one,
// across the mover's gravity (along it, the mover lands on it), not already moving it, and not
// past the longest row.
func (world *World) canPush(mover, other *Body, axis axis) bool {
	return other.Kind == Pushable && axis != mover.floorAxis() && len(world.passing) < longestPush &&
		!world.isPassing(other)
}

// isPassing tells whether a body is moving others now.
func (world *World) isPassing(body *Body) bool {
	for _, each := range world.passing {
		if each == body {
			return true
		}
	}
	return false
}

// stepsFor is how many steps a body takes to go a distance: none longer than half its smallest
// side, or than a point's step for a body with no size.
func stepsFor(body *Body, distance float64) int {
	longest := math.Min(body.Size.X, body.Size.Y) * longestStepShare
	if longest <= 0 {
		longest = pointStepMetres
	}
	return int(math.Ceil(math.Abs(distance) / longest))
}

// obstacleTo finds a body that a body overlaps and may not go through (blocks), but those moving
// it.
func (world *World) obstacleTo(body *Body) (*Body, bool) {
	for _, other := range world.bodies {
		if other != body && blocks(body, other) && body.overlaps(other) && !world.isPassing(other) {
			return other, true
		}
	}
	return nil, false
}

// blocks tells whether a body stops another moving into it: solids and platforms stop all;
// pushable bodies stop moving and pushable ones; moving bodies stop pushable ones (a box lands on
// whoever is under it), but not each other, as before pushable bodies were.
func blocks(mover, other *Body) bool {
	switch other.Kind {
	case Solid, Pushable, Platform:
		return true
	case Moving:
		return mover.Kind == Pushable
	}
	return false
}

// stopAgainst puts a body against the face of the solid it moved into, and stops it along that
// axis; stopped going down, it stands on the floor.
func (world *World) stopAgainst(body, solid *Body, axis axis, stride float64) {
	gap := (axis.of(body.Size) + axis.of(solid.Size)) / 2
	if stride > 0 {
		axis.set(&body.Position, axis.of(solid.Position)-gap)
	} else {
		axis.set(&body.Position, axis.of(solid.Position)+gap)
	}
	axis.set(&body.Velocity, 0)
	if body.floorIsAlong(axis, stride) {
		body.onFloor = true
	}
}

// findZones finds the triggers that have a gravity of their own, once a step: most worlds have
// none, and then a falling body looks at none.
func (world *World) findZones() {
	world.zones = world.zones[:0]
	for _, body := range world.bodies {
		if body.Kind == Trigger && body.ZonePull != nil {
			world.zones = append(world.zones, body)
		}
	}
}

// gravityAt is the gravity that pulls a body: the gravity of a zone it is in, or the world's.
// Where zones overlap, the one added to the world first pulls.
func (world *World) gravityAt(body *Body) space.Vector {
	if body.IgnoresZones {
		return world.gravity
	}
	for _, zone := range world.zones {
		if zone.overlaps(body) {
			return *zone.ZonePull
		}
	}
	return world.gravity
}

// floorIsAlong tells whether a body stopped going a way (a stride along an axis) stopped on its
// floor: the side its gravity pulls it to.
func (body *Body) floorIsAlong(axis axis, stride float64) bool {
	return axis == body.floorAxis() && (stride < 0) == (body.floorWay() < 0)
}

// floorAxis is the axis the body's gravity pulls it along most: up and down, unless it pulls more
// across. With nothing pulling it, it is up and down.
func (body *Body) floorAxis() axis {
	if math.Abs(body.pull.X) > math.Abs(body.pull.Y) {
		return horizontal
	}
	return vertical
}

// floorWay is which way along its floor axis the body's floor is: -1 or 1. With nothing pulling
// it, the floor is below.
func (body *Body) floorWay() float64 {
	if body.floorAxis().of(body.pull) > 0 {
		return 1
	}
	return -1
}

// axis is one of the two directions bodies move in.
type axis int

const (
	horizontal axis = iota
	vertical
)

// of is a vector's coordinate along the axis.
func (axis axis) of(vector space.Vector) float64 {
	if axis == horizontal {
		return vector.X
	}
	return vector.Y
}

// set changes a vector's coordinate along the axis.
func (axis axis) set(vector *space.Vector, value float64) {
	if axis == horizontal {
		vector.X = value
		return
	}
	vector.Y = value
}

// across is the other axis.
func (axis axis) across() axis {
	if axis == horizontal {
		return vertical
	}
	return horizontal
}
