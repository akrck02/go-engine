// Package game is what go-engine makes: a game of scenes laid out on a grid, with things in them
// that do what their kind says (the player walks, walls stop, boxes are pushed, coins are picked
// up, the goal leads on). It holds the game's description, kept as JSON, and its rules; it draws
// into a draw2d list, so the editor, the desktop and the Nintendo 3DS play it the same way.
package game

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Game is a whole game: its kinds of things and its scenes, played in order.
type Game struct {
	Title  string  `json:"title"`
	Kinds  []Kind  `json:"kinds"`
	Scenes []Scene `json:"scenes"`
}

// Kind is a kind of thing a scene holds: how it looks and what it does.
type Kind struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`         // "#rrggbb"
	Shape  Shape  `json:"shape"`          // how its cell is filled
	Icon   string `json:"icon,omitempty"` // a Lucide icon drawn over the shape
	Behaviour
}

// Behaviour is what a kind of thing does in play. A kind may do several things at once.
type Behaviour struct {
	Player      bool `json:"player,omitempty"`      // moves with the arrows (the first one placed)
	Solid       bool `json:"solid,omitempty"`       // nothing walks into it
	Pushable    bool `json:"pushable,omitempty"`    // the player pushes it, if the cell beyond is free
	Collectable bool `json:"collectable,omitempty"` // taken when the player steps on it
	Goal        bool `json:"goal,omitempty"`        // ends the scene, once everything is collected
	Hazard      bool `json:"hazard,omitempty"`      // starts the scene again when the player steps on it
}

// Shape is how a thing fills its cell.
type Shape string

const (
	Square  Shape = "square"
	Rounded Shape = "rounded"
	Circle  Shape = "circle"
	Small   Shape = "small" // a circle half the cell's size
)

// Shapes are the shapes in the order the editor offers them.
var Shapes = []Shape{Square, Rounded, Circle, Small}

// Scene is one screen of the game: a grid of cells with things placed in it.
type Scene struct {
	Name       string  `json:"name"`
	Width      int     `json:"width"`  // in cells
	Height     int     `json:"height"` // in cells
	Background string  `json:"background"`
	Things     []Thing `json:"things"`
}

// Thing is a thing placed in a scene: its kind, by name, and its cell.
type Thing struct {
	Kind string `json:"kind"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// Limits of a scene's size, in cells: large enough for a room, small enough to be seen whole on a
// 3DS's top screen at 8 pixels a cell.
const (
	MinimumSize = 3
	MaximumSize = 48
)

// Read reads a game kept as JSON and checks it can be played.
func Read(data []byte) (*Game, error) {
	var read Game
	if err := json.Unmarshal(data, &read); err != nil {
		return nil, err
	}
	if err := read.Check(); err != nil {
		return nil, err
	}
	return &read, nil
}

// Marshal writes the game as JSON, indented so it reads well in a repository.
func (game *Game) Marshal() ([]byte, error) {
	return json.MarshalIndent(game, "", "\t")
}

// Check tells what keeps the game from being played: no scenes, a scene too small or too large, a
// thing of a kind the game does not have, or two kinds with one name.
func (game *Game) Check() error {
	if len(game.Scenes) == 0 {
		return errors.New("the game has no scenes")
	}
	names := map[string]bool{}
	for _, kind := range game.Kinds {
		if names[kind.Name] {
			return fmt.Errorf("two kinds are called %q", kind.Name)
		}
		names[kind.Name] = true
	}
	for _, scene := range game.Scenes {
		if scene.Width < MinimumSize || scene.Height < MinimumSize || scene.Width > MaximumSize || scene.Height > MaximumSize {
			return fmt.Errorf("scene %q is %d × %d cells; it must be between %d and %d a side",
				scene.Name, scene.Width, scene.Height, MinimumSize, MaximumSize)
		}
		for _, thing := range scene.Things {
			if !names[thing.Kind] {
				return fmt.Errorf("scene %q holds a %q, which is not one of the game's kinds", scene.Name, thing.Kind)
			}
		}
	}
	return nil
}

// Kind finds a kind by its name.
func (game *Game) Kind(name string) (*Kind, bool) {
	for index := range game.Kinds {
		if game.Kinds[index].Name == name {
			return &game.Kinds[index], true
		}
	}
	return nil, false
}

// RenameKind gives a kind a new name, and the things of that kind with it. It does nothing if
// another kind already has the name.
func (game *Game) RenameKind(from, to string) bool {
	if from == to || to == "" {
		return false
	}
	if _, taken := game.Kind(to); taken {
		return false
	}
	kind, found := game.Kind(from)
	if !found {
		return false
	}
	kind.Name = to
	for sceneIndex := range game.Scenes {
		things := game.Scenes[sceneIndex].Things
		for index := range things {
			if things[index].Kind == from {
				things[index].Kind = to
			}
		}
	}
	return true
}

// RemoveKind takes a kind out of the game, and every thing of that kind out of its scenes.
func (game *Game) RemoveKind(name string) {
	kinds := game.Kinds[:0]
	for _, kind := range game.Kinds {
		if kind.Name != name {
			kinds = append(kinds, kind)
		}
	}
	game.Kinds = kinds
	for sceneIndex := range game.Scenes {
		scene := &game.Scenes[sceneIndex]
		things := scene.Things[:0]
		for _, thing := range scene.Things {
			if thing.Kind != name {
				things = append(things, thing)
			}
		}
		scene.Things = things
	}
}

// ThingAt finds the thing placed in a cell: the last one placed, which is drawn on top.
func (scene *Scene) ThingAt(x, y int) (int, bool) {
	for index := len(scene.Things) - 1; index >= 0; index-- {
		if scene.Things[index].X == x && scene.Things[index].Y == y {
			return index, true
		}
	}
	return -1, false
}

// Place puts a thing of a kind in a cell, in place of what was there. Cells outside the scene are
// left alone.
func (scene *Scene) Place(kind string, x, y int) {
	if !scene.Inside(x, y) {
		return
	}
	scene.Erase(x, y)
	scene.Things = append(scene.Things, Thing{Kind: kind, X: x, Y: y})
}

// Erase empties a cell.
func (scene *Scene) Erase(x, y int) {
	things := scene.Things[:0]
	for _, thing := range scene.Things {
		if thing.X != x || thing.Y != y {
			things = append(things, thing)
		}
	}
	scene.Things = things
}

// Inside tells whether a cell is in the scene.
func (scene *Scene) Inside(x, y int) bool {
	return x >= 0 && y >= 0 && x < scene.Width && y < scene.Height
}

// Resize changes the scene's size, forgetting the things that no longer fit.
func (scene *Scene) Resize(width, height int) {
	scene.Width = min(max(width, MinimumSize), MaximumSize)
	scene.Height = min(max(height, MinimumSize), MaximumSize)
	things := scene.Things[:0]
	for _, thing := range scene.Things {
		if scene.Inside(thing.X, thing.Y) {
			things = append(things, thing)
		}
	}
	scene.Things = things
}
