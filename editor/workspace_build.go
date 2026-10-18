package editor

import (
	"fmt"
	"time"

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

// playLogRefresh is how often the log of a game playing elsewhere is shown again.
const playLogRefresh = 250 * time.Millisecond

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
			view.playButton(context, target)
		})
	})
}

// playButton plays the game where the target is, or stops it there: this computer in a window,
// the 3DS sent to the console or in an emulator, Android on the device adb sees.
func (view *View) playButton(context *ui.Context, target BuildTarget) {
	editor := view.Editor
	if playedOn, playing := editor.PlayingOn(); playing && playedOn == target {
		if context.IconButton(iconStop, "Stop") {
			editor.Stop()
		}
		return
	}
	switch target {
	case ComputerTarget:
		if context.IconButton(iconInWindow, "Play") {
			editor.PlayInWindow()
		}
	case AndroidTarget:
		if context.IconButton(iconPhone, "Play on device") {
			view.build.shown = target
			editor.PlayOnDevice()
		}
		context.Tooltip("Build, install and start the game on the phone or emulator adb sees, with its log here")
	case ConsoleTarget:
		if context.IconButton(iconSend, "Send") {
			view.build.shown = target
			editor.SendToConsole()
		}
		context.Tooltip("Build and send the game to the 3DS: open the Homebrew Launcher and press Y first")
		if context.IconButton(iconInWindow, "Emulator") {
			view.build.shown = target
			editor.PlayInEmulator()
		}
		context.Tooltip("Build the game and open it in a 3DS emulator (Azahar or Citra)")
	}
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
	if !found && view.playOutput(target) == nil {
		note(context, "Build it to see what building says.")
		return
	}
	for _, line := range append(result.Log, view.playOutput(target)...) {
		context.Text(line, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
	}
	if playedOn, playing := view.Editor.PlayingOn(); playing && playedOn == target {
		context.RequestFrame(playLogRefresh) // the game's log grows while it plays
	}
}

// playOutput is what the game said, when it plays or last played on a target.
func (view *View) playOutput(target BuildTarget) []string {
	editor := view.Editor
	if editor.output == nil || editor.playedOn != target {
		return nil
	}
	return editor.output.Lines()
}
