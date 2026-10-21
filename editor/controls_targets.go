package editor

import (
	"strings"

	ui "github.com/akrck02/go-ui"
)

// Each target's device, for the Controls workspace: a 3DS, a phone with the touch buttons the
// engine draws, and a computer's keyboard, mouse and game pad.

// consoleDevice is a 3DS: its two screens, the circle pad and the cross on the left, the face
// buttons and the C-stick on the right, the shoulders above and Select and Start below.
func (view *View) consoleDevice(context *ui.Context) {
	body := ui.Layout{Width: ui.Fixed(consoleBodyWidth)} // both halves as wide, as a 3DS opened
	deviceBody(context, body, func() {
		screen(context, consoleTopWidth3, consoleTopTall3, "Top screen")
	})
	deviceBody(context, body, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			view.buttons(context, "ZL", "L")
			context.Box(ui.Layout{Width: ui.Grow()}, nil)
			view.buttons(context, "R", "ZR")
		})
		context.Box(ui.Layout{Gap: deviceGap * 2, AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Gap: deviceGap, AlignX: ui.Center}, func() {
				view.stick(context, "Stick", "Circle pad")
				view.cross(context, [4]string{"Up", "Left", "Right", "Down"}, [4]string{"↑", "←", "→", "↓"})
			})
			screen(context, consoleLowWidth3, consoleTopTall3, "Touch screen")
			context.Box(ui.Layout{Direction: ui.Vertical, Gap: deviceGap, AlignX: ui.Center}, func() {
				view.stick(context, "Right Stick", "C-stick")
				view.cross(context, [4]string{"X", "Y", "A", "B"}, [4]string{"X", "Y", "A", "B"})
			})
		})
		view.buttons(context, "Select", "Start")
	})
}

// phoneDevice is a phone held sideways in its body, with the touch buttons the engine draws on
// its screen (left and right, B and A) and the system's back under it.
func (view *View) phoneDevice(context *ui.Context) {
	deviceBody(context, ui.Layout{}, func() {
		look := ui.Look{Fill: screenColor, Radius: 6}
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(phoneTall), Height: ui.Fixed(phoneWidth), Padding: ui.Pad(deviceGap)}, look, func() {
			context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, func() {
				context.Text("The game", ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
			})
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				view.deviceButton(context, "Left", "←", 0)
				view.deviceButton(context, "Right", "→", 0)
				context.Box(ui.Layout{Width: ui.Grow()}, nil)
				view.deviceButton(context, "Escape", "B", 0)
				view.deviceButton(context, "Space", "A", 0)
			})
		})
		view.deviceButton(context, "Back", "Back", 0)
	})
}

// keyboardRows are a computer's keyboard as its rows, the keys' words; Space and the arrows are
// the last row's.
var keyboardRows = [][]string{
	{"Escape", "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12"},
	{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "Minus", "Plus", "Backspace"},
	{"Tab", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "Delete"},
	{"A", "S", "D", "F", "G", "H", "J", "K", "L", "Enter"},
	{"Shift", "Z", "X", "C", "V", "B", "N", "M", "Page Up", "Page Down"},
	{"Control", "Alt", "Space", "Home", "End", "Left", "Up", "Down", "Right"},
}

// keyLabels are the keys whose cap says something shorter than their word.
var keyLabels = map[string]string{"Escape": "Esc", "Backspace": "⌫", "Minus": "-", "Plus": "+", "Delete": "Del",
	"Control": "Ctrl", "Page Up": "PgUp", "Page Down": "PgDn", "Left": "←", "Up": "↑", "Down": "↓", "Right": "→", "Enter": "Enter"}

// computerDevices are a keyboard, a mouse and a game pad.
func (view *View) computerDevices(context *ui.Context) {
	deviceBody(context, ui.Layout{}, func() {
		context.Caption("Keyboard")
		for row, keys := range keyboardRows {
			context.Push(strings.Repeat("row", row+1))
			context.Box(ui.Layout{Gap: 4}, func() {
				for _, key := range keys {
					label, found := keyLabels[key]
					if !found {
						label = key
					}
					width := float32(keyWidth)
					if key == "Space" {
						width = keyWidth * 4
					}
					view.deviceButton(context, key, label, width)
				}
			})
			context.Pop()
		}
	})
	context.Box(ui.Layout{Gap: deviceGap * 2}, func() {
		deviceBody(context, ui.Layout{}, func() {
			context.Caption("Mouse")
			view.buttons(context, "Mouse Left", "Mouse Middle", "Mouse Right")
		})
		view.gamePadDevice(context)
	})
}

// gamePadDevice is a game pad in its one layout: shoulders above, the left stick and the cross on
// the left, the face buttons and the right stick on the right, the middle buttons between.
func (view *View) gamePadDevice(context *ui.Context) {
	deviceBody(context, ui.Layout{}, func() {
		context.Caption("Game pad")
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			view.buttons(context, "Pad Left Shoulder")
			context.Box(ui.Layout{Width: ui.Grow()}, nil)
			view.buttons(context, "Pad Right Shoulder")
		})
		context.Box(ui.Layout{Gap: deviceGap * 2, AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Gap: deviceGap, AlignX: ui.Center}, func() {
				view.stick(context, "Stick", "Left stick")
				view.cross(context, [4]string{"Pad Up", "Pad Left", "Pad Right", "Pad Down"}, [4]string{"↑", "←", "→", "↓"})
			})
			context.Box(ui.Layout{Direction: ui.Vertical, Gap: 4, AlignX: ui.Center}, func() {
				view.buttons(context, "Pad Back", "Pad Start")
				view.buttons(context, "Pad Left Thumb", "Pad Right Thumb")
			})
			context.Box(ui.Layout{Direction: ui.Vertical, Gap: deviceGap, AlignX: ui.Center}, func() {
				view.cross(context, [4]string{"Pad Y", "Pad X", "Pad B", "Pad A"}, [4]string{"Y", "X", "B", "A"})
				view.stick(context, "Right Stick", "Right stick")
			})
		})
	})
}

// inputLabel is a word as a button shows it: a pad's or a mouse's button without the device.
func inputLabel(word string) string {
	return strings.TrimPrefix(strings.TrimPrefix(word, "Pad "), "Mouse ")
}
