// Package editor is go-engine's editor: a window where a game is made, scene by scene, by painting
// things of its kinds on a grid, and played on the spot. This file holds what the editor does;
// view.go describes its window with go-ui.
package editor

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/akrck02/go-engine/game"
	"github.com/akrck02/go-engine/player"
)

// Editor is a game being made.
type Editor struct {
	Game  *game.Game
	Path  string // where the game is kept, as JSON
	Saved bool   // nothing changed since it was opened or saved

	scene   int    // the scene shown
	brush   string // the kind painted; "" paints nothing (the eraser)
	kind    int    // the kind shown in the inspector
	playing *player.Session

	Message string // the last thing the editor has to say: saved, exported, an error
	Failed  bool   // the message is an error

	// the inspector's fields, kept while they are typed in
	kindName  string
	sceneName string
}

// New edits a game kept at a path (which need not exist yet).
func New(edited *game.Game, path string) *Editor {
	editor := &Editor{Game: edited, Path: path, Saved: true}
	editor.chooseKind(0)
	editor.chooseScene(0)
	if len(edited.Kinds) > 0 {
		editor.brush = edited.Kinds[0].Name
	}
	return editor
}

// Open reads a game from a file, or starts the sample there if there is no file yet.
func Open(path string) (*Editor, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		editor := New(game.Sample(), path)
		editor.Saved = false
		editor.say("A new game: the sample, to start from", false)
		return editor, nil
	}
	if err != nil {
		return nil, err
	}
	opened, err := game.Read(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return New(opened, path), nil
}

// Save writes the game to its path.
func (editor *Editor) Save() error {
	if err := editor.Game.Check(); err != nil {
		editor.say(err.Error(), true)
		return err
	}
	data, err := editor.Game.Marshal()
	if err == nil {
		err = os.WriteFile(editor.Path, data, 0o644)
	}
	if err != nil {
		editor.say(err.Error(), true)
		return err
	}
	editor.Saved = true
	editor.say("Saved to "+editor.Path, false)
	return nil
}

// say puts a message on the status line, on one line.
func (editor *Editor) say(message string, failed bool) {
	editor.Message, editor.Failed = strings.Join(strings.Fields(message), " "), failed
}

func (editor *Editor) changed() { editor.Saved = false }

// Scene is the scene shown.
func (editor *Editor) Scene() *game.Scene { return &editor.Game.Scenes[editor.scene] }

// SceneIndex is the scene shown, from 0.
func (editor *Editor) SceneIndex() int { return editor.scene }

func (editor *Editor) chooseScene(index int) {
	editor.scene = min(max(index, 0), len(editor.Game.Scenes)-1)
	editor.sceneName = editor.Scene().Name
}

// Kind is the kind shown in the inspector, if the game has any.
func (editor *Editor) Kind() (*game.Kind, bool) {
	if editor.kind < 0 || editor.kind >= len(editor.Game.Kinds) {
		return nil, false
	}
	return &editor.Game.Kinds[editor.kind], true
}

func (editor *Editor) chooseKind(index int) {
	editor.kind = index
	if kind, found := editor.Kind(); found {
		editor.kindName = kind.Name
	}
}

// Brush is the kind painted, or "" for the eraser.
func (editor *Editor) Brush() string { return editor.brush }

// ChooseBrush paints with a kind (and shows it in the inspector), or erases with "".
func (editor *Editor) ChooseBrush(name string) {
	editor.brush = name
	for index, kind := range editor.Game.Kinds {
		if kind.Name == name {
			editor.chooseKind(index)
		}
	}
}

// Paint puts the brush's kind in a cell of the scene shown, or empties it with the eraser.
func (editor *Editor) Paint(column, row int) {
	scene := editor.Scene()
	index, occupied := scene.ThingAt(column, row)
	if editor.brush == "" {
		if occupied {
			scene.Erase(column, row)
			editor.changed()
		}
		return
	}
	if occupied && scene.Things[index].Kind == editor.brush {
		return // already painted: dragging over it changes nothing
	}
	scene.Place(editor.brush, column, row)
	editor.changed()
}

// AddScene adds an empty scene after the one shown, the same size, and shows it.
func (editor *Editor) AddScene() {
	shown := editor.Scene()
	added := game.Scene{Name: editor.unusedSceneName(), Width: shown.Width, Height: shown.Height, Background: shown.Background}
	at := editor.scene + 1
	editor.Game.Scenes = append(editor.Game.Scenes[:at], append([]game.Scene{added}, editor.Game.Scenes[at:]...)...)
	editor.chooseScene(at)
	editor.changed()
}

// RemoveScene removes the scene shown; the last one stays.
func (editor *Editor) RemoveScene() {
	if len(editor.Game.Scenes) <= 1 {
		editor.say("A game needs a scene", true)
		return
	}
	editor.Game.Scenes = append(editor.Game.Scenes[:editor.scene], editor.Game.Scenes[editor.scene+1:]...)
	editor.chooseScene(editor.scene)
	editor.changed()
}

// MoveScene moves the scene shown earlier (-1) or later (+1) in the game.
func (editor *Editor) MoveScene(by int) {
	to := editor.scene + by
	if to < 0 || to >= len(editor.Game.Scenes) {
		return
	}
	scenes := editor.Game.Scenes
	scenes[editor.scene], scenes[to] = scenes[to], scenes[editor.scene]
	editor.chooseScene(to)
	editor.changed()
}

// AddKind adds a kind that does nothing yet, and paints with it.
func (editor *Editor) AddKind() {
	name := editor.unusedKindName()
	editor.Game.Kinds = append(editor.Game.Kinds, game.Kind{Name: name, Colour: "#8d99ae", Shape: game.Rounded})
	editor.ChooseBrush(name)
	editor.changed()
}

// RemoveKind removes the kind shown, and its things from every scene.
func (editor *Editor) RemoveKind() {
	kind, found := editor.Kind()
	if !found {
		return
	}
	name := kind.Name
	editor.Game.RemoveKind(name)
	if editor.brush == name {
		editor.brush = ""
	}
	editor.chooseKind(min(editor.kind, len(editor.Game.Kinds)-1))
	editor.changed()
}

// RenameKind renames the kind shown, if the name is free; it tells whether it could.
func (editor *Editor) RenameKind(name string) bool {
	kind, found := editor.Kind()
	if !found {
		return false
	}
	from := kind.Name
	if !editor.Game.RenameKind(from, name) {
		return false
	}
	if editor.brush == from {
		editor.brush = name
	}
	editor.changed()
	return true
}

func (editor *Editor) unusedKindName() string {
	for number := len(editor.Game.Kinds) + 1; ; number++ {
		name := "Kind " + strconv.Itoa(number)
		if _, taken := editor.Game.Kind(name); !taken {
			return name
		}
	}
}

func (editor *Editor) unusedSceneName() string {
	return "Scene " + strconv.Itoa(len(editor.Game.Scenes)+1)
}

// Playing is the game being played, while it is.
func (editor *Editor) Playing() (*player.Session, bool) { return editor.playing, editor.playing != nil }

// Play plays the game from the scene shown.
func (editor *Editor) Play() {
	if err := editor.Game.Check(); err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.playing = player.NewSession(editor.Game)
	editor.playing.World.Start(editor.scene)
	editor.say("Playing: arrows walk, R starts the scene again", false)
}

// Stop goes back to editing.
func (editor *Editor) Stop() {
	editor.playing = nil
	editor.say("", false)
}
