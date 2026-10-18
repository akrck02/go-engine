package scenefile

import (
	"strconv"

	"github.com/akrck02/go-engine/failure"
)

// What a node of a 2D scene may have besides a sprite: a body in the scene's physics, a sheet of
// frames to animate its sprite, or a map of tiles.

// Body gives a node a body in the physics of a 2D scene.
type Body struct {
	Kind BodyKind   `json:"kind"`
	Size [2]float64 `json:"size,omitempty"` // width and height in metres; zero is the node's own sprite or shape
	// GravityScale is how much gravity pulls a moving body: 1 (the default) fully, 0 not at all.
	GravityScale *float64 `json:"gravityScale,omitempty"`
}

// BodyKind is what a body does: as physics.Kind, by name.
type BodyKind string

const (
	// Solid never moves; moving bodies stop against it.
	Solid BodyKind = "solid"
	// Moving falls, moves and is stopped by solids.
	Moving BodyKind = "moving"
	// Trigger stops nothing; it notices the moving bodies that touch it.
	Trigger BodyKind = "trigger"
)

// BodyKinds are the kinds of body, in the order the editor offers them.
var BodyKinds = []BodyKind{Solid, Moving, Trigger}

// Gravity is how much gravity pulls the body: its GravityScale, or 1.
func (body *Body) Gravity() float64 {
	if body.GravityScale == nil {
		return 1
	}
	return *body.GravityScale
}

// SpriteSheet cuts a node's sprite into frames of equal size, in rows read left to right and top to
// bottom (frame 0 is the top-left one), and names animations made of them.
type SpriteSheet struct {
	Columns    int         `json:"columns"`
	Rows       int         `json:"rows"`
	Animations []Animation `json:"animations,omitempty"`
	Start      string      `json:"start,omitempty"` // the animation playing when the scene starts
}

// Animation is a run of frames of a sheet, shown at a pace.
type Animation struct {
	Name            string  `json:"name"`
	Frames          []int   `json:"frames"`
	FramesPerSecond float64 `json:"framesPerSecond"`
	Loop            bool    `json:"loop,omitempty"` // starts again after its last frame; otherwise stays on it
}

// FrameCount is how many frames the sheet has.
func (sheet *SpriteSheet) FrameCount() int { return sheet.Columns * sheet.Rows }

// Find finds an animation by its name.
func (sheet *SpriteSheet) Find(name string) (Animation, bool) {
	for _, animation := range sheet.Animations {
		if animation.Name == name {
			return animation, true
		}
	}
	return Animation{}, false
}

// TileMap is a grid of tiles cut from a tileset. Its node's position is the map's top-left corner;
// rows go down from there. A tile is TileSize pixels a side, which the scene's pixels per metre
// make metres. The tileset is a tileset's file, which brings its tile size and solid tiles, or a
// picture, with the map's own.
type TileMap struct {
	Tileset  string `json:"tileset"`            // a tileset's file, or a picture cut into tiles in rows
	TileSize int    `json:"tileSize,omitempty"` // pixels a side, for a picture
	Columns  int    `json:"columns"`
	Rows     int    `json:"rows"`
	Tiles    []int  `json:"tiles"`           // a tile number per cell, row by row; EmptyTile for none
	Solid    []int  `json:"solid,omitempty"` // the tile numbers that bodies stand on, for a picture
}

// EmptyTile is a cell with no tile.
const EmptyTile = -1

// TileAt is the tile in a cell, or EmptyTile outside the map.
func (tiles *TileMap) TileAt(column, row int) int {
	if column < 0 || row < 0 || column >= tiles.Columns || row >= tiles.Rows {
		return EmptyTile
	}
	return tiles.Tiles[row*tiles.Columns+column]
}

// IsSolid tells whether a tile number is solid.
func (tiles *TileMap) IsSolid(tile int) bool {
	for _, solid := range tiles.Solid {
		if solid == tile {
			return true
		}
	}
	return false
}

// checkFlat tells what is wrong with a node's 2D parts, under a path for the message.
func (node *Node) checkFlat(path string) error {
	if node.Body != nil && !isBodyKind(node.Body.Kind) {
		return failure.Wrap(UnknownBodyKind, path+" is a "+string(node.Body.Kind), nil)
	}
	if node.Tiles != nil {
		if err := node.Tiles.check(path); err != nil {
			return err
		}
	}
	if node.Sheet != nil {
		return node.Sheet.check(path)
	}
	return nil
}

func isBodyKind(kind BodyKind) bool {
	for _, each := range BodyKinds {
		if each == kind {
			return true
		}
	}
	return false
}

func (tiles *TileMap) check(path string) error {
	if tiles.TileSize <= 0 && !IsTilesetFile(tiles.Tileset) { // a tileset's file has the size
		return failure.Wrap(TileSizeWrong, path, nil)
	}
	if tiles.Columns < 0 || tiles.Rows < 0 || len(tiles.Tiles) != tiles.Columns*tiles.Rows {
		count := strconv.Itoa(len(tiles.Tiles))
		size := strconv.Itoa(tiles.Columns) + " × " + strconv.Itoa(tiles.Rows)
		return failure.Wrap(TileCountWrong, path+": "+count+" tiles for "+size+" cells", nil)
	}
	return nil
}

func (sheet *SpriteSheet) check(path string) error {
	if sheet.Columns <= 0 || sheet.Rows <= 0 {
		return failure.Wrap(SheetSizeWrong, path, nil)
	}
	named := map[string]bool{}
	for _, animation := range sheet.Animations {
		if named[animation.Name] {
			return failure.Wrap(SameAnimationName, path+": "+animation.Name, nil)
		}
		named[animation.Name] = true
		if err := sheet.checkFrames(path, animation); err != nil {
			return err
		}
	}
	if sheet.Start != "" && !named[sheet.Start] {
		return failure.Wrap(UnknownAnimation, path+": "+sheet.Start, nil)
	}
	return nil
}

func (sheet *SpriteSheet) checkFrames(path string, animation Animation) error {
	for _, frame := range animation.Frames {
		if frame < 0 || frame >= sheet.FrameCount() {
			return failure.Wrap(FrameOutsideSheet, path+": "+animation.Name+" shows frame "+strconv.Itoa(frame), nil)
		}
	}
	return nil
}

// Copy is an independent copy of the body: changing one leaves the other alone.
func (body *Body) Copy() *Body {
	if body == nil {
		return nil
	}
	copied := *body
	if body.GravityScale != nil {
		scale := *body.GravityScale
		copied.GravityScale = &scale
	}
	return &copied
}

// Copy is an independent copy of the sheet and its animations.
func (sheet *SpriteSheet) Copy() *SpriteSheet {
	if sheet == nil {
		return nil
	}
	copied := *sheet
	copied.Animations = make([]Animation, len(sheet.Animations))
	for index, animation := range sheet.Animations {
		animation.Frames = append([]int(nil), animation.Frames...)
		copied.Animations[index] = animation
	}
	return &copied
}

// Copy is an independent copy of the tile map.
func (tiles *TileMap) Copy() *TileMap {
	if tiles == nil {
		return nil
	}
	copied := *tiles
	copied.Tiles = append([]int(nil), tiles.Tiles...)
	copied.Solid = append([]int(nil), tiles.Solid...)
	return &copied
}
