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
	// Trigger bodies don't move nor stop anything; they notice the moving bodies that touch them
	// (World.Begun).
	Trigger
)

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

	onFloor bool
}

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
}

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
	for _, body := range world.bodies {
		if body.Kind == Moving {
			world.move(body, seconds)
		}
	}
	world.findTouches()
}

// move pulls a body by gravity and moves it, across first and then up or down, so that blocked on
// one axis it still slides along the other.
func (world *World) move(body *Body, seconds float64) {
	body.onFloor = false
	body.Velocity = body.Velocity.Add(world.gravity.Scale(body.GravityScale * seconds))
	world.moveAlong(body, horizontal, body.Velocity.X*seconds)
	world.moveAlong(body, vertical, body.Velocity.Y*seconds)
}

// moveAlong moves a body a distance along one axis, in short steps, and stops it against the
// first solid it meets.
func (world *World) moveAlong(body *Body, axis axis, distance float64) {
	if distance == 0 {
		return
	}
	steps := stepsFor(body, distance)
	stride := distance / float64(steps)
	for step := 0; step < steps; step++ {
		axis.set(&body.Position, axis.of(body.Position)+stride)
		if blocker, found := world.solidOverlapping(body); found {
			world.stopAgainst(body, blocker, axis, stride)
			return
		}
	}
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

// solidOverlapping finds a solid body that a body overlaps.
func (world *World) solidOverlapping(body *Body) (*Body, bool) {
	for _, other := range world.bodies {
		if other.Kind == Solid && other != body && body.overlaps(other) {
			return other, true
		}
	}
	return nil, false
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
	if axis == vertical && stride < 0 {
		body.onFloor = true
	}
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
