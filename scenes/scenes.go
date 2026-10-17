// Package scenes is how a scene is kept: a JSON file of nodes, each with its transform (position in
// metres, rotation in degrees, scale), what it shows (a shape, a model from the project's assets,
// a camera) and the behaviours the game's own Go code gives it. The editor writes these files; the
// engine reads them.
package scenes

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Scene is a scene file.
type Scene struct {
	Name  string `json:"name"`
	Sky   string `json:"sky,omitempty"` // the background colour, "#rrggbb"; empty is a daylight sky
	Nodes []Node `json:"nodes"`
}

// Node is a thing in a scene, and its children, which move with it.
type Node struct {
	Name     string     `json:"name"`
	Position [3]float64 `json:"position"`
	Rotation [3]float64 `json:"rotation"` // degrees: pitch (X), yaw (Y), roll (Z), as space.Euler
	Scale    [3]float64 `json:"scale"`
	Hidden   bool       `json:"hidden,omitempty"`

	Shape  string  `json:"shape,omitempty"`  // one of Shapes
	Model  string  `json:"model,omitempty"`  // a model in the project, as "assets/models/tree.glb"
	Colour string  `json:"colour,omitempty"` // of a shape, "#rrggbb"
	Camera *Camera `json:"camera,omitempty"`

	Behaviours []Behaviour `json:"behaviours,omitempty"`
	Children   []Node      `json:"children,omitempty"`
}

// Shapes are the shapes a node may be, each a metre across.
var Shapes = []string{"cube", "sphere", "plane", "cylinder"}

// Camera makes a node a camera; the scene is seen through the first one marked Main (or the first
// camera, if none is).
type Camera struct {
	Projection  Projection `json:"projection"`
	FieldOfView float64    `json:"fieldOfView,omitempty"` // degrees, top to bottom, in perspective; 0 is 45
	Height      float64    `json:"height,omitempty"`      // metres seen top to bottom, orthographic and isometric; 0 is 10
	Main        bool       `json:"main,omitempty"`
}

// Projection is how a camera flattens the world.
type Projection string

const (
	Perspective  Projection = "perspective"  // far things look smaller: most 3D games
	Orthographic Projection = "orthographic" // no perspective: plans, side views, 2D in 3D
	Isometric    Projection = "isometric"    // orthographic, from the classic angle (looking down 35.264° from 45° round)
)

// Projections are the projections in the order the editor offers them.
var Projections = []Projection{Perspective, Orthographic, Isometric}

// IsometricAngles are the rotation, in degrees, of an isometric camera: the three axes look the
// same length on screen.
var IsometricAngles = [3]float64{-35.264389682754654, 45, 0}

// Behaviour is a piece of the game's own Go code attached to a node, by the name it was registered
// with, and the values of its fields.
type Behaviour struct {
	Name   string         `json:"name"`
	Fields map[string]any `json:"fields,omitempty"`
}

// NewNode is a node at the origin, unturned, at its own size.
func NewNode(name string) Node { return Node{Name: name, Scale: [3]float64{1, 1, 1}} }

// Read reads a scene file and checks it.
func Read(data []byte) (*Scene, error) {
	var scene Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, err
	}
	if err := scene.Check(); err != nil {
		return nil, err
	}
	return &scene, nil
}

// Marshal writes the scene as indented JSON.
func (scene *Scene) Marshal() ([]byte, error) { return json.MarshalIndent(scene, "", "\t") }

// Check tells what is wrong with a scene: a node without a name, a slash in a name (names make
// paths, as "car/wheel"), two siblings with one name, or a shape that does not exist.
func (scene *Scene) Check() error { return checkNodes(scene.Nodes, "") }

func checkNodes(nodes []Node, parent string) error {
	seen := map[string]bool{}
	for _, node := range nodes {
		path := strings.TrimPrefix(parent+"/"+node.Name, "/")
		switch {
		case node.Name == "":
			return fmt.Errorf("a node in %q has no name", parent)
		case strings.Contains(node.Name, "/"):
			return fmt.Errorf("node %q has a slash in its name", path)
		case seen[node.Name]:
			return fmt.Errorf("two nodes are called %q", path)
		case node.Shape != "" && !isShape(node.Shape):
			return fmt.Errorf("node %q is a %q, which is not a shape", path, node.Shape)
		}
		seen[node.Name] = true
		if err := checkNodes(node.Children, path); err != nil {
			return err
		}
	}
	return nil
}

func isShape(name string) bool {
	for _, shape := range Shapes {
		if shape == name {
			return true
		}
	}
	return false
}
