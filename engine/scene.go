package engine

import (
	"bytes"
	"io/fs"
	"log/slog"
	"math"
	"path"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/loaders"
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
	// Ambient is the light that reaches everything in a 2D scene; nil is daylight when the scene
	// has lights (see lights.go).
	Ambient *scenefile.Ambient

	// Editing shows what only the editor shows: a marker for cameras and empty nodes.
	Editing bool
	// Animate runs the scene's tweens; the engine moves it on every frame.
	Animate animation.Player

	world     *scene.Scene
	assets    *assets
	materials map[shapeLook]*scene.Material
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
	lit         []*Node                 // the lights a 2D frame draws
	ownLights   []scene.PointLight      // the program's own point lights (SetPointLights)
	flatColours map[string]draw2d.Color // shapes' colours, read from their text once
	drawnAspect float64                 // the screen's width over its height, as last drawn

	// what the scene was made from, to make it again, and the replacement a behaviour asked for
	description      *scenefile.Scene
	restartRequested bool
	changeRequested  string
}

// Load reads a scene file from a project's files.
func Load(files fs.FS, file string) (*Scene, error) { return loadWith(newAssets(files), file) }

// loadWith reads a scene file, with pictures and models already read.
func loadWith(library *assets, file string) (*Scene, error) {
	data, err := fs.ReadFile(library.files, file)
	if err != nil {
		return nil, failure.Wrap(SceneNotFound, file, err)
	}
	description, err := scenefile.Read(data)
	if err != nil {
		return nil, failure.Wrap(SceneUnreadable, file, err)
	}
	return buildWith(library, description)
}

// Build makes a scene from its description, finding its models in a project's files. A model
// that cannot be read leaves its node showing nothing; the first such error is returned with the
// scene.
func Build(files fs.FS, description *scenefile.Scene) (*Scene, error) {
	return buildWith(newAssets(files), description)
}

// buildWith makes a scene from its description, with pictures and models already read.
func buildWith(library *assets, description *scenefile.Scene) (*Scene, error) {
	built := newSceneWith(library)
	built.Name, built.Mode, built.PixelsPerMetre = description.Name, description.Mode, description.PixelsPerMetre
	built.Ambient = description.Ambient.Copy()
	built.description = description
	built.SetSky(description.Sky) // applies the ambient too
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
func NewScene(files fs.FS) *Scene { return newSceneWith(newAssets(files)) }

// newSceneWith is an empty scene, with pictures and models already read.
func newSceneWith(library *assets) *Scene {
	built := &Scene{world: scene.New(), assets: library, materials: map[shapeLook]*scene.Material{}, helpers: map[string]*scene.Mesh{},
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
	node.texture, node.textureScale = description.Texture, description.TextureScale
	node.camera = description.Camera
	node.body, node.sheet, node.tiles = description.Body.Copy(), description.Sheet.Copy(), description.Tiles.Copy()
	node.light = description.Light.Copy()
	node.startWith = description.Animation
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
	description := &scenefile.Scene{Name: owner.Name, Sky: owner.Sky, Mode: owner.Mode, PixelsPerMetre: owner.PixelsPerMetre,
		Ambient: owner.Ambient.Copy()}
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
		Shape:    node.shape, Colour: node.colour, Texture: node.texture, TextureScale: node.textureScale, Model: node.model, Sprite: node.sprite, Camera: node.camera,
		Body: node.body.Copy(), Sheet: node.sheet.Copy(), Tiles: node.tiles.Copy(), Animation: node.startWith,
		Light:      node.light.Copy(),
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
		if tint, valid := colourFromHex(colour); valid {
			environment.SkyZenith, environment.SkyHorizon, environment.FogColor = tint, tint, tint
			environment.Sky = &scene.Sky{}
		}
	}
	owner.world.Environment = environment
	owner.applyAmbient()
}

// World is what go-renderer draws, transforms brought up to date.
func (owner *Scene) World() *scene.Scene { return owner.world }

// Sync brings what go-renderer draws up to date with the nodes' transforms.
func (owner *Scene) Sync() {
	owner.Root.sync()
	owner.syncLights()
}

// Time is the seconds the scene has been playing.
func (owner *Scene) Time() float64 { return owner.clock }

// colourFromHex reads a colour written as on a web page ("#rrggbb", "#rgb") as an opaque one;
// false for anything else.
func colourFromHex(text string) (graphics.Vector, bool) {
	colour, err := draw2d.ParseHex(text)
	if err != nil {
		return graphics.Vector{}, false // the caller keeps its default colour
	}
	return graphics.Vector{colour[0], colour[1], colour[2], 1}, true
}

// shapeLook is how a shape looks: its colour, and the built-in texture it wears and how often.
type shapeLook struct {
	colour       string
	texture      string
	textureScale float64
}

// material is the shared material of a shape's look.
func (owner *Scene) material(look shapeLook) *scene.Material {
	if look.colour == "" {
		look.colour = "#c8c8c8"
	}
	if material, found := owner.materials[look]; found {
		return material
	}
	tint, valid := colourFromHex(look.colour)
	if !valid {
		tint = graphics.Vector{0.8, 0.8, 0.8, 1}
	}
	material := &scene.Material{Name: look.colour, Kind: scene.LitKind, BaseColor: tint}
	if texture, found := scene.BuiltInTexture(look.texture); found {
		repeats := float32(look.textureScale)
		material.Texture, material.TextureScale = texture, [2]float32{repeats, repeats}
	}
	owner.materials[look] = material
	return material
}

// helperFor is the marker the editor shows for a node with nothing else to show: a small dark box
// for a camera, a small sphere for an empty node.
func (owner *Scene) helperFor(node *Node) *scene.Node {
	if node.geometry != nil {
		return nil
	}
	kind, colour, size := "empty", graphics.Vector{0.95, 0.6, 0.1, 1}, float32(0.25)
	shape := "sphere"
	if node.camera != nil {
		kind, colour, size, shape = "camera", graphics.Vector{0.15, 0.15, 0.2, 1}, 0.4, "cube"
	}
	mesh, found := owner.helpers[kind]
	if !found {
		mesh = shapes.Mesh(shape, &scene.Material{Name: kind, Kind: scene.UnlitKind, BaseColor: colour})
		owner.helpers[kind] = mesh
	}
	helper := &scene.Node{Name: kind, Mesh: mesh, Hidden: !owner.Editing,
		Transform: graphics.Scale(graphics.Vector{size, size, size, 0})}
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
func (node *Node) View() camera.Lens {
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
func CameraAt(world space.Transform, settings scenefile.Camera) camera.Lens {
	position := world.Position
	target := position.Add(world.Forward().Scale(orDefault(settings.Focus, DefaultFocus)))
	up := world.Up()
	view := camera.Lens{
		Position: graphics.Vector{float32(position.X), float32(position.Y), float32(position.Z), 1},
		Target:   graphics.Vector{float32(target.X), float32(target.Y), float32(target.Z), 1},
		Up:       graphics.Vector{float32(up.X), float32(up.Y), float32(up.Z), 0},
		Near:     float32(orDefault(settings.Near, DefaultNear)),
		Far:      float32(orDefault(settings.Far, DefaultFar)),
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
		view.FieldOfViewRadians = float32(space.Radians(fieldOfView))
	}
	return view
}

// A camera's distances when its settings leave them at 0, in metres.
const (
	DefaultNear  = 0.05
	DefaultFar   = 2000
	DefaultFocus = 10 // a room's depth: most scenes look at things about this far
)

// orDefault is a value, or another when the value is not set (0 or less).
func orDefault(value, otherwise float64) float64 {
	if value <= 0 {
		return otherwise
	}
	return value
}

// assets reads a project's models, once each.
type assets struct {
	files  fs.FS
	models map[string]*scene.Scene
	images map[string]*draw2d.Image

	tilesets map[tilesetKey][]*draw2d.Image
	sheets   map[sheetKey][]*draw2d.Image

	sprites      map[string]*scenefile.Sprite  // sprites' files, read once
	tilesetFiles map[string]*scenefile.Tileset // tilesets' files, read once

	// failed are the pictures that could not be read, with why: tried once and reported once, not
	// on every frame that draws them
	failed map[string]error
}

func newAssets(files fs.FS) *assets {
	return &assets{files: files, models: map[string]*scene.Scene{}, images: map[string]*draw2d.Image{},
		tilesets: map[tilesetKey][]*draw2d.Image{}, sheets: map[sheetKey][]*draw2d.Image{},
		sprites: map[string]*scenefile.Sprite{}, tilesetFiles: map[string]*scenefile.Tileset{}, failed: map[string]error{}}
}

// image reads a picture of the project's (PNG or JPEG), once; one that cannot be read is
// reported once, and its error kept.
func (library *assets) image(file string) (*draw2d.Image, error) {
	if picture, found := library.images[file]; found {
		return picture, nil
	}
	if err, failed := library.failed[file]; failed {
		return nil, err
	}
	picture, err := library.readImage(file)
	if err != nil {
		library.failed[file] = err
		slog.Warn("a picture could not be read; what shows it draws nothing", "error", err)
		return nil, err
	}
	library.images[file] = picture
	return picture, nil
}

// pixelArt marks a picture to be drawn at the nearest pixel when scaled: a 2D scene's sprites and
// tiles are pixel art, whose pixels should stay square blocks, not blend into their neighbours.
func pixelArt(picture *draw2d.Image) *draw2d.Image {
	picture.Sharp = true
	return picture
}

// readImage reads and decodes a picture of the project's.
func (library *assets) readImage(file string) (*draw2d.Image, error) {
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
	return pixelArt(picture), nil
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
	model, err := loaders.ReadScene(bytes.NewReader(data))
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
	delete(owner.assets.sprites, file)
	delete(owner.assets.failed, file)
	delete(owner.assets.tilesetFiles, file)
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
