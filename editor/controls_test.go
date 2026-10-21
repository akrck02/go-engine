package editor

import (
	"os"
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
)

// controlsWith is a Controls workspace's work, read from some actions.
func controlsWith(actions map[string]map[string][]string) *controlsView {
	var controls controlsView
	controls.load(engine.Settings{Actions: actions})
	return &controls
}

func TestAControlDoesOneActionOnEveryTargetOfItsDevice(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"jump": {}, "fire": {}})
	controls.showTab(0) // the pad, played on a computer and in a browser
	controls.assign("Pad A", "jump")
	controls.assign("Pad A", "fire") // the other action replaces the first
	for _, target := range []string{engine.DesktopTarget, engine.WebTarget} {
		if !slices.Equal(controls.actions["fire"][target], []string{"Pad A"}) || len(controls.actions["jump"][target]) != 0 {
			test.Fatalf("on %s, fire has %v and jump %v", target, controls.actions["fire"][target], controls.actions["jump"][target])
		}
	}
	if name, _ := controls.actionOn("Pad A"); name != "fire" {
		test.Fatalf("Pad A says it does %q", name)
	}
	controls.assign("Pad A", "")
	if name, _ := controls.actionOn("Pad A"); name != "" {
		test.Fatal("choosing None leaves the button doing an action")
	}
}

func TestAStickGivesEachAxisItsOwnAction(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"turn": {}, "climb": {}})
	controls.assign("Stick X", "turn")
	controls.assign("Stick Y", "climb")
	stick := padDrawing().controls[4] // the left stick
	if label, does := controls.chipFor(stick, "turn"); !does || label != "Left stick ↔" {
		test.Fatalf("turn's card names the stick %q", label)
	}
	if name, _ := controls.actionOn("Stick Y"); name != "climb" {
		test.Fatalf("the stick's vertical axis does %q", name)
	}
}

func TestAButtonMovingAnAxisOneWayIsNamedWithItsSign(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"turn": {engine.DesktopTarget: {"Left-", "Stick X"}}})
	controls.showTab(1)
	if !controls.axes["turn"] {
		test.Fatal("an action bound to an axis is not read as one")
	}
	if _, label := controls.actionOn("Left"); label != "turn −" {
		test.Fatalf("the key Left says it does %q", label)
	}
}

func TestAnActionIsRenamedWithItsInputs(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"jump": {engine.DesktopTarget: {"Space"}}, "fire": {}})
	controls.renameAction("jump", "leap")
	controls.renameAction("leap", "fire") // taken: nothing changes
	if !slices.Equal(controls.actions["leap"][engine.DesktopTarget], []string{"Space"}) || controls.actions["jump"] != nil {
		test.Fatalf("the actions after renaming are %v", controls.names())
	}
}

func TestSwitchingAnActionsKindDropsItsInputs(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"jump": {engine.DesktopTarget: {"Space"}}})
	controls.switchKind("jump")
	if !controls.axes["jump"] || len(controls.actions["jump"][engine.DesktopTarget]) != 0 {
		test.Fatal("an action switched to an axis keeps a button's press")
	}
}

func TestEveryChangeOfTheControlsCanBeUndoneAndRedone(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{})
	controls.addAction("jump", false)
	controls.assign("Pad A", "jump")
	controls.undoChange()
	if len(controls.actions["jump"][engine.DesktopTarget]) != 0 {
		test.Fatal("undo leaves the button bound")
	}
	controls.undoChange()
	if controls.actions["jump"] != nil {
		test.Fatal("undo leaves the action added")
	}
	controls.redoChange()
	controls.redoChange()
	if !slices.Equal(controls.actions["jump"][engine.DesktopTarget], []string{"Pad A"}) {
		test.Fatal("redo does not bring the changes back")
	}
}

func TestTheControlsAreKeptInTheProjectAsTheyChange(test *testing.T) {
	editor := open(test)
	view := NewView(editor)
	view.controls.load(editor.Project.Settings)
	view.controls.addAction("jump", false)
	view.controls.assign("Space", "jump")
	view.saveControls()
	settings, err := engine.ReadSettings(os.DirFS(editor.Project.Folder))
	if err != nil {
		test.Fatal(err)
	}
	if !slices.Equal(settings.Actions["jump"][engine.DesktopTarget], []string{"Space"}) {
		test.Fatalf("engine.json keeps jump as %v", settings.Actions["jump"])
	}
}

func TestControlZInControlsUndoesTheControlsNotTheScene(test *testing.T) {
	editor := open(test)
	view := NewView(editor)
	editor.Add("Cube")
	editor.ShowWorkspace(ControlsWorkspace)
	view.controls.load(editor.Project.Settings)
	view.controls.addAction("jump", false)
	state := input.NewState()
	state.SetKey(input.KeyControl, true)
	state.SetKey(input.KeyZ, true)
	view.keys(state)
	if view.controls.actions["jump"] != nil {
		test.Fatal("Ctrl+Z in Controls does not undo the action added")
	}
	if _, found := editor.Scene.Find("Cube 2"); !found {
		test.Fatal("Ctrl+Z in Controls undid the scene")
	}
}

func TestAnInputPressedIsCaughtForTheActionHighlighted(test *testing.T) {
	editor := open(test)
	view := NewView(editor)
	editor.ShowWorkspace(ControlsWorkspace)
	view.controls.load(editor.Project.Settings)
	view.controls.showTab(1)
	view.controls.addAction("jump", false)
	view.controls.highlighted, view.controls.catching = "jump", true
	state := input.NewState()
	state.SetKey(input.Key1, true) // a key that would change the workspace
	view.keys(state)
	if words := view.controls.actions["jump"][engine.WebTarget]; !slices.Equal(words, []string{"1"}) || view.controls.catching {
		test.Fatalf("the key pressed is caught as %v, still catching %v", words, view.controls.catching)
	}
	if editor.ShownWorkspace() != ControlsWorkspace {
		test.Fatal("the key caught also changed the workspace")
	}
}

func TestAClickFindsTheControlUnderIt(test *testing.T) {
	device := consoleDrawing()
	// an area that shows the drawing at its own size, its units' origin at the window's
	area := draw2d.Rect{X: device.left - drawingMargin, Y: device.top - drawingMargin,
		Width: device.width + 2*drawingMargin, Height: device.height + 2*drawingMargin}
	index, found := device.controlAt(area, 480, 370) // the X button's centre
	if !found || device.controls[index].words[0] != "X" {
		test.Fatalf("a click on X finds %v", device.controls[index].words)
	}
	if _, found := device.controlAt(area, 320, 140); found {
		test.Fatal("a click on the top screen finds a control")
	}
}

func TestEveryDrawnControlIsAWordItsTargetsKnow(test *testing.T) {
	designed := controlsWith(nil)
	designed.startLayout()
	for tab := range deviceTabs {
		for _, controls := range []*controlsView{controlsWith(nil), designed} {
			controls.showTab(tab)
			for _, each := range controls.drawingShown().controls {
				for _, target := range deviceTabs[tab].targets {
					if problems := engine.CheckInputs(each.words, target); len(problems) != 0 {
						test.Errorf("the %s drawn for %s has %v", each.name, target, problems)
					}
				}
			}
		}
	}
}

func TestATouchControlIsAddedAndStaysWholeOnTheScreen(test *testing.T) {
	controls := controlsWith(nil)
	controls.startLayout()
	controls.addTouch(engine.TouchButton)
	added := controls.touch[len(controls.touch)-1]
	if added.Name != "C" || controls.design.selected != len(controls.touch)-1 {
		test.Fatalf("the button added is named %q and chosen %v", added.Name, controls.design.selected)
	}
	controls.moveTouch(controls.design.selected, 2, -1) // dragged past the screen's corner
	moved := controls.touch[controls.design.selected]
	if moved.X >= 1 || moved.Y-moved.Size/2 != 0 {
		test.Fatalf("the button dragged out lies at %v, %v", moved.X, moved.Y)
	}
}

func TestATouchControlsActionsFollowItsNameAndGoWithIt(test *testing.T) {
	controls := controlsWith(map[string]map[string][]string{"turn": {}})
	controls.startLayout()
	controls.showTab(mobileTab)
	controls.assign("Touch Stick X", "turn")
	controls.chooseTouch(controls.touchIndex("Stick"))
	controls.design.draft = "Move"
	controls.renameTouch()
	if words := controls.actions["turn"][engine.AndroidTarget]; !slices.Equal(words, []string{"Touch Move X"}) {
		test.Fatalf("after renaming the stick, turn is bound to %v", words)
	}
	controls.removeTouch()
	if len(controls.actions["turn"][engine.AndroidTarget]) != 0 || controls.touchIndex("Move") >= 0 {
		test.Fatal("a touch control removed leaves its words bound")
	}
}
