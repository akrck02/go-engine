package main

import ui "github.com/akrck02/go-ui"

// The game's surfaces: it paints them, and engine.json says where each goes on each device.

// showHeading shows where the lamp points, on a card that reads over the dark sea.
func showHeading(context *ui.Context) {
	card := ui.Look{Fill: context.Theme.Surface, Radius: context.Theme.ControlRadius}
	context.Frame(ui.Layout{Padding: ui.PadXY(10, 6)}, card, func() {
		context.Label(watch.heading)
	})
}

// showLog shows the keeper's log, the newest line last.
func showLog(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Padding: ui.Pad(8), Gap: 4}, func() {
		if len(watch.lines) == 0 {
			context.Caption("No ship yet: turn the lamp and mark the ones it finds")
			return
		}
		for _, line := range watch.lines {
			context.Label(line)
		}
	})
}
