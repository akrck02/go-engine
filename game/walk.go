package game

// Walker turns the arrows held into steps: one when an arrow is pressed, then, while it is held,
// another every repeat, so holding an arrow walks. Every platform feeds it what it reads.
type Walker struct {
	held    Direction
	waiting float64 // seconds until the next step of a held arrow
}

// Walking pace: the wait before a held arrow repeats, and between the repeats.
const (
	firstRepeat  = 0.3
	repeat       = 0.15
	longestFrame = 0.05
)

// Step tells which step to take this frame, given the arrow held (none: the zero Direction) and
// the seconds since the last frame.
func (walker *Walker) Step(held Direction, seconds float64) (Direction, bool) {
	if held == (Direction{}) {
		walker.held = held
		return held, false
	}
	if held != walker.held {
		walker.held, walker.waiting = held, firstRepeat
		return held, true
	}
	// A window that draws only when something happens can go long without a frame; that time
	// was not spent holding the arrow, so a frame counts for little.
	walker.waiting -= min(seconds, longestFrame)
	if walker.waiting > 0 {
		return held, false
	}
	walker.waiting += repeat
	if walker.waiting <= 0 {
		walker.waiting = repeat // a long frame steps once, not many times
	}
	return held, true
}
