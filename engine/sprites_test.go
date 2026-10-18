package engine

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
)

// sheetPicture is a sprite sheet of four 8-pixel frames in a row, 32 × 8; each frame is a shade of
// its own, and darker on its left half, so a mirrored frame shows it.
func sheetPicture() []byte {
	picture := image.NewNRGBA(image.Rect(0, 0, 32, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 32; x++ {
			shade := uint8(40 + 60*(x/8))
			if x%8 < 4 {
				shade /= 2
			}
			picture.SetNRGBA(x, y, color.NRGBA{R: shade, G: shade, B: shade, A: 255})
		}
	}
	var encoded bytes.Buffer
	png.Encode(&encoded, picture)
	return encoded.Bytes()
}

// walkerScene is a 2D scene at 8 pixels a metre with a hero whose sheet walks through its four
// frames at four a second, looping, and jumps on frame 2 once.
func walkerScene(test *testing.T) *Scene {
	description := &scenefile.Scene{Name: "Walk", Mode: scenefile.Mode2D, PixelsPerMetre: 8, Nodes: []scenefile.Node{{
		Name: "Hero", Sprite: "assets/hero.png", Scale: [3]float64{1, 1, 1},
		Sheet: &scenefile.SpriteSheet{Columns: 4, Rows: 1, Start: "walk", Animations: []scenefile.Animation{
			{Name: "walk", Frames: []int{0, 1, 2, 3}, FramesPerSecond: 4, Loop: true},
			{Name: "jump", Frames: []int{2, 3}, FramesPerSecond: 4},
		}},
	}}}
	built, err := Build(fstest.MapFS{"assets/hero.png": {Data: sheetPicture()}}, description)
	if err != nil {
		test.Fatal(err)
	}
	built.Start()
	return built
}

func TestAnAnimationShowsEachFrameForItsTimeAndLoops(test *testing.T) {
	scene := walkerScene(test)
	hero, _ := scene.Find("Hero")
	expected := []struct {
		seconds float64
		frame   int
	}{{0.1, 0}, {0.3, 1}, {0.55, 2}, {0.8, 3}, {1.1, 0}}
	elapsed := 0.0
	for _, step := range expected {
		playFor(scene, step.seconds-elapsed)
		elapsed = step.seconds
		if hero.Frame() != step.frame {
			test.Fatalf("at %vs the hero shows frame %d, not %d", step.seconds, hero.Frame(), step.frame)
		}
	}
}

func TestAnAnimationThatDoesNotLoopStaysOnItsLastFrame(test *testing.T) {
	scene := walkerScene(test)
	hero, _ := scene.Find("Hero")
	hero.Play("jump")
	playFor(scene, 2)
	if hero.Frame() != 3 || hero.Animation() != "jump" {
		test.Fatalf("after its end the jump shows frame %d of %q", hero.Frame(), hero.Animation())
	}
}

func TestPlayingTheAnimationAlreadyPlayingDoesNotStartItAgain(test *testing.T) {
	scene := walkerScene(test)
	hero, _ := scene.Find("Hero")
	playFor(scene, 0.55)
	hero.Play("walk")
	if hero.Frame() != 2 {
		test.Fatalf("playing walk again went back to frame %d", hero.Frame())
	}
	if hero.Play("fly") {
		test.Fatal("an animation the sheet does not have was played")
	}
}

func TestASpriteSheetNodeIsOneFrameLarge(test *testing.T) {
	scene := walkerScene(test)
	hero, _ := scene.Find("Hero")
	if size := hero.Bounds().Size(); size.X != 1 || size.Y != 1 {
		test.Fatalf("the hero is %v metres: a frame of 8 pixels at 8 a metre is 1", size)
	}
}

func TestASpriteFacingLeftIsDrawnMirrored(test *testing.T) {
	scene := walkerScene(test)
	hero, _ := scene.Find("Hero")
	area := draw2d.Rect{Width: 80, Height: 80}
	var right, left draw2d.List
	scene.DrawContent2D(&right, area, View2D{Height: 10})
	hero.FaceLeft(true)
	scene.DrawContent2D(&left, area, View2D{Height: 10})
	facingRight, facingLeft := right.Commands[0].Picture.Pixels, left.Commands[0].Picture.Pixels
	if !hero.FacingLeft() || facingRight.NRGBAAt(0, 0) != facingLeft.NRGBAAt(7, 0) || facingRight.NRGBAAt(0, 0) == facingLeft.NRGBAAt(0, 0) {
		test.Fatal("facing left did not draw the frame mirrored")
	}
}

func TestAnimatingAFrameAllocatesNothing(test *testing.T) {
	scene := walkerScene(test)
	var list draw2d.List
	area := draw2d.Rect{Width: 80, Height: 80}
	playFor(scene, 1)
	scene.DrawContent2D(&list, area, View2D{Height: 10})
	allocations := testing.AllocsPerRun(100, func() {
		scene.Update(1.0/60, nil)
		list.Reset()
		scene.DrawContent2D(&list, area, View2D{Height: 10})
	})
	if allocations != 0 {
		test.Fatalf("an animated frame allocates %v times", allocations)
	}
}
