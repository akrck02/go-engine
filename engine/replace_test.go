package engine

import (
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// Leaver asks for another scene, or to start this one again, once a tenth of a second has gone.
type Leaver struct {
	Next  string `json:"next"` // the scene to go to; empty starts this one again
	asked bool
}

func (leaver *Leaver) Update(context *Context) {
	context.Node.Transform.Move(space.NewVector(1, 0, 0))
	if leaver.asked || context.Time < 0.1 {
		return
	}
	leaver.asked = true
	if leaver.Next == "" {
		context.RestartScene()
		return
	}
	context.ChangeScene(leaver.Next)
}

// twoLevels are a project's two scenes: the first's runner goes on to a scene of a name.
func twoLevels(test *testing.T, next string) fstest.MapFS {
	RegisterAs("Leaver", Leaver{})
	write := func(scene *scenefile.Scene) []byte {
		data, err := scene.Marshal()
		if err != nil {
			test.Fatal(err)
		}
		return data
	}
	first := &scenefile.Scene{Name: "First", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{{Name: "Runner", Scale: [3]float64{1, 1, 1},
		Behaviours: []scenefile.Behaviour{{Name: "Leaver", Fields: map[string]any{"next": next}}}}}}
	second := &scenefile.Scene{Name: "Second", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{{Name: "Finish", Scale: [3]float64{1, 1, 1}}}}
	return fstest.MapFS{"scenes/first.scene.json": {Data: write(first)}, "scenes/second.scene.json": {Data: write(second)}}
}

// playUntilReplaced plays a scene until it asks to be replaced, and gives what replaces it.
func playUntilReplaced(test *testing.T, scene *Scene) (*Scene, error) {
	for frame := 0; frame < 60; frame++ {
		scene.Update(1.0/60, nil)
		next, replaced, err := scene.Replacement()
		if replaced || err != nil {
			return next, err
		}
	}
	test.Fatal("the scene never asked to be replaced")
	return nil, nil
}

func TestChangingSceneStartsTheNewOne(test *testing.T) {
	files := twoLevels(test, "scenes/second.scene.json")
	first, err := Load(files, "scenes/first.scene.json")
	if err != nil {
		test.Fatal(err)
	}
	next, err := playUntilReplaced(test, first)
	if err != nil || next.Name != "Second" {
		test.Fatalf("the next scene is %v: %v", next, err)
	}
	if _, found := next.Find("Finish"); !found || !next.started {
		test.Fatal("the new scene was not started")
	}
}

func TestRestartingASceneStartsItAsItWasWritten(test *testing.T) {
	files := twoLevels(test, "")
	first, err := Load(files, "scenes/first.scene.json")
	if err != nil {
		test.Fatal(err)
	}
	again, err := playUntilReplaced(test, first)
	if err != nil || again == first || again.Name != "First" {
		test.Fatalf("the restart gave %v: %v", again, err)
	}
	runner, _ := again.Find("Runner")
	if runner.Transform.Position != space.Zero {
		test.Fatalf("the restarted runner starts at %v, not where it was written", runner.Transform.Position)
	}
}

func TestChangingToAMissingSceneKeepsTheOneThereIs(test *testing.T) {
	files := twoLevels(test, "scenes/nowhere.scene.json")
	first, err := Load(files, "scenes/first.scene.json")
	if err != nil {
		test.Fatal(err)
	}
	if _, err := playUntilReplaced(test, first); !failure.Is(err, SceneNotFound) {
		test.Fatalf("changing to a missing scene: %v", err)
	}
	if _, replaced, _ := first.Replacement(); replaced {
		test.Fatal("the failed change was asked for again")
	}
}
