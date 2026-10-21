package editor

import (
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/engine"
)

// The devices of the Controls workspace, drawn in lines as Lucide draws its icons: one thin grey
// stroke, rounded, no fills. Each drawing has its own units (scaled to the room it gets), its
// outline (a path, drawn once into a picture and kept), and its controls, each a place a click
// finds and the words it is bound with. The shapes follow each device's real proportions.

// side is where a control's actions are named: on the drawing's left edge, its right edge, or
// under the control (a key, whose neighbours leave no room for a line).
type side int

const (
	leftSide side = iota
	rightSide
	belowSide
)

// control is a button or a stick of a drawing: its name and what it is (for the panel that
// chooses its actions), its shape, the label on it, its words (one for a button, two axes for a
// stick) and where its actions are named. A control with no words only draws. A live control is
// outlined every frame instead of in the drawing's outline: one that moves (a touch control).
type control struct {
	name, about                 string
	label                       string
	x, y, width, height, radius float32
	words                       []string
	stick                       bool
	tags                        side
	live                        bool
}

// drawing is a device: the part of its units that is shown, where the actions' names go on its
// left and right, its outline and its controls.
type drawing struct {
	name                string
	left, top           float32
	width, height       float32
	tagsLeft, tagsRight float32
	outline             *draw2d.Path
	controls            []control
}

// newDrawing is a drawing with nothing in it yet.
func newDrawing(name string, left, top, width, height, tagsLeft, tagsRight float32) drawing {
	return drawing{name: name, left: left, top: top, width: width, height: height,
		tagsLeft: tagsLeft, tagsRight: tagsRight, outline: &draw2d.Path{}}
}

// add adds controls, each outlined in the drawing's outline unless it is live.
func (device *drawing) add(controls ...control) {
	for _, each := range controls {
		device.controls = append(device.controls, each)
		if !each.live {
			device.outline.RoundedRectangle(each.x, each.y, each.width, each.height, each.radius)
		}
	}
}

// addArea adds controls that are only places to click: the drawing's outline draws them its own
// way (a cross's arms, a pad's shoulders, a mouse's buttons).
func (device *drawing) addArea(controls ...control) {
	device.controls = append(device.controls, controls...)
}

// button is a round button's control.
func button(name, about, label string, centreX, centreY, radius float32, tags side, word string) control {
	return control{name: name, about: about, label: label, x: centreX - radius, y: centreY - radius,
		width: 2 * radius, height: 2 * radius, radius: radius, words: []string{word}, tags: tags}
}

// key is a rounded rectangle's control.
func key(name, about, label string, x, y, width, height, radius float32, tags side, words ...string) control {
	return control{name: name, about: about, label: label, x: x, y: y, width: width, height: height, radius: radius, words: words, tags: tags}
}

// stick is a stick's control, which moves two axes; its knob is drawn inside it.
func stick(name, about string, centreX, centreY, radius float32, tags side, across, up string) control {
	return control{name: name, about: about, x: centreX - radius, y: centreY - radius, width: 2 * radius,
		height: 2 * radius, radius: radius, words: []string{across, up}, stick: true, tags: tags}
}

// stickKnobShare is a stick's knob, as a share of its width.
const stickKnobShare = 0.6

// addCross outlines a cross about a centre and adds its four arms: up, down, left, right.
func (device *drawing) addCross(centreX, centreY, arm, thickness float32, words [4]string, about string) {
	half, reach := thickness/2, thickness/2+arm
	path := device.outline
	path.MoveTo(centreX-half, centreY-reach)
	for _, corner := range [][2]float32{{half, -reach}, {half, -half}, {reach, -half}, {reach, half}, {half, half},
		{half, reach}, {-half, reach}, {-half, half}, {-reach, half}, {-reach, -half}, {-half, -half}} {
		path.LineTo(centreX+corner[0], centreY+corner[1])
	}
	path.Close()
	names := [4]string{"Up", "Down", "Left", "Right"}
	places := [4][4]float32{{-half, -reach, thickness, arm}, {-half, half, thickness, arm}, {-reach, -half, arm, thickness}, {half, -half, arm, thickness}}
	for index, place := range places {
		device.addArea(key(names[index], about+", "+lowerFirst(names[index])+".", "",
			centreX+place[0], centreY+place[1], place[2], place[3], 2, leftSide, words[index]))
	}
}

// lowerFirst is a word with its first letter in lower case: "Up" as "up".
func lowerFirst(word string) string { return string(word[0]+'a'-'A') + word[1:] }

// consoleDrawing is a New Nintendo 3DS opened, from the front: the top screen, the hinge, and
// under it the circle pad, the cross, the touch screen, the C-stick, the face buttons, Start and
// Select, with the shoulders on the lower half's top edge. Its units are about half millimetres.
func consoleDrawing() drawing {
	device := newDrawing("3DS", 0, 10, 640, 525, 70, 570)
	path := device.outline
	path.RoundedRectangle(110, 20, 420, 238, 18)
	path.RoundedRectangle(194, 64, 252, 150, 4)
	path.Circle(320, 42, 4)
	for _, speaker := range [2]float32{0, 340} {
		for _, hole := range [5][2]float32{{146, 131}, {154, 131}, {150, 139}, {146, 147}, {154, 147}} {
			path.Circle(hole[0]+speaker, hole[1], 0.8)
		}
	}
	path.RoundedRectangle(190, 262, 260, 20, 10)
	path.RoundedRectangle(110, 286, 420, 238, 22)
	path.RoundedRectangle(220, 300, 200, 150, 4)
	path.RoundedRectangle(306, 475, 28, 12, 6) // HOME: the console's own
	path.Circle(160, 336, 24*stickKnobShare)
	path.Circle(480, 324, 6*stickKnobShare)
	device.addCross(160, 425, 16, 18, [4]string{"Up", "Down", "Left", "Right"}, "The cross")
	device.add(
		stick("Circle pad", "The stick on the left: it moves two axes.", 160, 336, 24, leftSide, "Stick X", "Stick Y"),
		stick("C-stick", "The small stick on the right, above the face buttons.", 480, 324, 6, rightSide, "Right Stick X", "Right Stick Y"),
		button("X button", "The top face button.", "X", 480, 370, 11, rightSide, "X"),
		button("Y button", "The left face button.", "Y", 454, 396, 11, rightSide, "Y"),
		button("A button", "The right face button.", "A", 506, 396, 11, rightSide, "A"),
		button("B button", "The bottom face button.", "B", 480, 422, 11, rightSide, "B"),
		button("Start", "The upper of the two small buttons under the face buttons.", "", 466, 467, 5, rightSide, "Start"),
		button("Select", "The lower of the two small buttons under the face buttons.", "", 466, 486, 5, rightSide, "Select"),
		key("L", "The left shoulder, on the lower half's top edge.", "", 114, 278, 40, 8, 4, leftSide, "L"),
		key("ZL", "The inner left shoulder, beside L.", "", 160, 278, 24, 8, 4, leftSide, "ZL"),
		key("ZR", "The inner right shoulder, beside R.", "", 456, 278, 24, 8, 4, rightSide, "ZR"),
		key("R", "The right shoulder, on the lower half's top edge.", "", 486, 278, 40, 8, 4, rightSide, "R"),
	)
	return device
}

// padDrawing is a game pad in the one layout every pad is read as: shoulders on top, the cross
// and the left stick on the left, the face buttons and the right stick on the right, Back and
// Start between.
func padDrawing() drawing {
	device := newDrawing("pad", 0, 50, 720, 400, 60, 660)
	path := device.outline
	path.MoveTo(230, 90)
	path.CurveTo(290, 80, 430, 80, 490, 90)
	path.CurveTo(560, 100, 600, 150, 615, 230)
	path.CurveTo(628, 300, 630, 360, 612, 395)
	path.CurveTo(595, 430, 555, 435, 530, 405)
	path.CurveTo(505, 375, 490, 345, 460, 340)
	path.LineTo(260, 340)
	path.CurveTo(230, 345, 215, 375, 190, 405)
	path.CurveTo(165, 435, 125, 430, 108, 395)
	path.CurveTo(90, 360, 92, 300, 105, 230)
	path.CurveTo(120, 150, 160, 100, 230, 90)
	path.MoveTo(196, 86)
	path.CurveTo(226, 70, 266, 64, 300, 66)
	path.MoveTo(524, 86)
	path.CurveTo(494, 70, 454, 64, 420, 66)
	path.Circle(290, 268, 30*stickKnobShare)
	path.Circle(430, 268, 30*stickKnobShare)
	device.addCross(230, 185, 18, 18, [4]string{"Pad Up", "Pad Down", "Pad Left", "Pad Right"}, "The pad's cross")
	device.add(
		stick("Left stick", "The pad's left stick: it moves two axes.", 290, 268, 30, leftSide, "Stick X", "Stick Y"),
		stick("Right stick", "The pad's right stick.", 430, 268, 30, rightSide, "Right Stick X", "Right Stick Y"),
		button("Y", "The pad's top face button.", "Y", 490, 159, 12, rightSide, "Pad Y"),
		button("X", "The pad's left face button.", "X", 464, 185, 12, rightSide, "Pad X"),
		button("B", "The pad's right face button.", "B", 516, 185, 12, rightSide, "Pad B"),
		button("A", "The pad's bottom face button.", "A", 490, 211, 12, rightSide, "Pad A"),
		key("Back", "The pad's left middle button.", "", 319, 179, 24, 12, 6, leftSide, "Pad Back"),
		key("Start", "The pad's right middle button.", "", 377, 179, 24, 12, 6, rightSide, "Pad Start"),
	)
	device.addArea(
		key("Left shoulder", "The pad's left shoulder.", "", 196, 60, 104, 26, 8, leftSide, "Pad Left Shoulder"),
		key("Right shoulder", "The pad's right shoulder.", "", 420, 60, 104, 26, 8, rightSide, "Pad Right Shoulder"),
	)
	return device
}

// keyboardDrawing is a compact keyboard (its rows in keyboardRows) and a mouse beside it.
func keyboardDrawing() drawing {
	const unit, gap, left, top = 36, 5, 30, 30
	device := newDrawing("keyboard", 0, 0, 790, 260, 30, 760)
	device.outline.RoundedRectangle(left, top, 16*unit+20, 5*unit+20, 14)
	for row, keys := range keyboardRows {
		x := float32(left + 10)
		y := float32(top + 10 + row*unit)
		for _, each := range keys {
			width := each.units*unit - gap
			if each.label != "" || each.word != "" {
				device.add(keyboardKey(each, x, y, width, unit-gap))
			}
			x += each.units * unit
		}
	}
	device.addMouse(left+16*unit+20+50, top+5)
	return device
}

// keyboardKey is a key's control; a key the engine does not read only draws.
func keyboardKey(each keyCap, x, y, width, height float32) control {
	name := each.word + " key"
	if each.word == "" {
		name = each.label
	}
	var words []string
	if each.word != "" {
		words = []string{each.word}
	}
	return key(name, "A key of the keyboard.", each.label, x, y, width, height, 6, belowSide, words...)
}

// addMouse adds a mouse with its top left corner at a place: two buttons and the wheel.
func (device *drawing) addMouse(x, y float32) {
	path := device.outline
	path.MoveTo(x, y+40)
	path.CurveTo(x, y+15, x+18, y, x+40, y)
	path.CurveTo(x+62, y, x+80, y+15, x+80, y+40)
	path.LineTo(x+80, y+110)
	path.CurveTo(x+80, y+140, x+62, y+155, x+40, y+155)
	path.CurveTo(x+18, y+155, x, y+140, x, y+110)
	path.Close()
	path.MoveTo(x+40, y)
	path.LineTo(x+40, y+62)
	path.MoveTo(x, y+62)
	path.LineTo(x+80, y+62)
	device.addArea(
		key("Left mouse button", "The mouse's main button.", "", x+4, y+4, 32, 56, 14, belowSide, "Mouse Left"),
		key("Right mouse button", "The mouse's second button.", "", x+44, y+4, 32, 56, 14, belowSide, "Mouse Right"),
	)
	device.add(key("Mouse wheel", "Pressing the wheel.", "", x+35, y+16, 10, 24, 5, rightSide, "Mouse Middle"))
}

// keyCap is a key on the keyboard drawn: its label, its width in keys, and its word (empty for a
// key the engine does not read, and for a gap).
type keyCap struct {
	label string
	units float32
	word  string
}

// keyboardRows are a compact keyboard's rows, with the arrows and a few keys on its right.
var keyboardRows = [][]keyCap{
	append(append([]keyCap{{"Esc", 1, "Escape"}}, letters("1234567890")...),
		keyCap{"-", 1, "Minus"}, keyCap{"=", 1, "Plus"}, keyCap{"⌫", 2, "Backspace"}, keyCap{"Home", 1, "Home"}),
	append(append([]keyCap{{"Tab", 1.5, "Tab"}}, letters("QWERTYUIOP")...),
		keyCap{"[", 1, ""}, keyCap{"]", 1, ""}, keyCap{"\\", 1.5, ""}, keyCap{"PgUp", 1, "Page Up"}),
	append(append([]keyCap{{"Caps", 1.75, ""}}, letters("ASDFGHJKL")...),
		keyCap{";", 1, ""}, keyCap{"'", 1, ""}, keyCap{"Enter", 2.25, "Enter"}, keyCap{"PgDn", 1, "Page Down"}),
	append(append([]keyCap{{"Shift", 2.25, "Shift"}}, letters("ZXCVBNM")...),
		keyCap{",", 1, ""}, keyCap{".", 1, ""}, keyCap{"/", 1, ""}, keyCap{"Shift", 1.75, "Shift"}, keyCap{"↑", 1, "Up"}, keyCap{"End", 1, "End"}),
	{{"Ctrl", 1.25, "Control"}, {"Super", 1.25, ""}, {"Alt", 1.25, "Alt"}, {"", 6.25, "Space"}, {"Alt", 1.25, "Alt"},
		{"Ctrl", 1.25, "Control"}, {"", 0.5, ""}, {"←", 1, "Left"}, {"↓", 1, "Down"}, {"→", 1, "Right"}},
}

// letters are keys one wide whose label is their word, one for each character.
func letters(characters string) []keyCap {
	keys := make([]keyCap, 0, len(characters))
	for _, character := range characters {
		keys = append(keys, keyCap{string(character), 1, string(character)})
	}
	return keys
}

// The phone drawn: its screen, where the touch controls lie, in its units.
const (
	phoneScreenX      = 97
	phoneScreenY      = 57
	phoneScreenWidth  = 506
	phoneScreenHeight = 231
)

// phoneDrawing is a phone held sideways, with the system's back on its edge; its touch controls
// are added by withTouchControls.
func phoneDrawing() drawing {
	device := newDrawing("phone", 0, 30, 700, 280, 70, 630)
	path := device.outline
	path.RoundedRectangle(90, 50, 520, 245, 30)
	path.RoundedRectangle(phoneScreenX, phoneScreenY, phoneScreenWidth, phoneScreenHeight, 24)
	path.MoveTo(251, 45)
	path.LineTo(282, 45)
	path.MoveTo(308, 45)
	path.LineTo(371, 45)
	path.MoveTo(121, 164)
	path.LineTo(115, 172)
	path.LineTo(121, 180)
	device.addArea(button("Back", "The system's back: the gesture from the edge, or its button.", "", 118, 172, 10, leftSide, "Back"))
	return device
}

// withTouchControls is the phone with the touch controls the game draws on its screen. With none
// of its own the game has the engine's four buttons, which hold keys.
func withTouchControls(phone drawing, touch []engine.TouchControl) drawing {
	phone.controls = phone.controls[:1:1] // the back; a copy from here on
	if len(touch) == 0 {
		for _, each := range engineButtons {
			each.live = true
			phone.controls = append(phone.controls, each)
		}
		return phone
	}
	for _, each := range touch {
		phone.controls = append(phone.controls, touchControlDrawn(each))
	}
	return phone
}

// touchControlDrawn is a touch control on the phone drawn.
func touchControlDrawn(each engine.TouchControl) control {
	centreX := phoneScreenX + each.X*phoneScreenWidth
	centreY := phoneScreenY + each.Y*phoneScreenHeight
	radius := each.Size * phoneScreenHeight / 2
	tags := rightSide
	if each.X < 0.5 {
		tags = leftSide
	}
	words := engine.TouchWords(each)
	drawn := button(each.Name, "A button drawn on the screen.", each.Name, centreX, centreY, radius, tags, words[0])
	if each.Kind == engine.TouchStick {
		drawn = stick(each.Name, "A stick drawn on the screen: the thumb drags it each way.", centreX, centreY, radius, tags, words[0], words[1])
	}
	drawn.live = true
	return drawn
}

// engineButtons are the engine's four touch buttons, which hold the keys Left, Right, Escape and
// Space.
var engineButtons = []control{
	button("Left", "The engine's touch button on the left: it holds the key Left.", "←", 170, 246, 18, leftSide, "Left"),
	button("Right", "The engine's touch button beside it: it holds the key Right.", "→", 220, 246, 18, leftSide, "Right"),
	button("B", "The engine's touch button on the right: it holds the key Escape.", "B", 476, 246, 18, rightSide, "Escape"),
	button("A", "The engine's last touch button: it holds the key Space.", "A", 526, 246, 18, rightSide, "Space"),
}
