package editor

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
)

// The inspector's section for a tile map: its tileset, its size, and the tile tool's palette.

var (
	tileSizeLimits = numberLimits{1, 512, 1}
	mapCellLimits  = numberLimits{1, 1024, 1}
)

// paletteSide is the side, in pixels, of a tile in the palette.
const paletteSide = 30

func (view *View) tileMapInspector(context *ui.Context, node *engine.Node, tiles *scenefile.TileMap) {
	context.Separator()
	context.Heading("Tile map")
	view.tilesetSelect(context, tiles)
	view.tileMapSizeFields(context, tiles)
	context.Separator()
	context.Row(func() {
		context.IconSized("paintbrush", 16, context.Theme.Primary)
		context.Heading("Paint")
	})
	if context.Checkbox("Paint in the view", &view.Editor.PaintingTiles) {
		view.Editor.Brush = max(view.Editor.Brush, 0)
	}
	note(context, "Left button paints the chosen tile, right button erases; the middle button moves the view.")
	view.tilePalette(context, node)
	view.solidCheckbox(context, tiles)
}

// tilesetSelect chooses the tileset among the project's pictures.
func (view *View) tilesetSelect(context *ui.Context, tiles *scenefile.TileMap) {
	options := append([]string{"No tileset yet"}, pictures(view.assetTree)...)
	chosen := max(0, indexOf(options, tiles.Tileset))
	context.Caption("Tileset: a picture of the project, cut into tiles in rows")
	if context.Select("Tileset", &chosen, options) && chosen > 0 {
		view.Editor.SetTileset(options[chosen])
	}
}

// pictures are the paths of the pictures among the project's assets.
func pictures(entry project.AssetEntry) []string {
	var found []string
	for _, child := range entry.Children {
		if child.Folder {
			found = append(found, pictures(child)...)
			continue
		}
		if project.KindOf(child.Path) == project.ImageAsset {
			found = append(found, child.Path)
		}
	}
	return found
}

func (view *View) tileMapSizeFields(context *ui.Context, tiles *scenefile.TileMap) {
	editor := view.Editor
	size := float64(tiles.TileSize)
	context.Caption("Pixels a tile, a side")
	if numberField(context, "Tile size", &size, tileSizeLimits, "%.0f px") {
		editor.SetTileSize(int(size))
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

// tilePalette shows the eraser and the tileset's tiles, to choose the one painted.
func (view *View) tilePalette(context *ui.Context, node *engine.Node) {
	editor := view.Editor
	tileset, err := editor.Scene.Tileset(node)
	if err != nil || len(tileset) == 0 {
		note(context, "Choose a tileset to paint with its tiles.")
		return
	}
	context.Push("Palette")
	context.Grid(8, func() {
		view.paletteCell(context, "eraser", scenefile.EmptyTile, func() {
			context.IconSized("eraser", 16, context.Theme.InkSoft)
		})
		for tile, picture := range tileset {
			view.paletteCell(context, "tile "+strconv.Itoa(tile), tile, func() {
				context.Image(picture, ui.Layout{Width: ui.Fixed(paletteSide - 6), Height: ui.Fixed(paletteSide - 6)}, 2)
			})
		}
	})
	context.Pop()
}

// paletteCell is one choice of the palette, outlined when it is the brush.
func (view *View) paletteCell(context *ui.Context, name string, tile int, content func()) {
	look := ui.Look{Fill: context.Theme.Background, Radius: 4, Border: context.Theme.ControlLine, BorderWidth: 1}
	if view.Editor.Brush == tile {
		look.Border, look.BorderWidth = context.Theme.Primary, 3
	}
	layout := ui.Layout{Width: ui.Fixed(paletteSide), Height: ui.Fixed(paletteSide), Padding: ui.Pad(3)}
	if context.Clickable(name, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, content) {
		view.Editor.Brush, view.Editor.PaintingTiles = tile, true
	}
}

// solidCheckbox makes the brush's tile solid, or not, and lists the solid ones.
func (view *View) solidCheckbox(context *ui.Context, tiles *scenefile.TileMap) {
	brush := view.Editor.Brush
	if brush >= 0 && context.Checkbox("Bodies stand on tile "+strconv.Itoa(brush), pointerTo(tiles.IsSolid(brush))) {
		view.Editor.ToggleSolid(brush)
	}
	solid := "none"
	if len(tiles.Solid) > 0 {
		solid = writtenFrames(tiles.Solid)
	}
	context.Text("Solid tiles: "+solid, ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
}

// hoveredCell is the outline of the cell under the pointer while painting.
var hoveredCell = draw2d.Color{1, 1, 1, 0.9}
