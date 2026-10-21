package main

import (
	"fmt"
	"math"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/space"
)

// Lamp turns the lighthouse's lamp as the keeper asks (the action turn, from -1 to 1), and marks
// in the log where it points when the keeper asks (the action mark).
type Lamp struct {
	DegreesPerSecond float64 `json:"degreesPerSecond"`
}

func (lamp *Lamp) Update(context *engine.Context) {
	turn := context.Input.Action("turn").Value
	if turn != 0 {
		degrees := turn * lamp.DegreesPerSecond * context.Delta
		context.Node.Transform.Turn(space.Up, -degrees) // the right arrow turns it clockwise, seen from above
		watch.turn(degrees)
	}
	if context.Input.Action("mark").Pressed {
		watch.mark()
	}
}

// watch is what the keeper has seen tonight: where the lamp points, and the log.
var watch = keeperWatch{heading: "Lamp at 0°"}

// longestLog is how many lines the log keeps: the oldest go first.
const longestLog = 6

// keeperWatch is the lamp's heading, in whole degrees and as the line that shows it, and the log's
// lines. The lines are written when they change, not on every frame the surfaces paint them.
type keeperWatch struct {
	degrees float64
	shown   int
	heading string
	lines   []string
	marked  int // ships marked tonight, the lines gone included
}

// turn turns the lamp by some degrees, rewriting its heading's line when its whole degree changes.
func (keeper *keeperWatch) turn(degrees float64) {
	keeper.degrees = math.Mod(keeper.degrees+degrees+360, 360)
	if whole := int(keeper.degrees); whole != keeper.shown {
		keeper.shown = whole
		keeper.heading = fmt.Sprintf("Lamp at %d°", whole)
	}
}

// mark writes a line in the log: a ship seen where the lamp points.
func (keeper *keeperWatch) mark() {
	keeper.marked++
	line := fmt.Sprintf("Ship %d sighted at %d°", keeper.marked, keeper.shown)
	if len(keeper.lines) == longestLog {
		keeper.lines = append(keeper.lines[:0], keeper.lines[1:]...)
	}
	keeper.lines = append(keeper.lines, line)
}
