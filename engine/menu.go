package engine

import (
	"io/fs"
	"path"
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/scenefile"
)

// A game whose settings ask for it (Settings.SceneMenu) lists its scenes on the second screen of a
// device that has one (a 3DS's), to play any of them with a touch: a level select, handy while a
// game is tried. A game draws that screen its own way with SetInterface, given the same scenes
// (Menu).

// scenesFolder is where a project keeps its scenes, and sceneSuffix what their files end with.
const (
	scenesFolder = "scenes"
	sceneSuffix  = ".scene.json"
)

// MenuScene is a scene the menu offers: its file, and the name it shows.
type MenuScene struct {
	File, Name string
}

// scenesOf lists a project's scenes, in the order of their files, each by its name (or its file's,
// when it has none or cannot be read).
func scenesOf(files fs.FS) []MenuScene {
	found, err := fs.Glob(files, path.Join(scenesFolder, "*"+sceneSuffix))
	if err != nil {
		return nil // the pattern is fixed: Glob fails only on a bad one
	}
	scenes := make([]MenuScene, 0, len(found))
	for _, file := range found {
		scenes = append(scenes, MenuScene{File: file, Name: sceneName(files, file)})
	}
	return scenes
}

// sceneName is the name a scene's file gives it, or the file's own without its suffix.
func sceneName(files fs.FS, file string) string {
	if data, err := fs.ReadFile(files, file); err == nil {
		if described, err := scenefile.Read(data); err == nil && described.Name != "" {
			return described.Name
		}
	}
	return strings.TrimSuffix(path.Base(file), sceneSuffix)
}

// choose plays a scene of the menu next, as a behaviour's ChangeScene does.
func (game *runner) choose(file string) {
	game.played.changeRequested = file
}

// Menu is what a game's own bottom screen is given: its scenes, the one playing, and how to play
// another.
type Menu struct{ game *runner }

// Scenes are the game's scenes, in the order of their files.
func (menu Menu) Scenes() []MenuScene { return menu.game.menu }

// Playing is the scene playing now: its name, and its pictures (Scene.Picture) to draw with.
func (menu Menu) Playing() *Scene { return menu.game.played }

// Play plays a scene next, by its file.
func (menu Menu) Play(file string) { menu.game.choose(file) }

// SetInterface makes a game show an interface of its own, with go-ui, in its own look: a menu of
// its scenes (Menu), a map, its controls, a login form. It is the surface named "interface"
// (SetSurface): unless engine.json places it elsewhere, on the Nintendo 3DS it is the bottom
// screen, and elsewhere it is painted over the game's view. nil removes it. Call it before Run.
func SetInterface(paint func(context *ui.Context, menu Menu)) { setSurface(InterfaceSurface, paint) }

// InterfaceSurface is the name of the game's interface among its surfaces (SetInterface).
const InterfaceSurface = "interface"
