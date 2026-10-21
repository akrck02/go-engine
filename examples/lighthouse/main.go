// Lighthouse, a sample of go-engine for 3D games driven by code: the beam, the fog and the buoy are
// meshes the behaviours build, lit additively, translucent and unlit; the game draws its own 2D
// over the scene; and the 3DS's bottom screen shows a chart, a scene of its own, under the
// game's interface. Play it with `go run .`; edit it with `go run ../../cmd/go-engine .`.
package main

import (
	"embed"
	"log"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// files are the project's settings and scenes, carried inside the game.
//
//go:embed engine.json scenes
var files embed.FS

func main() {
	engine.Register(Beam{}, Fog{}, Marker{})
	chart, err := engine.Load(files, "scenes/chart.scene.json")
	if err != nil {
		log.Fatal(err)
	}
	engine.SetBottomScene(chart)
	engine.SetOverlay(drawTitle)
	engine.SetInterface(chartCaption)
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}

// The title's place and size, in the screen's pixels.
const (
	titleMargin = 12
	titleSize   = 18
)

// drawTitle is the game's own 2D over the scene: a band with the sample's name.
func drawTitle(list *draw2d.List, width, height float32) {
	band := draw2d.Rect{X: titleMargin, Y: titleMargin, Width: width - 2*titleMargin, Height: titleSize + titleMargin}
	list.Rectangle(band, 6, draw2d.Color{0, 0, 0, 0.45}, draw2d.Transparent, 0)
	list.Text("Lighthouse · the beam turns", band.X+titleMargin, band.Y+titleSize+2, titleSize, draw2d.Medium, draw2d.Hex("#f5ead0"))
}

// chartCaption is the game's interface: a caption on the 3DS's chart, or over the window.
func chartCaption(context *ui.Context, _ engine.Menu) {
	context.Box(ui.Layout{Padding: ui.Pad(8)}, func() {
		context.Caption("The chart turns with the lamp")
	})
}
