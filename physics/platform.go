package physics

import "github.com/akrck02/go-engine/space"

// ridingTolerance is how far apart, in metres, a body's floor side and what it stands on may be for
// it to still stand there: placed against each other they meet exactly, or nearly, after rounding.
const ridingTolerance = 1e-6

// movePlatform moves a platform by its velocity, one axis after the other, carrying what stands on
// it (and what stands on that), pushing out of its way what it runs into, and stopping against
// what it could not move: it never leaves a body inside it.
func (world *World) movePlatform(platform *Body, seconds float64) {
	delta := platform.Velocity.Scale(seconds)
	if delta == space.Zero {
		return
	}
	world.findRiders(platform)
	world.carryAlong(platform, horizontal, delta.X)
	world.carryAlong(platform, vertical, delta.Y)
}

// carryAlong moves a platform a distance along an axis with its riders, the ones furthest ahead
// first, so that one stopped stops those behind it.
func (world *World) carryAlong(platform *Body, axis axis, distance float64) {
	if distance == 0 {
		return
	}
	axis.set(&platform.Position, axis.of(platform.Position)+distance)
	world.passing = append(world.passing, platform)
	sortAhead(world.riders, axis, distance)
	for _, rider := range world.riders {
		world.moveAlong(rider, axis, distance)
	}
	world.pushOutOf(platform, axis, distance)
	world.passing = world.passing[:len(world.passing)-1]
	world.stopPlatform(platform, axis, distance)
}

// sortAhead orders bodies by how far ahead they are along an axis, the way a distance goes: the
// first is the furthest. An insertion sort: there are a few, and it allocates nothing.
func sortAhead(bodies []*Body, axis axis, distance float64) {
	ahead := func(body *Body) float64 { return axis.of(body.Position) * distance }
	for index := 1; index < len(bodies); index++ {
		for back := index; back > 0 && ahead(bodies[back]) > ahead(bodies[back-1]); back-- {
			bodies[back], bodies[back-1] = bodies[back-1], bodies[back]
		}
	}
}

// findRiders finds the falling bodies standing on a platform, and those standing on them, into
// the world's list of riders.
func (world *World) findRiders(platform *Body) {
	world.riders = world.riders[:0]
	for grew := true; grew; {
		grew = false
		for _, body := range world.bodies {
			if !body.Kind.falls() || world.isRider(body) {
				continue
			}
			if body.standsOn(platform) || world.standsOnRider(body) {
				world.riders = append(world.riders, body)
				grew = true
			}
		}
	}
}

// isRider tells whether a body was found standing on the platform moving.
func (world *World) isRider(body *Body) bool {
	for _, rider := range world.riders {
		if rider == body {
			return true
		}
	}
	return false
}

// standsOnRider tells whether a body stands on a body standing on the platform moving.
func (world *World) standsOnRider(body *Body) bool {
	for _, rider := range world.riders {
		if body.standsOn(rider) {
			return true
		}
	}
	return false
}

// standsOn tells whether a body stands on another: its floor side (the one its gravity pulls it to)
// against the other's facing side, and across it.
func (body *Body) standsOn(base *Body) bool {
	axis, way := body.floorAxis(), body.floorWay()
	feet := axis.of(body.Position) + way*axis.of(body.Size)/2
	face := axis.of(base.Position) - way*axis.of(base.Size)/2
	return feet-face <= ridingTolerance && face-feet <= ridingTolerance && overlapAlong(body, base, axis.across())
}

// pushOutOf moves the falling bodies a platform ran into, along the way it went, out of its way.
func (world *World) pushOutOf(platform *Body, axis axis, distance float64) {
	for _, body := range world.bodies {
		if body.Kind.falls() && body.overlaps(platform) {
			world.moveAlong(body, axis, depthAlong(body, platform, axis, distance))
		}
	}
}

// stopPlatform moves a platform back against the falling bodies it could not move out of its way.
func (world *World) stopPlatform(platform *Body, axis axis, distance float64) {
	for _, body := range world.bodies {
		if !body.Kind.falls() || !body.overlaps(platform) {
			continue
		}
		reach := (axis.of(body.Size) + axis.of(platform.Size)) / 2
		if distance > 0 {
			axis.set(&platform.Position, axis.of(body.Position)-reach)
		} else {
			axis.set(&platform.Position, axis.of(body.Position)+reach)
		}
	}
}

// depthAlong is how far a body must move along an axis, the way a platform goes, to leave it.
func depthAlong(body, platform *Body, axis axis, way float64) float64 {
	reach := (axis.of(body.Size) + axis.of(platform.Size)) / 2
	if way > 0 {
		return axis.of(platform.Position) + reach - axis.of(body.Position)
	}
	return axis.of(platform.Position) - reach - axis.of(body.Position)
}
