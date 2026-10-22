package engine

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Touch buttons are the console's buttons drawn on a touch screen with no keyboard (a phone): the
// cross's left and right on the left, A and B on the right. A finger on one holds its key, as the
// console's own button does. A phone held upright shows the game at the top, at the shape its
// camera was made for (when the camera names a width), and the buttons under it, as a handheld
// console does; held sideways, or with a camera that names no width, the game fills the screen
// and the buttons lie over its bottom corners.
//
// A person who cannot see them uses a screen reader (TalkBack), which takes the touches itself:
// a finger moved over the screen reads what is under it, and a double tap presses it. While one
// is on, each button is also a control of the interface, named, which a double tap presses for a
// moment; a double tap held passes the touch through, and holds the button as a finger does.

// The touch buttons' measures, in layout pixels.
const (
	touchButtonSide   = 72
	touchButtonMargin = 24
	touchButtonGap    = 16
	touchButtonIcon   = 36
	touchButtonAlpha  = 0.35
	touchBandHeight   = touchButtonSide + 2*touchButtonMargin
)

// readerPressFrames is how many frames a screen reader's press holds a button's key: a fifth of a
// second at sixty frames a second, long enough for a game to see a jump start.
const readerPressFrames = 12

// readerPressArea is a touch button's control for a screen reader, filling the button.
var readerPressArea = ui.Layout{Width: ui.Grow(), Height: ui.Grow()}

// touchButton is a button on the screen and the key it holds.
type touchButton struct {
	key  input.Key
	icon string
	name string // what a screen reader calls it
	area draw2d.Rectangle
}

// touchLayout is where a phone shows the game and its buttons.
type touchLayout struct {
	scene   draw2d.Rectangle
	buttons [4]touchButton
}

// touchLayoutFor places the game and the buttons on a screen: area is the whole screen, safe the
// part the system draws nothing over, and aspect the shape the game's camera was made for (width
// over height; 0 when it says none).
func touchLayoutFor(area, safe draw2d.Rectangle, aspect float64) touchLayout {
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
func touchButtonsIn(band draw2d.Rectangle) [4]touchButton {
	top := band.Y + (band.Height-touchButtonSide)/2
	square := func(x float32) draw2d.Rectangle {
		return draw2d.Rectangle{X: x, Y: top, Width: touchButtonSide, Height: touchButtonSide}
	}
	step := float32(touchButtonSide + touchButtonGap)
	left := band.X + touchButtonMargin
	right := band.X + band.Width - touchButtonMargin - touchButtonSide
	return [4]touchButton{
		{input.KeyLeft, "arrow-left", "Left", square(left)},
		{input.KeyRight, "arrow-right", "Right", square(left + step)},
		{input.KeyEscape, "circle-x", "B", square(right - step)},
		{input.KeySpace, "circle-arrow-up", "A", square(right)},
	}
}

// touchKeys sets a game's keys for this frame, kept from frame to frame as on the console (the
// window hands the scene a fresh copy of its input every frame): those held on a keyboard, and
// those whose button a finger is on, and those a screen reader holds (readerHeld, by button).
func touchKeys(keys, window *input.State, buttons [4]touchButton, readerHeld [4]bool) *input.State {
	keys.Feed().EndFrame()
	for key := input.KeyUnknown + 1; key <= input.KeySuper; key++ { // KeySuper is the last key
		keys.Feed().SetKey(key, window.Down(key) || buttonHeld(window, buttons, readerHeld, key))
	}
	keys.Feed().SetStick(window.Stick())
	return keys
}

// buttonHeld tells whether a finger is on a button of a key, or a screen reader holds it.
func buttonHeld(window *input.State, buttons [4]touchButton, readerHeld [4]bool, key input.Key) bool {
	for index, button := range buttons {
		if button.key == key && (readerHeld[index] || touched(window, button.area)) {
			return true
		}
	}
	return false
}

// describeTouchButtons makes each touch button a control a screen reader reads by its name, while
// one is on, and notes the ones it pressed (readerHeld). Nothing is added while none is.
func (console *consoleInput) describeTouchButtons(context *ui.Context) {
	if !context.ScreenReaderOn() {
		return
	}
	for index, button := range console.buttons {
		options := ui.LayerOptions{Padding: ui.Padding{Left: button.area.X, Top: button.area.Y},
			Width: ui.Fixed(button.area.Width), Height: ui.Fixed(button.area.Height)}
		context.Overlay(button.icon, options, func() {
			if context.Clickable("press", readerPressArea, ui.Looks{}, nil) {
				console.readerFrames[index] = readerPressFrames
			}
			context.Describe(ui.Description{Role: ui.ButtonRole, Label: button.name})
		})
	}
}

// readerHeld tells which buttons a screen reader's press still holds, and counts that frame off.
func (console *consoleInput) readerHeld() [4]bool {
	var held [4]bool
	for index, frames := range console.readerFrames {
		if frames > 0 {
			held[index] = true
			console.readerFrames[index]--
		}
	}
	return held
}

// touched tells whether a finger is on an area.
func touched(window *input.State, area draw2d.Rectangle) bool {
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
