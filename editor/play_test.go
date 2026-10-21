package editor

import (
	"log/slog"
	"math"
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/engine"
)

// Drifter moves its node along X at a speed, in metres a second.
type Drifter struct {
	Speed float64 `json:"speed"`
	moved int
}

func (drifter *Drifter) Update(context *engine.Context) {
	context.Node.Transform.Position.X += drifter.Speed * context.Delta
	drifter.moved++
}

// playDrifting plays the starter scene with its cube drifting, and gives the cube played.
func playDrifting(test *testing.T) (*Editor, *engine.Node) {
	engine.RegisterAs("Drifter", Drifter{})
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	editor.AttachBehaviour("Drifter")
	editor.Selected.Behaviours[len(editor.Selected.Behaviours)-1].Fields = map[string]any{"speed": 1.0}
	editor.ShowWorkspace(PlayWorkspace)
	played, _ := editor.PlayingHere()
	moving, _ := played.Find("Cube")
	return editor, moving
}

func TestAPausedGameStandsStillAndStepsOneFrame(test *testing.T) {
	editor, cube := playDrifting(test)
	editor.Paused = true
	from := cube.Transform.Position.X
	editor.AdvancePlay(0.5, nil)
	if cube.Transform.Position.X != from {
		test.Fatal("a paused game moved")
	}
	editor.StepFrame()
	editor.AdvancePlay(0.5, nil)
	stepped := cube.Transform.Position.X
	editor.AdvancePlay(0.5, nil)
	if stepped <= from || cube.Transform.Position.X != stepped {
		test.Fatalf("a step went from %v to %v, then to %v", from, stepped, cube.Transform.Position.X)
	}
}

func TestRestartStartsTheSceneAsWritten(test *testing.T) {
	editor, cube := playDrifting(test)
	written := cube.Transform.Position.X
	editor.AdvancePlay(1, nil)
	editor.RestartPlay()
	played, _ := editor.PlayingHere()
	again, _ := played.Find("Cube")
	if again == cube || again.Transform.Position.X != written {
		test.Fatalf("restarted, the cube is at %v, not %v", again.Transform.Position.X, written)
	}
}

func TestTheWatchShowsABehavioursFields(test *testing.T) {
	editor, _ := playDrifting(test)
	editor.AdvancePlay(0.25, nil)
	lines := editor.Watch()
	found := map[string]string{}
	for _, line := range lines {
		found[line.Name] = line.Value
	}
	if found["Drifter · Speed"] != "1" || !strings.HasPrefix(found["Position"], "0.25") {
		test.Fatalf("the watch shows %v", lines)
	}
}

func TestWhatIsLoggedWhilePlayingIsKept(test *testing.T) {
	editor, _ := playDrifting(test)
	slog.Warn("the lamp is lit", "colour", "amber")
	editor.LeavePlay()
	slog.Warn("after playing")
	log := editor.PlayLog()
	if len(log) != 1 || !strings.Contains(log[0], "the lamp is lit") || !strings.Contains(log[0], "amber") {
		test.Fatalf("the play log is %v", log)
	}
}

func TestTheNintendo3DSShapeShowsTheSecondScreenUnderTheTopOne(test *testing.T) {
	top, bottom := consoleScreens(draw2d.Rect{X: 100, Width: 1000, Height: 520})
	if math.Abs(float64(top.Width/top.Height)-400.0/240) > 1e-3 || math.Abs(float64(bottom.Width/bottom.Height)-320.0/240) > 1e-3 {
		test.Fatalf("the screens are %v and %v, not the 3DS's shapes", top, bottom)
	}
	if bottom.Y < top.Y+top.Height || bottom.Y+bottom.Height > 520 {
		test.Fatalf("the bottom screen %v is not under the top one %v, inside the area", bottom, top)
	}
	if middle := func(area draw2d.Rect) float32 { return area.X + area.Width/2 }; math.Abs(float64(middle(top)-middle(bottom))) > 1e-3 || math.Abs(float64(middle(top)-600)) > 1e-3 {
		test.Fatalf("the screens %v and %v are not centred one over the other", top, bottom)
	}
}
