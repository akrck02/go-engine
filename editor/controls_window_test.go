package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-ui/uitest"
)

// padControl is where a control of the device shown was drawn in the last frame.
func padControl(test *testing.T, view *View, word string) (x, y float32) {
	test.Helper()
	device := view.controls.drawingShown()
	for _, each := range device.controls {
		if slices.Contains(each.words, word) {
			shape := device.placed(view.controls.area, each.x, each.y, each.width, each.height)
			return shape.X + shape.Width/2, shape.Y + shape.Height/2
		}
	}
	test.Fatalf("the device shown has no control for %q", word)
	return 0, 0
}

func TestAPadButtonIsGivenANewActionInTheControlsWorkspace(test *testing.T) {
	editor := open(test)
	view := NewView(editor)
	window, err := uitest.New(view.Build, uitest.Size(1280, 800), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	if err := window.Click("Controls"); err != nil {
		test.Fatal(err)
	}
	if err := window.Click("Add an action"); err != nil { // the + is called by its tooltip
		test.Fatal(err)
	}
	window.PressKey(input.KeyEnter) // keep the name the new action is given
	name := view.controls.names()[len(view.controls.names())-1]
	if card, err := window.Element(name); err != nil || !card.Visible {
		test.Fatalf("the new action's card is not shown: %v", err)
	}
	x, y := padControl(test, view, "Pad A")
	window.ClickAt(x, y)
	choice, err := window.ElementWithin("A", name) // the row of the panel titled with the button's name
	if err != nil {
		test.Fatal(err)
	}
	window.ClickAt(choice.Rect.X+choice.Rect.Width/2, choice.Rect.Y+choice.Rect.Height/2)
	contents, err := os.ReadFile(filepath.Join(editor.Project.Folder, "engine.json"))
	if err != nil {
		test.Fatal(err)
	}
	var settings engine.Settings
	if err := json.Unmarshal(contents, &settings); err != nil {
		test.Fatal(err)
	}
	if !slices.Contains(settings.Actions[name][engine.DesktopTarget], "Pad A") {
		test.Fatalf("engine.json gives %q %v", name, settings.Actions[name])
	}
}
