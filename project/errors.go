package project

import "github.com/akrck02/go-engine/failure"

// Error codes of the project package.
var (
	NotAProject        = failure.Define("project.not-a-project", "the folder is not a go-engine project; make one with: go-engine new <folder>")
	FolderNotEmpty     = failure.Define("project.folder-not-empty", "a new project needs an empty folder")
	SceneExists        = failure.Define("project.scene-exists", "a scene of that name exists already")
	ImportFormat       = failure.Define("project.import-format", "only .glb models can be imported for now (export glTF as binary .glb)")
	ImportFailed       = failure.Define("project.import-failed", "the file could not be copied into the project")
	ZedNotFound        = failure.Define("project.zed-not-found", "Zed was not found (zed, zeditor): install it from zed.dev, or open the project's folder yourself")
	CodeEditorNotFound = failure.Define("project.code-editor-not-found", "no code editor found: set codeEditor in engine.json (\"code\", \"zed\", \"nvim\"…) or $VISUAL")
	TerminalNotFound   = failure.Define("project.terminal-not-found", "the code editor runs in a terminal, and no terminal was found")
)
