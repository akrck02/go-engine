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

func TestABodysBoxIsItsSizeOrItsNodes(test *testing.T) {
	scene := platform(test)
	ground, _ := scene.Find("Ground")
	box, found := ground.BodyBox()
	if !found || !box.Size().Near(space.NewVector(10, 1, 0), 1e-9) || !box.Centre().Near(space.NewVector(0, -0.5, 0), 1e-9) {
		test.Fatalf("the ground's body is %v", box)
	}
	ground.SetBodySettings(&scenefile.Body{Kind: scenefile.Solid, Size: [2]float64{4, 0.5}})
	if box, _ := ground.BodyBox(); box.Size().X != 4 || box.Size().Y != 0.5 {
		test.Fatalf("a body of its own size is %v", box.Size())
	}
	if _, found := scene.Root.BodyBox(); found {
		test.Fatal("the root has a body")
	}
}

func TestAPushedBoxsNodeMovesWithIt(test *testing.T) {
	description := &scenefile.Scene{Name: "Push", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "Ground", Shape: "cube", Position: [3]float64{0, -0.5, 0}, Scale: [3]float64{20, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Solid}},
		{Name: "Box", Shape: "cube", Position: [3]float64{2, 0.5, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Pushable}},
		{Name: "Walker", Shape: "cube", Position: [3]float64{0, 0.5, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}},
	}}
	scene, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	scene.Start()
	walker, _ := scene.Find("Walker")
	box, _ := scene.Find("Box")
	const frame = 1.0 / 60
	for elapsed := 0.0; elapsed < 1; elapsed += frame {
		body, _ := walker.Body()
		body.Velocity.X = 3
		scene.Update(frame, nil)
	}
	if x := box.World().Position.X; x <= 2.5 {
		test.Fatalf("the box's node stayed at %v: it does not move with its body", x)
	}
}

func TestAPlatformsNodeCarriesTheNodeOnIt(test *testing.T) {
	description := &scenefile.Scene{Name: "Lift", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "Lift", Shape: "cube", Position: [3]float64{0, -0.25, 0}, Scale: [3]float64{4, 0.5, 1},
			Body: &scenefile.Body{Kind: scenefile.Platform}},
		{Name: "Rider", Shape: "cube", Position: [3]float64{0, 0.5, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}},
	}}
	scene, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	scene.Start()
	lift, _ := scene.Find("Lift")
	rider, _ := scene.Find("Rider")
	body, _ := lift.Body()
	body.Velocity = space.NewVector(1, 0, 0)
	playFor(scene, 1)
	if liftX, riderX := lift.World().Position.X, rider.World().Position.X; liftX < 0.9 || math.Abs(riderX-liftX) > 1e-6 {
		test.Fatalf("the lift's node is at %v and the rider's at %v: the rider was not carried", liftX, riderX)
	}
}

func TestANodeInAGravityZoneFallsItsWay(test *testing.T) {
	description := &scenefile.Scene{Name: "Field", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "Ceiling", Shape: "cube", Position: [3]float64{0, 6.5, 0}, Scale: [3]float64{20, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Solid}},
		{Name: "Field", Shape: "cube", Position: [3]float64{0, 3, 0}, Scale: [3]float64{10, 10, 1},
			Body: &scenefile.Body{Kind: scenefile.Trigger, Pull: &[2]float64{0, 9.81}}},
		{Name: "Crate", Shape: "cube", Position: [3]float64{0, 1, 0}, Scale: [3]float64{1, 1, 1},
			Body: &scenefile.Body{Kind: scenefile.Moving}},
	}}
	scene, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	scene.Start()
	playFor(scene, 2)
	crate, _ := scene.Find("Crate")
	if top := crate.World().Position.Y + 0.5; math.Abs(top-6) > 1e-6 {
		test.Fatalf("the crate's top is at %v: the field should pull it up to the ceiling at 6", top)
	}
}

func TestMarkingAnotherCameraMainChangesTheView(test *testing.T) {
	description := &scenefile.Scene{Name: "Two rooms", Mode: scenefile.Mode2D, Nodes: []scenefile.Node{
		{Name: "First room", Position: [3]float64{0, 0, 0}, Camera: &scenefile.Camera{Projection: scenefile.Orthographic, Height: 8, Main: true}},
		{Name: "Second room", Position: [3]float64{30, 0, 0}, Camera: &scenefile.Camera{Projection: scenefile.Orthographic, Height: 8}},
	}}
	scene, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	scene.Start()
	if centre := scene.View2D().Centre.X; centre != 0 {
		test.Fatalf("the scene is seen from %v, not the first room's camera", centre)
	}
	first, _ := scene.Find("First room")
	second, _ := scene.Find("Second room")
	firstCamera, _ := first.Camera()
	secondCamera, _ := second.Camera()
	firstCamera.Main, secondCamera.Main = false, true
	if centre := scene.View2D().Centre.X; centre != 30 {
		test.Fatalf("with the second camera marked Main the scene is seen from %v, not 30", centre)
	}
}
