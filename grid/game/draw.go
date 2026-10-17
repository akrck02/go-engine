package game

import (
	"math"

	"github.com/akrck02/go-renderer/draw2d"
)

// Grid is where a scene's cells fall in an area of the screen: as large as fits, in whole pixels so
// the cells stay sharp, and centred.
type Grid struct {
	X, Y          float32 // the top-left corner of the first cell
	Cell          float32 // a cell's side
	Columns, Rows int
}

// Fit lays a scene's cells out in an area.
func Fit(area draw2d.Rect, scene *Scene) Grid {
	columns, rows := max(scene.Width, 1), max(scene.Height, 1)
	cell := float32(math.Floor(float64(min(area.Width/float32(columns), area.Height/float32(rows)))))
	cell = max(cell, 1)
	width, height := cell*float32(columns), cell*float32(rows)
	return Grid{
		X:    float32(math.Round(float64(area.X + (area.Width-width)/2))),
		Y:    float32(math.Round(float64(area.Y + (area.Height-height)/2))),
		Cell: cell, Columns: columns, Rows: rows,
	}
}

// CellAt finds the cell under a point, if the point is on the grid.
func (grid Grid) CellAt(x, y float32) (column, row int, found bool) {
	column = int(math.Floor(float64((x - grid.X) / grid.Cell)))
	row = int(math.Floor(float64((y - grid.Y) / grid.Cell)))
	return column, row, column >= 0 && row >= 0 && column < grid.Columns && row < grid.Rows
}

// CellArea is where a cell is drawn.
func (grid Grid) CellArea(column, row int) draw2d.Rect {
	return draw2d.Rect{X: grid.X + float32(column)*grid.Cell, Y: grid.Y + float32(row)*grid.Cell, Width: grid.Cell, Height: grid.Cell}
}

// Area is the whole grid.
func (grid Grid) Area() draw2d.Rect {
	return draw2d.Rect{X: grid.X, Y: grid.Y, Width: grid.Cell * float32(grid.Columns), Height: grid.Cell * float32(grid.Rows)}
}

// gridLine is the faint line between cells the editor draws.
var gridLine = draw2d.Color{0, 0, 0, 0.08}

// DrawScene draws a scene as it was laid out, with lines between its cells when asked (the
// editor's view).
func DrawScene(list *draw2d.List, area draw2d.Rect, game *Game, scene *Scene, lines bool) Grid {
	grid := Fit(area, scene)
	list.Rectangle(grid.Area(), 0, draw2d.Hex(scene.Background), draw2d.Color{}, 0)
	if lines && grid.Cell >= 6 {
		drawLines(list, grid)
	}
	for _, thing := range scene.Things {
		if kind, found := game.Kind(thing.Kind); found {
			DrawThing(list, grid.CellArea(thing.X, thing.Y), kind)
		}
	}
	return grid
}

// DrawWorld draws a scene in play.
func DrawWorld(list *draw2d.List, area draw2d.Rect, world *World) Grid {
	scene := world.Scene()
	grid := Fit(area, scene)
	list.Rectangle(grid.Area(), 0, draw2d.Hex(scene.Background), draw2d.Color{}, 0)
	world.Visit(func(kind *Kind, x, y int) {
		DrawThing(list, grid.CellArea(x, y), kind)
	})
	return grid
}

func drawLines(list *draw2d.List, grid Grid) {
	whole := grid.Area()
	for column := 1; column < grid.Columns; column++ {
		x := grid.X + float32(column)*grid.Cell
		list.Rectangle(draw2d.Rect{X: x, Y: whole.Y, Width: 1, Height: whole.Height}, 0, gridLine, draw2d.Color{}, 0)
	}
	for row := 1; row < grid.Rows; row++ {
		y := grid.Y + float32(row)*grid.Cell
		list.Rectangle(draw2d.Rect{X: whole.X, Y: y, Width: whole.Width, Height: 1}, 0, gridLine, draw2d.Color{}, 0)
	}
}

// DrawThing draws a thing of a kind in a cell: its shape in its colour, and its icon over it in a
// colour that stands out from it.
func DrawThing(list *draw2d.List, cell draw2d.Rect, kind *Kind) {
	colour := draw2d.Hex(kind.Colour)
	shape, radius := cell, float32(0)
	switch kind.Shape {
	case Rounded:
		shape, radius = cell.Inset(cell.Width/16), cell.Width/5
	case Circle:
		shape = cell.Inset(cell.Width / 10)
		radius = shape.Width / 2
	case Small:
		shape = cell.Inset(cell.Width / 4)
		radius = shape.Width / 2
	}
	list.Rectangle(shape, radius, colour, draw2d.Color{}, 0)
	if kind.Icon == "" || cell.Width < 8 {
		return
	}
	size := float32(math.Round(float64(cell.Width * 0.7)))
	list.Icon(kind.Icon, cell.X+(cell.Width-size)/2, cell.Y+(cell.Height-size)/2, size, inkOn(colour))
}

// inkOn is dark ink on a light colour and light ink on a dark one.
func inkOn(colour draw2d.Color) draw2d.Color {
	brightness := 0.299*colour[0] + 0.587*colour[1] + 0.114*colour[2]
	if brightness > 0.6 {
		return draw2d.Color{0, 0, 0, 0.7}
	}
	return draw2d.Color{1, 1, 1, 0.9}
}
