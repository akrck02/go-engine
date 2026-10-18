package engine

import (
	"io/fs"
	"math"
	"path"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/loaders"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/animation"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/physics"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/shapes"
	"github.com/akrck02/go-engine/space"
)

// Scene is a scene loaded from its file: a tree of nodes under a root, and what go-renderer draws
// of it.
type Scene struct {
	Name string
	Sky  string // the background colour, "#rrggbb"; empty for a daylight sky
	Root *Node
	// Mode is 3D or 2D; a 2D scene draws its sprites and shapes flat (see Draw2D).
	Mode scenefile.Mode
	// PixelsPerMetre is how many of a sprite's pixels make a metre, in a 2D scene.
	PixelsPerMetre float64

	// Editing shows what only the editor shows: a marker for cameras and empty nodes.
	Editing bool
	// Animate runs the scene's tweens; the engine moves it on every frame.
	Animate animation.Player

	world     *scene.Scene
	assets    *assets
	materials map[string]*scene.Material
	helpers   map[string]*scene.Mesh
	clock     float64
	started   bool

	pendingStarts []startOf // behaviours made and not started yet

	// while it plays: the physics of a 2D scene, the nodes with bodies, and whose body is whose
	physics    *physics.World
	bodied     []*Node
	bodyOwners map[*physics.Body]*Node

	// what each frame reuses, so a frame allocates nothing
	context     Context
	input       Input
	updating    []*Node
	drawn       []*Node                 // the nodes a 2D frame draws, furthest first
	flatColours map[string]draw2d.Color // shapes' colours, read from their text once
	drawnAspect float64                 // the screen's width over its height, as last drawn

	// what the scene was made from, to make it again, and the replacement a behaviour asked for
	description      *scenefile.Scene
	restartRequested bool
	changeRequested  string
}

// Load reads a scene file from a project's files.
func Load(files fs.FS, file string) (*Scene, error) {
	data, err := fs.ReadFile(files, file)
	if err != nil {
		return nil, failure.Wrap(SceneNotFound, file, err)
	}
	description, err := scenefile.Read(data)
	if err != nil {
		return nil, failure.Wrap(SceneUnreadable, file, err)
	}
	return Build(files, description)
}

// Build makes a scene from its description, finding its models in a project's files. A model
// that cannot be read leaves its node showing nothing; the first such error is returned with the
// scene.
func Build(files fs.FS, description *scenefile.Scene) (*Scene, error) {
	built := NewScene(files)
	built.Name, built.Mode, built.PixelsPerMetre = description.Name, description.Mode, description.PixelsPerMetre
	built.description = description
	built.SetSky(description.Sky)
	var firstError error
	for _, node := range description.Nodes {
		child, err := built.buildNode(node)
		if err != nil && firstError == nil {
			firstError = err
		}
		built.Root.Add(child)
	}
	built.Sync()
	return built, firstError
}

// NewScene is an empty scene, whose models come from a project's files.
func NewScene(files fs.FS) *Scene {
	built := &Scene{world: scene.New(), assets: newAssets(files), materials: map[string]*scene.Material{}, helpers: map[string]*scene.Mesh{},
		flatColours: map[string]draw2d.Color{}}
	built.Root = newNode(built, "")
	built.world.Nodes = []*scene.Node{built.Root.render}
	return built
}

// NewNode makes a node of the scene from a description (and its children), not yet in the scene:
// Add it to a node.
func (owner *Scene) NewNode(description scenefile.Node) (*Node, error) {
	return owner.buildNode(description)
}

func (owner *Scene) buildNode(description scenefile.Node) (*Node, error) {
	node := newNode(owner, description.Name)
	node.Transform = space.Transform{
		Position: space.NewVector(description.Position[0], description.Position[1], description.Position[2]),
		Rotation: space.Angles{X: description.Rotation[0], Y: description.Rotation[1], Z: description.Rotation[2]}.Rotation(),
		Scale:    space.NewVector(description.Scale[0], description.Scale[1], description.Scale[2]),
	}
	node.Visible = !description.Hidden
	node.shape, node.colour, node.model, node.sprite = description.Shape, description.Colour, description.Model, description.Sprite
	node.camera = description.Camera
	node.body, node.sheet, node.tiles = description.Body.Copy(), description.Sheet.Copy(), description.Tiles.Copy()
	node.Behaviours = append([]scenefile.Behaviour(nil), description.Behaviours...)
	var firstError error
	if node.model != "" {
		firstError = node.rebuildModel()
		if node.helper == nil && node.geometry == nil {
			node.rebuild()
		}
	} else {
		node.rebuild()
	}
	for _, childDescription := range description.Children {
		child, err := owner.buildNode(childDescription)
		if err != nil && firstError == nil {
			firstError = err
		}
		node.Add(child)
	}
	return node, firstError
}

// Describe writes the scene back as its file keeps it.
func (owner *Scene) Describe() *scenefile.Scene {
	description := &scenefile.Scene{Name: owner.Name, Sky: owner.Sky, Mode: owner.Mode, PixelsPerMetre: owner.PixelsPerMetre}
	for _, child := range owner.Root.children {
		description.Nodes = append(description.Nodes, child.Describe())
	}
	return description
}

// Describe writes a node, and its children, as a scene file keeps them.
func (node *Node) Describe() scenefile.Node {
	angles := node.Transform.Angles()
	description := scenefile.Node{
		Name:     node.Name,
		Position: [3]float64{node.Transform.Position.X, node.Transform.Position.Y, node.Transform.Position.Z},
		Rotation: [3]float64{rounded(angles.X), rounded(angles.Y), rounded(angles.Z)},
		Scale:    [3]float64{node.Transform.Scale.X, node.Transform.Scale.Y, node.Transform.Scale.Z},
		Hidden:   !node.Visible,
		Shape:    node.shape, Colour: node.colour, Model: node.model, Sprite: node.sprite, Camera: node.camera,
		Body: node.body.Copy(), Sheet: node.sheet.Copy(), Tiles: node.tiles.Copy(),
		Behaviours: append([]scenefile.Behaviour(nil), node.Behaviours...),
	}
	for _, child := range node.children {
		description.Children = append(description.Children, child.Describe())
	}
	return description
}

// rounded drops the last digits that turning angles into a rotation and back leaves behind.
func rounded(degrees float64) float64 { return math.Round(degrees*1e6)/1e6 + 0 } // + 0 turns -0 into 0

// SetMode makes the scene 3D or 2D: its nodes show their shapes and models in 3D, or their sprites
// and flat shapes in 2D.
func (owner *Scene) SetMode(mode scenefile.Mode) {
	owner.Mode = mode
	owner.Root.Walk(func(node *Node) bool {
		if node != owner.Root {
			node.rebuild()
		}
		return true
	})
	owner.Sync()
}

// Find finds a node by its path from the root, as "Car/Wheel".
func (owner *Scene) Find(path string) (*Node, bool) { return owner.Root.Find(path) }

// SetSky sets the background colour ("" for a daylight sky).
func (owner *Scene) SetSky(colour string) {
	owner.Sky = colour
	environment := scene.DefaultEnvironment()
	if colour != "" {
		if tint, valid := scene.ColorFromHex(colour); valid {
			environment.SkyZenith, environment.SkyHorizon, environment.FogColor = tint, tint, tint
			environment.Sky = &scene.Sky{}
		}
	}
	owner.world.Environment = environment
}

// World is what go-renderer draws, transforms brought up to date.
func (owner *Scene) World() *scene.Scene { return owner.world }

// Sync brings what go-renderer draws up to date with the nodes' transforms.
func (owner *Scene) Sync() { owner.Root.sync() }

// Time is the seconds the scene has been playing.
func (owner *Scene) Time() float64 { return owner.clock }

// material is the shared material of a colour.
func (owner *Scene) material(colour string) *scene.Material {
	if colour == "" {
		colour = "#c8c8c8"
	}
	if material, found := owner.materials[colour]; found {
		return material
	}
	tint, valid := scene.ColorFromHex(colour)
	if !valid {
		tint = graphics.Vec4{0.8, 0.8, 0.8, 1}
	}
	material := &scene.Material{Name: colour, Kind: scene.KindLit, BaseColor: tint}
	owner.materials[colour] = material
	return material
}

// helperFor is the marker the editor shows for a node with nothing else to show: a small dark box
// for a camera, a small sphere for an empty node.
func (owner *Scene) helperFor(node *Node) *scene.Node {
	if node.geometry != nil {
		return nil
	}
	kind, colour, size := "empty", graphics.Vec4{0.95, 0.6, 0.1, 1}, float32(0.25)
	shape := "sphere"
	if node.camera != nil {
		kind, colour, size, shape = "camera", graphics.Vec4{0.15, 0.15, 0.2, 1}, 0.4, "cube"
	}
	mesh, found := owner.helpers[kind]
	if !found {
		mesh = shapes.Mesh(shape, &scene.Material{Name: kind, Kind: scene.KindUnlit, BaseColor: colour})
		owner.helpers[kind] = mesh
	}
	helper := &scene.Node{Name: kind, Mesh: mesh, Hidden: !owner.Editing,
		Transform: graphics.Scale(graphics.Vec4{size, size, size, 0})}
	return helper
}

// MainCamera is the camera the scene is seen through: the first marked Main, or the first camera.
func (owner *Scene) MainCamera() (*Node, bool) {
	var first, main *Node
	owner.Root.Walk(func(node *Node) bool {
		if node.camera == nil || main != nil {
			return main == nil
		}
		if first == nil {
			first = node
		}
		if node.camera.Main {
			main = node
		}
		return true
	})
	if main != nil {
		return main, true
	}
	return first, first != nil
}

// View is what a camera node sees, as go-renderer's camera.
func (node *Node) View() models.Camera {
	world := node.World()
	settings := scenefile.Camera{Projection: scenefile.Perspective}
	if node.camera != nil {
		settings = *node.camera
	}
	if settings.Projection == scenefile.Isometric {
		// always the classic angle, whatever way the node is turned: isometric is a fixed view
		world.Rotation = space.Angles{X: scenefile.IsometricAngles[0], Y: scenefile.IsometricAngles[1]}.Rotation()
	}
	return CameraAt(world, settings)
}

// CameraAt is go-renderer's camera for a camera placed at a transform with some settings.
func CameraAt(world space.Transform, settings scenefile.Camera) models.Camera {
	position := world.Position
	target := position.Add(world.Forward())
	up := world.Up()
	view := models.Camera{
		Position: graphics.Vec4{float32(position.X), float32(position.Y), float32(position.Z), 1},
		Target:   graphics.Vec4{float32(target.X), float32(target.Y), float32(target.Z), 1},
		Up:       graphics.Vec4{float32(up.X), float32(up.Y), float32(up.Z), 0},
		Near:     0.05,
		Far:      2000,
	}
	switch settings.Projection {
	case scenefile.Orthographic, scenefile.Isometric:
		view.OrthographicHeight = float32(settings.Height)
		if view.OrthographicHeight <= 0 {
			view.OrthographicHeight = 10
		}
		view.Near = -1000 // an orthographic camera sees what is behind it as well
	default:
		fieldOfView := settings.FieldOfView
		if fieldOfView <= 0 {
			fieldOfView = 45
		}
		view.FieldOfView = space.Radians(fieldOfView)
	}
	return view
}

// assets reads a project's models, once each.
type assets struct {
	files  fs.FS
	models map[string]*scene.Scene
	images map[string]*draw2d.Image

	tilesets map[tilesetKey][]*draw2d.Image
	sheets   map[sheetKey][]*draw2d.Image
}

func newAssets(files fs.FS) *assets {
	return &assets{files: files, models: map[string]*scene.Scene{}, images: map[string]*draw2d.Image{},
		tilesets: map[tilesetKey][]*draw2d.Image{}, sheets: map[sheetKey][]*draw2d.Image{}}
}

// image reads a picture of the project's (PNG or JPEG), once.
func (library *assets) image(file string) (*draw2d.Image, error) {
	if picture, found := library.images[file]; found {
		return picture, nil
	}
	if library.files == nil {
		return nil, failure.Wrap(NoFiles, file, nil)
	}
	data, err := library.files.Open(file)
	if err != nil {
		return nil, failure.Wrap(PictureNotFound, file, err)
	}
	defer data.Close()
	picture, err := draw2d.DecodeImage(data)
	if err != nil {
		return nil, failure.Wrap(PictureUnreadable, file, err)
	}
	library.images[file] = picture
	return picture, nil
}

func (library *assets) model(file string) (*scene.Scene, error) {
	if model, found := library.models[file]; found {
		return model, nil
	}
	if library.files == nil {
		return nil, failure.Wrap(NoFiles, file, nil)
	}
	if path.Ext(file) != ".glb" {
		return nil, failure.Wrap(ModelFormat, file, nil)
	}
	data, err := fs.ReadFile(library.files, file)
	if err != nil {
		return nil, failure.Wrap(ModelNotFound, file, err)
	}
	model, err := loaders.ParseSceneGLB(data)
	if err != nil {
		return nil, failure.Wrap(ModelUnreadable, file, err)
	}
	library.models[file] = model
	return model, nil
}

// Forget makes the scene read a model or a picture again the next time a node shows it: after the
// file changed.
func (owner *Scene) Forget(file string) {
	delete(owner.assets.models, file)
	delete(owner.assets.images, file)
	for key := range owner.assets.tilesets {
		if key.file == file {
			delete(owner.assets.tilesets, key)
		}
	}
	for key := range owner.assets.sheets {
		if key.file == file {
			delete(owner.assets.sheets, key)
		}
	}
}
