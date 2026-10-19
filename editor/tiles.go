package editor

import (
	"slices"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// The tile tool: a tile map node is painted in the 2D view, a tile of its tileset at a time.

// New tile maps' size: cells across and down, and pixels a tile.
const (
	newTileMapColumns  = 32
	newTileMapRows     = 16
	newTileMapTileSize = 16
)

// newTileMap is an empty tile map, with no tileset yet.
func newTileMap() *scenefile.TileMap {
	tiles := &scenefile.TileMap{TileSize: newTileMapTileSize, Columns: newTileMapColumns, Rows: newTileMapRows}
	tiles.Tiles = emptyCells(newTileMapColumns * newTileMapRows)
	return tiles
}

func emptyCells(count int) []int {
	cells := make([]int, count)
	for index := range cells {
		cells[index] = scenefile.EmptyTile
	}
	return cells
}

func (editor *Editor) selectedTiles() (*scenefile.TileMap, bool) {
	if editor.Selected == nil {
		return nil, false
	}
	return editor.Selected.Tiles()
}

// PaintingHere tells whether the 2D view paints tiles now: a tile tool is taken and a tile map is
// chosen.
func (editor *Editor) PaintingHere() bool {
	_, isTileMap := editor.selectedTiles()
	return editor.Tool.paintsTiles() && isTileMap
}

// PaintTileAt puts a tile (scenefile.EmptyTile erases) in the chosen tile map's cell at a point of
// the world; it tells whether the map changed.
func (editor *Editor) PaintTileAt(point space.Vector, tile int) bool {
	tiles, found := editor.selectedTiles()
	if !found {
		return false
	}
	column, row, inside := editor.Selected.CellAt(point)
	if !inside || tiles.TileAt(column, row) == tile {
		return false
	}
	tiles.Tiles[row*tiles.Columns+column] = tile
	editor.Changed()
	return true
}

// SetTileset sets the picture the chosen tile map's tiles are cut from.
func (editor *Editor) SetTileset(picture string) {
	tiles, found := editor.selectedTiles()
	if !found {
		return
	}
	tiles.Tileset = picture
	editor.Changed()
}

// SetTileSize sets how many pixels a side the chosen tile map's tiles are, at least one.
func (editor *Editor) SetTileSize(pixels int) {
	tiles, found := editor.selectedTiles()
	if !found {
		return
	}
	tiles.TileSize = max(pixels, 1)
	editor.Changed()
}

// ResizeTileMap gives the chosen tile map columns and rows, at least one of each. The tiles stay
// where they were from its top-left corner; the cells cut off are lost, the new ones empty.
func (editor *Editor) ResizeTileMap(columns, rows int) {
	tiles, found := editor.selectedTiles()
	if !found {
		return
	}
	columns, rows = max(columns, 1), max(rows, 1)
	cells := emptyCells(columns * rows)
	for row := 0; row < min(rows, tiles.Rows); row++ {
		for column := 0; column < min(columns, tiles.Columns); column++ {
			cells[row*columns+column] = tiles.TileAt(column, row)
		}
	}
	tiles.Columns, tiles.Rows, tiles.Tiles = columns, rows, cells
	editor.Changed()
}

// ToggleSolid makes a tile of the chosen map's tileset solid, for bodies to stand on, or not.
func (editor *Editor) ToggleSolid(tile int) {
	tiles, found := editor.selectedTiles()
	if !found || tile < 0 {
		return
	}
	if at := slices.Index(tiles.Solid, tile); at >= 0 {
		tiles.Solid = slices.Delete(tiles.Solid, at, at+1)
	} else {
		tiles.Solid = append(tiles.Solid, tile)
		slices.Sort(tiles.Solid)
	}
	editor.Changed()
}
