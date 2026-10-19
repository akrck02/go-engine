package engine

import "log/slog"

// A game moves from scene to scene: a behaviour asks to start the scene again or to change to
// another (Context.RestartScene, Context.ChangeScene), and whoever runs the scene — the game's
// window, the editor's view — asks it for its replacement after each frame. The change waits for
// the frame's end, so nothing changes under the behaviours still updating.

// RestartScene asks for the scene to start again as it was written, at the end of the frame.
func (context *Context) RestartScene() { context.Scene.restartRequested = true }

// ChangeScene asks for another scene of the project, by its file, at the end of the frame.
func (context *Context) ChangeScene(file string) { context.Scene.changeRequested = file }

// TakeReplacement builds and starts the scene a behaviour asked for in the last frame, and
// forgets the asking; replaced is false when nothing was asked. The new scene shares this one's
// pictures and models, already read. A scene that cannot be made gives its error, and the game
// goes on in this one.
func (owner *Scene) TakeReplacement() (next *Scene, replaced bool, err error) {
	if owner.restartRequested {
		owner.restartRequested = false
		if owner.description == nil {
			return nil, false, SceneNotRestartable
		}
		return startedReplacement(buildWith(owner.assets, owner.description))
	}
	if file := owner.changeRequested; file != "" {
		owner.changeRequested = ""
		return startedReplacement(loadWith(owner.assets, file))
	}
	return nil, false, nil
}

// startedReplacement starts a scene just made, telling what it could not make of it.
func startedReplacement(next *Scene, err error) (*Scene, bool, error) {
	if next == nil {
		return nil, false, err
	}
	if err != nil {
		slog.Warn("the next scene is missing a part", "error", err)
	}
	for _, problem := range next.Start() {
		slog.Warn("a behaviour of the next scene could not be made", "error", problem)
	}
	return next, true, nil
}
