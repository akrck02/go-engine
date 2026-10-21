package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On Android a 2D scene is painted by the window's interface, as on the desktop, with touch
// buttons (touchLayoutFor): a phone has no keyboard.

// flatNeedsScene is false on Android: a window given a scene would draw it over the interface.
const flatNeedsScene = false

// drawOver draws nothing over the scene: the interface does it (interfaceFor).
func drawOver(*Scene, *input.State, *draw2d.List) {}

// interfaceFor is what the window's interface shows: a 2D scene, the touch buttons, and the game's
// surfaces over them.
func interfaceFor(game *runner) func(context *ui.Context) {
	return func(context *ui.Context) {
		insets := context.Insets()
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
			safe := draw2d.Rect{X: area.X + insets.Left, Y: area.Y + insets.Top,
				Width: area.Width - insets.Left - insets.Right, Height: area.Height - insets.Top - insets.Bottom}
			played := game.played
			layout := touchLayoutFor(area, safe, played.designedAspect())
			if len(touchPanel.controls) > 0 { // the game's own controls lie over the whole screen
				layout = touchLayout{scene: area}
				placeTouchControls(safe)
			}
			game.console.buttons = layout.buttons
			if played.Is2D() { // a 3D scene is drawn under the interface: covering it would hide it
				list.Rectangle(area, 0, draw2d.Color{0, 0, 0, 1}, draw2d.Color{}, 0)
				list.PushClip(layout.scene)
				played.Draw2D(list, layout.scene, played.View2D())
				list.PopClip()
			}
			if len(touchPanel.controls) > 0 {
				drawTouchControls(list)
				return
			}
			drawTouchButtons(list, layout.buttons)
		})
		// the surfaces keep off the system's bars and cut-outs, as the touch controls do
		screen := game.screen
		safe := draw2d.Rect{X: insets.Left, Y: insets.Top,
			Width: screen.Width - insets.Left - insets.Right, Height: screen.Height - insets.Top - insets.Bottom}
		paintSurfaces(context, Menu{game}, safe, true)
		if len(touchPanel.controls) == 0 {
			// TODO(akrck02): a screen reader reads the engine's four buttons, not a game's own touch controls yet
			game.console.describeTouchButtons(context)
		}
	}
}

// mainScreen is nothing on Android: the window's interface paints the surfaces.
type mainScreen struct{}

// paintOnMainScreen does nothing on Android: the window's interface paints the surfaces.
func (game *runner) paintOnMainScreen() {}

// consoleKeys is the game's keys on Android: a keyboard's, if one is plugged in, and the touch
// buttons'.
func consoleKeys(console *consoleInput, window *input.State) *input.State {
	readTouchControls(window.Touches())
	return touchKeys(&console.keys, window, console.buttons, console.readerHeld())
}

// readConsole is an empty reading: there is no 3DS here.
func readConsole() ctr.Reading { return ctr.Reading{} }

// startConsole does nothing: there is no 3DS here.
func startConsole() {}

// stopConsole does nothing: there is no 3DS here.
func stopConsole() {}

// runningTarget is the target this build runs on.
const runningTarget = AndroidTarget
