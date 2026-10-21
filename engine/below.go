package engine

import (
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/window"
)

// The 3DS's bottom screen can show a 3D scene of its own under the game's interface: a map, a
// model turning in a shop. It plays as the main scene does, its behaviours included, and is seen
// from its main camera. Elsewhere there is no bottom screen and it is not drawn, though it still
// plays.

// bottomScene is the scene the bottom screen shows, when the game sets one.
var bottomScene *Scene

// SetBottomScene shows a scene on the 3DS's bottom screen, under the game's interface
// (SetInterface), seen from its main camera; nil leaves the screen to the interface. Make it
// with Load or NewScene.
func SetBottomScene(below *Scene) {
	if below != nil {
		below.Start()
	}
	bottomScene = below
}

// playBelow moves the bottom screen's scene on a frame, and says how it is seen.
func (game *runner) playBelow(seconds float64, keys *input.State) *window.BelowView {
	if bottomScene == nil {
		return nil
	}
	bottomScene.Update(seconds, keys)
	game.below = window.BelowView{World: bottomScene.World(), Camera: CameraAt(fallbackCamera, defaultCamera)}
	if camera, found := bottomScene.MainCamera(); found {
		game.below.Camera = camera.View()
	}
	return &game.below
}
