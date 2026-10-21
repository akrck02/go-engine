package editor

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"

	"github.com/akrck02/go-engine/engine"
)

// inside tells whether a rectangle lies in another.
func inside(inner, outer draw2d.Rect) bool {
	return inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.Width <= outer.X+outer.Width && inner.Y+inner.Height <= outer.Y+outer.Height
}

func TestPlayShowsTheGamesSurfacesWhereEachTargetPlacesThem(test *testing.T) {
	editor := open(test)
	editor.Project.Settings.Surfaces = map[string]map[string]engine.SurfacePlace{"Map": {
		engine.Nintendo3DSTarget: {Place: engine.SecondScreenPlace},
		engine.DesktopTarget:     {Place: engine.OverPlace, Align: "bottom-right"},
		engine.AndroidTarget:     {Place: engine.OverPlace, Align: "top-left"},
	}}
	engine.SetSurface("Map", func(context *ui.Context) { context.Caption("Map shown") })
	defer engine.SetSurface("Map", nil)
	editor.ShowWorkspace(PlayWorkspace)
	view := NewView(editor)
	window, err := uitest.New(view.Build, uitest.Size(1280, 800), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	for _, target := range []PlayTarget{WideScreen, Nintendo3DSScreens, PhoneScreen} {
		editor.SetTarget(target)
		window.Frame()
		window.Frame() // the stage is laid out, then the surfaces placed in it
		shown, err := window.Element("Map shown")
		if err != nil || !shown.Visible {
			test.Fatalf("on %v the map is not shown: %v", target, err)
		}
		screen := view.staged.top
		if target == Nintendo3DSScreens {
			screen = view.staged.bottom
		}
		if !inside(shown.Rect, screen) {
			test.Fatalf("on %v the map is at %+v, not on its screen %+v", target, shown.Rect, screen)
		}
		middle := screen.X + screen.Width/2
		if target == WideScreen && shown.Rect.X < middle || target == PhoneScreen && shown.Rect.X > middle {
			test.Fatalf("on %v the map is at %+v, not on its side of %+v", target, shown.Rect, screen)
		}
	}
}
