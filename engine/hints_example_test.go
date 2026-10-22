package engine_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// This is the code of the guide to actions (docs/5-guides/9-actions-and-controls.md, "Write the
// hints from the bindings"); a test checks the guide still shows it.

// menuHint and menuPrompt are what the game shows for its menu, worked out again only when the
// hint changes, so a frame allocates nothing.
var menuHint, menuPrompt string

// showMenuHint runs before the behaviours, every frame.
func showMenuHint(context *engine.Context) {
	if hint := context.Input.ActionHint("menu"); hint != menuHint {
		menuHint, menuPrompt = hint, hint+" menu"
	}
}

// drawMenuHint draws the prompt at the bottom left of the screen.
func drawMenuHint(list *draw2d.List, width, height float32) {
	list.Text(menuPrompt, 12, height-12, 18, draw2d.Medium, draw2d.Hex("#f5ead0"))
}

func ExampleInput_ActionHint() {
	engine.BindActions(map[string]map[string][]string{"menu": {"desktop": {"Tab", "Pad Start"}}}, engine.DesktopTarget)
	engine.BeforeBehaviours(showMenuHint)
	defer engine.BeforeBehaviours(nil)
	engine.SetOverlay(drawMenuHint)
	defer engine.SetOverlay(nil)
	played := engine.NewScene(fstest.MapFS{})

	keyboard := input.NewState()
	keyboard.SetKey(input.KeyEscape, true)
	played.PlayFrame(0.1, keyboard)
	fmt.Println(menuPrompt)

	pad := input.NewState()
	pad.SetPadLayout(input.PlayStationLayout)
	pad.SetPadButton(input.PadA, true)
	played.PlayFrame(0.1, pad)
	fmt.Println(menuPrompt)
	// Output:
	// Tab menu
	// Options menu
}

// hintGuide is the guide that shows ExampleInput_ActionHint's code.
const hintGuide = "../docs/5-guides/9-actions-and-controls.md"

func TestTheGuideShowsTheHintCodeThatIsTested(test *testing.T) {
	guide, err := os.ReadFile(hintGuide)
	if err != nil {
		test.Fatal(err)
	}
	example, err := os.ReadFile("hints_example_test.go")
	if err != nil {
		test.Fatal(err)
	}
	_, section, _ := strings.Cut(string(guide), "## Write the hints from the bindings")
	_, code, found := strings.Cut(section, "```go\n")
	code, _, _ = strings.Cut(code, "```")
	if !found || strings.TrimSpace(code) == "" {
		test.Fatal("the guide shows no code for hints")
	}
	if !strings.Contains(string(example), code) {
		test.Fatalf("the guide's code is not the example's:\n%s", code)
	}
}
