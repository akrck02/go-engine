package editor

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
)

// The Play workspace: the game in the middle at the shape of its screen, the controls of a player
// over it, the scene's nodes on the left to choose one to watch, what that node is doing on the
// right, and what the game says under it. What happens while playing is not kept in the scene.

// The Play workspace's measures, in pixels.
const (
	playSideWidth  = 240
	watchWidth     = 300
	playLogHeight  = 150
	transportTall  = 44
	letterboxColor = 0.12
)

func (view *View) playWorkspace(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Live scene", ui.Fixed(playSideWidth), func() { view.liveTree(context) })
		context.Separator()
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.transport(context)
			context.Separator()
			view.stage(context)
			context.Separator()
			view.logPanel(context)
		})
		context.Separator()
		panel(context, "Watch", ui.Fixed(watchWidth), func() { view.watchPanel(context) })
	})
}

// liveTree is the nodes of the scene playing (the game may have gone on to another), to choose
// the one watched.
func (view *View) liveTree(context *ui.Context) {
	played, playing := view.Editor.PlayingHere()
	context.Heading("Scene playing")
	if !playing {
		return
	}
	context.Caption(played.Name)
	note(context, "Choose a node to watch it.")
	for _, child := range played.Root.Children() {
		view.liveNode(context, child, 0)
	}
}

// liveNode is a node of the scene playing, and those under it; choosing one watches it.
func (view *View) liveNode(context *ui.Context, node *engine.Node, depth int) {
	editor := view.Editor
	context.Push(node.Name)
	defer context.Pop()
	watched := editor.Watched == node.Path()
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 16}}, func() {
		if context.Selectable(nodeIcon(node), node.Name, watched) {
			editor.Watched = node.Path()
		}
	})
	for _, child := range node.Children() {
		view.liveNode(context, child, depth+1)
	}
}

// transport is the strip over the game: play or pause, a step, starting again, the screen's
// shape, playing in a window, and stopping.
func (view *View) transport(context *ui.Context) {
	editor := view.Editor
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(transportTall), Padding: ui.PadXY(10, 0), Gap: 6, AlignX: ui.Center, AlignY: ui.Center}
	bar(context, layout, func() {
		view.pauseButton(context)
		if context.IconButton(iconStep, "Step") {
			editor.Paused = true
			editor.StepFrame()
		}
		context.Tooltip("Move the game on one frame, paused")
		if context.IconButton(iconRestart, "Restart") {
			editor.RestartPlay()
		}
		context.Tooltip("Start the scene again as it is written")
		if played, playing := editor.PlayingHere(); playing && played.Is2D() {
			view.targetTabs(context) // a 3D game fills the view: its shape is the window's
		}
		view.windowButton(context)
		if context.IconButton(iconStop, "Stop") {
			editor.LeavePlay()
		}
		context.Tooltip("Back to the workspace before (Escape)")
	})
}

// pauseButton pauses the game, or plays it on.
func (view *View) pauseButton(context *ui.Context) {
	editor := view.Editor
	if editor.Paused {
		if context.PrimaryIconButton(iconPlay, "Play") {
			editor.Paused = false
		}
		context.Tooltip("Play on")
		return
	}
	if context.IconButton(iconPause, "Pause") {
		editor.Paused = true
	}
	context.Tooltip("Stop the game where it is, to look at it")
}

// targetTabs choose the shape of the screen the game is shown at.
func (view *View) targetTabs(context *ui.Context) {
	target := int(view.Editor.Target)
	context.Box(ui.Layout{Width: ui.Fixed(190)}, func() {
		if context.Tabs("Screen", &target, "16:9", "3DS") {
			view.Editor.Target = PlayTarget(target)
		}
	})
	context.Tooltip("The shape of the screen: a computer's, or the 3DS's two screens")
}

// windowButton plays the game in a window of its own, or closes it.
func (view *View) windowButton(context *ui.Context) {
	editor := view.Editor
	if target, playing := editor.PlayingOn(); playing && target == ComputerTarget {
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

// stage is the game played, at its target's shape; a 3D game is shown behind the interface.
func (view *View) stage(context *ui.Context) {
	played, playing := view.Editor.PlayingHere()
	if !playing {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16)}, func() {
			note(context, "The scene could not be played: see the line at the bottom.")
		})
		return
	}
	if !played.Is2D() {
		view.stage3D(context, played)
		return
	}
	target := view.Editor.Target
	ink := context.Theme.InkSoft
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
		if target == ConsoleScreen {
			paintConsole(played, area, ink, list)
			return
		}
		paintLetterboxed(played, area, target.Aspect(), list)
	})
}

// The 3DS's screens and the gap of its hinge between them, in the console's pixels.
const (
	consoleTopWidth    = 400
	consoleBottomWidth = 320
	consoleScreenTall  = 240
	consoleHingeTall   = 24
	// consoleCaptionSize is the size of the bottom screen's caption, in the editor's pixels.
	consoleCaptionSize = 14
)

// bottomScreenColor is the bottom screen's, while the game has no interface on it.
var bottomScreenColor = draw2d.Color{0.18, 0.18, 0.2, 1}

// consoleScreens are the 3DS's two screens at their shapes, as large as fit in an area: the top
// one, and the bottom one centred under it.
func consoleScreens(area draw2d.Rect) (top, bottom draw2d.Rect) {
	scale := min(area.Width/consoleTopWidth, area.Height/(2*consoleScreenTall+consoleHingeTall))
	middle := area.X + area.Width/2
	tall := consoleScreenTall * scale
	first := area.Y + (area.Height-(2*consoleScreenTall+consoleHingeTall)*scale)/2
	top = draw2d.Rect{X: middle - consoleTopWidth*scale/2, Y: first, Width: consoleTopWidth * scale, Height: tall}
	bottom = draw2d.Rect{X: middle - consoleBottomWidth*scale/2, Y: first + tall + consoleHingeTall*scale, Width: consoleBottomWidth * scale, Height: tall}
	return top, bottom
}

// paintConsole draws a 2D scene played on the 3DS's top screen, and its bottom screen, which the
// game's interface will have, under it.
func paintConsole(played *engine.Scene, area draw2d.Rect, ink draw2d.Color, list *draw2d.List) {
	list.Rectangle(area, 0, draw2d.Color{letterboxColor, letterboxColor, letterboxColor, 1}, draw2d.Color{}, 0)
	top, bottom := consoleScreens(area)
	paintPlaying(played, top, list)
	list.Rectangle(bottom, 0, bottomScreenColor, ink, 1)
	baseline := bottom.Y + bottom.Height/2 + consoleCaptionSize/3
	list.Text("Bottom screen: the game's interface", bottom.X+consoleCaptionSize, baseline, consoleCaptionSize, draw2d.Regular, ink)
}

// paintLetterboxed draws a 2D scene played at a screen's shape, centred in an area, with bars
// where the shapes differ.
func paintLetterboxed(played *engine.Scene, area draw2d.Rect, aspect float64, list *draw2d.List) {
	list.Rectangle(area, 0, draw2d.Color{letterboxColor, letterboxColor, letterboxColor, 1}, draw2d.Color{}, 0)
	paintPlaying(played, fitted(area, aspect), list)
}

// fitted is the largest rectangle of a shape (width over height) centred in an area.
func fitted(area draw2d.Rect, aspect float64) draw2d.Rect {
	screen := area
	if float64(area.Width)/float64(max(area.Height, 1)) > aspect {
		screen.Width = float32(float64(area.Height) * aspect)
		screen.X += (area.Width - screen.Width) / 2
		return screen
	}
	screen.Height = float32(float64(area.Width) / aspect)
	screen.Y += (area.Height - screen.Height) / 2
	return screen
}

// watchPanel shows what the node chosen is doing while the game plays.
func (view *View) watchPanel(context *ui.Context) {
	context.Row(func() {
		context.IconSized(iconWatch, 16, context.Theme.Primary)
		context.Heading("Watch")
	})
	lines := view.Editor.Watch()
	if len(lines) == 0 {
		note(context, "Choose a node on the left to see where it is, how it moves and what its behaviours hold.")
		return
	}
	for _, line := range lines {
		context.Caption(line.Name)
		context.Label(line.Value)
	}
	context.RequestFrame(0) // what is watched changes every frame
}

// logPanel is what the game and the engine said while playing, newest last.
func (view *View) logPanel(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(playLogHeight), Padding: ui.PadXY(10, 6), Gap: 2}
	bar(context, layout, func() {
		context.Row(func() {
			context.IconSized(iconLog, 16, context.Theme.Primary)
			context.Text("Log", ui.TextStyle{Weight: draw2d.Bold})
		})
		inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 1}
		context.ScrollArea("Log", inside, func() {
			lines := view.Editor.PlayLog()
			if len(lines) == 0 {
				note(context, "Nothing yet: what the game and the engine log (log/slog) shows here.")
			}
			for _, line := range lines {
				context.Text(line, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
			}
		})
	})
}
