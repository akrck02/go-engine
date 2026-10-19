package editor

import (
	"math"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// What the 2D view can draw over the scene, to see what the game will do with it: the bodies
// (and the solid tiles), the box a following camera keeps its view in, and the cells of the tile
// map chosen.

// Colours of the overlays.
var (
	solidFill    = draw2d.Color{0.25, 0.45, 0.85, 0.22}
	solidLine    = draw2d.Color{0.25, 0.45, 0.85, 0.9}
	movingLine   = draw2d.Color{0.2, 0.75, 0.35, 0.95}
	triggerFill  = draw2d.Color{0.95, 0.55, 0.1, 0.18}
	triggerLine  = draw2d.Color{0.95, 0.55, 0.1, 0.95}
	pushFill     = draw2d.Color{0.2, 0.75, 0.35, 0.18}
	platformFill = draw2d.Color{0.6, 0.35, 0.85, 0.22}
	platformLine = draw2d.Color{0.6, 0.35, 0.85, 0.95}
	boundsLine   = draw2d.Color{0.85, 0.2, 0.5, 0.9}
	cellLine     = draw2d.Color{1, 1, 1, 0.25}
)

// smallestCellPixels is the size under which a tile map's cells are not outlined: too many lines.
const smallestCellPixels = 6

// drawOverlays draws the overlays switched on, and the chosen tile map's cells.
func (flat *flatView) drawOverlays(editor *Editor) {
	if flat.showBodies {
		editor.Scene.Root.Walk(func(node *engine.Node) bool {
			flat.drawBody(node)
			flat.drawSolidCells(node)
			return node.Visible || node == editor.Scene.Root
		})
	}
	if flat.showBounds {
		editor.Scene.Root.Walk(func(node *engine.Node) bool {
			flat.drawCameraBounds(editor.Scene, node)
			return true
		})
	}
	if editor.Selected != nil {
		flat.drawCells(editor.Selected)
	}
}

// drawBody outlines a node's body: solid ones filled blue, moving ones in green (filled when they
// can be pushed), platforms purple, triggers orange with an arrow where they pull.
func (flat *flatView) drawBody(node *engine.Node) {
	box, found := node.BodyBox()
	settings, _ := node.BodySettings()
	if !found {
		return
	}
	rect := flat.screenRect(box)
	switch settings.Kind {
	case scenefile.Solid:
		flat.drawing.Rectangle(rect, 0, solidFill, solidLine, 1.5)
	case scenefile.Trigger:
		flat.drawing.Rectangle(rect, 0, triggerFill, triggerLine, 1.5)
		flat.drawPull(rect, settings.Pull)
	case scenefile.Pushable:
		flat.drawing.Rectangle(rect, 0, pushFill, movingLine, 2)
	case scenefile.Platform:
		flat.drawing.Rectangle(rect, 0, platformFill, platformLine, 2)
	default:
		flat.drawing.Rectangle(rect, 0, draw2d.Color{}, movingLine, 2)
	}
}

// pullArrowSize is the side of the arrow drawn in a gravity zone, in pixels.
const pullArrowSize = 28

// drawPull draws, in the middle of a gravity zone, an arrow the way it pulls.
func (flat *flatView) drawPull(rect draw2d.Rect, pull *[2]float64) {
	if pull == nil || (pull[0] == 0 && pull[1] == 0) {
		return
	}
	icon := pullIcon(pull[0], pull[1])
	x, y := rect.X+(rect.Width-pullArrowSize)/2, rect.Y+(rect.Height-pullArrowSize)/2
	flat.drawing.Icon(icon, x, y, pullArrowSize, triggerLine)
}

// pullIcon is the arrow of a pull across and up: the way it pulls most.
func pullIcon(across, up float64) string {
	if math.Abs(across) > math.Abs(up) {
		if across > 0 {
			return "arrow-right"
		}
		return "arrow-left"
	}
	if up > 0 {
		return "arrow-up"
	}
	return "arrow-down"
}

// drawSolidCells fills a tile map's solid cells that are in sight.
func (flat *flatView) drawSolidCells(node *engine.Node) {
	tiles, isTileMap := node.Tiles()
	if !isTileMap {
		return
	}
	first, last, found := flat.cellsInSight(node, tiles)
	if !found {
		return
	}
	for row := first[1]; row <= last[1]; row++ {
		for column := first[0]; column <= last[0]; column++ {
			if node.SolidCell(column, row) {
				flat.drawing.Rectangle(flat.screenRect(node.CellBox(column, row)), 0, solidFill, draw2d.Color{}, 0)
			}
		}
	}
}

// drawCells outlines the cells of a tile map in sight, and while it is chosen.
func (flat *flatView) drawCells(node *engine.Node) {
	tiles, isTileMap := node.Tiles()
	if !isTileMap {
		return
	}
	first, last, found := flat.cellsInSight(node, tiles)
	if !found || flat.screenRect(node.CellBox(0, 0)).Width < smallestCellPixels {
		return
	}
	for row := first[1]; row <= last[1]; row++ {
		for column := first[0]; column <= last[0]; column++ {
			flat.drawing.Rectangle(flat.screenRect(node.CellBox(column, row)), 0, draw2d.Color{}, cellLine, 1)
		}
	}
}

// cellsInSight are the first and last cells of a tile map the view shows, and whether any is.
func (flat *flatView) cellsInSight(node *engine.Node, tiles *scenefile.TileMap) (first, last [2]int, found bool) {
	topLeft := flat.view.FromScreen(flat.area, flat.area.X, flat.area.Y)
	bottomRight := flat.view.FromScreen(flat.area, flat.area.X+flat.area.Width, flat.area.Y+flat.area.Height)
	firstColumn, firstRow, _ := node.CellAt(topLeft)
	lastColumn, lastRow, _ := node.CellAt(bottomRight)
	first = [2]int{max(firstColumn, 0), max(firstRow, 0)}
	last = [2]int{min(lastColumn, tiles.Columns-1), min(lastRow, tiles.Rows-1)}
	return first, last, first[0] <= last[0] && first[1] <= last[1]
}

// drawCameraBounds draws the box a following camera keeps its view in; an open top is left open.
func (flat *flatView) drawCameraBounds(scene *engine.Scene, node *engine.Node) {
	settings, isCamera := node.Camera()
	if !isCamera || settings.Follow == "" {
		return
	}
	bounds, found := scene.CameraBounds(node)
	if !found {
		return
	}
	top := bounds.Maximum.Y
	open := math.IsInf(top, 1)
	if open {
		top = flat.view.FromScreen(flat.area, 0, flat.area.Y).Y // the top of the view
	}
	left, upper := flat.view.ToScreen(flat.area, space.NewVector(bounds.Minimum.X, top, 0))
	right, lower := flat.view.ToScreen(flat.area, space.NewVector(bounds.Maximum.X, bounds.Minimum.Y, 0))
	const width = 2
	flat.drawing.Rectangle(draw2d.Rect{X: left, Y: lower - width/2, Width: right - left, Height: width}, 0, boundsLine, draw2d.Color{}, 0)
	flat.drawing.Rectangle(draw2d.Rect{X: left - width/2, Y: upper, Width: width, Height: lower - upper}, 0, boundsLine, draw2d.Color{}, 0)
	flat.drawing.Rectangle(draw2d.Rect{X: right - width/2, Y: upper, Width: width, Height: lower - upper}, 0, boundsLine, draw2d.Color{}, 0)
	if !open {
		flat.drawing.Rectangle(draw2d.Rect{X: left, Y: upper - width/2, Width: right - left, Height: width}, 0, boundsLine, draw2d.Color{}, 0)
	}
}

// screenRect is where a box of the world falls in the view.
func (flat *flatView) screenRect(box space.Box) draw2d.Rect {
	left, top := flat.view.ToScreen(flat.area, space.NewVector(box.Minimum.X, box.Maximum.Y, 0))
	right, bottom := flat.view.ToScreen(flat.area, space.NewVector(box.Maximum.X, box.Minimum.Y, 0))
	return draw2d.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
}
