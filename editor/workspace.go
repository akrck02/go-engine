package editor

import (
	"strings"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/scenefile"
)

// The editor is six workspaces, one job each, as the pages of an editing suite: gathering the
// project's material, drawing it, building levels with it, writing the code, trying the game and
// shipping it. A bar along the bottom changes between them (or the keys 1 to 6); all of them see
// the same scene and the same chosen node.

// Workspace is one of the editor's screens.
type Workspace int

const (
	// ProjectWorkspace gathers the project's assets, scenes and settings.
	ProjectWorkspace Workspace = iota
	// ArtWorkspace makes pictures into sprites and tilesets.
	ArtWorkspace
	// LevelWorkspace places things in a scene and paints its tiles.
	LevelWorkspace
	// CodeWorkspace shows the behaviours the project's Go code defines.
	CodeWorkspace
	// PlayWorkspace plays the scene, to try it.
	PlayWorkspace
	// BuildWorkspace makes the game for each place it runs.
	BuildWorkspace
)

// Workspaces are the workspaces in the order of the bar.
var Workspaces = []Workspace{ProjectWorkspace, ArtWorkspace, LevelWorkspace, CodeWorkspace, PlayWorkspace, BuildWorkspace}

// workspaceParts are each workspace's word, icon and key, in the order of Workspaces.
var workspaceParts = []struct {
	name, icon string
	key        input.Key
}{
	{"Project", iconProject, input.Key1}, {"Art", iconArt, input.Key2}, {"Level", iconScene, input.Key3},
	{"Code", iconCode, input.Key4}, {"Play", iconPlay, input.Key5}, {"Build", iconBuild, input.Key6},
}

// Name is the workspace's word on the bar.
func (workspace Workspace) Name() string { return workspaceParts[workspace].name }

// Icon is the workspace's icon on the bar.
func (workspace Workspace) Icon() string { return workspaceParts[workspace].icon }

// Key is the key that shows the workspace.
func (workspace Workspace) Key() input.Key { return workspaceParts[workspace].key }

// WorkspaceNamed finds a workspace by its word, in any case: "art" is Art.
func WorkspaceNamed(name string) (Workspace, bool) {
	for _, workspace := range Workspaces {
		if strings.EqualFold(workspace.Name(), name) {
			return workspace, true
		}
	}
	return LevelWorkspace, false
}

// ShownWorkspace is the workspace the editor shows.
func (editor *Editor) ShownWorkspace() Workspace { return editor.workspace }

// ShowWorkspace changes to a workspace. Play plays the scene from where it is; leaving Play stops
// it, and what was played is not kept.
func (editor *Editor) ShowWorkspace(workspace Workspace) {
	if workspace == editor.workspace {
		return
	}
	if editor.workspace == PlayWorkspace {
		editor.StopHere()
	}
	editor.before, editor.workspace = editor.workspace, workspace
	switch workspace {
	case PlayWorkspace:
		editor.PlayHere()
	case ArtWorkspace:
		editor.openChosenSprite()
	}
}

// openChosenSprite opens the chosen node's sprite file, or its tile map's tileset file, in Art,
// when nothing is open there yet.
func (editor *Editor) openChosenSprite() {
	if editor.art.path != "" || editor.Selected == nil {
		return
	}
	file := editor.Selected.Sprite()
	if tiles, isTileMap := editor.Selected.Tiles(); isTileMap {
		file = tiles.Tileset
	}
	if !scenefile.IsSpriteFile(file) && !scenefile.IsTilesetFile(file) {
		return
	}
	if err := editor.OpenArt(file); err != nil {
		editor.say(err.Error(), true)
	}
}

// LeavePlay goes back from Play to the workspace it was entered from.
func (editor *Editor) LeavePlay() {
	if editor.workspace != PlayWorkspace {
		return
	}
	editor.ShowWorkspace(editor.before)
}
