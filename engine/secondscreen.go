package engine

import (
	"log/slog"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/window"
)

// A device with a second screen (a 3DS's bottom one) can show a 3D scene of its own there, under
// the game's interface: a map, a model turning in a shop. It plays as the main scene does, its
// behaviours included, and is seen from its main camera. Without a second screen it still plays,
// and the editor's Play shows it where the device would.

// secondScene is the scene the second screen shows, when the game sets one.
var secondScene *Scene

// secondPlace is where a device with no second screen shows its scene, in window pixels, and
// whether the game placed it anywhere (PlaceSecondScreen).
var secondPlace struct {
	area   draw2d.Rect
	placed bool
}

// PlaceSecondScreen shows the second screen's scene in an area of the window, in window pixels, on
// a device that has no second screen: a radar in a corner of a computer's window, the same scene
// a 3DS shows below. nil shows it nowhere, as by default. A device with a second screen ignores
// it. Call it again when the window's size changes (each frame is fine: it allocates nothing).
func PlaceSecondScreen(area *draw2d.Rect) {
	if area == nil {
		secondPlace.placed = false
		return
	}
	secondPlace.area, secondPlace.placed = *area, true
}

// SetSecondScreen shows a scene on the second screen, under the game's interface (SetInterface),
// seen from its main camera; nil leaves the screen to the interface. Make it with Load or
// NewScene.
func SetSecondScreen(shown *Scene) {
	if shown != nil {
		for _, problem := range shown.Start() {
			slog.Warn("a behaviour of the second screen's scene could not be made", "error", problem)
		}
	}
	secondScene = shown
}

// SecondScreen is the scene the second screen shows (SetSecondScreen), if any: one who plays the
// game elsewhere than Run moves it on each frame (Scene.Update) and shows it from Scene.Camera.
func SecondScreen() (*Scene, bool) { return secondScene, secondScene != nil }

// playSecondScreen moves the second screen's scene on a frame, and says how it is seen.
func (game *runner) playSecondScreen(seconds float64, keys *input.State) *window.BelowView {
	if secondScene == nil {
		return nil
	}
	secondScene.Update(seconds, keys)
	game.below = window.BelowView{World: secondScene.World(), Camera: secondScene.Camera()}
	if secondPlace.placed {
		game.belowArea = secondPlace.area
		game.below.Area = &game.belowArea
	}
	return &game.below
}
