package engine

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Touch controls are the sticks and buttons a game draws over its view on a touch screen with no
// keyboard (a phone), as its author placed them in the editor (engine.json's touchControls).
// Their places and sizes are shares of the screen, so one layout fits every phone. Actions bind
// them by name: "Touch Jump" presses, "Touch Move X" and "Touch Move Y" are a stick's axes.
// A game that places none keeps the engine's four touch buttons (touch.go).

// TouchKind is what a touch control is.
type TouchKind string

// The kinds of touch control.
const (
	TouchButton TouchKind = "button" // pressed while a finger is on it
	TouchStick  TouchKind = "stick"  // moves two axes as a finger drags it
)

// TouchControl is a stick or a button drawn on a touch screen: its name, what it is, where its
// centre sits (X across the screen and Y down it, from 0 to 1) and its size (its width, as a
// share of the screen's shorter side).
type TouchControl struct {
	Name string    `json:"name"`
	Kind TouchKind `json:"kind"`
	X    float32   `json:"x"`
	Y    float32   `json:"y"`
	Size float32   `json:"size"`
}

// The words of touch controls.
const (
	touchWordPrefix = "Touch "
	touchAxisX      = " X"
	touchAxisY      = " Y"
)

// TouchWords are the words a touch control is bound with: a button's name, a stick's two axes.
func TouchWords(control TouchControl) []string {
	word := touchWordPrefix + control.Name
	if control.Kind == TouchStick {
		return []string{word + touchAxisX, word + touchAxisY}
	}
	return []string{word}
}

// touchAxisNone is the code of a touch input that presses; a stick's axes are 0 (across) and 1
// (up).
const touchAxisNone = -1

// readTouchWord is the touch control a word names: its name, and the axis for a stick's.
func readTouchWord(word string) (boundInput, bool) {
	name, found := strings.CutPrefix(word, touchWordPrefix)
	if !found || name == "" {
		return boundInput{}, false
	}
	if stick, isX := strings.CutSuffix(name, touchAxisX); isX {
		return boundInput{kind: touchInput, name: stick, code: 0}, true
	}
	if stick, isY := strings.CutSuffix(name, touchAxisY); isY {
		return boundInput{kind: touchInput, name: stick, code: 1}, true
	}
	return boundInput{kind: touchInput, name: name, code: touchAxisNone}, true
}

// The touch controls' look.
const (
	touchReachShare  = 1.5 // a stick follows a finger this far from its centre, as a share of its radius
	touchKnobShare   = 0.45
	touchLabelShare  = 0.4 // a button's name, as a share of its width
	touchLetterShare = 0.56
	touchLineWidth   = 2
)

// touchPanel is the game's touch controls and what the fingers do with them this frame; it is
// read by the actions and kept from frame to frame without allocating.
var touchPanel struct {
	controls []TouchControl
	areas    []draw2d.Rectangle // where each was drawn last, in the window
	held     []bool             // a finger is on a button, or on a stick
	axes     [][2]float64       // a stick's axes, from -1 to 1, up positive
}

// useTouchControls makes some touch controls the game's.
func useTouchControls(controls []TouchControl) {
	touchPanel.controls = controls
	touchPanel.areas = make([]draw2d.Rectangle, len(controls))
	touchPanel.held = make([]bool, len(controls))
	touchPanel.axes = make([][2]float64, len(controls))
}

// placeTouchControls works out where each touch control goes on a screen's safe part: its size a
// share of the screen's shorter side, so it is as big held upright as held sideways, and kept
// whole on the screen.
func placeTouchControls(safe draw2d.Rectangle) {
	shorter := min(safe.Width, safe.Height)
	for index, control := range touchPanel.controls {
		side := min(control.Size*shorter, shorter)
		centreX := min(max(safe.X+control.X*safe.Width, safe.X+side/2), safe.X+safe.Width-side/2)
		centreY := min(max(safe.Y+control.Y*safe.Height, safe.Y+side/2), safe.Y+safe.Height-side/2)
		touchPanel.areas[index] = draw2d.Rectangle{X: centreX - side/2, Y: centreY - side/2, Width: side, Height: side}
	}
}

// readTouchControls reads the fingers on the screen: which buttons they hold, and how far they
// drag each stick from its centre.
func readTouchControls(touches []input.Touch) {
	for index, control := range touchPanel.controls {
		touchPanel.held[index], touchPanel.axes[index] = false, [2]float64{}
		area := touchPanel.areas[index]
		radius := float64(area.Width) / 2
		centreX, centreY := float64(area.X)+radius, float64(area.Y)+radius
		for _, finger := range touches {
			acrossX, downY := finger.X-centreX, finger.Y-centreY
			if control.Kind == TouchButton && area.Contains(float32(finger.X), float32(finger.Y)) {
				touchPanel.held[index] = true
				break
			}
			if control.Kind == TouchStick && math.Hypot(acrossX, downY) <= radius*touchReachShare {
				touchPanel.held[index] = true
				touchPanel.axes[index] = [2]float64{clampUnit(acrossX / radius), clampUnit(-downY / radius)}
				break
			}
		}
	}
}

// clampUnit keeps a value between -1 and 1.
func clampUnit(value float64) float64 { return math.Max(-1, math.Min(1, value)) }

// touchControlIndex is where a touch control is among the game's, by its name.
func touchControlIndex(name string) (int, bool) {
	for index, control := range touchPanel.controls {
		if control.Name == name {
			return index, true
		}
	}
	return 0, false
}

// touchValue is a touch input this frame: whether it is held, and a stick axis's value.
func touchValue(bound boundInput) (held bool, value float64) {
	index, found := touchControlIndex(bound.name)
	if !found {
		return false, 0
	}
	if bound.code == touchAxisNone {
		return touchPanel.held[index], 0
	}
	return false, touchPanel.axes[index][bound.code]
}

// drawTouchControls draws the touch controls: rings, a stick's knob where the finger holds it, a
// button's name in it.
func drawTouchControls(list *draw2d.List) {
	ink, shade := draw2d.Color{1, 1, 1, touchButtonAlpha * 2}, draw2d.Color{0, 0, 0, touchButtonAlpha}
	for index, control := range touchPanel.controls {
		area := touchPanel.areas[index]
		fill := shade
		if touchPanel.held[index] {
			fill = draw2d.Color{1, 1, 1, touchButtonAlpha}
		}
		list.Rectangle(area, area.Width/2, fill, ink, touchLineWidth)
		if control.Kind == TouchStick {
			drawTouchKnob(list, area, touchPanel.axes[index], ink)
			continue
		}
		size := area.Width * touchLabelShare / float32(max(1, utf8.RuneCountInString(control.Name)))
		size = min(size*2, area.Width*touchLabelShare)
		width := float32(utf8.RuneCountInString(control.Name)) * size * touchLetterShare
		list.Text(control.Name, area.X+(area.Width-width)/2, area.Y+area.Height/2+size*0.35, size, draw2d.Medium, ink)
	}
}

// drawTouchKnob draws a stick's knob, moved as far as its axes say.
func drawTouchKnob(list *draw2d.List, area draw2d.Rectangle, axes [2]float64, ink draw2d.Color) {
	knob := area.Width * touchKnobShare
	radius := area.Width / 2
	centreX := area.X + radius + float32(axes[0])*(radius-knob/2)
	centreY := area.Y + radius - float32(axes[1])*(radius-knob/2)
	list.Rectangle(draw2d.Rectangle{X: centreX - knob/2, Y: centreY - knob/2, Width: knob, Height: knob}, knob/2, ink, draw2d.Color{}, 0)
}
