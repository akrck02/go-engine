package editor

import "github.com/akrck02/go-engine/failure"

// Error codes of the editor package.
var (
	RestartFailed = failure.Define("editor.restart-failed", "the editor could not start again with the project's new code")
	FramesWrong   = failure.Define("editor.frames-wrong", "an animation's frames are not numbers and runs of the sheet's frames, as 0-3, 6")
)
