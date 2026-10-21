package project

import "github.com/akrck02/go-engine/failure"

// Error codes of the project package.
var (
	NotAProject                 = failure.Define("project.not-a-project", "the folder is not a go-engine project; make one with: go-engine new <folder>")
	FolderNotEmpty              = failure.Define("project.folder-not-empty", "a new project needs an empty folder")
	SceneExists                 = failure.Define("project.scene-exists", "a scene of that name exists already")
	ImportFormat                = failure.Define("project.import-format", "only .glb models can be imported for now (export glTF as binary .glb)")
	ImportFailed                = failure.Define("project.import-failed", "the file could not be copied into the project")
	CodeEditorNotFound          = failure.Define("project.code-editor-not-found", "no code editor found: set codeEditor in engine.json (\"code\", \"zed\", \"nvim\"…) or $VISUAL")
	TerminalNotFound            = failure.Define("project.terminal-not-found", "the code editor runs in a terminal, and no terminal was found")
	SettingsNotSaved            = failure.Define("project.settings-not-saved", "engine.json could not be written")
	ApplicationIDWrong          = failure.Define("project.application-id-wrong", "an application id is two words or more joined by dots, each starting with a letter, as io.example.game")
	TargetCannot                = failure.Define("project.target-cannot", "the target cannot do that")
	EngineNotRead               = failure.Define("project.engine-not-read", "the version of go-engine the project requires could not be read from its go.mod")
	ReleaseNeedsPublishedEngine = failure.Define("project.release-needs-published-engine", "the game's go.mod uses go-engine from a folder of this computer, which GitHub's machines do not have: update it first (go-engine -update)")
	EngineNotUpdated            = failure.Define("project.engine-not-updated", "go-engine could not be updated: check the network, and that the version exists")
	NoAndroidDevice             = failure.Define("project.no-android-device", "adb sees no Android device: plug a phone in with USB debugging on, or start the emulator")
)
