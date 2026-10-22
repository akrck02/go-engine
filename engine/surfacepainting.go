package engine

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Each screen paints the surfaces placed on it, in the order the game set them: the main screen
// those over the view, in windows and in sheets (paintSurfaces), the second screen its own in a
// column (paintSecondScreen).

// paintSurfaces paints the surfaces placed on the main screen, in an area of it: the whole window
// in a game, or where a preview shows the game's screen (within is then true, and a surface on the
// whole view is as large as what it holds, so the rest of the area keeps the mouse).
func paintSurfaces(context *ui.Context, menu Menu, area draw2d.Rectangle, within bool) {
	paintOnWholeView(context, menu, area, within)
	for index := range surfaces.list {
		each := &surfaces.list[index]
		switch each.place.Place {
		case OverPlace:
			if !onWholeView(each.place) {
				paintOver(context, each, menu, area)
			}
		case WindowPlace:
			if each.shown {
				paintWindow(context, each, menu, area)
			}
		case SheetPlace:
			if each.shown && paintSheet(context, each, menu, area) {
				each.shown = false // the player dismissed it (back, a touch outside)
			}
		}
	}
}

// onWholeView tells whether a place lays a surface out on the whole view, as a game's interface
// is: over it, with no alignment or size.
func onWholeView(place SurfacePlace) bool {
	return place.Place == OverPlace && place.Align == "" && place.Width == 0 && place.Height == 0
}

// paintOnWholeView paints the surfaces laid out on the whole view, one under another: straight on
// the window's screen, or, within an area, in one layer as large as what they hold.
func paintOnWholeView(context *ui.Context, menu Menu, area draw2d.Rectangle, within bool) {
	if !within {
		paintEachOnWholeView(context, menu)
		return
	}
	for _, each := range surfaces.list {
		if onWholeView(each.place) {
			context.Within(area, func() {
				context.Overlay("Surfaces over the view", ui.LayerOptions{}, func() { paintEachOnWholeView(context, menu) })
			})
			return
		}
	}
}

// paintEachOnWholeView paints each surface laid out on the whole view.
func paintEachOnWholeView(context *ui.Context, menu Menu) {
	for index := range surfaces.list {
		if each := &surfaces.list[index]; onWholeView(each.place) {
			each.paint(context, menu)
		}
	}
}

// paintOver paints a surface over the view, aligned to an edge or a corner, or sized.
func paintOver(context *ui.Context, each *surface, menu Menu, area draw2d.Rectangle) {
	place := each.place
	across, down, _ := alignment(place.Align)
	options := ui.LayerOptions{AlignX: across, AlignY: down, Width: sizing(place.Width), Height: sizing(place.Height)}
	if place.Align != "" {
		options.Padding = ui.Pad(surfaceMargin)
	}
	context.Within(area, func() {
		context.Overlay(each.name, options, func() { each.paint(context, menu) })
	})
}

// paintWindow paints a surface in a floating window, titled with its name, which the player can
// drag; its action closes it.
func paintWindow(context *ui.Context, each *surface, menu Menu, area draw2d.Rectangle) {
	place := each.place
	across, down, _ := alignment(place.Align)
	x, y := alignedCorner(across, down, area, place.Width, place.Height)
	options := ui.PanelOptions{X: x, Y: y, Width: sizing(place.Width), Height: sizing(place.Height)}
	context.Within(area, func() {
		context.PanelWith(each.name, options, func() { each.paint(context, menu) })
	})
}

// paintSheet paints a surface in a sheet from the bottom of the view, over the rest dimmed, and
// tells whether the player dismissed it.
func paintSheet(context *ui.Context, each *surface, menu Menu, area draw2d.Rectangle) (dismissed bool) {
	place := each.place
	width := sizing(place.Width)
	if place.Width == 0 {
		width = ui.Grow()
	}
	options := ui.LayerOptions{AlignX: ui.Center, AlignY: ui.End, Width: width, Height: sizing(place.Height)}
	theme := &context.Theme
	card := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: theme.Gap}
	look := ui.Look{Fill: theme.Surface, Radius: theme.PanelRadius, Shadow: true}
	context.Within(area, func() {
		dismissed = context.Modal(each.name, options, func() {
			context.Frame(card, look, func() { each.paint(context, menu) })
		})
	})
	return dismissed
}

// paintSecondScreen paints the surfaces placed on the second screen, one under another, and tells
// whether there were any. One with a height is given that much of the screen; the rest are as
// tall as what they hold.
func paintSecondScreen(context *ui.Context, menu Menu) bool {
	painted := false
	for index := range surfaces.list {
		each := &surfaces.list[index]
		if each.place.Place != SecondScreenPlace {
			continue
		}
		painted = true
		if each.place.Height == 0 {
			each.paint(context, menu)
			continue
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(each.place.Height)}, func() {
			each.paint(context, menu)
		})
	}
	return painted
}

// hasMainScreenSurfaces tells whether any surface is placed on the main screen.
func hasMainScreenSurfaces() bool {
	for _, each := range surfaces.list {
		switch each.place.Place {
		case OverPlace, WindowPlace, SheetPlace:
			return true
		}
	}
	return false
}
