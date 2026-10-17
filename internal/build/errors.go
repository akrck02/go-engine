package build

import "github.com/akrck02/go-engine/failure"

// Error codes of the build package.
var (
	SourceNotFound = failure.Define("build.source-not-found", "go-engine's source, which programs build against, is not where the editor was built")
	ToolFailed     = failure.Define("build.tool-failed", "a Go tool failed")
	TinyGoNotFound = failure.Define("build.tinygo-not-found", "building for the 3DS needs TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds): set $TINYGO")
)
