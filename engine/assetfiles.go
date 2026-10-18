package engine

import (
	"io/fs"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/scenefile"
)

// A node's sprite may be a picture or a sprite's file, and a tile map's tileset a picture or a
// tileset's file (scenefile.Sprite, scenefile.Tileset). These read the files once and give the
// node what it draws with: a sheet the node sets itself wins over its sprite's, so scenes written
// before sprites had files of their own still play.

// spriteFile reads a sprite's file of the project's, once.
func (library *assets) spriteFile(file string) (*scenefile.Sprite, error) {
	if sprite, found := library.sprites[file]; found {
		return sprite, nil
	}
	data, err := fs.ReadFile(library.files, file)
	if err != nil {
		return nil, failure.Wrap(SpriteNotFound, file, err)
	}
	sprite, err := scenefile.ReadSprite(data)
	if err != nil {
		return nil, err
	}
	library.sprites[file] = sprite
	return sprite, nil
}

// tilesetFile reads a tileset's file of the project's, once.
func (library *assets) tilesetFile(file string) (*scenefile.Tileset, error) {
	if tileset, found := library.tilesetFiles[file]; found {
		return tileset, nil
	}
	data, err := fs.ReadFile(library.files, file)
	if err != nil {
		return nil, failure.Wrap(TilesetNotFound, file, err)
	}
	tileset, err := scenefile.ReadTileset(data)
	if err != nil {
		return nil, err
	}
	library.tilesetFiles[file] = tileset
	return tileset, nil
}

// picture is the picture the node's sprite shows: its own, or its sprite file's.
func (node *Node) picture() string {
	if !scenefile.IsSpriteFile(node.sprite) {
		return node.sprite
	}
	sprite, err := node.scene.assets.spriteFile(node.sprite)
	if err != nil {
		return ""
	}
	return sprite.Picture
}

// frames is the sheet the node's sprite is cut into: the node's own, its sprite file's, or none.
func (node *Node) frames() *scenefile.SpriteSheet {
	if node.sheet != nil || !scenefile.IsSpriteFile(node.sprite) {
		return node.sheet
	}
	sprite, err := node.scene.assets.spriteFile(node.sprite)
	if err != nil {
		return nil
	}
	return &sprite.SpriteSheet
}

// tileSource is the picture a tile map's tiles are cut from, and their size in pixels: the map's
// own, or its tileset file's.
func (node *Node) tileSource() (picture string, tileSize int) {
	if !scenefile.IsTilesetFile(node.tiles.Tileset) {
		return node.tiles.Tileset, node.tiles.TileSize
	}
	tileset, err := node.scene.assets.tilesetFile(node.tiles.Tileset)
	if err != nil {
		return "", 0
	}
	return tileset.Picture, tileset.TileSize
}

// isSolidTile tells whether bodies stand on a tile of the node's map: by the map's own list, or
// its tileset file's.
func (node *Node) isSolidTile(tile int) bool {
	if !scenefile.IsTilesetFile(node.tiles.Tileset) {
		return node.tiles.IsSolid(tile)
	}
	tileset, err := node.scene.assets.tilesetFile(node.tiles.Tileset)
	return err == nil && tileset.IsSolid(tile)
}

// Picture is a picture of the project's, read once: what the editor's Art workspace shows.
func (owner *Scene) Picture(file string) (*draw2d.Image, error) { return owner.assets.image(file) }

// Frames are a picture cut into a sheet's frames, numbered from the top left, read and cut once.
func (owner *Scene) Frames(picture string, sheet *scenefile.SpriteSheet) ([]*draw2d.Image, error) {
	return owner.assets.sheetFrames(picture, sheet, false)
}
