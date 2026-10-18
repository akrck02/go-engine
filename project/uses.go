package project

import (
	"io/fs"

	"github.com/akrck02/go-engine/scenefile"
)

// Use is a place an asset is used: a node of a scene, by its path, or a sprite's or tileset's
// file that cuts a picture (Node is then empty).
type Use struct {
	File string // the scene, or the sprite's or tileset's file
	Node string // the node's path in the scene, as "Harbour/Lamp"
}

// UsesOf finds where an asset is used: the nodes of every scene that show it (as a sprite, a
// model or a tile map's tileset), and the sprites and tilesets cut from it. A file that cannot be
// read is skipped: it uses nothing.
func (project *Project) UsesOf(asset string) []Use {
	var uses []Use
	for _, file := range project.Scenes() {
		scene, err := project.ReadScene(file)
		if err != nil {
			continue
		}
		uses = append(uses, nodesUsing(asset, file, scene.Nodes, "")...)
	}
	return append(uses, project.filesCutting(asset)...)
}

// nodesUsing are the nodes, and the nodes under them, that show an asset.
func nodesUsing(asset, file string, nodes []scenefile.Node, parent string) []Use {
	var uses []Use
	for _, node := range nodes {
		path := node.Name
		if parent != "" {
			path = parent + "/" + node.Name
		}
		if node.Sprite == asset || node.Model == asset || node.Tiles != nil && node.Tiles.Tileset == asset {
			uses = append(uses, Use{File: file, Node: path})
		}
		uses = append(uses, nodesUsing(asset, file, node.Children, path)...)
	}
	return uses
}

// filesCutting are the project's sprites and tilesets cut from a picture.
func (project *Project) filesCutting(picture string) []Use {
	var uses []Use
	for _, file := range project.filesUnder(AssetsFolder, isArtFile) {
		data, err := fs.ReadFile(project.Files(), file)
		if err != nil {
			continue
		}
		if cutFrom(file, data) == picture {
			uses = append(uses, Use{File: file})
		}
	}
	return uses
}

func isArtFile(name string) bool {
	return scenefile.IsSpriteFile(name) || scenefile.IsTilesetFile(name)
}

// cutFrom is the picture a sprite's or tileset's file cuts, or "" for one that cannot be read.
func cutFrom(file string, data []byte) string {
	if scenefile.IsSpriteFile(file) {
		if sprite, err := scenefile.ReadSprite(data); err == nil {
			return sprite.Picture
		}
		return ""
	}
	if tileset, err := scenefile.ReadTileset(data); err == nil {
		return tileset.Picture
	}
	return ""
}
