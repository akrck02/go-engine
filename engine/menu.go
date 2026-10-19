package engine

import (
	"io/fs"
	"path"
	"strings"

	"github.com/akrck02/go-engine/scenefile"
)

// A game whose settings ask for it (Settings.SceneMenu) lists its scenes on the Nintendo 3DS's
// bottom screen, to play any of them with a touch: a level select, handy while a game is tried.

// scenesFolder is where a project keeps its scenes, and sceneSuffix what their files end with.
const (
	scenesFolder = "scenes"
	sceneSuffix  = ".scene.json"
)

// menuScene is a scene the menu offers: its file, and the name it shows.
type menuScene struct {
	File, Name string
}

// scenesOf lists a project's scenes, in the order of their files, each by its name (or its file's,
// when it has none or cannot be read).
func scenesOf(files fs.FS) []menuScene {
	found, err := fs.Glob(files, path.Join(scenesFolder, "*"+sceneSuffix))
	if err != nil {
		return nil // the pattern is fixed: Glob fails only on a bad one
	}
	scenes := make([]menuScene, 0, len(found))
	for _, file := range found {
		scenes = append(scenes, menuScene{File: file, Name: sceneName(files, file)})
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
