package engine

import (
	"image"
	"image/draw"
	"math"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/physics"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// A node of a 2D scene may be a map of tiles (scenefile.TileMap): pictures cut from a tileset,
// laid on a grid from the node's position, its top-left corner, rows going down. A tile is
// TileSize pixels a side, made metres by the scene's pixels per metre. Only the tiles in sight
// are drawn, and the solid ones, row by row, become solid boxes while the game plays.

// Tiles is the node's tile map, to edit in place; a node that is not one has none.
func (node *Node) Tiles() (*scenefile.TileMap, bool) { return node.tiles, node.tiles != nil }

// SetTiles makes the node a tile map, or not one (nil).
func (node *Node) SetTiles(tiles *scenefile.TileMap) { node.tiles = tiles }

// CellAt finds the cell of the node's tile map at a point of the world; inside is false outside
// the map, or for a node that is not one.
func (node *Node) CellAt(point space.Vector) (column, row int, inside bool) {
	if node.tiles == nil {
		return 0, 0, false
	}
	if _, tileSize := node.tileSource(); tileSize <= 0 {
		return 0, 0, false
	}
	width, height := node.tileMetres()
	corner := node.World().Position
	column = int(math.Floor((point.X - corner.X) / width))
	row = int(math.Floor((corner.Y - point.Y) / height))
	inside = column >= 0 && row >= 0 && column < node.tiles.Columns && row < node.tiles.Rows
	return column, row, inside
}

// SolidCell tells whether bodies stand on a cell of the node's tile map.
func (node *Node) SolidCell(column, row int) bool {
	return node.tiles != nil && node.isSolidTile(node.tiles.TileAt(column, row))
}

// SolidTile tells whether bodies stand on a tile of the node's tileset, by its number.
func (node *Node) SolidTile(tile int) bool { return node.tiles != nil && node.isSolidTile(tile) }

// CellBox is the box of the world one cell of the node's tile map fills.
func (node *Node) CellBox(column, row int) space.Box { return node.cellBox(column, row) }

// Tileset is the pictures of a tile map's tileset, cut once, numbered as its tiles are.
func (owner *Scene) Tileset(node *Node) ([]*draw2d.Image, error) {
	if node.tiles == nil {
		return nil, nil
	}
	return owner.assets.tileset(node.tileSource())
}

// TilesetColumns is how many tiles across a tile map's tileset picture is: the palette shows
// them in its rows, as they were drawn. It is 0 for a map whose tileset cannot be read.
func (owner *Scene) TilesetColumns(node *Node) int {
	if node.tiles == nil {
		return 0
	}
	picture, tileSize := node.tileSource()
	whole, err := owner.assets.image(picture)
	if err != nil || tileSize <= 0 {
		return 0
	}
	width, _ := whole.Size()
	return width / tileSize
}

// tileMetres is how many metres a tile of a node's map takes, across and down, its node's scale
// counted.
func (node *Node) tileMetres() (width, height float64) {
	scale := node.World().Scale
	_, tileSize := node.tileSource()
	side := float64(tileSize) / node.scene.pixelsPerMetre()
	return side * math.Abs(scale.X), side * math.Abs(scale.Y)
}

// tileBounds is the box the whole map fills.
func (node *Node) tileBounds() space.Box {
	width, height := node.tileMetres()
	corner := node.World().Position
	return space.Box{
		Minimum: space.NewVector(corner.X, corner.Y-height*float64(node.tiles.Rows), corner.Z-0.005),
		Maximum: space.NewVector(corner.X+width*float64(node.tiles.Columns), corner.Y, corner.Z+0.005),
	}
}

// cellBox is the box of one cell of the map.
func (node *Node) cellBox(column, row int) space.Box {
	width, height := node.tileMetres()
	corner := node.World().Position
	left, top := corner.X+width*float64(column), corner.Y-height*float64(row)
	return space.Box{Minimum: space.NewVector(left, top-height, corner.Z), Maximum: space.NewVector(left+width, top, corner.Z)}
}

// drawTiles draws the map's tiles that fall in the area.
func (owner *Scene) drawTiles(list *draw2d.List, area draw2d.Rectangle, view View2D, node *Node) {
	pictures, err := owner.assets.tileset(node.tileSource())
	if err != nil {
		return
	}
	first, last := node.cellsInSight(area, view)
	for row := first.row; row <= last.row; row++ {
		for column := first.column; column <= last.column; column++ {
			tile := node.tiles.TileAt(column, row)
			if tile < 0 || tile >= len(pictures) {
				continue
			}
			list.Image(pictures[tile], screenRect(area, view, node.cellBox(column, row)), 0, draw2d.Color{1, 1, 1, 1})
		}
	}
}

// cell is a cell of a tile map, by its column and row.
type cell struct{ column, row int }

// cellsInSight are the first and last cells of the map the area shows (the last before the first
// when none is in sight).
func (node *Node) cellsInSight(area draw2d.Rectangle, view View2D) (first, last cell) {
	width, height := node.tileMetres()
	corner := node.World().Position
	topLeft := view.FromScreen(area, area.X, area.Y)
	bottomRight := view.FromScreen(area, area.X+area.Width, area.Y+area.Height)
	first = cell{column: max(int(math.Floor((topLeft.X-corner.X)/width)), 0), row: max(int(math.Floor((corner.Y-topLeft.Y)/height)), 0)}
	last = cell{
		column: min(int(math.Floor((bottomRight.X-corner.X)/width)), node.tiles.Columns-1),
		row:    min(int(math.Floor((corner.Y-bottomRight.Y)/height)), node.tiles.Rows-1),
	}
	return first, last
}

// screenRect is where a box of the world falls on the screen.
func screenRect(area draw2d.Rectangle, view View2D, box space.Box) draw2d.Rectangle {
	left, top := view.ToScreen(area, space.NewVector(box.Minimum.X, box.Maximum.Y, 0))
	right, bottom := view.ToScreen(area, space.NewVector(box.Maximum.X, box.Minimum.Y, 0))
	return draw2d.Rectangle{X: left, Y: top, Width: right - left, Height: bottom - top}
}

// addTileBodies makes the solid boxes of a tile map: one for each run of solid tiles side by side
// in a row.
func (owner *Scene) addTileBodies(node *Node) {
	for row := 0; row < node.tiles.Rows; row++ {
		for column := 0; column < node.tiles.Columns; {
			run := node.solidRunFrom(column, row)
			if run == 0 {
				column++
				continue
			}
			box := node.cellBox(column, row).Union(node.cellBox(column+run-1, row))
			owner.addTileBody(node, box)
			column += run
		}
	}
}

// solidRunFrom is how many solid tiles there are side by side in a row, from a column.
func (node *Node) solidRunFrom(column, row int) int {
	run := 0
	for column+run < node.tiles.Columns && node.isSolidTile(node.tiles.TileAt(column+run, row)) {
		run++
	}
	return run
}

func (owner *Scene) addTileBody(node *Node, box space.Box) {
	body := &physics.Body{Kind: physics.Solid, Position: box.Centre(), Size: box.Size()}
	node.tileBodies = append(node.tileBodies, body)
	owner.bodyOwners[body] = node
	owner.physics.Add(body)
}

// tileset reads a tileset picture and cuts it into tiles of a size, in rows, once.
func (library *assets) tileset(file string, tileSize int) ([]*draw2d.Image, error) {
	key := tilesetKey{file: file, tileSize: tileSize}
	if tiles, found := library.tilesets[key]; found {
		return tiles, nil
	}
	whole, err := library.image(file)
	if err != nil {
		return nil, err
	}
	if tileSize <= 0 {
		return nil, failure.Wrap(scenefile.TileSizeWrong, file, nil)
	}
	tiles := cutIntoTiles(whole.Pixels, tileSize)
	library.tilesets[key] = tiles
	return tiles, nil
}

// tilesetKey names a tileset cut at a size.
type tilesetKey struct {
	file     string
	tileSize int
}

// cutIntoTiles cuts a picture into squares of a size, left to right and top to bottom.
func cutIntoTiles(picture *image.NRGBA, tileSize int) []*draw2d.Image {
	bounds := picture.Bounds()
	var tiles []*draw2d.Image
	for top := bounds.Min.Y; top+tileSize <= bounds.Max.Y; top += tileSize {
		for left := bounds.Min.X; left+tileSize <= bounds.Max.X; left += tileSize {
			tile := image.NewNRGBA(image.Rect(0, 0, tileSize, tileSize))
			draw.Draw(tile, tile.Bounds(), picture, image.Pt(left, top), draw.Src)
			tiles = append(tiles, pixelArt(draw2d.NewImage(tile)))
		}
	}
	return tiles
}
