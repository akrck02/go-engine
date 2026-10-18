package editor

import (
	"slices"
	"strconv"
	"strings"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/scenefile"
)

// The parts a node of a 2D scene may have besides its sprite or shape: a body in the physics, and
// a sheet that cuts its sprite into animated frames. The inspector shows them; these change them,
// on the chosen node, and keep the scene file valid (an animation never shows a frame the sheet
// does not have, nor shares a name).

// newAnimationFramesPerSecond is the pace of a new animation.
const newAnimationFramesPerSecond = 8

// GiveBody gives the chosen node a body of a kind, or changes the kind of the one it has. Its size
// is the node's own until it is set.
func (editor *Editor) GiveBody(kind scenefile.BodyKind) {
	node := editor.Selected
	if node == nil {
		return
	}
	body, found := node.BodySettings()
	if !found {
		body = &scenefile.Body{}
		node.SetBodySettings(body)
	}
	body.Kind = kind
	editor.Changed()
}

// TakeBody takes the chosen node's body away.
func (editor *Editor) TakeBody() {
	if editor.Selected == nil {
		return
	}
	editor.Selected.SetBodySettings(nil)
	editor.Changed()
}

// SetGravityScale sets how much gravity pulls the chosen node's body: 1 fully, 0 not at all.
func (editor *Editor) SetGravityScale(scale float64) {
	body, found := editor.selectedBody()
	if !found {
		return
	}
	body.GravityScale = &scale
	if scale == 1 {
		body.GravityScale = nil // 1 is the default, which the scene file leaves out
	}
	editor.Changed()
}

func (editor *Editor) selectedBody() (*scenefile.Body, bool) {
	if editor.Selected == nil {
		return nil, false
	}
	return editor.Selected.BodySettings()
}

// ResizeSheet cuts the chosen node's sprite into columns and rows of frames, at least one of each.
// The animations keep only the frames the sheet still has.
func (editor *Editor) ResizeSheet(columns, rows int) {
	node := editor.Selected
	if node == nil {
		return
	}
	sheet, found := node.Sheet()
	if !found {
		sheet = &scenefile.SpriteSheet{}
		node.SetSheet(sheet)
	}
	sheet.Columns, sheet.Rows = max(columns, 1), max(rows, 1)
	keepFramesInside(sheet)
	editor.Changed()
}

// keepFramesInside drops the frames a sheet does not have from its animations; one left with none
// shows the first frame.
func keepFramesInside(sheet *scenefile.SpriteSheet) {
	count := sheet.FrameCount()
	for index := range sheet.Animations {
		animation := &sheet.Animations[index]
		animation.Frames = slices.DeleteFunc(animation.Frames, func(frame int) bool { return frame >= count })
		if len(animation.Frames) == 0 {
			animation.Frames = []int{0}
		}
	}
}

// TakeSheet shows the chosen node's sprite whole again.
func (editor *Editor) TakeSheet() {
	if editor.Selected == nil {
		return
	}
	editor.Selected.SetSheet(nil)
	editor.Changed()
}

func (editor *Editor) selectedSheet() (*scenefile.SpriteSheet, bool) {
	if editor.Selected == nil {
		return nil, false
	}
	return editor.Selected.Sheet()
}

// AddAnimation adds an animation of the first frame to the chosen node's sheet, with a name of its
// own, and tells where it is in the sheet's list (-1 without a sheet). The first one added plays
// when the scene starts.
func (editor *Editor) AddAnimation() int {
	sheet, found := editor.selectedSheet()
	if !found {
		return -1
	}
	name := freeAnimationName(sheet)
	sheet.Animations = append(sheet.Animations, scenefile.Animation{
		Name: name, Frames: []int{0}, FramesPerSecond: newAnimationFramesPerSecond, Loop: true,
	})
	if sheet.Start == "" {
		sheet.Start = name
	}
	editor.Changed()
	return len(sheet.Animations) - 1
}

// freeAnimationName is "animation", or "animation 2", 3… the first a sheet has not taken.
func freeAnimationName(sheet *scenefile.SpriteSheet) string {
	name := "animation"
	for number := 2; ; number++ {
		if _, taken := sheet.Find(name); !taken {
			return name
		}
		name = "animation " + strconv.Itoa(number)
	}
}

// RenameAnimation names an animation of the chosen node's sheet; it refuses an empty name or one
// another animation has, and tells whether it took it.
func (editor *Editor) RenameAnimation(index int, name string) bool {
	animation, sheet, found := editor.selectedAnimation(index)
	name = strings.TrimSpace(name)
	if !found || name == "" {
		return false
	}
	if _, taken := sheet.Find(name); taken {
		return name == animation.Name
	}
	if sheet.Start == animation.Name {
		sheet.Start = name
	}
	animation.Name = name
	editor.Changed()
	return true
}

// SetAnimationFrames sets the frames an animation shows, written as numbers and runs ("0-3, 6").
func (editor *Editor) SetAnimationFrames(index int, written string) error {
	animation, sheet, found := editor.selectedAnimation(index)
	if !found {
		return nil
	}
	frames, err := ReadFrames(written)
	if err != nil {
		return err
	}
	if last := slices.Max(frames); last >= sheet.FrameCount() {
		return failure.Wrap(FramesWrong, "the sheet has "+strconv.Itoa(sheet.FrameCount())+" frames, from 0", nil)
	}
	animation.Frames = frames
	editor.Changed()
	return nil
}

// RemoveAnimation takes an animation from the chosen node's sheet; if it was the one the scene
// starts with, the first one left starts it.
func (editor *Editor) RemoveAnimation(index int) {
	animation, sheet, found := editor.selectedAnimation(index)
	if !found {
		return
	}
	removed := animation.Name
	sheet.Animations = slices.Delete(sheet.Animations, index, index+1)
	if sheet.Start == removed {
		sheet.Start = ""
		if len(sheet.Animations) > 0 {
			sheet.Start = sheet.Animations[0].Name
		}
	}
	editor.Changed()
}

func (editor *Editor) selectedAnimation(index int) (*scenefile.Animation, *scenefile.SpriteSheet, bool) {
	sheet, found := editor.selectedSheet()
	if !found || index < 0 || index >= len(sheet.Animations) {
		return nil, nil, false
	}
	return &sheet.Animations[index], sheet, true
}

// ReadFrames reads frames written as numbers and runs, separated by commas: "0-3, 6" is 0, 1, 2,
// 3 and 6.
func ReadFrames(written string) ([]int, error) {
	var frames []int
	for _, part := range strings.Split(written, ",") {
		first, last, err := readRun(strings.TrimSpace(part))
		if err != nil {
			return nil, failure.Wrap(FramesWrong, written, err)
		}
		for frame := first; frame <= last; frame++ {
			frames = append(frames, frame)
		}
	}
	return frames, nil
}

// readRun reads "4" as 4 to 4, and "0-3" as 0 to 3.
func readRun(part string) (first, last int, err error) {
	firstText, lastText, isRun := strings.Cut(part, "-")
	if !isRun {
		lastText = firstText
	}
	if first, err = strconv.Atoi(firstText); err != nil {
		return 0, 0, err
	}
	if last, err = strconv.Atoi(lastText); err != nil {
		return 0, 0, err
	}
	if first < 0 || last < first {
		return 0, 0, FramesWrong
	}
	return first, last, nil
}

// Followable are the paths of the nodes a camera may follow: all but itself.
func (editor *Editor) Followable(camera *engine.Node) []string {
	var paths []string
	editor.Scene.Root.Walk(func(node *engine.Node) bool {
		if node != editor.Scene.Root && node != camera {
			paths = append(paths, node.Path())
		}
		return true
	})
	return paths
}

// SetAnimationFramesIfChanged sets an animation's frames from what is written, when it says other
// frames than the ones it shows; it tells why what is written cannot be.
func (editor *Editor) SetAnimationFramesIfChanged(index int, written string, shown []int) error {
	if frames, err := ReadFrames(written); err == nil && slices.Equal(frames, shown) {
		return nil
	}
	return editor.SetAnimationFrames(index, written)
}

// writtenFrames writes frames as ReadFrames reads them, runs joined: 0, 1, 2, 3, 6 is "0-3, 6".
func writtenFrames(frames []int) string {
	var written strings.Builder
	for start := 0; start < len(frames); {
		end := start
		for end+1 < len(frames) && frames[end+1] == frames[end]+1 {
			end++
		}
		if written.Len() > 0 {
			written.WriteString(", ")
		}
		written.WriteString(strconv.Itoa(frames[start]))
		if end > start {
			written.WriteString("-" + strconv.Itoa(frames[end]))
		}
		start = end + 1
	}
	return written.String()
}
