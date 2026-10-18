package scenefile

import (
	"encoding/json"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// A picture can be an asset of its own beside the scenes that use it: a sprite (the picture cut
// into frames, with named animations) or a tileset (the picture cut into tiles, some of them
// solid). Nodes and tile maps point to these files by path, so every coin of a game shares one
// spin and every map of a forest the same solid ground.

// The files of sprites and tilesets end with these.
const (
	SpriteSuffix  = ".sprite.json"
	TilesetSuffix = ".tileset.json"
)

// IsSpriteFile tells whether a path names a sprite's file.
func IsSpriteFile(path string) bool { return strings.HasSuffix(path, SpriteSuffix) }

// IsTilesetFile tells whether a path names a tileset's file.
func IsTilesetFile(path string) bool { return strings.HasSuffix(path, TilesetSuffix) }

// Sprite is a picture cut into frames, with the animations made of them.
type Sprite struct {
	Picture string `json:"picture"` // a picture in the project, as "assets/sprites/coin.png"
	SpriteSheet
}

// Tileset is a picture cut into square tiles, in rows, and which of them bodies stand on.
type Tileset struct {
	Picture  string `json:"picture"`
	TileSize int    `json:"tileSize"`        // pixels a side
	Solid    []int  `json:"solid,omitempty"` // the tile numbers that bodies stand on
}

// ReadSprite reads a sprite's file and checks it.
func ReadSprite(data []byte) (*Sprite, error) {
	var sprite Sprite
	if err := json.Unmarshal(data, &sprite); err != nil {
		return nil, failure.Wrap(Unreadable, "a sprite", err)
	}
	return &sprite, sprite.Check()
}

// Check tells what is wrong with a sprite, if anything.
func (sprite *Sprite) Check() error {
	if sprite.Picture == "" {
		return failure.Wrap(NoPicture, "a sprite", nil)
	}
	return sprite.SpriteSheet.check(sprite.Picture)
}

// Marshal writes the sprite as its file keeps it.
func (sprite *Sprite) Marshal() ([]byte, error) { return json.MarshalIndent(sprite, "", "\t") }

// ReadTileset reads a tileset's file and checks it.
func ReadTileset(data []byte) (*Tileset, error) {
	var tileset Tileset
	if err := json.Unmarshal(data, &tileset); err != nil {
		return nil, failure.Wrap(Unreadable, "a tileset", err)
	}
	return &tileset, tileset.Check()
}

// Check tells what is wrong with a tileset, if anything.
func (tileset *Tileset) Check() error {
	if tileset.Picture == "" {
		return failure.Wrap(NoPicture, "a tileset", nil)
	}
	if tileset.TileSize <= 0 {
		return failure.Wrap(TileSizeWrong, tileset.Picture, nil)
	}
	return nil
}

// Marshal writes the tileset as its file keeps it.
func (tileset *Tileset) Marshal() ([]byte, error) { return json.MarshalIndent(tileset, "", "\t") }

// IsSolid tells whether bodies stand on a tile of the tileset.
func (tileset *Tileset) IsSolid(tile int) bool {
	for _, solid := range tileset.Solid {
		if solid == tile {
			return true
		}
	}
	return false
}
