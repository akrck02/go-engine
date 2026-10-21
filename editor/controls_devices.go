package editor

import (
	"slices"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// The devices drawn in the Controls workspace: their buttons where they sit, each saying the
// actions it does and choosing them when clicked. Screens are dark boxes, a body is a rounded
// frame: enough to recognise the device, not a picture of it.

// The devices' measures, in pixels.
const (
	deviceGap        = 12
	deviceRound      = 22
	buttonRound      = 8
	keyWidth         = 52
	consoleTopWidth3 = 320 // the 3DS's top screen, at four fifths of its 400 × 240
	consoleTopTall3  = 192
	consoleLowWidth3 = 256 // its touch screen, 320 × 240
	consoleBodyWidth = 560
	phoneWidth       = 280
	phoneTall        = 500
)

// screenColor is a device's screen: dark glass.
var screenColor = draw2d.Color{0.12, 0.13, 0.16, 1}

// deviceButton is a button of a device: its name, the actions it does under it, and on a click
// the actions to choose from.
func (view *View) deviceButton(context *ui.Context, word, label string, width float32) {
	controls := &view.controls
	bound := controls.bindingsOf(word)
	normal := ui.Look{Fill: context.Theme.Background, Radius: buttonRound, Border: context.Theme.Line, BorderWidth: 1}
	if len(bound) > 0 {
		normal.Border, normal.BorderWidth = context.Theme.Primary, 2
	}
	hovered := normal
	hovered.Fill = context.Theme.Sunken
	layout := ui.Layout{Direction: ui.Vertical, Padding: ui.PadXY(8, 5), Gap: 2, AlignX: ui.Center}
	if width > 0 {
		layout.Width = ui.Fixed(width)
	}
	context.Push(word)
	defer context.Pop()
	if context.Clickable("button", layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		context.Text(label, ui.TextStyle{Weight: draw2d.Bold})
		for _, action := range bound {
			context.Text(action, ui.TextStyle{Color: context.Theme.Primary, Size: context.Theme.SmallTextSize})
		}
	}) {
		controls.opened = word
	}
	open := controls.opened == word
	context.Popover("Actions", &open, ui.Look{Fill: context.Theme.Surface, Radius: buttonRound, Border: context.Theme.Line, BorderWidth: 1}, func() {
		view.actionChoices(context, word)
	})
	if !open && controls.opened == word {
		controls.opened = ""
	}
}

// actionChoices are the actions a button can do, each ticked when it does it: an action that
// presses once, one that moves an axis twice (towards − and towards +). A stick moves axes only.
func (view *View) actionChoices(context *ui.Context, word string) {
	controls := &view.controls
	names := controls.names()
	if len(names) == 0 {
		note(context, "Add an action on the left first.")
		return
	}
	for _, name := range names {
		for _, choice := range choicesFor(word, name, controls.axes[name]) {
			icon := ""
			if slices.Contains(controls.actions[name][controls.shownTarget()], choice.word) {
				icon = iconFound
			}
			if context.MenuItem(icon, choice.label, "") {
				controls.toggleBinding(name, choice.word)
			}
		}
	}
}

// actionChoice is one way a button can do an action: the word bound, and how the list says it.
type actionChoice struct{ word, label string }

// choicesFor are the ways a button does an action: pressing it, or moving its axis one way or the
// other; a stick moves an axis, or nothing.
func choicesFor(word, name string, axis bool) []actionChoice {
	switch {
	case isStick(word) && axis:
		return []actionChoice{{word, name}}
	case isStick(word):
		return nil
	case axis:
		return []actionChoice{{word + "-", name + " −"}, {word + "+", name + " +"}}
	}
	return []actionChoice{{word, name}}
}

// stick is a stick, its two axes as buttons.
func (view *View) stick(context *ui.Context, prefix, title string) {
	look := ui.Look{Fill: context.Theme.Sunken, Radius: deviceRound, Border: context.Theme.Line, BorderWidth: 1}
	context.Frame(ui.Layout{Direction: ui.Vertical, Padding: ui.Pad(8), Gap: 4, AlignX: ui.Center}, look, func() {
		context.Caption(title)
		context.Box(ui.Layout{Gap: 4}, func() {
			view.deviceButton(context, prefix+" X", "←→", 0)
			view.deviceButton(context, prefix+" Y", "↑↓", 0)
		})
	})
}

// cross is four buttons in a cross: up above, left and right beside, down below.
func (view *View) cross(context *ui.Context, words, labels [4]string) {
	context.Box(ui.Layout{Direction: ui.Vertical, AlignX: ui.Center, Gap: 4}, func() {
		view.deviceButton(context, words[0], labels[0], 0)
		context.Box(ui.Layout{Gap: 4}, func() {
			view.deviceButton(context, words[1], labels[1], 0)
			view.deviceButton(context, words[2], labels[2], 0)
		})
		view.deviceButton(context, words[3], labels[3], 0)
	})
}

// buttons are some buttons side by side, each named by its word.
func (view *View) buttons(context *ui.Context, words ...string) {
	context.Box(ui.Layout{Gap: 4}, func() {
		for _, word := range words {
			view.deviceButton(context, word, inputLabel(word), 0)
		}
	})
}

// screen is a device's screen, with what it is written on it.
func screen(context *ui.Context, width, height float32, caption string) {
	look := ui.Look{Fill: screenColor, Radius: 4}
	context.Frame(ui.Layout{Width: ui.Fixed(width), Height: ui.Fixed(height), AlignX: ui.Center, AlignY: ui.Center}, look, func() {
		context.Text(caption, ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
	})
}

// deviceBody is a device's rounded body, its content in a column.
func deviceBody(context *ui.Context, layout ui.Layout, content func()) {
	layout.Direction, layout.Padding, layout.Gap, layout.AlignX = ui.Vertical, ui.Pad(deviceGap*1.5), deviceGap, ui.Center
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface, Radius: deviceRound, Border: context.Theme.Line, BorderWidth: 2}, content)
}
