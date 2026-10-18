package engine

import (
	"image"
	"image/draw"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
)

// A sprite with a sheet (scenefile.SpriteSheet) is cut into frames of equal size and shows one at
// a time: the frame its animation is on. A behaviour changes the animation with Play and the way
// the sprite faces with FaceLeft; the engine moves every animation on, frame by frame.

// Sheet is the sprite sheet the node's sprite is cut into, to edit in place; a sprite shown whole
// has none.
func (node *Node) Sheet() (*scenefile.SpriteSheet, bool) { return node.sheet, node.sheet != nil }

// SetSheet cuts the node's sprite into a sheet's frames, or shows it whole again (nil).
func (node *Node) SetSheet(sheet *scenefile.SpriteSheet) {
	node.sheet = sheet
	node.playing, node.playingFor = scenefile.Animation{}, 0
}

// Play starts an animation of the node's sheet by its name. The animation already playing goes on
// where it is, so a behaviour may ask for "walk" every frame it walks. It tells whether the sheet
// has the animation.
func (node *Node) Play(name string) bool {
	if node.sheet == nil {
		return false
	}
	if name == node.playing.Name {
		return true
	}
	animation, found := node.sheet.Find(name)
	if !found {
		return false
	}
	node.playing, node.playingFor = animation, 0
	return true
}

// Animation is the name of the animation the node plays, or "".
func (node *Node) Animation() string { return node.playing.Name }

// Frame is the frame of the sheet the node shows now, from 0, the top-left one.
func (node *Node) Frame() int {
	frames := node.playing.Frames
	if len(frames) == 0 {
		return 0
	}
	step := int(node.playingFor * node.playing.FramesPerSecond)
	if node.playing.Loop {
		return frames[step%len(frames)]
	}
	return frames[min(step, len(frames)-1)]
}

// FaceLeft turns the node's sprite to face left (true) or right (false, as it was drawn).
func (node *Node) FaceLeft(left bool) { node.facingLeft = left }

// FacingLeft tells whether the node's sprite faces left.
func (node *Node) FacingLeft() bool { return node.facingLeft }

// startAnimations plays each sheet's first animation.
func (owner *Scene) startAnimations(node *Node) {
	if node.sheet != nil && node.sheet.Start != "" {
		node.Play(node.sheet.Start)
	}
	for _, child := range node.children {
		owner.startAnimations(child)
	}
}

// advanceAnimations moves every playing animation on by some seconds.
func (owner *Scene) advanceAnimations(node *Node, seconds float64) {
	if node.playing.Name != "" {
		node.playingFor += seconds
	}
	for _, child := range node.children {
		owner.advanceAnimations(child, seconds)
	}
}

// spritePicture is the picture a sprite shows now: the whole picture, or its sheet's frame, turned
// to face left when it does.
func (owner *Scene) spritePicture(node *Node) (*draw2d.Image, bool) {
	if node.sheet == nil {
		picture, err := owner.assets.image(node.sprite)
		return picture, err == nil
	}
	frames, err := owner.assets.sheetFrames(node.sprite, node.sheet, node.facingLeft)
	if err != nil || len(frames) == 0 {
		return nil, false
	}
	return frames[min(node.Frame(), len(frames)-1)], true
}

// framePixels is how many pixels wide and high a sprite's frame is.
func (owner *Scene) framePixels(node *Node) (width, height int, found bool) {
	picture, err := owner.assets.image(node.sprite)
	if err != nil {
		return 0, 0, false
	}
	width, height = picture.Size()
	if node.sheet == nil || node.sheet.Columns <= 0 || node.sheet.Rows <= 0 {
		return width, height, true
	}
	return width / node.sheet.Columns, height / node.sheet.Rows, true
}

// sheetKey names a picture cut into a sheet's frames, facing one way.
type sheetKey struct {
	file          string
	columns, rows int
	facingLeft    bool
}

// sheetFrames reads a picture and cuts it into a sheet's frames, mirrored when facing left, once.
func (library *assets) sheetFrames(file string, sheet *scenefile.SpriteSheet, facingLeft bool) ([]*draw2d.Image, error) {
	key := sheetKey{file: file, columns: sheet.Columns, rows: sheet.Rows, facingLeft: facingLeft}
	if frames, found := library.sheets[key]; found {
		return frames, nil
	}
	whole, err := library.image(file)
	if err != nil {
		return nil, err
	}
	frames := cutIntoFrames(whole.Pixels, sheet.Columns, sheet.Rows, facingLeft)
	library.sheets[key] = frames
	return frames, nil
}

// cutIntoFrames cuts a picture into a grid of equal frames, left to right and top to bottom, each
// mirrored when asked.
func cutIntoFrames(picture *image.NRGBA, columns, rows int, mirrored bool) []*draw2d.Image {
	bounds := picture.Bounds()
	width, height := bounds.Dx()/max(columns, 1), bounds.Dy()/max(rows, 1)
	var frames []*draw2d.Image
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			frame := image.NewNRGBA(image.Rect(0, 0, width, height))
			draw.Draw(frame, frame.Bounds(), picture, image.Pt(bounds.Min.X+column*width, bounds.Min.Y+row*height), draw.Src)
			if mirrored {
				mirror(frame)
			}
			frames = append(frames, draw2d.NewImage(frame))
		}
	}
	return frames
}

// mirror turns a picture left to right, in place.
func mirror(picture *image.NRGBA) {
	bounds := picture.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for left, right := bounds.Min.X, bounds.Max.X-1; left < right; left, right = left+1, right-1 {
			leftColour, rightColour := picture.NRGBAAt(left, y), picture.NRGBAAt(right, y)
			picture.SetNRGBA(left, y, rightColour)
			picture.SetNRGBA(right, y, leftColour)
		}
	}
}
