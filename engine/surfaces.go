package engine

import (
	"sort"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/failure"
)

// A game paints what it shows besides its scene as named surfaces ("map", "inventory", "chat"),
// each with Go UI, and says nothing of where they go: engine.json's surfaces place each one on
// each target. On a 3DS a map may sit on the second screen, touchable; on a computer, in a corner
// of the window or in a floating window opened by an action; on a phone, in a sheet from the
// bottom. A surface engine.json does not place goes where a game's interface always went: on the
// second screen where there is one, over the view elsewhere.

// The places a surface can go.
const (
	OverPlace         = "over"          // over the game's view, aligned to an edge or a corner
	WindowPlace       = "window"        // a floating window over the view, opened and closed by an action
	SecondScreenPlace = "second-screen" // the second screen (a 3DS's), touchable, in a column with the rest there
	SheetPlace        = "sheet"         // a sheet from the bottom (a phone's), opened and closed by an action
	NoPlace           = "none"          // not shown on that target
)

// SurfacePlace is where a surface goes on one target, as engine.json says.
type SurfacePlace struct {
	Place string `json:"place"`
	// Align is where an over or window surface sits in the view: top-left, top, top-right, left,
	// center, right, bottom-left, bottom or bottom-right. Empty lays an over surface out on the
	// whole view, as a game's interface is, and a window in the top left corner.
	Align string `json:"align,omitempty"`
	// Width and Height are its size in pixels; 0 is its content's.
	Width  float32 `json:"width,omitempty"`
	Height float32 `json:"height,omitempty"`
	// Action opens and closes a window or a sheet; Open shows it when the game starts.
	Action string `json:"action,omitempty"`
	Open   bool   `json:"open,omitempty"`
}

// surface is one the game paints: its name, how, where it goes on the target played, and whether
// a window or a sheet is open.
type surface struct {
	name  string
	paint func(context *ui.Context, menu Menu)
	place SurfacePlace
	shown bool
}

// surfaces are the game's, in the order it set them, and the places engine.json gives them.
var surfaces struct {
	list   []surface
	places map[string]map[string]SurfacePlace
	target string
}

// surfaceMargin is how far a surface aligned to an edge keeps from it, in pixels.
const surfaceMargin = 12

// SetSurface makes a game show a surface of its own, by name, painted with Go UI every frame;
// engine.json's surfaces say where it goes on each target. Setting a name again replaces how it
// is painted; nil removes it. Call it before Run.
func SetSurface(name string, paint func(context *ui.Context)) {
	if paint == nil {
		setSurface(name, nil)
		return
	}
	setSurface(name, func(context *ui.Context, _ Menu) { paint(context) })
}

// setSurface sets or removes a surface painted with the game's menu.
func setSurface(name string, paint func(context *ui.Context, menu Menu)) {
	for index := range surfaces.list {
		if surfaces.list[index].name != name {
			continue
		}
		if paint == nil {
			surfaces.list = append(surfaces.list[:index], surfaces.list[index+1:]...)
			return
		}
		surfaces.list[index].paint = paint
		return
	}
	if paint == nil {
		return
	}
	place, _ := placeOn(surfaces.places[name], surfaces.target)
	surfaces.list = append(surfaces.list, surface{name: name, paint: paint, place: place, shown: place.Open})
}

// UseSurfaces places the game's surfaces on a target, as engine.json's surfaces say (by surface,
// then by target); Run does it, and the editor for the target it previews. Places it does not know
// are reported, and the surface keeps where it goes by default.
func UseSurfaces(places map[string]map[string]SurfacePlace, target string) []error {
	surfaces.places, surfaces.target = places, target
	var problems []error
	for index := range surfaces.list {
		each := &surfaces.list[index]
		place, err := placeOn(places[each.name], target)
		if err != nil {
			problems = append(problems, err)
		}
		each.place, each.shown = place, place.Open
	}
	for _, name := range sortedNames(places) {
		if _, found := surfaceNamed(name); !found {
			_, err := placeOn(places[name], target) // a place for a surface not set yet is checked too
			problems = appendIfAny(problems, err)
		}
	}
	return problems
}

// placeOn is where a surface goes on a target, from its places by target: as given, the nearest
// place the target has, or its default.
func placeOn(byTarget map[string]SurfacePlace, target string) (SurfacePlace, error) {
	place, given := byTarget[target]
	if !given {
		return defaultPlace(target), nil
	}
	if !knownPlace(place) {
		return defaultPlace(target), failure.Wrap(SurfacePlaceUnknown, place.Place+" "+place.Align+" on "+target, nil)
	}
	return nearestPlace(place, target), nil
}

// defaultPlace is where a surface engine.json does not place goes: on a second screen, or over the
// whole view.
func defaultPlace(target string) SurfacePlace {
	if hasSecondScreen(target) {
		return SurfacePlace{Place: SecondScreenPlace}
	}
	return SurfacePlace{Place: OverPlace}
}

// nearestPlace is a place the target has: a second screen elsewhere than a 3DS is over the view,
// and a sheet on a 3DS (whose top screen is not touched) is a window.
func nearestPlace(place SurfacePlace, target string) SurfacePlace {
	if place.Place == SecondScreenPlace && !hasSecondScreen(target) {
		place.Place = OverPlace
	}
	if place.Place == SheetPlace && hasSecondScreen(target) {
		place.Place = WindowPlace
	}
	return place
}

// hasSecondScreen tells whether a target has a second screen.
func hasSecondScreen(target string) bool { return target == Nintendo3DSTarget }

// knownPlace tells whether a place and its alignment are ones the engine knows.
func knownPlace(place SurfacePlace) bool {
	_, _, aligned := alignment(place.Align)
	switch place.Place {
	case OverPlace, WindowPlace, SecondScreenPlace, SheetPlace, NoPlace:
		return aligned || place.Align == ""
	}
	return false
}

// surfacePlace is where a surface goes on the target played.
func surfacePlace(name string) SurfacePlace {
	found, _ := surfaceNamed(name)
	return found.place
}

// surfaceNamed is a surface by its name.
func surfaceNamed(name string) (surface, bool) {
	for _, each := range surfaces.list {
		if each.name == name {
			return each, true
		}
	}
	return surface{}, false
}

// forgetSurfaces removes every surface, and the places read for them.
func forgetSurfaces() {
	surfaces.list, surfaces.places, surfaces.target = nil, nil, ""
}

// toggleSurfaces opens or closes each window and sheet whose action was pressed this frame.
func toggleSurfaces() {
	for index := range surfaces.list {
		each := &surfaces.list[index]
		if each.place.Action != "" && actionPressed(each.place.Action) {
			each.shown = !each.shown
		}
	}
}

// actionPressed tells whether an action was pressed this frame.
func actionPressed(name string) bool {
	index, found := actionTable.index[name]
	return found && actionTable.actions[index].state.Pressed
}

// sortedNames are a map's keys in order.
func sortedNames[Value any](named map[string]Value) []string {
	names := make([]string, 0, len(named))
	for name := range named {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// appendIfAny adds an error to some, when there is one.
func appendIfAny(problems []error, err error) []error {
	if err == nil {
		return problems
	}
	return append(problems, err)
}

// alignment is an alignment's word as Go UI's alignments across and down.
func alignment(word string) (across, down ui.Align, known bool) {
	switch word {
	case "top-left":
		return ui.Start, ui.Start, true
	case "top":
		return ui.Center, ui.Start, true
	case "top-right":
		return ui.End, ui.Start, true
	case "left":
		return ui.Start, ui.Center, true
	case "center":
		return ui.Center, ui.Center, true
	case "right":
		return ui.End, ui.Center, true
	case "bottom-left":
		return ui.Start, ui.End, true
	case "bottom":
		return ui.Center, ui.End, true
	case "bottom-right":
		return ui.End, ui.End, true
	}
	return ui.Start, ui.Start, false
}

// sizing is a size in pixels as Go UI's: 0 is its content's.
func sizing(pixels float32) ui.Sizing {
	if pixels <= 0 {
		return ui.Sizing{}
	}
	return ui.Fixed(pixels)
}

// alignedCorner is where a box of a size aligned in an area has its top left corner, counted from
// the area's, keeping a margin from its edges.
func alignedCorner(across, down ui.Align, area draw2d.Rect, width, height float32) (x, y float32) {
	between := func(align ui.Align, room, size float32) float32 {
		switch align {
		case ui.Center:
			return (room - size) / 2
		case ui.End:
			return room - size - surfaceMargin
		}
		return surfaceMargin
	}
	return between(across, area.Width, width), between(down, area.Height, height)
}
