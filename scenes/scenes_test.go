package scenes

import "testing"

func TestASceneComesBackAsWritten(t *testing.T) {
	ground := NewNode("Ground")
	ground.Shape, ground.Scale = "plane", [3]float64{20, 1, 20}
	car := NewNode("Car")
	car.Position = [3]float64{1, 0.5, -2}
	car.Behaviours = []Behaviour{{Name: "Drive", Fields: map[string]any{"speed": 4.0}}}
	car.Children = []Node{NewNode("Wheel")}
	scene := &Scene{Name: "Main", Nodes: []Node{ground, car}}
	data, err := scene.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if read.Nodes[1].Position != car.Position || read.Nodes[1].Behaviours[0].Fields["speed"] != 4.0 || read.Nodes[1].Children[0].Name != "Wheel" {
		t.Fatalf("the scene came back different: %+v", read.Nodes[1])
	}
}

func TestBrokenScenesAreRefused(t *testing.T) {
	for name, data := range map[string]string{
		"no name":   `{"nodes": [{"name": ""}]}`,
		"slash":     `{"nodes": [{"name": "a/b"}]}`,
		"twins":     `{"nodes": [{"name": "A", "children": [{"name": "B"}, {"name": "B"}]}]}`,
		"bad shape": `{"nodes": [{"name": "A", "shape": "teapot"}]}`,
	} {
		if _, err := Read([]byte(data)); err == nil {
			t.Errorf("%s: read", name)
		}
	}
	if _, err := Read([]byte(`{"nodes": [{"name": "A"}, {"name": "B", "children": [{"name": "A"}]}]}`)); err != nil {
		t.Errorf("the same name in two places was refused: %v", err)
	}
}
