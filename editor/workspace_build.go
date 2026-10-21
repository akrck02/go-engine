package editor

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/install"
	"github.com/akrck02/go-engine/project"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

// The Build workspace: two tabs, one with a card for each place the game runs (its last build and
// what can be done), one with the tools of this computer that build for them; the chosen target's
// log beside both, kept until its next build.

// The Build workspace's measures, in pixels.
const (
	targetCardWidth = 340
	targetCardGap   = 12
	toolsWidth      = 2*targetCardWidth + targetCardGap // under two cards
	cardIconSize    = 22
	stateIconSize   = 16
	buildLogWidth   = 460
	buildTabsWidth  = 240
	bytesInMegabyte = 1 << 20
)

// playLogRefresh is how often the log of a game playing elsewhere is shown again.
const playLogRefresh = 250 * time.Millisecond

// buildTab is a tab of the Build workspace.
type buildTab int

// The Build workspace's tabs, in the order they are shown.
const (
	platformsTab buildTab = iota
	toolsTab
)

// buildView is what the Build workspace remembers between frames.
type buildView struct {
	tab         buildTab       // the tab shown
	shown       project.Target // the target whose log is shown
	tools       build.Tools    // the tools' places as typed, kept when saved
	toolsFilled bool           // tools was filled with the places kept, to be typed over
	toolsKept   int            // the editor's toolsKept when tools was filled: an install keeps new ones
}

func (view *View) buildWorkspace(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Targets", ui.Grow(), func() {
			context.Heading("Build the game")
			note(context, "Every build leaves the editor out of the game (-tags release) and writes, beside the program, the notices of the code it carries.")
			view.buildTabs(context)
			switch view.build.tab {
			case platformsTab:
				view.targetCards(context)
			case toolsTab:
				view.toolsSection(context)
			}
			view.licenceSheet(context)
		})
		context.Separator()
		panel(context, "Build log", ui.Fixed(buildLogWidth), func() { view.buildLog(context) })
	})
}

// buildTabs choose between the places the game is built for and the tools that build it.
func (view *View) buildTabs(context *ui.Context) {
	tab := int(view.build.tab)
	context.Box(ui.Layout{Width: ui.Fixed(buildTabsWidth)}, func() {
		if context.Tabs("Build tab", &tab, "Platforms", "Tools") {
			view.build.tab = buildTab(tab)
		}
	})
}

// targetCards are a card for each place the game runs.
func (view *View) targetCards(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: targetCardGap, Wrap: true}, func() {
		for _, target := range project.Targets {
			view.targetCard(context, target)
		}
	})
}

// targetCard is a target with its last build, the button that builds it, and what else it can do.
func (view *View) targetCard(context *ui.Context, target project.Target) {
	look := ui.Look{Fill: context.Theme.Background, Radius: 10, Border: context.Theme.Line, BorderWidth: 1}
	if view.build.shown == target {
		look.Border, look.BorderWidth = context.Theme.Primary, 2
	}
	context.Push(targetName(target))
	defer context.Pop()
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(targetCardWidth), Padding: ui.Pad(14), Gap: 6}, look, func() {
		context.Row(func() {
			context.IconSized(targetIcon(target), cardIconSize, context.Theme.Primary)
			context.Heading(targetName(target))
		})
		context.Caption(targetMakes(target))
		view.buildState(context, target)
		context.Row(func() {
			if context.PrimaryIconButton(iconBuild, "Build") {
				view.build.shown = target
				view.Editor.BuildFor(target)
			}
			if context.IconButton(iconLog, "Log") {
				view.build.shown = target
			}
		})
		context.Row(func() { view.playButton(context, target) }) // where to try what was built
	})
}

// playButton plays the game where the target is, or stops it there: this computer in a window,
// the 3DS sent to the console or in an emulator, Android on the device adb sees.
func (view *View) playButton(context *ui.Context, target project.Target) {
	editor := view.Editor
	if playedOn, playing := editor.PlayingOn(); playing && playedOn == target {
		if context.IconButton(iconStop, "Stop") {
			editor.Stop()
		}
		return
	}
	switch target {
	case project.Desktop:
		if context.IconButton(iconInWindow, "Play") {
			editor.PlayOn(project.Desktop)
		}
		if context.IconButton(iconRelease, "Release workflow") {
			editor.WriteReleaseWorkflow()
		}
		context.Tooltip("Write a GitHub workflow that builds the game for Windows, macOS and Linux on each version tag (v1.0.0)")
	case project.Android:
		if context.IconButton(iconPhone, "Play on device") {
			view.build.shown = target
			editor.PlayOn(project.Android)
		}
		context.Tooltip("Build, install and start the game on the phone or emulator adb sees, with its log here")
	case project.Nintendo3DS:
		if context.IconButton(iconSend, "Send") {
			view.build.shown = target
			editor.PlayOn(project.Nintendo3DS)
		}
		context.Tooltip("Build and send the game to the 3DS: open the Homebrew Launcher and press Y first")
		if context.IconButton(iconInWindow, "Emulator") {
			view.build.shown = target
			editor.PlayInEmulator(project.Nintendo3DS)
		}
		context.Tooltip("Build the game and open it in a 3DS emulator (Azahar or Citra)")
	}
}

// buildState says how the target's last build went.
func (view *View) buildState(context *ui.Context, target project.Target) {
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
		context.Text(targetName(target)+" · log", ui.TextStyle{Weight: draw2d.Bold})
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
func (view *View) playOutput(target project.Target) []string {
	editor := view.Editor
	if editor.output == nil || editor.playedOn != target {
		return nil
	}
	return editor.output.Lines()
}

// toolsSection is where the tools that build for consoles and phones are: set by the person, and
// found or not.
func (view *View) toolsSection(context *ui.Context) {
	editor := view.Editor
	if !view.build.toolsFilled || view.build.toolsKept != editor.toolsKept {
		view.build.tools, view.build.toolsFilled, view.build.toolsKept = editor.Project.Tools, true, editor.toolsKept
	}
	look := ui.Look{Fill: context.Theme.Background, Radius: 10, Border: context.Theme.Line, BorderWidth: 1}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(toolsWidth), Padding: ui.Pad(14), Gap: 6}, look, func() {
		context.Row(func() {
			context.IconSized(iconTools, cardIconSize, context.Theme.Primary)
			context.Heading("Tools on this computer")
		})
		note(context, "Kept for every project on this computer, not in the project. An empty place looks for the tool in its variable and where it is usually installed.")
		if unread := editor.Project.ToolsUnread; unread != nil {
			wrapped(context, "The places kept could not be read, and saving writes them anew: "+explained(unread), ui.TextStyle{Color: context.Theme.Danger, Size: context.Theme.SmallTextSize})
		}
		for index, found := range editor.ToolsFound() {
			view.toolRow(context, index, found)
		}
		context.Row(func() {
			if context.PrimaryIconButton(iconSave, "Save tools") {
				view.saveTools()
			}
			if context.IconButton(iconRestart, "Check") {
				editor.LookForTools(view.build.tools)
			}
			context.Tooltip("Look for the tools where the places typed say, before saving them, or again after installing one")
		})
	})
}

// toolRow is a tool: what needs it, the place typed, and where it was found or why not.
func (view *View) toolRow(context *ui.Context, index int, found ToolFound) {
	context.Push(found.Name)
	defer context.Pop()
	context.Separator()
	context.Row(func() {
		if found.Err != nil {
			context.IconSized(iconMissing, stateIconSize, context.Theme.Danger)
		} else {
			context.IconSized(iconFound, stateIconSize, context.Theme.Positive)
		}
		context.Text(found.Name, ui.TextStyle{Weight: draw2d.Bold})
	})
	context.Caption(found.Needs)
	place := buildToolParts[index].place(&view.build.tools)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.TextField(found.Name, place) })
		if context.IconButton(iconBrowse, "Browse") {
			view.browseForTool(index, found)
		}
		if install, offered := toolInstalls[found.Name]; offered && context.IconButton(iconInstall, "Install") {
			install(view)
		}
	})
	view.installProgress(context, found.Name)
	small := ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize}
	if found.Found != "" {
		wrapped(context, "Found: "+found.Found, small)
	}
	if found.Err != nil {
		small.Color = context.Theme.Danger
		wrapped(context, explained(found.Err), small)
	}
}

// saveTools keeps the tools' places as typed, and says how it went.
func (view *View) saveTools() {
	if err := view.Editor.SaveTools(view.build.tools); err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("Saved the tools' places, for every project on this computer", false)
}

// explained is an error with its code's meaning: a code wrapped with details says only its name.
func explained(err error) string {
	var failed *failure.Error
	if errors.As(err, &failed) && failed.Details != "" {
		return failed.Code.Error() + " (" + failed.Details + ")"
	}
	return err.Error()
}

// toolPlaces are the folders the tools are usually installed in, offered in the file picker.
var toolPlaces = []string{"/opt", "/usr/local", "/usr/bin"}

// browseForTool asks for a tool's place with the file picker, starting where it was found; the
// place chosen is typed in its field, to check and save.
func (view *View) browseForTool(index int, found ToolFound) {
	part := buildToolParts[index]
	place := part.place(&view.build.tools)
	start := *place
	if start == "" {
		start = found.Found
	}
	options := ui.FilePickerOptions{Title: "Choose " + part.name, Folder: part.folder, Start: start, Places: toolPlaces}
	view.pickFile(options, func(path string) {
		*part.place(&view.build.tools) = path
		view.Editor.LookForTools(view.build.tools)
	})
}

// toolInstalls are how each tool that can be installed from the editor is, by its name.
var toolInstalls = map[string]func(view *View){
	"TinyGo":       func(view *View) { view.Editor.InstallTinyGo() },
	"devkitPro":    (*View).installDevkitPro,
	"3DS emulator": func(view *View) { view.Editor.InstallEmulator() },
	"Android SDK":  func(view *View) { view.Editor.ReadLicence("Android SDK") },
	"Android NDK":  func(view *View) { view.Editor.ReadLicence("Android SDK") }, // installed with the SDK
	"JDK":          (*View).installJava,
}

// The licence sheet's measures, in pixels.
const (
	licenceWidth  = 640
	licenceHeight = 420
)

// licenceSheet shows a tool's licence while it waits: accepting it installs the tool.
func (view *View) licenceSheet(context *ui.Context) {
	editor := view.Editor
	licence, waiting := editor.Licence()
	if !waiting {
		return
	}
	accepted := false
	if context.Sheet(licence.Tool+" licence", ui.SheetOptions{Width: ui.Fixed(licenceWidth)}, func() {
		note(context, "Installing "+licence.Tool+" means accepting its licence:")
		inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(licenceHeight), Gap: 8}
		context.ScrollArea("Licence", inside, func() {
			for _, paragraph := range strings.Split(licence.Text, "\n\n") {
				wrapped(context, paragraph, ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize})
			}
		})
		context.Row(func() {
			if context.Button("Decline") {
				editor.DeclineLicence()
			}
			accepted = context.PrimaryButton("Accept and install")
		})
	}) {
		editor.DeclineLicence()
	}
	if accepted {
		editor.AcceptLicence()
	}
}

// installDevkitPro runs devkitPro's own installer for the 3DS's packages in a terminal, where the
// person types their password; without it, it opens devkitPro's guide to get it.
func (view *View) installDevkitPro() {
	command, found := install.DevkitProCommand()
	if !found {
		window.OpenLink(install.DevkitProGuide)
		view.Editor.say("devkitPro's installer is not here yet: follow its guide, then press Install again", false)
		return
	}
	if err := project.RunInTerminal(command); err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("devkitPro installs in the terminal (it asks for your password): press Check when it ends", false)
}

// installJava opens the page of a JDK's installers: a JDK is the system's to install (and to keep
// up to date), not the editor's.
func (view *View) installJava() {
	window.OpenLink(javaDownloads)
	view.Editor.say("Install a JDK, 17 or newer, from the page opened: then press Check", false)
}

// javaDownloads is the page of Adoptium's Temurin, a free JDK for every system.
const javaDownloads = "https://adoptium.net/temurin/releases/"

// The progress bar's measures, in pixels.
const (
	progressHeight = 6
	progressRadius = 3
)

// installProgress shows how the install of a tool goes, while it is installed.
func (view *View) installProgress(context *ui.Context, name string) {
	tool, done, total, busy := view.Editor.Installing()
	if !busy || tool != name {
		return
	}
	small := ui.TextStyle{Color: context.Theme.InkSoft, Size: context.Theme.SmallTextSize}
	if total <= 0 {
		context.Text(fmt.Sprintf("Installing… %.1f MB", float64(done)/bytesInMegabyte), small)
		return
	}
	share := float32(done) / float32(total)
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(progressHeight)}, func(area draw2d.Rect, list *draw2d.List) {
		list.Rectangle(area, progressRadius, context.Theme.Line, draw2d.Color{}, 0)
		filled := area
		filled.Width *= share
		list.Rectangle(filled, progressRadius, context.Theme.Primary, draw2d.Color{}, 0)
	})
	context.Text(fmt.Sprintf("Installing… %.0f of %.0f MB", float64(done)/bytesInMegabyte, float64(total)/bytesInMegabyte), small)
}
