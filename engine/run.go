package engine

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// Settings are a project's, kept in its engine.json.
type Settings struct {
	Name       string `json:"name"`
	StartScene string `json:"startScene"` // the scene the game opens, as "scenes/main.scene.json"
	// CodeEditor is the command the editor opens the project's code with ("code", "zed", "nvim"…);
	// empty tries $VISUAL, $EDITOR and the usual ones.
	CodeEditor string `json:"codeEditor,omitempty"`
}

// SettingsFile is the settings' file in a project.
const SettingsFile = "engine.json"

// ReadSettings reads a project's settings.
func ReadSettings(files fs.FS) (Settings, error) {
	data, err := fs.ReadFile(files, SettingsFile)
	if err != nil {
		return Settings{}, failure.Wrap(SettingsNotFound, "", err)
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, failure.Wrap(SettingsUnreadable, "", err)
	}
	if settings.StartScene == "" {
		return Settings{}, NoStartScene
	}
	return settings, nil
}

// Run plays a project, from its files (embedded in the game, or a folder), until its window is
// closed: on the desktop in a window, on the Nintendo 3DS on the top screen.
func Run(files fs.FS) error {
	if editing() {
		return editor()
	}
	settings, err := ReadSettings(files)
	if err != nil {
		return err
	}
	played, err := Load(files, settings.StartScene)
	if played == nil {
		return err
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // a model that would not load: the rest plays
	}
	for _, problem := range played.Start() {
		fmt.Fprintln(os.Stderr, problem)
	}
	return play(played, settings.Name)
}

// play shows a scene in a window, updating it every frame.
func play(played *Scene, title string) error {
	var last time.Time
	var flat draw2d.List // a 2D scene's drawing, made again every frame
	fallback := space.Transform{Position: space.NewVector(0, 4, 10), Rotation: space.Angles{X: -20}.Rotation(), Scale: space.One}
	options := window.Options{Title: title, Width: 1280, Height: 720, Continuous: true}
	options.Scene = func(state *input.State, step float64) *window.SceneView {
		// the time between frames by the input's clock: a window's step is the display's frame,
		// which a window drawing faster or slower than the display does not keep
		now := state.Time()
		seconds := step
		if !last.IsZero() {
			seconds = min(now.Sub(last).Seconds(), 0.1)
		}
		last = now
		consoleKeys(state)
		played.Update(seconds, state)
		view := &window.SceneView{World: played.World()}
		drawOver(played, state, &flat)
		if camera, found := played.MainCamera(); found {
			view.Camera = camera.View()
		} else {
			view.Camera = CameraAt(fallback, defaultCamera)
		}
		return view
	}
	return window.Run(options, interfaceFor(played))
}

// defaultCamera is how a scene with no camera is seen: in perspective, from a little above.
var defaultCamera = scenefile.Camera{Projection: scenefile.Perspective}
