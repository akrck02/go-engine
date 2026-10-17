// Package player plays a game made with go-engine, in a go-ui window: on the desktop the scene
// fills the window, under a bar with the steps; on the Nintendo 3DS the scene is on the top screen
// and the bar, touched, on the bottom one. A program exported from the editor is a call to Run.
package player

import (
	"strconv"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/grid/game"
)

// Session is a game being played: the world, the arrows held, and the texts the bar shows, kept
// written while what they say does not change.
type Session struct {
	World    *game.World
	walker   game.Walker
	lastStep time.Time // when Step last ran, by the input's clock

	shownSteps, shownCollected, shownToCollect int
	stepsText, collectedText                   string
}

// NewSession starts a game.
func NewSession(played *game.Game) *Session {
	return &Session{World: game.NewWorld(played), shownSteps: -1}
}

// Step moves the game on a frame: an arrow held walks, R starts the scene again, and once the
// game is won Enter plays it again. It tells whether an arrow is held, so the window keeps
// drawing while it is.
func (session *Session) Step(state *input.State, seconds float64) (walking bool) {
	// The time held is measured on the input's clock: a window's step is the display's frame,
	// which a window drawing faster (no vsync) or slower than the display does not keep.
	now := state.Time()
	if !session.lastStep.IsZero() {
		seconds = now.Sub(session.lastStep).Seconds()
	}
	session.lastStep = now
	held := HeldDirection(state)
	if direction, step := session.walker.Step(held, seconds); step {
		session.World.Move(direction)
	}
	if state.Pressed(input.KeyR) {
		session.World.Restart()
	}
	if session.World.State == game.Won && state.Pressed(input.KeyEnter) {
		session.World.Start(0)
	}
	return held != (game.Direction{})
}

// HeldDirection is the arrow held, the first of up, down, left and right.
func HeldDirection(state *input.State) game.Direction {
	switch {
	case state.Down(input.KeyUp):
		return game.Up
	case state.Down(input.KeyDown):
		return game.Down
	case state.Down(input.KeyLeft):
		return game.Left
	case state.Down(input.KeyRight):
		return game.Right
	}
	return game.Direction{}
}

// Bar describes the strip of the game's state: the scene's name, the steps, what is collected, a
// line for the player, and a button to start again (or to play again, once won).
func (session *Session) Bar(context *ui.Context) {
	world := session.World
	session.writeTexts()
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(12, 8), Gap: 12, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		if world.State == game.Won {
			context.Heading("You won!")
			context.Spacer()
			if context.PrimaryButton("Play again") {
				world.Start(0)
			}
			return
		}
		context.Heading(world.Scene().Name)
		context.Caption(session.stepsText)
		if world.ToCollect > 0 {
			context.Caption(session.collectedText)
		}
		if world.Notice != "" {
			context.Text(world.Notice, ui.TextStyle{Color: context.Theme.Warning})
		}
		context.Spacer()
		if context.Button("Restart") {
			world.Restart()
		}
	})
}

// writeTexts writes the bar's figures again only when they change.
func (session *Session) writeTexts() {
	world := session.World
	if world.Steps != session.shownSteps {
		session.shownSteps = world.Steps
		session.stepsText = "Steps " + strconv.Itoa(world.Steps)
	}
	if world.Collected != session.shownCollected || world.ToCollect != session.shownToCollect || session.collectedText == "" {
		session.shownCollected, session.shownToCollect = world.Collected, world.ToCollect
		session.collectedText = "Collected " + strconv.Itoa(world.Collected) + " of " + strconv.Itoa(world.ToCollect)
	}
}

// Draw draws the scene in play into an area.
func (session *Session) Draw(list *draw2d.List, area draw2d.Rect) {
	game.DrawWorld(list, area, session.World)
}
