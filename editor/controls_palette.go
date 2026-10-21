package editor

import (
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The inputs of the target shown, group by group, to add to the action chosen with a click: a 3DS's
// buttons and a game pad's laid out as they sit on the device, the rest as rows of chips.

// The devices' drawings' measures, in pixels.
const (
	deviceGap   = 10
	deviceRound = 14
)

// inputPalette is every input of the target shown, by group.
func (view *View) inputPalette(context *ui.Context) {
	controls := &view.controls
	context.Heading("Inputs")
	if controls.chosen == "" {
		note(context, "Choose an action on the left, then click its inputs here.")
	} else {
		note(context, "Click an input to add it to "+controls.chosen+".")
	}
	groups := map[string][]string{}
	var order []string
	for _, choice := range engine.InputChoices(controls.shownTarget()) {
		if groups[choice.Group] == nil {
			order = append(order, choice.Group)
		}
		groups[choice.Group] = append(groups[choice.Group], choice.Word)
	}
	for _, group := range order {
		view.inputGroup(context, group, groups[group])
	}
}

// inputGroup is a group's heading and its inputs.
func (view *View) inputGroup(context *ui.Context, group string, words []string) {
	context.Push(group)
	defer context.Pop()
	context.Separator()
	context.Row(func() {
		context.IconSized(groupIcon(group), stateIconSize, context.Theme.Primary)
		context.Text(group, ui.TextStyle{Weight: weightFor(true)})
	})
	switch group {
	case engine.ButtonInputs:
		view.deviceDrawing(context, consoleLayout)
	case engine.PadInputs:
		view.deviceDrawing(context, padLayout)
	default:
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
			for _, word := range words {
				view.paletteChip(context, word)
			}
		})
	}
}

// deviceLayout is a device's buttons as they sit on it: its shoulders above, its left hand's and
// right hand's buttons (each a cross: up, then left and right, then down), and the middle ones
// below.
type deviceLayout struct {
	shoulders   [2][]string // the left's, the right's
	left, right [4]string   // up, left, right, down
	middle      []string
}

// consoleLayout is a 3DS: the cross on the left, A B X Y on the right.
var consoleLayout = deviceLayout{
	shoulders: [2][]string{{"ZL", "L"}, {"R", "ZR"}},
	left:      [4]string{"Up", "Left", "Right", "Down"},
	right:     [4]string{"X", "Y", "A", "B"},
	middle:    []string{"Select", "Start"},
}

// padLayout is a game pad, in its one layout: the cross on the left, A B X Y on the right.
var padLayout = deviceLayout{
	shoulders: [2][]string{{"Pad Left Shoulder"}, {"Pad Right Shoulder"}},
	left:      [4]string{"Pad Up", "Pad Left", "Pad Right", "Pad Down"},
	right:     [4]string{"Pad Y", "Pad X", "Pad B", "Pad A"},
	middle:    []string{"Pad Back", "Pad Start", "Pad Left Thumb", "Pad Right Thumb"},
}

// deviceDrawing is a device's body with its buttons where they are, each one a chip to click.
func (view *View) deviceDrawing(context *ui.Context, layout deviceLayout) {
	look := ui.Look{Fill: context.Theme.Sunken, Radius: deviceRound}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(deviceGap), Gap: deviceGap}, look, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			view.chipRow(context, layout.shoulders[0])
			context.Box(ui.Layout{Width: ui.Grow()}, nil)
			view.chipRow(context, layout.shoulders[1])
		})
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			view.crossOfChips(context, layout.left)
			context.Box(ui.Layout{Width: ui.Grow()}, nil)
			view.crossOfChips(context, layout.right)
		})
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.Center, Gap: 6, Wrap: true}, func() {
			for _, word := range layout.middle {
				view.paletteChip(context, word)
			}
		})
	})
}

// crossOfChips is four buttons in a cross: up above, left and right beside, down below.
func (view *View) crossOfChips(context *ui.Context, buttons [4]string) {
	context.Box(ui.Layout{Direction: ui.Vertical, AlignX: ui.Center, Gap: 4}, func() {
		view.paletteChip(context, buttons[0])
		view.chipRow(context, buttons[1:3])
		view.paletteChip(context, buttons[3])
	})
}

// chipRow is some buttons side by side.
func (view *View) chipRow(context *ui.Context, words []string) {
	context.Box(ui.Layout{Gap: 4}, func() {
		for _, word := range words {
			view.paletteChip(context, word)
		}
	})
}

// paletteChip is an input to add to the action chosen; it is shown chosen when the action has it.
func (view *View) paletteChip(context *ui.Context, word string) {
	controls := &view.controls
	target := controls.shownTarget()
	look := ui.ChipLook{Fill: context.Theme.Background, Ink: context.Theme.Ink} // a key, raised from the panel
	for _, bound := range controls.actions[controls.chosen][target] {
		if bound == word {
			look = ui.ChipLook{Fill: context.Theme.Primary, Ink: context.Theme.OnPrimary}
		}
	}
	context.Push(word)
	defer context.Pop()
	if context.Chip("", inputLabel(word), look) {
		controls.addInput(word)
	}
	context.Tooltip(word)
}
