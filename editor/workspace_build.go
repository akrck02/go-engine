package editor

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// The Build workspace: one card for each place the game runs, with its last build and what can be
// done; the chosen target's log beside them, kept until its next build.

// The Build workspace's measures, in pixels.
const (
	targetCardWidth = 340
	buildLogWidth   = 460
	bytesInMegabyte = 1 << 20
)

// buildView is what the Build workspace remembers between frames.
type buildView struct {
	shown BuildTarget // the target whose log is shown
}

func (view *View) buildWorkspace(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Targets", ui.Grow(), func() {
			context.Heading("Build the game")
			note(context, "Every build leaves the editor out of the game (-tags release) and writes, beside the program, the notices of the code it carries.")
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, Wrap: true}, func() {
				for _, target := range BuildTargets {
					view.targetCard(context, target)
				}
			})
		})
		context.Separator()
		panel(context, "Build log", ui.Fixed(buildLogWidth), func() { view.buildLog(context) })
	})
}

// targetCard is a target with its last build, the button that builds it, and what else it can do.
func (view *View) targetCard(context *ui.Context, target BuildTarget) {
	look := ui.Look{Fill: context.Theme.Background, Radius: 10, Border: context.Theme.Line, BorderWidth: 1}
	if view.build.shown == target {
		look.Border, look.BorderWidth = context.Theme.Primary, 2
	}
	context.Push(target.Name())
	defer context.Pop()
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(targetCardWidth), Padding: ui.Pad(14), Gap: 6}, look, func() {
		context.Row(func() {
			context.IconSized(target.Icon(), 22, context.Theme.Primary)
			context.Heading(target.Name())
		})
		context.Caption(target.Makes())
		view.buildState(context, target)
		context.Row(func() {
			if context.PrimaryIconButton(iconBuild, "Build") {
				view.build.shown = target
				view.Editor.BuildFor(target)
			}
			if context.IconButton(iconLog, "Log") {
				view.build.shown = target
			}
			if target == ComputerTarget && context.IconButton(iconInWindow, "Play") {
				view.Editor.PlayInWindow()
			}
		})
	})
}

// buildState says how the target's last build went.
func (view *View) buildState(context *ui.Context, target BuildTarget) {
	result, found := view.Editor.LastBuild(target)
	switch {
	case !found:
		context.Text("Not built in this session", ui.TextStyle{Color: context.Theme.InkSoft})
	case result.Err != nil:
		context.Text("Failed at "+result.At.Format("15:04")+": see its log", ui.TextStyle{Color: context.Theme.Danger})
	default:
		size := fmt.Sprintf("%.1f MB", float64(result.Bytes)/bytesInMegabyte)
		context.Text("Built at "+result.At.Format("15:04")+" · "+size, ui.TextStyle{Color: context.Theme.Positive})
		context.Caption(result.Program)
	}
}

// buildLog is what the shown target's last build said.
func (view *View) buildLog(context *ui.Context) {
	target := view.build.shown
	context.Row(func() {
		context.IconSized(iconLog, 16, context.Theme.Primary)
		context.Text(target.Name()+" · log", ui.TextStyle{Weight: draw2d.Bold})
	})
	result, found := view.Editor.LastBuild(target)
	if !found {
		note(context, "Build it to see what building says.")
		return
	}
	for _, line := range result.Log {
		context.Text(line, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
	}
}
