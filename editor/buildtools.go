package editor

import "github.com/akrck02/go-engine/internal/build"

// The tools of this computer that build for consoles and phones: where the person set each, and
// where it was found. They are looked for when the editor first shows them and when they are
// saved, not every frame: looking runs programs' searches through the path.

// ToolFound is a tool, what needs it, and where it was found or why it was not.
type ToolFound struct {
	Name, Needs string
	Found       string
	Err         error
}

// buildToolParts are each tool's name, what needs it, where it is set among the Tools, and how it is
// found.
var buildToolParts = []struct {
	name, needs string
	folder      bool // its place is a folder, not a program
	place       func(tools *build.Tools) *string
	find        func(tools build.Tools) (string, error)
}{
	{"TinyGo", "Building for the 3DS: TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds), the tinygo program", false,
		func(tools *build.Tools) *string { return &tools.TinyGo }, build.Tools.FindTinyGo},
	{"devkitPro", "Sending to the 3DS with 3dslink: devkitPro's folder", true,
		func(tools *build.Tools) *string { return &tools.DevkitPro }, build.Tools.FindThreeDSLink},
	{"3DS emulator", "Trying the game without a console: Azahar's or Citra's program", false,
		func(tools *build.Tools) *string { return &tools.ConsoleEmulator }, build.Tools.FindConsoleEmulator},
	{"Android SDK", "Building for Android and playing through adb: the SDK's folder (Android Studio installs it)", true,
		func(tools *build.Tools) *string { return &tools.AndroidSDK }, build.Tools.FindAndroidSDK},
	{"Android NDK", "Building for Android: one version's folder, from the SDK Manager", true,
		func(tools *build.Tools) *string { return &tools.AndroidNDK }, findNDK},
}

// findNDK finds the NDK, in the SDK found when none is set.
func findNDK(tools build.Tools) (string, error) {
	sdk, _ := tools.FindAndroidSDK() // without an SDK the NDK can still be set or in its variable
	return tools.FindAndroidNDK(sdk)
}

// ToolsFound are the tools and where they were found, in the order of buildToolParts: where the
// tools kept say, until LookForTools looks where others say.
func (editor *Editor) ToolsFound() []ToolFound {
	if editor.toolsFound == nil {
		editor.LookForTools(editor.Project.Tools)
	}
	return editor.toolsFound
}

// SaveTools keeps the tools' places on this computer, for the builds and every project, and looks
// for them again. Not while something builds or plays: it reads them.
func (editor *Editor) SaveTools(tools build.Tools) error {
	if editor.running {
		return ToolsInUse
	}
	if err := tools.Save(); err != nil {
		return err
	}
	editor.Project.Tools, editor.Project.ToolsUnread = tools, nil
	editor.LookForTools(tools)
	return nil
}

// LookForTools looks for every tool where some places say (the ones typed, before saving them).
func (editor *Editor) LookForTools(tools build.Tools) {
	found := make([]ToolFound, 0, len(buildToolParts))
	for _, part := range buildToolParts {
		place, err := part.find(tools)
		found = append(found, ToolFound{Name: part.name, Needs: part.needs, Found: place, Err: err})
	}
	editor.toolsFound = found
}
