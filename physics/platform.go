package physics

import "github.com/akrck02/go-engine/space"

// ridingTolerance is how far apart, in metres, a body's bottom and a platform's top may be for the
// body to still stand on it: placed against each other they meet exactly, or nearly, after rounding.
const ridingTolerance = 1e-6

// movePlatform moves a platform by its velocity, carrying what stands on it and pushing out of its
// way what it runs into.
func (world *World) movePlatform(platform *Body, seconds float64) {
	delta := platform.Velocity.Scale(seconds)
	if delta == space.Zero {
		return
	}
	riders := world.ridersOf(platform)
	platform.Position = platform.Position.Add(delta)
	for _, rider := range riders {
		world.moveAlong(rider, horizontal, delta.X, platform)
		world.moveAlong(rider, vertical, delta.Y, platform)
	}
	world.pushOutOf(platform, delta)
}

// ridersOf are the falling bodies standing on a platform: their bottom on its top, across it.
// The list is the world's own, reused from step to step.
func (world *World) ridersOf(platform *Body) []*Body {
	world.riders = world.riders[:0]
	top := platform.Position.Y + platform.Size.Y/2
	for _, body := range world.bodies {
		if !body.Kind.falls() || !overlapAlong(body, platform, horizontal) {
			continue
		}
		if bottom := body.Position.Y - body.Size.Y/2; bottom-top <= ridingTolerance && top-bottom <= ridingTolerance {
			world.riders = append(world.riders, body)
		}
	}
	return world.riders
}

// pushOutOf moves the falling bodies a platform ran into out of its way, along the way it went.
func (world *World) pushOutOf(platform *Body, delta space.Vector) {
	for _, body := range world.bodies {
		if !body.Kind.falls() || !body.overlaps(platform) {
			continue
		}
		if delta.X != 0 {
			world.moveAlong(body, horizontal, depthAlong(body, platform, horizontal, delta.X), platform)
		}
		if delta.Y != 0 && body.overlaps(platform) {
			world.moveAlong(body, vertical, depthAlong(body, platform, vertical, delta.Y), platform)
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
