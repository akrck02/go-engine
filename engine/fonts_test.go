package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl/overlay"
	"golang.org/x/image/font/gofont/gomono"
)

func TestAGamesFontsMeasureItsTextAndReachItsWindow(test *testing.T) {
	mono := overlay.Fonts{Regular: gomono.TTF}
	SetFonts(mono)
	defer SetFonts(overlay.Fonts{})
	measurer, err := TextMeasurer()
	if err != nil {
		test.Fatal(err)
	}
	narrow := measurer.MeasureText("iii", 16, draw2d.Regular, 1).Width
	wide := measurer.MeasureText("mmm", 16, draw2d.Regular, 1).Width
	if narrow != wide {
		test.Fatalf("in a monospaced font the lines measure %v and %v: the game's font is not used", narrow, wide)
	}
	game := &runner{played: load(test)}
	if options := game.windowOptions("Game"); string(options.Fonts.Regular) != string(mono.Regular) {
		test.Fatal("the window opens without the game's fonts")
	}
}
