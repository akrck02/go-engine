package engine

import "os"

// The editor runs inside the game: a project's editor.go installs it, and go-engine starts the
// game with EditorVariable set, so Run opens the editor instead of playing. That way the editor
// plays the game's own code in its view, since the game's behaviours are compiled into it. Builds
// for the 3DS and finished builds (the tag release) leave editor.go out.

// EditorVariable is set, in the environment, when go-engine starts a game to edit it.
const EditorVariable = "GO_ENGINE_EDITOR"

var editor func() error

// SetEditor installs the editor (the editor package does, from a project's editor.go).
func SetEditor(run func() error) { editor = run }

// editing tells whether the game was started to be edited, with an editor installed.
func editing() bool { return editor != nil && os.Getenv(EditorVariable) != "" }
