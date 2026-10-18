package editor

import "github.com/akrck02/go-engine/failure"

// Error codes of the editor package.
var (
	RestartFailed = failure.Define("editor.restart-failed", "the editor could not start again with the project's new code")
	FramesWrong   = failure.Define("editor.frames-wrong", "an animation's frames are not numbers and runs of the sheet's frames, as 0-3, 6")
	ArtNotFound   = failure.Define("editor.art-not-found", "the sprite or tileset is not a file of the project's")
	ArtNotSaved   = failure.Define("editor.art-not-saved", "the sprite or tileset could not be written")
)
