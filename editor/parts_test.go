package editor

import (
	"errors"
	"slices"
	"testing"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// openFlat opens a new 2D scene of a new project, with its starter square chosen.
func openFlat(test *testing.T) *Editor {
	editor := open(test)
	file, err := editor.Project.NewScene("Flat", scenefile.Mode2D)
	if err != nil || editor.OpenScene(file) != nil {
		test.Fatal(err)
	}
	editor.Select(editor.Add("Square"))
	return editor
}

// saveAndOpenAgain saves the scene, opens it again and finds a node of it.
func saveAndOpenAgain(test *testing.T, editor *Editor, path string) *engine.Node {
	if editor.Save() != nil {
		test.Fatal("the scene was not saved")
	}
	if err := editor.OpenScene(editor.SceneFile); err != nil {
		test.Fatal(err)
	}
	node, found := editor.Scene.Find(path)
	if !found {
		test.Fatalf("%s is gone", path)
	}
	return node
}

func TestABodySetInTheInspectorIsSaved(test *testing.T) {
	editor := openFlat(test)
	editor.GiveBody(scenefile.Solid)
	editor.GiveBody(scenefile.Moving)
	body, _ := editor.Selected.BodySettings()
	body.Size = [2]float64{0.8, 0.6}
	editor.SetGravityScale(0.5)
	if editor.Saved {
		test.Fatal("giving a body did not change the scene")
	}
	square := saveAndOpenAgain(test, editor, editor.Selected.Path())
	saved, found := square.BodySettings()
	if !found || saved.Kind != scenefile.Moving || saved.Size != [2]float64{0.8, 0.6} || saved.Gravity() != 0.5 {
		test.Fatalf("the body came back as %+v", saved)
	}
	editor.Select(square)
	editor.TakeBody()
	if _, found := square.BodySettings(); found {
		test.Fatal("the body was not taken")
	}
}

func TestASpriteSheetAndItsAnimationsAreSaved(test *testing.T) {
	editor := openFlat(test)
	editor.ResizeSheet(4, 2)
	walk := editor.AddAnimation()
	if !editor.RenameAnimation(walk, "walk") || editor.SetAnimationFrames(walk, "0-3") != nil {
		test.Fatal("the animation was not set")
	}
	jump := editor.AddAnimation()
	if editor.RenameAnimation(jump, "walk") || !editor.RenameAnimation(jump, "jump") || editor.SetAnimationFrames(jump, "4, 7") != nil {
		test.Fatal("two animations took the same name")
	}
	square := saveAndOpenAgain(test, editor, editor.Selected.Path())
	sheet, found := square.Sheet()
	if !found || sheet.Columns != 4 || sheet.Rows != 2 || sheet.Start != "walk" || len(sheet.Animations) != 2 {
		test.Fatalf("the sheet came back as %+v", sheet)
	}
	if !slices.Equal(sheet.Animations[0].Frames, []int{0, 1, 2, 3}) || !slices.Equal(sheet.Animations[1].Frames, []int{4, 7}) {
		test.Fatalf("the frames came back as %v", sheet.Animations)
	}
}

func TestASmallerSheetKeepsOnlyTheFramesItHas(test *testing.T) {
	editor := openFlat(test)
	editor.ResizeSheet(4, 2)
	jump := editor.AddAnimation()
	editor.SetAnimationFrames(jump, "2, 6, 7")
	editor.ResizeSheet(2, 2)
	sheet, _ := editor.Selected.Sheet()
	if !slices.Equal(sheet.Animations[0].Frames, []int{2}) {
		test.Fatalf("a sheet of 4 frames kept %v", sheet.Animations[0].Frames)
	}
	editor.RemoveAnimation(jump)
	if len(sheet.Animations) != 0 || sheet.Start != "" {
		test.Fatalf("the animation was not removed: %+v", sheet)
	}
}

func TestFramesAreReadAsTyped(test *testing.T) {
	read, err := ReadFrames(" 0-3, 6 ,5 ", 8)
	if err != nil || !slices.Equal(read, []int{0, 1, 2, 3, 6, 5}) {
		test.Fatalf("read %v, %v", read, err)
	}
	for _, wrong := range []string{"", "a", "3-1", "-2", "1,,2", "9", "0-300000000"} {
		if _, err := ReadFrames(wrong, 8); !errors.Is(err, FramesWrong) {
			test.Errorf("%q: %v", wrong, err)
		}
	}
}

func TestAFollowingCameraIsSaved(test *testing.T) {
	editor := openFlat(test)
	followed := editor.Selected.Path()
	camera, _ := editor.Scene.Find("Camera")
	if !slices.Contains(editor.Followable(camera), followed) || slices.Contains(editor.Followable(camera), camera.Path()) {
		test.Fatalf("a camera may follow %v", editor.Followable(camera))
	}
	settings, _ := camera.Camera()
	settings.Follow, settings.SmoothingSeconds, settings.Ceiling = followed, 0.25, true
	settings.Bounds = &[4]float64{-10, -5, 40, 15}
	saved, _ := saveAndOpenAgain(test, editor, "Camera").Camera()
	if saved.Follow != followed || saved.SmoothingSeconds != 0.25 || !saved.Ceiling || *saved.Bounds != [4]float64{-10, -5, 40, 15} {
		test.Fatalf("the camera came back as %+v", saved)
	}
}

func TestFramesAreWrittenAsTheyAreRead(test *testing.T) {
	for _, written := range []string{"0-3, 6", "4", "2, 1, 0", "0-1, 5-7"} {
		frames, _ := ReadFrames(written, 8)
		if writtenFrames(frames) != written {
			test.Errorf("%q came back as %q", written, writtenFrames(frames))
		}
	}
}

func TestFramesWrittenAnotherWayAreNotAChange(test *testing.T) {
	editor := openFlat(test)
	editor.ResizeSheet(4, 1)
	walk := editor.AddAnimation()
	editor.SetAnimationFrames(walk, "0-2")
	editor.Save()
	sheet, _ := editor.Selected.Sheet()
	if editor.SetAnimationFramesIfChanged(walk, "0, 1, 2", sheet.Animations[walk].Frames) != nil || !editor.Saved {
		test.Fatal("the same frames written another way changed the scene")
	}
}
