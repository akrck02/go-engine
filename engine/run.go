package engine

import (
	"encoding/json"
	"io/fs"
	"log/slog"
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
		slog.Warn("the scene is missing a part; the rest plays", "error", err)
	}
	for _, problem := range played.Start() {
		slog.Warn("a behaviour could not be made", "error", problem)
	}
	return (&runner{played: played}).run(settings.Name)
}

// The game's window.
const (
	windowWidth  = 1280
	windowHeight = 720
	// longestFrame is the most seconds a frame moves the game on: after a pause (a window dragged,
	// a debugger) the game goes on from where it was instead of leaping.
	longestFrame = 0.1
)

// fallbackCamera is where a scene with no camera is seen from: four metres up and ten back,
// looking a little down at the origin.
var fallbackCamera = space.Transform{Position: space.NewVector(0, 4, 10), Rotation: space.Angles{X: -20}.Rotation(), Scale: space.One}

// runner plays scenes in a window, one after another, as the game asks.
type runner struct {
	played  *Scene
	last    time.Time
	flat    draw2d.List      // a 2D scene's drawing on the console, made again every frame
	console consoleInput     // the console's buttons, as keys
	view    window.SceneView // what the window shows this frame
}

// run opens the window and plays until it is closed.
func (game *runner) run(title string) error {
	options := window.Options{Title: title, Width: windowWidth, Height: windowHeight, Continuous: true}
	options.Scene = game.step
	return window.Run(options, interfaceFor(game.current))
}

// current is the scene playing now.
func (game *runner) current() *Scene { return game.played }

// step moves the game on a frame and gives the window what to show.
func (game *runner) step(state *input.State, step float64) *window.SceneView {
	seconds := game.secondsSince(state, step)
	keys := consoleKeys(&game.console, state)
	game.played.Update(seconds, keys)
	game.replaceIfAsked()
	drawOver(game.played, state, &game.flat)
	game.view = window.SceneView{World: game.played.World(), Camera: CameraAt(fallbackCamera, defaultCamera)}
	if camera, found := game.played.MainCamera(); found {
		game.view.Camera = camera.View()
	}
	return &game.view
}

// secondsSince is the time since the last frame by the input's clock: a window's step is the
// display's frame, which a window drawing faster or slower than the display does not keep.
func (game *runner) secondsSince(state *input.State, step float64) float64 {
	now := state.Time()
	seconds := step
	if !game.last.IsZero() {
		seconds = min(now.Sub(game.last).Seconds(), longestFrame)
	}
	game.last = now
	return seconds
}

// replaceIfAsked goes on to the scene a behaviour asked for, if it can be read.
func (game *runner) replaceIfAsked() {
	next, replaced, err := game.played.Replacement()
	if err != nil {
		slog.Error("the scene asked for could not be read; the game stays in this one", "error", err)
		return
	}
	if replaced {
		game.played = next
	}
}

// defaultCamera is how a scene with no camera is seen: in perspective, from a little above.
var defaultCamera = scenefile.Camera{Projection: scenefile.Perspective}
