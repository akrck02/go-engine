package editor

import "github.com/akrck02/go-engine/failure"

// Error codes of the editor package.
var (
	RestartFailed  = failure.Define("editor.restart-failed", "the editor could not start again with the project's new code")
	FramesWrong    = failure.Define("editor.frames-wrong", "an animation's frames are not numbers and runs of the sheet's frames, as 0-3, 6")
	ArtNotFound    = failure.Define("editor.art-not-found", "the sprite or tileset is not a file of the project's")
	ArtNotSaved    = failure.Define("editor.art-not-saved", "the sprite or tileset could not be written")
	ArtWrongKind   = failure.Define("editor.art-wrong-kind", "the file is not a sprite's or a tileset's")
	ArtUnsaved     = failure.Define("editor.art-unsaved", "the sprite or tileset open in Art has changes not saved: save it first")
	NoOwnSheet     = failure.Define("editor.no-own-sheet", "the node does not cut its picture into frames itself")
	UnsavedChanges = failure.Define("editor.unsaved-changes", "the scene has changes not saved: save it first (Ctrl+S)")
)
