package editor

import (
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// The Level workspace's tools: what the left button does in the 2D view. The middle button always
// moves the view, and the right one only ever erases.

// Tool is what the left button does in the 2D view.
type Tool int

const (
	// SelectTool chooses a node, and drags it.
	SelectTool Tool = iota
	// PaintTool paints the chosen tile map with the brush.
	PaintTool
	// EraseTool takes tiles off the chosen tile map.
	EraseTool
	// FillTool paints a region of the same tile, as far as it goes.
	FillTool
	// HandTool moves the view.
	HandTool
)

// Tools are the tools in the order of the rail.
var Tools = []Tool{SelectTool, PaintTool, EraseTool, FillTool, HandTool}

// toolParts are each tool's word, icon, key and what it does, in the order of Tools.
var toolParts = []struct {
	name, icon, does string
	key              input.Key
}{
	{"Select", iconSelect, "Choose a node, and drag it", input.KeyV},
	{"Paint", iconPaint, "Paint the chosen tile map with the tile chosen in the palette", input.KeyB},
	{"Erase", iconErase, "Take tiles off the chosen tile map", input.KeyE},
	{"Fill", iconFill, "Paint a region of the same tile, as far as it goes", input.KeyG},
	{"Hand", iconHand, "Move the view", input.KeyH},
}

// Name is the tool's word.
func (tool Tool) Name() string { return toolParts[tool].name }

// Icon is the tool's icon on the rail.
func (tool Tool) Icon() string { return toolParts[tool].icon }

// Does says what the tool does, for its tooltip.
func (tool Tool) Does() string { return toolParts[tool].does }

// Key is the key that takes the tool.
func (tool Tool) Key() input.Key { return toolParts[tool].key }

// paintsTiles tells whether the tool works on a tile map.
func (tool Tool) paintsTiles() bool {
	return tool == PaintTool || tool == EraseTool || tool == FillTool
}

// FillTileAt paints a tile (scenefile.EmptyTile erases) over the region of the chosen tile map
// that holds the same tile as the cell at a point, cell by cell to the sides and up and down; it
// tells whether the map changed.
func (editor *Editor) FillTileAt(point space.Vector, tile int) bool {
	tiles, found := editor.selectedTiles()
	if !found {
		return false
	}
	column, row, inside := editor.Selected.CellAt(point)
	replaced := tiles.TileAt(column, row)
	if !inside || replaced == tile {
		return false
	}
	fillFrom(tiles, column, row, replaced, tile)
	editor.Changed()
	return true
}

// fillFrom paints the region of replaced tiles that a cell is part of, with a list of cells still
// to look at instead of recursion, so a large map cannot run out of stack.
func fillFrom(tiles *scenefile.TileMap, column, row, replaced, tile int) {
	pending := [][2]int{{column, row}}
	for len(pending) > 0 {
		cell := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if tiles.TileAt(cell[0], cell[1]) != replaced || !insideMap(tiles, cell[0], cell[1]) {
			continue
		}
		tiles.Tiles[cell[1]*tiles.Columns+cell[0]] = tile
		pending = append(pending, [2]int{cell[0] + 1, cell[1]}, [2]int{cell[0] - 1, cell[1]}, [2]int{cell[0], cell[1] + 1}, [2]int{cell[0], cell[1] - 1})
	}
}

func insideMap(tiles *scenefile.TileMap, column, row int) bool {
	return column >= 0 && row >= 0 && column < tiles.Columns && row < tiles.Rows
}

// PaletteColumns is how many tiles across the chosen tile map's tileset is, as it was drawn: the
// palette keeps that shape, so corners and edges stay beside each other.
func (editor *Editor) PaletteColumns() int {
	if _, found := editor.selectedTiles(); !found {
		return 0
	}
	return editor.Scene.TilesetColumns(editor.Selected)
}
