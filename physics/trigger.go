package physics

// Touch is a moving body touching a trigger: a coin taken, spikes stepped on.
type Touch struct {
	Trigger *Body
	Other   *Body // the moving body that touches it
}

// Begun are the touches that began in the last step: each is reported once, when a moving body
// first overlaps a trigger, and again only after it has left and comes back. The slice is the
// world's own and is reused by the next step: read it, don't keep it.
func (world *World) Begun() []Touch { return world.begun }

// findTouches finds which moving bodies overlap each trigger now, and which of those touches are
// new since the last step. Its lists are kept and reused, so a step allocates nothing once they
// have grown.
func (world *World) findTouches() {
	world.touching, world.wasTouching = world.wasTouching[:0], world.touching
	world.begun = world.begun[:0]
	for _, trigger := range world.bodies {
		if trigger.Kind == Trigger {
			world.findTouchesOf(trigger)
		}
	}
}

func (world *World) findTouchesOf(trigger *Body) {
	for _, other := range world.bodies {
		if other.Kind != Moving || !trigger.overlaps(other) {
			continue
		}
		touch := Touch{Trigger: trigger, Other: other}
		world.touching = append(world.touching, touch)
		if !contains(world.wasTouching, touch) {
			world.begun = append(world.begun, touch)
		}
	}
}

func contains(touches []Touch, wanted Touch) bool {
	for _, touch := range touches {
		if touch == wanted {
			return true
		}
	}
	return false
}
