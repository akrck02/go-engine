package engine

import (
	"sync"

	"github.com/akrck02/go-renderer/opengl/overlay"
)

// A game's own fonts: its text in the look it was drawn for (pixel fonts, a brand's), wherever the
// engine paints text for it: the game's drawing over the scene, its interface and surfaces, and
// the second screen, on every target.

// gameFonts are the fonts the game chose (SetFonts); empty ones keep Go UI's own.
var gameFonts overlay.Fonts

// SetFonts makes the game's text be drawn and measured in these fonts, a font file for each weight
// (overlay.Fonts: an empty weight keeps the default one). Call it before Run.
func SetFonts(fonts overlay.Fonts) {
	gameFonts = fonts
	measurerOnce = sync.Once{} // a measurer made before measures in the old fonts
}

// measurer is the measurer of the game's fonts, made once when first asked for.
var (
	measurerOnce sync.Once
	measurer     *overlay.Measurer
	measurerErr  error
)

// TextMeasurer measures text in the game's fonts (SetFonts), as the engine draws it: for a layout
// of the game's own that has to know a line's width. It is made once and kept.
func TextMeasurer() (*overlay.Measurer, error) {
	measurerOnce.Do(func() { measurer, measurerErr = overlay.NewMeasurer(gameFonts) })
	return measurer, measurerErr
}
