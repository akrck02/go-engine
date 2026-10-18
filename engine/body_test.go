package engine

import (
	"math"
	"testing"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// Taker counts what touches its node, and what its node touches.
type Taker struct {
	touches []string
}

func (taker *Taker) Touched(context *Context, other *Node) {
	taker.touches = append(taker.touches, other.Name)
}

// platform is a 2D scene: a solid ground whose top is at height 0, a crate that falls onto it, and
// a marker, a trigger, in the crate's way.
func platform(test *testing.T) *Scene {
	RegisterAs("Taker", Taker{})
	description := &scenefile.Scene{Name: "Platform", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "Ground", Shape: "cube", Position: [3]float64{0, -0.5, 0}, Scale: [3]float64{10, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Solid}},
		{Name: "Crate", Shape: "cube", Position: [3]float64{0, 4, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}, Behaviours: []scenefile.Behaviour{{Name: "Taker"}}},
		{Name: "Marker", Shape: "sphere", Position: [3]float64{0, 2, 0}, Scale: [3]float64{0.5, 0.5, 1},
			Body: &scenefile.Body{Kind: scenefile.Trigger}, Behaviours: []scenefile.Behaviour{{Name: "Taker"}}},
	}}
	built, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	if problems := built.Start(); len(problems) > 0 {
		test.Fatal(problems)
	}
	return built
}

func playFor(scene *Scene, seconds float64) {
	const frame = 1.0 / 60
	for elapsed := 0.0; elapsed < seconds; elapsed += frame {
		scene.Update(frame, nil)
	}
}

func TestANodeWithABodyFallsOntoASolidNode(test *testing.T) {
	scene := platform(test)
	playFor(scene, 3)
	crate, _ := scene.Find("Crate")
	if math.Abs(crate.Transform.Position.Y-0.5) > 1e-6 {
		test.Fatalf("the crate rests at height %v, not on the ground", crate.Transform.Position.Y)
	}
	body, found := crate.Body()
	if !found || !body.OnFloor() {
		test.Fatal("the crate on the ground is not on the floor")
	}
}

func TestBehavioursHearWhatTheirNodesTouch(test *testing.T) {
	scene := platform(test)
	playFor(scene, 3)
	crate, _ := scene.Find("Crate")
	marker, _ := scene.Find("Marker")
	crateTaker, _ := Behaviour[Taker](crate)
	markerTaker, _ := Behaviour[Taker](marker)
	if len(markerTaker.touches) != 1 || markerTaker.touches[0] != "Crate" {
		test.Fatalf("the marker heard %v", markerTaker.touches)
	}
	if len(crateTaker.touches) != 1 || crateTaker.touches[0] != "Marker" {
		test.Fatalf("the crate heard %v", crateTaker.touches)
	}
}

func TestARemovedNodesBodyLeavesTheWorld(test *testing.T) {
	scene := platform(test)
	ground, _ := scene.Find("Ground")
	ground.Remove()
	playFor(scene, 1)
	crate, _ := scene.Find("Crate")
	if crate.Transform.Position.Y > -0.5 {
		test.Fatalf("the crate still stands on the removed ground, at %v", crate.Transform.Position.Y)
	}
}

func TestAMovedNodeTakesItsBodyAlong(test *testing.T) {
	scene := platform(test)
	crate, _ := scene.Find("Crate")
	crate.Transform.Position = space.NewVector(3, 1, 0) // a teleport by the game's code
	scene.Update(1.0/60, nil)
	if math.Abs(crate.Transform.Position.X-3) > 1e-9 {
		test.Fatalf("the body pulled the node back to x %v", crate.Transform.Position.X)
	}
}

func TestAFrameOf2DPlayAllocatesNothing(test *testing.T) {
	scene := platform(test)
	playFor(scene, 3)
	allocations := testing.AllocsPerRun(100, func() { scene.Update(1.0/60, nil) })
	if allocations != 0 {
		test.Fatalf("a frame allocates %v times", allocations)
	}
}

func TestANodeMovedOrAddedWhilePlayingKeepsOrGetsItsBody(test *testing.T) {
	scene := platform(test)
	ground, _ := scene.Find("Ground")
	crate, _ := scene.Find("Crate")
	ground.Add(crate)
	if _, found := crate.Body(); !found {
		test.Fatal("moved under another node, the crate lost its body")
	}
	box, err := scene.NewNode(scenefile.Node{Name: "Box", Shape: "cube", Position: [3]float64{3, 4, 0}, Scale: [3]float64{1, 1, 1},
		Body: &scenefile.Body{Kind: scenefile.Moving}})
	if err != nil {
		test.Fatal(err)
	}
	scene.Root.Add(box)
	playFor(scene, 2)
	if body, found := box.Body(); !found || !body.OnFloor() {
		test.Fatal("a node added while playing did not fall onto the ground")
	}
	box.Remove()
	if _, found := box.Body(); found {
		test.Fatal("a removed node kept its body")
	}
}
