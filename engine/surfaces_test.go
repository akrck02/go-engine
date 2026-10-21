package engine

import (
	"errors"
	"testing"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// A made-up game's window, in pixels.
const (
	gameWidth  = 1280
	gameHeight = 720
)

// captionSurface is a surface that says it is shown, by its name (a window's title is its name
// alone).
func captionSurface(name string) func(context *ui.Context) {
	return func(context *ui.Context) { context.Caption(name + " shown") }
}

// showSurfaces starts afresh with some surfaces placed on a target, and opens a made-up game's
// window that shows them over its view, with no second screen.
func showSurfaces(test *testing.T, places map[string]map[string]SurfacePlace, target string, names ...string) *uitest.Window {
	test.Helper()
	forgetSurfaces()
	for _, name := range names {
		SetSurface(name, captionSurface(name))
	}
	if problems := UseSurfaces(places, target); len(problems) != 0 {
		test.Fatal(problems)
	}
	whole := draw2d.Rect{Width: gameWidth, Height: gameHeight}
	window, err := uitest.New(func(context *ui.Context) { paintSurfaces(context, Menu{}, whole, false) },
		uitest.Size(gameWidth, gameHeight), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	return window
}

// shownAt is where a surface's caption was shown, or nothing if it was not.
func shownAt(window *uitest.Window, name string) (draw2d.Rect, bool) {
	element, err := window.Element(name + " shown")
	if err != nil || !element.Visible {
		return draw2d.Rect{}, false
	}
	return element.Rect, true
}

func TestASurfaceOverTheViewIsAlignedWhereEngineJsonSays(test *testing.T) {
	places := map[string]map[string]SurfacePlace{
		"Map": {DesktopTarget: {Place: OverPlace, Align: "bottom-right", Width: 200, Height: 100}},
	}
	window := showSurfaces(test, places, DesktopTarget, "Map")
	shown, found := shownAt(window, "Map")
	if !found {
		test.Fatal("the map is not shown")
	}
	if shown.X < gameWidth-200-surfaceMargin-1 || shown.Y < gameHeight-100-surfaceMargin-1 {
		test.Fatalf("the map is at %+v, not in the bottom right corner", shown)
	}
}

func TestASurfaceNotPlacedIsOnTheSecondScreenOfAConsoleAndOverTheViewElsewhere(test *testing.T) {
	forgetSurfaces()
	SetSurface("Map", captionSurface("Map"))
	UseSurfaces(nil, Nintendo3DSTarget)
	if place := surfacePlace("Map"); place.Place != SecondScreenPlace {
		test.Fatalf("on the 3DS the map is %q", place.Place)
	}
	UseSurfaces(nil, DesktopTarget)
	if place := surfacePlace("Map"); place.Place != OverPlace {
		test.Fatalf("on the desktop the map is %q", place.Place)
	}
}

func TestAPlaceTheTargetLacksIsTheNearestItHas(test *testing.T) {
	places := map[string]map[string]SurfacePlace{
		"Map": {DesktopTarget: {Place: SecondScreenPlace}, Nintendo3DSTarget: {Place: SheetPlace, Action: "map"}},
	}
	forgetSurfaces()
	SetSurface("Map", captionSurface("Map"))
	UseSurfaces(places, DesktopTarget)
	if place := surfacePlace("Map"); place.Place != OverPlace {
		test.Fatalf("a second screen on the desktop is %q", place.Place)
	}
	UseSurfaces(places, Nintendo3DSTarget)
	if place := surfacePlace("Map"); place.Place != WindowPlace || place.Action != "map" {
		test.Fatalf("a sheet on the 3DS is %+v", place)
	}
}

func TestASurfacePlacedNowhereIsNotShown(test *testing.T) {
	places := map[string]map[string]SurfacePlace{"Map": {DesktopTarget: {Place: NoPlace}}}
	window := showSurfaces(test, places, DesktopTarget, "Map", "Lamp")
	if _, found := shownAt(window, "Map"); found {
		test.Fatal("the map is shown")
	}
	if _, found := shownAt(window, "Lamp"); !found {
		test.Fatal("the lamp, placed by default, is not shown")
	}
}

func TestAWindowSurfaceOpensAndClosesWithItsAction(test *testing.T) {
	BindActions(map[string]map[string][]string{"map": {DesktopTarget: {"M"}}}, DesktopTarget)
	places := map[string]map[string]SurfacePlace{"Map": {DesktopTarget: {Place: WindowPlace, Action: "map"}}}
	window := showSurfaces(test, places, DesktopTarget, "Map")
	if _, found := shownAt(window, "Map"); found {
		test.Fatal("the map's window is open before its action")
	}
	pressed := input.NewState()
	pressed.SetKey(input.KeyM, true)
	frameWith(pressed, ctr.Reading{})
	toggleSurfaces()
	window.Frame()
	if _, found := shownAt(window, "Map"); !found {
		test.Fatal("the map's window does not open with its action")
	}
	frameWith(input.NewState(), ctr.Reading{})
	frameWith(pressed, ctr.Reading{})
	toggleSurfaces()
	window.Frame()
	if _, found := shownAt(window, "Map"); found {
		test.Fatal("the map's window does not close with its action")
	}
}

func TestSurfacesOnTheSecondScreenShareItInAColumn(test *testing.T) {
	forgetSurfaces()
	SetSurface("Map", captionSurface("Map"))
	SetSurface("Lamp", captionSurface("Lamp"))
	UseSurfaces(nil, Nintendo3DSTarget)
	window, err := uitest.New(func(context *ui.Context) { paintSecondScreen(context, Menu{}) },
		uitest.Size(320, 240), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	mapShown, mapFound := shownAt(window, "Map")
	lampShown, lampFound := shownAt(window, "Lamp")
	if !mapFound || !lampFound || lampShown.Y <= mapShown.Y {
		test.Fatalf("the map is at %+v (%v) and the lamp at %+v (%v)", mapShown, mapFound, lampShown, lampFound)
	}
}

func TestAnUnknownPlaceIsReportedAndTheSurfaceKeepsItsDefault(test *testing.T) {
	forgetSurfaces()
	SetSurface("Map", captionSurface("Map"))
	problems := UseSurfaces(map[string]map[string]SurfacePlace{"Map": {DesktopTarget: {Place: "ceiling"}}}, DesktopTarget)
	if len(problems) != 1 || !errors.Is(problems[0], SurfacePlaceUnknown) {
		test.Fatalf("the problems are %v", problems)
	}
	if place := surfacePlace("Map"); place.Place != OverPlace {
		test.Fatalf("the map is %q", place.Place)
	}
}

func TestPaintingSurfacesAllocatesNothingOfItsOwn(test *testing.T) {
	places := map[string]map[string]SurfacePlace{
		"Map":  {DesktopTarget: {Place: OverPlace, Align: "top-right"}},
		"Lamp": {DesktopTarget: {Place: WindowPlace, Open: true}},
	}
	window := showSurfaces(test, places, DesktopTarget, "Map", "Lamp")
	window.Frame()
	// the same layers built with Go UI alone: what Go UI allocates is its own business
	mapShown, lampShown := captionSurface("Map"), captionSurface("Lamp")
	byHand, err := uitest.New(func(context *ui.Context) {
		context.Overlay("Map", ui.LayerOptions{AlignX: ui.End, Padding: ui.Pad(surfaceMargin)}, func() { mapShown(context) })
		context.PanelWith("Lamp", ui.PanelOptions{X: surfaceMargin, Y: surfaceMargin}, func() { lampShown(context) })
	}, uitest.Size(gameWidth, gameHeight))
	if err != nil {
		test.Fatal(err)
	}
	byHand.Frame()
	surfaced, alone := testing.AllocsPerRun(20, window.Frame), testing.AllocsPerRun(20, byHand.Frame)
	if surfaced > alone {
		test.Fatalf("a frame of surfaces allocates %v times, and the same with Go UI alone %v", surfaced, alone)
	}
}

func TestTheGamesInterfaceIsTheSurfaceNamedInterface(test *testing.T) {
	forgetSurfaces()
	SetInterface(func(context *ui.Context, _ Menu) { context.Caption("interface shown") })
	if !HasInterface() {
		test.Fatal("the game has no interface")
	}
	UseSurfaces(nil, Nintendo3DSTarget)
	if place := surfacePlace(InterfaceSurface); place.Place != SecondScreenPlace {
		test.Fatalf("on the 3DS the interface is %q", place.Place)
	}
	places := map[string]map[string]SurfacePlace{InterfaceSurface: {DesktopTarget: {Place: OverPlace, Align: "bottom"}}}
	UseSurfaces(places, DesktopTarget)
	window, err := uitest.New(func(context *ui.Context) {
		paintSurfaces(context, Menu{}, draw2d.Rect{Width: gameWidth, Height: gameHeight}, false)
	}, uitest.Size(gameWidth, gameHeight))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	if shown, found := shownAt(window, InterfaceSurface); !found || shown.Y < gameHeight/2 {
		test.Fatalf("the interface is at %+v (%v), not at the bottom", shown, found)
	}
}

func TestASheetOpensWithItsActionAndClosesWhenThePlayerDismissesIt(test *testing.T) {
	BindActions(map[string]map[string][]string{"map": {AndroidTarget: {"M"}}}, AndroidTarget)
	places := map[string]map[string]SurfacePlace{"Map": {AndroidTarget: {Place: SheetPlace, Action: "map"}}}
	window := showSurfaces(test, places, AndroidTarget, "Map")
	pressed := input.NewState()
	pressed.SetKey(input.KeyM, true)
	frameWith(input.NewState(), ctr.Reading{})
	frameWith(pressed, ctr.Reading{})
	toggleSurfaces()
	window.Frame()
	shown, found := shownAt(window, "Map")
	if !found || shown.Y < gameHeight/2 {
		test.Fatalf("the sheet is at %+v (%v), not at the bottom", shown, found)
	}
	window.PressKey(input.KeyEscape)
	window.Frame()
	if _, found := shownAt(window, "Map"); found {
		test.Fatal("the sheet stays open when the player dismisses it")
	}
}
