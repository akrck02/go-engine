package editor

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
)

// A tile map in the Level workspace: the inspector chooses its tileset and its size; the palette,
// under the view, chooses the tile the paint and fill tools put down. Which tiles are solid
// belongs to a tileset's file, made in the Art workspace; a map of a bare picture keeps its own.

var (
	tileSizeLimits = numberLimits{1, 512, 1}
	mapCellLimits  = numberLimits{1, 1024, 1}
)

// The palette's look: the side of a tile in it, in pixels, and the gap and padding around one.
const (
	paletteSide    = 30
	paletteGap     = 3
	paletteHeight  = 170
	solidMarkSide  = 7
	paletteColumns = 12 // for a tileset whose columns cannot be read
)

func (view *View) tileMapInspector(context *ui.Context, tiles *scenefile.TileMap) {
	context.Separator()
	context.Heading("Tile map")
	view.tilesetSelect(context, tiles)
	view.tileMapSizeFields(context, tiles)
	note(context, "Paint it with the tools beside the view (B paints, E erases, G fills); choose the tile in the palette under it.")
}

// tilesetSelect chooses the tileset among the project's tilesets and pictures.
func (view *View) tilesetSelect(context *ui.Context, tiles *scenefile.TileMap) {
	options := append([]string{"No tileset yet"}, tilesetChoices(view.assetTree)...)
	chosen := max(0, indexOf(options, tiles.Tileset))
	context.Caption("Tileset: a tileset made in Art, or a picture cut into tiles in rows")
	if context.Select("Tileset", &chosen, options) && chosen > 0 {
		view.Editor.SetTileset(options[chosen])
	}
}

// tilesetChoices are the project's tilesets first, then its pictures.
func tilesetChoices(entry project.AssetEntry) []string {
	var tilesets, pictures []string
	for _, path := range assetPaths(entry) {
		switch {
		case scenefile.IsTilesetFile(path):
			tilesets = append(tilesets, path)
		case project.KindOf(path) == project.ImageAsset:
			pictures = append(pictures, path)
		}
	}
	return append(tilesets, pictures...)
}

// assetPaths are the paths of every file among the project's assets.
func assetPaths(entry project.AssetEntry) []string {
	var found []string
	for _, child := range entry.Children {
		if child.Folder {
			found = append(found, assetPaths(child)...)
			continue
		}
		found = append(found, child.Path)
	}
	return found
}

// pictures are the paths of the pictures among the project's assets.
func pictures(entry project.AssetEntry) []string {
	var found []string
	for _, path := range assetPaths(entry) {
		if project.KindOf(path) == project.ImageAsset {
			found = append(found, path)
		}
	}
	return found
}

func (view *View) tileMapSizeFields(context *ui.Context, tiles *scenefile.TileMap) {
	editor := view.Editor
	if !scenefile.IsTilesetFile(tiles.Tileset) {
		size := float64(tiles.TileSize)
		context.Caption("Pixels a tile, a side")
		if numberField(context, "Tile size", &size, tileSizeLimits, "%.0f px") {
			editor.SetTileSize(int(size))
		}
	}
	context.Caption("Cells across and down, from the top-left corner")
	columns, rows := float64(tiles.Columns), float64(tiles.Rows)
	context.Push("Map size")
	context.Grid(2, func() {
		if numberField(context, "Columns", &columns, mapCellLimits, "%.0f across") ||
			numberField(context, "Rows", &rows, mapCellLimits, "%.0f down") {
			editor.ResizeTileMap(int(columns), int(rows))
		}
	})
	context.Pop()
}

// paletteStrip is the strip under the view while a tile map is chosen: its tileset as it was
// drawn, the eraser first, the brush outlined and solid tiles marked in a corner.
func (view *View) paletteStrip(context *ui.Context, node *engine.Node, tiles *scenefile.TileMap) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(paletteHeight), Padding: ui.PadXY(10, 6), Gap: 4}
	bar(context, layout, func() {
		context.Row(func() {
			context.IconSized(iconTileset, 16, context.Theme.Primary)
			context.Text("Palette", ui.TextStyle{Weight: draw2d.Bold})
			context.Caption(tiles.Tileset)
			context.Spacer()
			view.solidToggle(context, tiles)
		})
		view.paletteTiles(context, node, tiles)
	})
}

// paletteTiles are the tileset's tiles in its own columns.
func (view *View) paletteTiles(context *ui.Context, node *engine.Node, tiles *scenefile.TileMap) {
	editor := view.Editor
	if tiles.Tileset == "" {
		note(context, "Choose a tileset in the inspector to paint with its tiles.")
		return
	}
	tileset, err := editor.Scene.Tileset(node)
	if err != nil {
		context.Text(err.Error(), ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
		return
	}
	columns := editor.PaletteColumns()
	if columns <= 0 {
		columns = paletteColumns
	}
	inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: paletteGap}
	context.ScrollArea("Palette", inside, func() {
		context.Row(func() {
			view.paletteCell(context, "Eraser", scenefile.EmptyTile, false, func() { context.IconSized(iconErase, 16, context.Theme.InkSoft) })
		})
		context.Grid(columns, func() {
			for tile, picture := range tileset {
				view.paletteCell(context, "Tile "+strconv.Itoa(tile), tile, node.SolidTile(tile), func() {
					context.Image(picture, ui.Layout{Width: ui.Fixed(paletteSide - 6), Height: ui.Fixed(paletteSide - 6)}, 2)
				})
			}
		})
	})
}

// paletteCell is one choice of the palette: the brush is outlined, a solid tile marked.
func (view *View) paletteCell(context *ui.Context, name string, tile int, solid bool, content func()) {
	look := ui.Look{Fill: context.Theme.Background, Radius: 4, Border: context.Theme.ControlLine, BorderWidth: 1}
	if view.Editor.Brush == tile {
		look.Border, look.BorderWidth = context.Theme.Primary, 3
	}
	if solid {
		look.Fill = context.Theme.Sunken
	}
	layout := ui.Layout{Width: ui.Fixed(paletteSide), Height: ui.Fixed(paletteSide), Padding: ui.Pad(3)}
	if context.Clickable(name, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, content) {
		view.Editor.Brush = tile
		if !view.Editor.Tool.paintsTiles() {
			view.Editor.Tool = PaintTool
		}
	}
	tip := name
	if solid {
		tip += " · solid"
	}
	context.Tooltip(tip)
}

// solidToggle makes the brush's tile solid, or not, for a map of a bare picture; a tileset's file
// keeps its solid tiles, set in Art.
func (view *View) solidToggle(context *ui.Context, tiles *scenefile.TileMap) {
	brush := view.Editor.Brush
	if brush < 0 {
		return
	}
	if scenefile.IsTilesetFile(tiles.Tileset) {
		context.Caption("Solid tiles are set in Art")
		return
	}
	solid := tiles.IsSolid(brush)
	if context.Checkbox("Bodies stand on tile "+strconv.Itoa(brush), &solid) {
		view.Editor.ToggleSolid(brush)
	}
}

// hoveredCell is the outline of the cell under the pointer while painting.
var hoveredCell = draw2d.Color{1, 1, 1, 0.9}
