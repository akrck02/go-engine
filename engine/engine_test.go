package engine

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-scene-builder/gltf"
	"github.com/akrck02/go-scene-builder/mesh"

	"github.com/akrck02/go-engine/scenes"
	"github.com/akrck02/go-engine/space"
)

// tower is a .glb of one box, 2 metres high, standing on its origin.
func tower(t *testing.T) []byte {
	piece := &mesh.Piece{Name: "tower", Material: mesh.Material{Name: "stone", Kind: "lit"}}
	white := [4]float32{1, 1, 1, 1}
	corners := []mesh.Vector{{-0.5, 0, -0.5}, {0.5, 0, -0.5}, {0.5, 2, -0.5}, {-0.5, 2, -0.5}, {-0.5, 0, 0.5}, {0.5, 0, 0.5}, {0.5, 2, 0.5}, {-0.5, 2, 0.5}}
	for _, corner := range corners {
		piece.AddVertex(corner, mesh.Vector{0, 1, 0}, white)
	}
	piece.Indices = []uint32{0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7}
	data, err := gltf.Encode([]*mesh.Piece{piece}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func project(t *testing.T) fstest.MapFS {
	description := &scenes.Scene{Name: "Main", Nodes: []scenes.Node{
		{Name: "Ground", Shape: "plane", Colour: "#557755", Scale: [3]float64{10, 1, 10}},
		{Name: "Tower", Model: "assets/models/tower.glb", Position: [3]float64{3, 0, 0}, Scale: [3]float64{1, 1, 1},
			Children: []scenes.Node{{Name: "Flag", Shape: "cube", Position: [3]float64{0, 2.5, 0}, Scale: [3]float64{0.5, 0.5, 0.5}}}},
		{Name: "Spinner", Shape: "cube", Scale: [3]float64{1, 1, 1}, Behaviours: []scenes.Behaviour{{Name: "Spin", Fields: map[string]any{"speed": 90}}}},
		{Name: "Eye", Position: [3]float64{0, 5, 10}, Rotation: [3]float64{-20, 0, 0}, Scale: [3]float64{1, 1, 1},
			Camera: &scenes.Camera{Projection: scenes.Perspective, Main: true}},
	}}
	data, err := description.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return fstest.MapFS{
		"engine.json":             {Data: []byte(`{"name": "Test", "startScene": "scenes/main.scene.json"}`)},
		"scenes/main.scene.json":  {Data: data},
		"assets/models/tower.glb": {Data: tower(t)},
	}
}

type Spin struct {
	Speed   float64 `json:"speed"`
	started bool
}

func (spin *Spin) Start(context *Context) { spin.started = true }
func (spin *Spin) Update(context *Context) {
	context.Node.Transform.Turn(space.Up, spin.Speed*context.Delta)
}

func load(t *testing.T) *Scene {
	files := project(t)
	settings, err := ReadSettings(files)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(files, settings.StartScene)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func TestASceneIsBuiltFromItsFileAndFound(t *testing.T) {
	loaded := load(t)
	flag, found := loaded.Find("Tower/Flag")
	if !found || flag.Path() != "Tower/Flag" {
		t.Fatal("Tower/Flag was not found")
	}
	if world := flag.World().Position; !world.Near(space.V(3, 2.5, 0), 1e-9) {
		t.Fatalf("the flag is at %v in the world", world)
	}
	if _, found := loaded.Find("Tower/Nothing"); found {
		t.Fatal("a node that does not exist was found")
	}
}

func TestAnImportedModelHasItsSizeAndMovesWithItsNode(t *testing.T) {
	loaded := load(t)
	tower, _ := loaded.Find("Tower")
	bounds := tower.Bounds()
	if !bounds.Size().Near(space.V(1, 2, 1), 1e-3) || !bounds.Centre().Near(space.V(3, 1, 0), 1e-3) {
		t.Fatalf("the tower spans %v", bounds)
	}
	tower.Transform.Scale = space.V(2, 2, 2)
	if size := tower.Bounds().Size(); !size.Near(space.V(2, 4, 2), 1e-3) {
		t.Fatalf("scaled twice, the tower is %v", size)
	}
}

func TestBehavioursAreMadeWithTheirFieldsAndRun(t *testing.T) {
	Register(Spin{})
	loaded := load(t)
	if problems := loaded.Start(); len(problems) != 0 {
		t.Fatal(problems)
	}
	for frame := 0; frame < 60; frame++ {
		loaded.Update(1.0/60, nil)
	}
	spinner, _ := loaded.Find("Spinner")
	spin, found := Behaviour[Spin](spinner)
	if !found || spin.Speed != 90 || !spin.started {
		t.Fatalf("the behaviour was not made with its fields: %+v", spin)
	}
	if yaw := spinner.Transform.Euler().Y; math.Abs(yaw-90) > 1e-6 {
		t.Fatalf("after a second at 90° a second, the yaw is %v", yaw)
	}
}

func TestAnUnregisteredBehaviourIsReportedAndTheRestPlays(t *testing.T) {
	loaded := load(t)
	spinner, _ := loaded.Find("Spinner")
	spinner.Behaviours = append(spinner.Behaviours, scenes.Behaviour{Name: "Missing"})
	problems := loaded.Start()
	if len(problems) == 0 || !strings.Contains(problems[len(problems)-1].Error(), "Missing") {
		t.Fatalf("problems: %v", problems)
	}
}

func TestASceneIsDescribedBackAsItWasRead(t *testing.T) {
	loaded := load(t)
	described := loaded.Describe()
	files := project(t)
	original, _ := scenes.Read(files["scenes/main.scene.json"].Data)
	again, _ := described.Marshal()
	first, _ := original.Marshal()
	if string(again) != string(first) {
		t.Fatalf("the scene changed on the way:\n%s\n---\n%s", first, again)
	}
}

func TestTheMainCameraLooksWhereItsNodeLooks(t *testing.T) {
	loaded := load(t)
	camera, found := loaded.MainCamera()
	if !found || camera.Name != "Eye" {
		t.Fatal("the main camera was not found")
	}
	view := camera.View()
	direction := space.V(float64(view.Target[0]-view.Position[0]), float64(view.Target[1]-view.Position[1]), float64(view.Target[2]-view.Position[2]))
	if !direction.Near(space.V(0, -math.Sin(space.Radians(20)), -math.Cos(space.Radians(20))), 1e-6) {
		t.Fatalf("the camera looks along %v", direction)
	}
	camera.SetCamera(&scenes.Camera{Projection: scenes.Isometric, Height: 12})
	camera.Transform.SetEuler(space.Euler{Y: 45})
	iso := camera.View()
	if iso.OrthographicHeight != 12 {
		t.Fatalf("an isometric camera sees %v metres", iso.OrthographicHeight)
	}
}

func TestNodesMoveFromParentToParent(t *testing.T) {
	loaded := load(t)
	tower, _ := loaded.Find("Tower")
	spinner, _ := loaded.Find("Spinner")
	tower.Add(spinner)
	if _, found := loaded.Find("Tower/Spinner"); !found || len(loaded.Root.Children()) != 3 {
		t.Fatal("the spinner did not move under the tower")
	}
	if len(tower.render.Children) != 3 { // the model, the flag and the spinner
		t.Fatalf("the tower draws %d children", len(tower.render.Children))
	}
	spinner.Add(tower) // a node cannot hang from what hangs from it
	if tower.Parent() != loaded.Root {
		t.Fatal("a node was hung from its own child")
	}
	spinner.Remove()
	if _, found := loaded.Find("Tower/Spinner"); found {
		t.Fatal("the removed node is still there")
	}
}

// flatProject is a 2D scene: a sprite 64 pixels wide and 32 high, a red square in front of it,
// and a camera seeing 8 metres.
func flatProject(t *testing.T) fstest.MapFS {
	picture := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	var encoded bytes.Buffer
	png.Encode(&encoded, picture)
	description := &scenes.Scene{Name: "Flat", Mode: scenes.Mode2D, PixelsPerMetre: 32, Nodes: []scenes.Node{
		{Name: "Square", Shape: "cube", Colour: "#ff0000", Position: [3]float64{0, 0, 1}, Scale: [3]float64{1, 1, 1}},
		{Name: "Hero", Sprite: "assets/hero.png", Position: [3]float64{2, 1, 0}, Scale: [3]float64{1, 1, 1}},
		{Name: "Eye", Position: [3]float64{2, 1, 0}, Scale: [3]float64{1, 1, 1}, Camera: &scenes.Camera{Projection: scenes.Orthographic, Height: 8}},
	}}
	data, _ := description.Marshal()
	return fstest.MapFS{"scenes/flat.scene.json": {Data: data}, "assets/hero.png": {Data: encoded.Bytes()}}
}

func TestA2DSceneDrawsSpritesAtTheirSizeFurthestFirst(t *testing.T) {
	flat, err := Load(flatProject(t), "scenes/flat.scene.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, _ := flat.Find("Hero")
	if size := hero.Bounds().Size(); math.Abs(size.X-2) > 1e-9 || math.Abs(size.Y-1) > 1e-9 {
		t.Fatalf("a sprite of 64 × 32 pixels at 32 a metre is %v metres", size)
	}
	view := flat.View2D()
	if view.Height != 8 || !view.Centre.Near(space.V(2, 1, 0), 1e-9) {
		t.Fatalf("the camera sees %+v", view)
	}
	area := draw2d.Rect{Width: 800, Height: 400} // 50 pixels a metre
	if x, y := view.ToScreen(area, space.V(3, 2, 0)); x != 450 || y != 150 {
		t.Fatalf("a metre right and up of the centre is at %v, %v", x, y)
	}
	if back := view.FromScreen(area, 450, 150); !back.Near(space.V(3, 2, 0), 1e-9) {
		t.Fatalf("the screen point came back as %v", back)
	}
	order := flat.drawOrder()
	if len(order) != 2 || order[0].Name != "Hero" || order[1].Name != "Square" {
		t.Fatalf("drawn in the order %v, %v", order[0].Name, order[1].Name)
	}
	var list draw2d.List
	flat.Draw2D(&list, area, view)
	if len(list.Commands) < 2 {
		t.Fatalf("%d commands drawn", len(list.Commands))
	}
	if len(flat.World().Nodes[0].Children) != 3 || flat.World().Nodes[0].Children[0].Children != nil {
		t.Fatal("a 2D scene made 3D geometry")
	}
}
