package editor

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/project"
)

// The editor runs inside the project's game (see engine.SetEditor): Install, from the project's
// editor.go, makes the game open the editor when go-engine starts it to edit it.

// Variables go-engine passes to the game it starts, and the editor to itself when it reloads.
const (
	sceneVariable    = "GO_ENGINE_SCENE"    // the scene to open
	selectedVariable = "GO_ENGINE_SELECTED" // the node to choose, by its path
)

// Install makes the game open the editor when go-engine starts it to edit it.
func Install() { engine.SetEditor(RunHere) }

// RunHere opens the editor on the project in the folder the game runs in, with the window's
// flags from the command line.
func RunHere() error {
	options := window.Options{Width: 1440, Height: 900, MinWidth: 1000, MinHeight: 600}
	flags := flag.NewFlagSet("go-engine", flag.ContinueOnError)
	options.RegisterFlags(flags)
	choose := flags.String("select", "", "choose a node by its path when the editor opens (for captures)")
	play := flags.Bool("play", false, "play the scene in the view when the editor opens (for captures)")
	shown := flags.String("workspace", "", "the workspace to open in: project, art, level, code, play or build (for captures)")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	opened, err := project.Open(".")
	if err != nil {
		return err
	}
	slog.Info("Go Engine's editor", "version", engine.Version(), "project", opened.Settings.Name)
	edited, err := Open(opened)
	if err != nil {
		return err
	}
	edited.builtAt = startedAt()
	if file := os.Getenv(sceneVariable); file != "" && file != edited.SceneFile {
		if err := edited.OpenScene(file); err != nil {
			edited.say(err.Error(), true)
		}
	}
	if path := firstOf(*choose, os.Getenv(selectedVariable)); path != "" {
		if node, found := edited.Scene.Find(path); found {
			edited.Select(node)
		}
	}
	if *play {
		edited.ShowWorkspace(PlayWorkspace)
	}
	if workspace, found := WorkspaceNamed(*shown); found {
		edited.ShowWorkspace(workspace)
	}
	view := NewView(edited)
	view.DarkByDefault = options.Dark
	return window.Run(view.Options(options), view.Build)
}

func firstOf(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// startedAt is when the running game was built: its file's time.
func startedAt() time.Time {
	executable, err := os.Executable()
	if err != nil {
		return time.Now()
	}
	details, err := os.Stat(executable)
	if err != nil {
		return time.Now()
	}
	return details.ModTime()
}

// CodeChanged tells whether the project's code changed since the game the editor runs in was
// built: Reload brings the changes in.
func (editor *Editor) CodeChanged() bool {
	if time.Since(editor.codeCheckedAt) < 2*time.Second {
		return editor.codeChanged
	}
	editor.codeCheckedAt = time.Now()
	editor.codeChanged = editor.Project.CodeChangedSince(editor.builtAt)
	return editor.codeChanged
}

// Reload saves the scene, builds the game again with the project's code as it is now, and starts
// the editor again in it, on the same scene and node.
func (editor *Editor) Reload() {
	if !editor.Saved && editor.Save() != nil {
		return
	}
	if err := editor.saveArtIfChanged(); err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.StopHere()
	editor.Run("Building the project's code…", func() (string, error) {
		if err := editor.Project.BuildEditor(); err != nil {
			fmt.Fprintln(&editor.buildOutput, err.Error()) // the Code workspace shows it
			return "", err
		}
		selected := ""
		if editor.Selected != nil {
			selected = editor.Selected.Path()
		}
		os.Setenv(sceneVariable, editor.SceneFile)
		os.Setenv(selectedVariable, selected)
		os.Setenv(engine.EditorVariable, "1")
		if err := restart(editor.Project.EditorProgram()); err != nil {
			return "", failure.Wrap(RestartFailed, "", err)
		}
		return "", nil
	})
}
