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

// ReadAndroidLicence reads Android's licence in the background, for the person to read and accept
// before the SDK is installed (AndroidLicence).
func (editor *Editor) ReadAndroidLicence() {
	if editor.running {
		editor.say(ToolsInUse.Error(), true)
		return
	}
	folder, err := install.DataFolder()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.Run("Reading Android's licence…", func() (string, error) {
		licence, err := install.Installer{Folder: folder}.AndroidLicence()
		if err != nil {
			return "", err
		}
		editor.send(func(editor *Editor) { editor.androidLicence = licence })
		return "Read Android's licence: accept it to install the SDK", nil
	})
}

// AndroidLicence is Android's licence while it waits for the person to accept it or not.
func (editor *Editor) AndroidLicence() (string, bool) {
	return editor.androidLicence, editor.androidLicence != ""
}

// DeclineAndroidLicence leaves the SDK uninstalled.
func (editor *Editor) DeclineAndroidLicence() { editor.androidLicence = "" }

// AcceptAndroidLicence installs Android's SDK, with the NDK and what a game's build needs, the
// licence the person read (AndroidLicence) accepted, and sets the SDK's place. With no licence
// read, nothing is installed.
func (editor *Editor) AcceptAndroidLicence() {
	accepted := editor.androidLicence
	editor.androidLicence = ""
	if accepted == "" {
		editor.say(install.LicenceNotAccepted.Error(), true)
		return
	}
	android := func(installer install.Installer) (string, error) { return installer.AndroidSDK(accepted) }
	editor.installTool("Android SDK", android, func(tools *build.Tools, place string) { tools.AndroidSDK = place })
}

// InstallEmulator installs Azahar's latest release for this system and sets the 3DS emulator's
// place.
func (editor *Editor) InstallEmulator() {
	editor.installTool("3DS emulator", install.Installer.Azahar, func(tools *build.Tools, place string) { tools.ConsoleEmulator = place })
}
