package editor

import "github.com/akrck02/go-engine/internal/build"

// The tools of this computer that build for consoles and phones: where the person set each, and
// where it was found. They are looked for when the editor first shows them and when they are
// saved, not every frame: looking runs programs' searches through the path.

// ToolFound is a tool, what needs it, where it was found, and why it was not or why it will not do
// (found, but missing parts).
type ToolFound struct {
	Name, Needs string
	Found       string
	Err         error
}

// buildToolParts are each tool's name, what needs it, where it is set among the Tools, how it is
// found, and how what was found is checked.
var buildToolParts = []struct {
	name, needs string
	folder      bool // its place is a folder, not a program
	place       func(tools *build.Tools) *string
	find        func(tools build.Tools) (string, error)
	check       func(tools build.Tools, found string) error
}{
	{"TinyGo", "Building for the 3DS: TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds), the tinygo program", false,
		func(tools *build.Tools) *string { return &tools.TinyGo }, build.Tools.FindTinyGo, checkFound(build.CheckTinyGo)},
	{"devkitPro", "Building for the 3DS and sending to it with 3dslink: devkitPro's folder", true,
		func(tools *build.Tools) *string { return &tools.DevkitPro }, build.Tools.FindThreeDSLink, checkDevkitPro},
	{"3DS emulator", "Trying the game without a console: Azahar's or Citra's program", false,
		func(tools *build.Tools) *string { return &tools.Nintendo3DSEmulator }, build.Tools.FindNintendo3DSEmulator, nothingToCheck},
	{"Android SDK", "Building for Android and playing through adb: the SDK's folder (Android Studio installs it)", true,
		func(tools *build.Tools) *string { return &tools.AndroidSDK }, build.Tools.FindAndroidSDK, checkFound(build.CheckAndroidSDK)},
	{"Android NDK", "Building for Android: one version's folder, from the SDK Manager", true,
		func(tools *build.Tools) *string { return &tools.AndroidNDK }, findNDK, checkFound(build.CheckAndroidNDK)},
	{"JDK", "Building for Android and installing its SDK: a JDK's folder, 17 or newer (Android Studio brings one)", true,
		func(tools *build.Tools) *string { return &tools.Java }, build.Tools.FindJava, checkFound(build.CheckJava)},
}

// checkFound checks the place found with a check of the build package.
func checkFound(check func(place string) error) func(build.Tools, string) error {
	return func(_ build.Tools, found string) error { return check(found) }
}

// checkDevkitPro checks devkitPro's folder, as 3dslink is what is found.
func checkDevkitPro(tools build.Tools, _ string) error {
	return build.CheckDevkitPro(tools.DevkitProFolder())
}

// nothingToCheck is the check of a tool that is only run: an emulator that is there will do.
func nothingToCheck(build.Tools, string) error { return nil }

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
		if err == nil {
			err = part.check(tools, place)
		}
		found = append(found, ToolFound{Name: part.name, Needs: part.needs, Found: place, Err: err})
	}
	editor.toolsFound = found
}
