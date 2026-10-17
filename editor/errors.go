package editor

import "github.com/akrck02/go-engine/failure"

// Error codes of the editor package.
var (
	RestartFailed = failure.Define("editor.restart-failed", "the editor could not start again with the project's new code")
)
