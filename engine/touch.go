package engine

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Touch buttons are the console's buttons drawn on a touch screen with no keyboard (a phone): the
// cross's left and right on the left, A and B on the right. A finger on one holds its key, as the
// console's own button does. A phone held upright shows the game at the top, at the shape its
// camera was made for, and the buttons under it, as a handheld console does; held sideways, the
// game fills the screen and the buttons lie over its bottom corners.

// The touch buttons' measures, in layout pixels.
const (
	touchButtonSide   = 72
	touchButtonMargin = 24
	touchButtonGap    = 16
	touchButtonIcon   = 36
	touchButtonAlpha  = 0.35
	touchBandHeight   = touchButtonSide + 2*touchButtonMargin
)

// touchButton is a button on the screen and the key it holds.
type touchButton struct {
	key  input.Key
	icon string
	area draw2d.Rect
}

// touchLayout is where a phone shows the game and its buttons.
type touchLayout struct {
	scene   draw2d.Rect
	buttons [4]touchButton
}

// touchLayoutFor places the game and the buttons on a screen: area is the whole screen, safe the
// part the system draws nothing over, and aspect the shape the game's camera was made for (width
// over height; 0 when it says none).
func touchLayoutFor(area, safe draw2d.Rect, aspect float64) touchLayout {
	if aspect <= 0 || safe.Width >= safe.Height {
		band := safe
		band.Y, band.Height = safe.Y+safe.Height-touchBandHeight, touchBandHeight
		return touchLayout{scene: area, buttons: touchButtonsIn(band)}
	}
	scene := safe
	scene.Height = min(float32(float64(safe.Width)/aspect), safe.Height-touchBandHeight)
	band := safe
	band.Y, band.Height = scene.Y+scene.Height, safe.Height-scene.Height
	return touchLayout{scene: scene, buttons: touchButtonsIn(band)}
}

// touchButtonsIn places the buttons in a band across the screen: left and right on its left, B and
// A on its right, in the middle of its height.
func touchButtonsIn(band draw2d.Rect) [4]touchButton {
	top := band.Y + (band.Height-touchButtonSide)/2
	square := func(x float32) draw2d.Rect {
		return draw2d.Rect{X: x, Y: top, Width: touchButtonSide, Height: touchButtonSide}
	}
	step := float32(touchButtonSide + touchButtonGap)
	left := band.X + touchButtonMargin
	right := band.X + band.Width - touchButtonMargin - touchButtonSide
	return [4]touchButton{
		{input.KeyLeft, "arrow-left", square(left)},
		{input.KeyRight, "arrow-right", square(left + step)},
		{input.KeyEscape, "circle-x", square(right - step)}, // B
		{input.KeySpace, "circle-arrow-up", square(right)},  // A
	}
}

// touchKeys sets a game's keys for this frame, kept from frame to frame as on the console (the
// window hands the scene a fresh copy of its input every frame): those held on a keyboard, and
// those whose button a finger is on.
func touchKeys(keys, window *input.State, buttons [4]touchButton) *input.State {
	keys.EndFrame()
	for _, button := range buttons {
		keys.SetKey(button.key, window.Down(button.key) || touched(window, button.area))
	}
	keys.StickX, keys.StickY = window.StickX, window.StickY
	return keys
}

// touched tells whether a finger is on an area.
func touched(window *input.State, area draw2d.Rect) bool {
	for _, finger := range window.Touches() {
		if area.Contains(float32(finger.X), float32(finger.Y)) {
			return true
		}
	}
	return false
}

// drawTouchButtons draws the touch buttons.
func drawTouchButtons(list *draw2d.List, buttons [4]touchButton) {
	for _, button := range buttons {
		shade := draw2d.Color{0, 0, 0, touchButtonAlpha}
		list.Rectangle(button.area, touchButtonSide/2, shade, draw2d.Color{1, 1, 1, touchButtonAlpha}, 2)
		inset := float32(touchButtonSide-touchButtonIcon) / 2
		list.Icon(button.icon, button.area.X+inset, button.area.Y+inset, touchButtonIcon, draw2d.Color{1, 1, 1, 0.9})
	}
}
