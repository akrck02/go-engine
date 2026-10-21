//go:build nintendo3ds

package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On the console a 2D scene is drawn on the top screen, as the scene's own drawing, which go-ui's
// window paints over the (empty) 3D scene; the interface is the bottom screen's.

// drawOver draws a 2D scene over its empty world, on the top screen.
func drawOver(played *Scene, _ *input.State, flat *draw2d.List) {
	if !played.Is2D() {
		return
	}
	flat.Reset()
	played.Draw2D(flat, draw2d.Rect{Width: 400, Height: 240}, played.View2D())
	played.World().Overlay.Drawing = flat
}

// flatNeedsScene is true on the console: a 2D scene is drawn over the top screen's scene.
const flatNeedsScene = true

// interfaceFor is the bottom screen's interface: the game's own (SetInterface), or the menu of
// scenes when the settings ask for it.
func interfaceFor(game *runner) func(context *ui.Context) {
	return func(context *ui.Context) {
		if gameInterface != nil {
			gameInterface(context, Menu{game})
			return
		}
		if len(game.menu) == 0 {
			return
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(menuPadding), Gap: menuGap}, func() {
			context.Heading(game.name)
			context.Caption("Touch a scene to play it")
			for _, scene := range game.menu {
				context.Push(scene.File)
				if context.Selectable("map", scene.Name, game.played.Name == scene.Name) {
					game.choose(scene.File)
				}
				context.Pop()
			}
		})
	}
}

// The menu's measures, in the bottom screen's pixels.
const (
	menuPadding = 10
	menuGap     = 4
)

// consoleKeys is the game's keys on the console: the cross as the arrow keys and A as Space, read
// from the console itself, whatever the bottom screen's interface did with them (consoleInput).
func consoleKeys(console *consoleInput, window *input.State) *input.State {
	return console.keysFrom(window, console.reading.Held)
}

// readConsole reads the console's controls for this frame.
func readConsole() ctr.Reading { return ctr.ReadControls() }

// startConsole starts the C-stick and the gyroscope, for the game to read. A console without one
// (or an emulator) reads it as still, so which started does not matter here.
func startConsole() { ctr.StartMotionControls() }

// stopConsole stops them when the game ends: the gyroscope uses power.
func stopConsole() { ctr.StopMotionControls() }
