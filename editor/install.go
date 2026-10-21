package editor

import (
	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/install"
)

// Installing a tool from the Build workspace: it runs in the background as a build does (one
// thing at a time), shows how much of its download has arrived, and when it ends the tool's place
// is set and kept, as if the person had chosen it.

// installing is the install going on: the tool, and how much of its download has arrived.
type installing struct {
	tool        string
	done, total int64
}

// percentStep is how much a download must move on, in hundredths, before the editor hears it.
const percentStep = 1

// Installing tells which tool is being installed, and how much of its download has arrived
// (total is 0 when unknown).
func (editor *Editor) Installing() (tool string, done, total int64, busy bool) {
	if editor.installing == nil {
		return "", 0, 0, false
	}
	return editor.installing.tool, editor.installing.done, editor.installing.total, true
}

// InstallTinyGo installs tinygo-3ds's release for this system and sets TinyGo's place.
func (editor *Editor) InstallTinyGo() {
	editor.installTool("TinyGo", install.Installer.TinyGo, func(tools *build.Tools, place string) { tools.TinyGo = place })
}

// installTool installs a tool in the background into go-engine's data folder; set puts the place
// it returns among the tools, which are then kept.
func (editor *Editor) installTool(name string, run func(installer install.Installer) (string, error), set func(tools *build.Tools, place string)) {
	if editor.running {
		editor.say(ToolsInUse.Error(), true)
		return
	}
	folder, err := install.DataFolder()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.installing = &installing{tool: name}
	installer := install.Installer{Folder: folder, Progress: editor.downloadProgress()}
	editor.Run("Installing "+name+"…", func() (string, error) {
		place, err := run(installer)
		editor.send(func(editor *Editor) { editor.installEnded(set, place, err) })
		if err != nil {
			return "", err
		}
		return "Installed " + name + ": " + place, nil
	})
}

// downloadProgress is a Progress that tells the editor how a download goes, a hundredth at a time:
// every chunk would flood it.
func (editor *Editor) downloadProgress() install.Progress {
	lastPercent := int64(-percentStep)
	return func(done, total int64) {
		percent := int64(0)
		if total > 0 {
			percent = done * 100 / total
		}
		if total > 0 && percent < lastPercent+percentStep && done != total {
			return
		}
		lastPercent = percent
		editor.send(func(editor *Editor) {
			if editor.installing != nil {
				editor.installing.done, editor.installing.total = done, total
			}
		})
	}
}

// installEnded keeps the place of a tool installed, and looks for the tools again.
func (editor *Editor) installEnded(set func(tools *build.Tools, place string), place string, err error) {
	editor.installing = nil
	if err != nil {
		return
	}
	tools := editor.Project.Tools
	set(&tools, place)
	if err := tools.Save(); err != nil {
		editor.say(err.Error(), true)
	}
	editor.Project.Tools, editor.Project.ToolsUnread = tools, nil
	editor.toolsKept++
	editor.LookForTools(tools)
}

// Some tools are installed only once the person has read and accepted their licence: the
// licence is read first, shown, and the tool installed with the text accepted.

// licensedTool is how a tool with a licence is installed: getting ready (what it needs to run),
// reading its licence, and installing it with the licence accepted.
type licensedTool struct {
	ready   func(editor *Editor) bool
	read    func(installer install.Installer) (string, error)
	install func(editor *Editor, accepted string)
}

// licensedTools are the tools with a licence, by their names among the tools.
var licensedTools = map[string]licensedTool{
	"Android SDK": {
		ready: func(editor *Editor) bool { _, found := editor.javaForAndroid(); return found },
		read:  install.Installer.AndroidLicence,
		install: func(editor *Editor, accepted string) {
			java, found := editor.javaForAndroid()
			if !found {
				return // javaForAndroid has said why
			}
			android := func(installer install.Installer) (string, error) { return installer.AndroidSDK(accepted, java) }
			editor.installTool("Android SDK", android, func(tools *build.Tools, place string) { tools.AndroidSDK = place })
		},
	},
}

// PendingLicence is a tool's licence, read and waiting for the person to accept it or not.
type PendingLicence struct {
	Tool, Text string
}

// ReadLicence reads a tool's licence in the background, for the person to read and accept before
// the tool is installed (Licence). A tool without one is refused.
func (editor *Editor) ReadLicence(tool string) {
	licensed, found := licensedTools[tool]
	if !found {
		editor.say(tool+": "+NoLicence.Error(), true)
		return
	}
	if editor.running {
		editor.say(ToolsInUse.Error(), true)
		return
	}
	if !licensed.ready(editor) {
		return // no need to read a licence for a tool that could not run
	}
	folder, err := install.DataFolder()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.Run("Reading the licence of "+tool+"…", func() (string, error) {
		text, err := licensed.read(install.Installer{Folder: folder})
		if err != nil {
			return "", err
		}
		editor.send(func(editor *Editor) { editor.licence = PendingLicence{Tool: tool, Text: text} })
		return "Read the licence of " + tool + ": accept it to install it", nil
	})
}

// Licence is the licence waiting for the person to accept it or not, if one is.
func (editor *Editor) Licence() (PendingLicence, bool) {
	return editor.licence, editor.licence.Text != ""
}

// DeclineLicence leaves the tool whose licence waits uninstalled.
func (editor *Editor) DeclineLicence() { editor.licence = PendingLicence{} }

// AcceptLicence installs the tool whose licence the person read (Licence), with that licence
// accepted, and sets its place. With no licence read, nothing is installed.
func (editor *Editor) AcceptLicence() {
	accepted := editor.licence
	editor.licence = PendingLicence{}
	licensed, found := licensedTools[accepted.Tool]
	if accepted.Text == "" || !found {
		editor.say(install.LicenceNotAccepted.Error(), true)
		return
	}
	licensed.install(editor, accepted.Text)
}

// javaForAndroid is the JDK Android's sdkmanager runs on, as the tools kept say; it says why when
// there is none.
func (editor *Editor) javaForAndroid() (string, bool) {
	java, err := editor.Project.Tools.FindJava()
	if err != nil {
		editor.say(err.Error(), true)
		return "", false
	}
	return java, true
}

// InstallEmulator installs Azahar's latest release for this system and sets the 3DS emulator's
// place.
func (editor *Editor) InstallEmulator() {
	editor.installTool("3DS emulator", install.Installer.Azahar, func(tools *build.Tools, place string) { tools.Nintendo3DSEmulator = place })
}
