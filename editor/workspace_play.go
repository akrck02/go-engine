package editor

import ui "github.com/akrck02/go-ui"

// The Play workspace: the scene played, and what controls it.

func (view *View) playWorkspace(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		view.transport(context)
		context.Separator()
		view.middle(context)
	})
}

// transport is the strip over the game: stop, and play in a window of its own.
func (view *View) transport(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(44), Padding: ui.PadXY(10, 0), Gap: 6, AlignX: ui.Center, AlignY: ui.Center}
	bar(context, layout, func() {
		if context.PrimaryIconButton(iconStop, "Stop") {
			view.Editor.LeavePlay()
		}
		context.Tooltip("Back to the workspace before (Escape)")
		view.windowButton(context)
	})
}

// windowButton plays the game in a window of its own, or closes it.
func (view *View) windowButton(context *ui.Context) {
	editor := view.Editor
	if editor.Running() && editor.playing != nil {
		if context.IconButton(iconClose, "Close window") {
			editor.Stop()
		}
		context.Tooltip("Close the game's window")
		return
	}
	if context.IconButton(iconInWindow, "In a window") {
		editor.PlayInWindow()
	}
	context.Tooltip("Build the game and play it in a window of its own")
}
