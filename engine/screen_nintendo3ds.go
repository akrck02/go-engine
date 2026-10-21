//go:build nintendo3ds

package engine

import (
	"log/slog"

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

// interfaceFor is the bottom screen's interface: the surfaces placed there (the game's interface,
// SetInterface, among them), or the menu of scenes when the settings ask for it.
func interfaceFor(game *runner) func(context *ui.Context) {
	return func(context *ui.Context) {
		if paintSecondScreen(context, Menu{game}) {
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

// mainScreen paints the surfaces placed on the top screen, which the window's interface does not
// reach, with a Go UI context of its own; no finger touches it, so it is given no pointer.
type mainScreen struct {
	context *ui.Context
	state   *input.State
}

// The top screen's size, in pixels.
const (
	topWidth  = 400
	topHeight = 240
)

// paintOnMainScreen paints the surfaces placed on the top screen over the scene, after the game's
// own 2D.
func (game *runner) paintOnMainScreen() {
	if !hasMainScreenSurfaces() {
		return
	}
	top := &game.top
	if top.context == nil && !top.start(game.context) {
		return
	}
	top.context.Begin(top.state)
	paintSurfaces(top.context, Menu{game}, draw2d.Rect{Width: topWidth, Height: topHeight}, false)
	top.context.End()
	drawn := top.context.Drawing().Commands
	if game.played.Is2D() {
		game.flat.Commands = append(game.flat.Commands, drawn...) // a 2D scene's own drawing (drawOver)
		return
	}
	if game.view.Drawing == nil {
		game.overlay.Reset()
		game.view.Drawing = &game.overlay
	}
	game.view.Drawing.Commands = append(game.view.Drawing.Commands, drawn...)
}

// start makes the top screen's context, looking as the bottom screen's, and tells whether it could.
func (top *mainScreen) start(bottom *ui.Context) bool {
	context, err := ui.New()
	if err != nil {
		slog.Error("the top screen's surfaces cannot be painted", "error", err)
		return false
	}
	top.context = context
	top.context.Theme = bottom.Theme
	top.state = input.NewState()
	top.state.WindowWidth, top.state.WindowHeight = topWidth, topHeight
	top.state.FramebufferWidth, top.state.FramebufferHeight = topWidth, topHeight
	top.state.CursorX, top.state.CursorY = offScreen, offScreen
	return true
}

// offScreen is where the top screen's pointer is: nowhere a control could be.
const offScreen = -1e6

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

// runningTarget is the target this build runs on.
const runningTarget = Nintendo3DSTarget
