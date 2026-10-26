// Package engine runs a game: a scene of nodes, read from the project's scene files, drawn by
// go-renderer, and moved by the game's own Go code through behaviours. A node is a thing in the
// world: a shape, a model, a sprite, a camera, or an empty point to hang others from. Its
// Transform places it relative to its parent, in metres and degrees (see package space).
//
// A game's main registers its behaviours and runs the project it embeds:
//
//	//go:embed engine.json scenes all:assets
//	var files embed.FS
//
//	func main() {
//		engine.Register(Spin{})
//		if err := engine.Run(files); err != nil {
//			log.Fatal(err)
//		}
//	}
package engine

import (
	"math"
	"strings"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/physics"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/shapes"
	"github.com/akrck02/go-engine/space"
)

// Node is a thing in a scene.
type Node struct {
	Name      string
	Transform space.Transform
	Visible   bool

	shape        string
	colour       string
	texture      string     // a built-in texture of the shape, or ""
	textureScale float64    // its repeats a metre; 0 is 1
	textureSize  [3]float64 // the scale the shape's UVs were laid for (fitTexture)
	model        string
	sprite       string
	camera       *scenefile.Camera

	// a 2D scene's parts, kept as the scene file has them
	body  *scenefile.Body
	sheet *scenefile.SpriteSheet
	tiles *scenefile.TileMap
	light *scenefile.Light

	startWith  string              // the animation it starts with, when not its sprite's own
	playing    scenefile.Animation // the sheet's animation playing, while the game plays
	playingFor float64             // seconds it has played
	facingLeft bool

	physicsBody *physics.Body   // the body in the scene's physics, while it plays
	tileBodies  []*physics.Body // a tile map's solid boxes, while it plays

	// Behaviours are the game's code attached to the node, as the scene file keeps them; while the
	// game runs, behaviours holds them made.
	Behaviours []scenefile.Behaviour
	behaviours []any

	parent   *Node
	children []*Node
	scene    *Scene

	render    *scene.Node // what go-renderer draws: the node's own geometry, then its children's
	geometry  *scene.Node // the shape or the model, inside render
	helper    *scene.Node // what the editor shows for a node with nothing to show (a camera, an empty)
	modelBase string      // the model the geometry was made from
	codeMesh  *scene.Mesh // a mesh the game's code gave the node, shown instead of its shape or model
	moving    bool        // the game moves the node every frame: lit on the GPU (SetMoving)
}

// newNode makes a node of a scene with nothing to show yet.
func newNode(owner *Scene, name string) *Node {
	node := &Node{Name: name, Transform: space.NewTransform(), Visible: true, scene: owner}
	node.render = &scene.Node{Name: name, Transform: graphics.Identity()}
	return node
}

// Parent is the node this one hangs from; nil for the scene's root.
func (node *Node) Parent() *Node { return node.parent }

// Children are the nodes that hang from this one. The slice is the node's own: do not change it.
func (node *Node) Children() []*Node { return node.children }

// Scene is the scene the node is in.
func (node *Node) Scene() *Scene { return node.scene }

// Path is the node's names from the root, as "Car/Wheel".
func (node *Node) Path() string {
	if node.parent == nil || node.parent.parent == nil {
		return node.Name
	}
	return node.parent.Path() + "/" + node.Name
}

// Find finds a node under this one by its path: "Wheel", or "Car/Wheel" for a deeper one.
func (node *Node) Find(path string) (*Node, bool) {
	name, rest, deeper := strings.Cut(path, "/")
	for _, child := range node.children {
		if child.Name != name {
			continue
		}
		if !deeper {
			return child, true
		}
		return child.Find(rest)
	}
	return nil, false
}

// Walk calls a function for the node and everything under it, parents before their children; it
// stops going down a branch when the function returns false.
func (node *Node) Walk(visit func(node *Node) bool) {
	if !visit(node) {
		return
	}
	for _, child := range node.children {
		child.Walk(visit)
	}
}

// Add hangs a node from this one, taking it from where it hung before. Its transform stays the
// same, so it moves with its new parent from where the parent is.
func (node *Node) Add(child *Node) {
	if child == node || child.contains(node) {
		return
	}
	child.detach()
	child.parent = node
	node.children = append(node.children, child)
	node.render.Children = append(node.render.Children, child.render)
	child.Walk(func(each *Node) bool { each.scene = node.scene; return true })
	if node.scene.physics != nil {
		node.scene.addBodies(child) // made while the scene plays, or moved from another: bodies it lacks
	}
}

// Insert hangs a node from this one at a place among its children.
func (node *Node) Insert(child *Node, at int) {
	node.Add(child)
	at = min(max(at, 0), len(node.children)-1)
	copy(node.children[at+1:], node.children[at:len(node.children)-1])
	node.children[at] = child
	node.rebuildRenderChildren()
}

// Remove takes the node, and what hangs from it, out of its scene, and their bodies out of its
// physics.
func (node *Node) Remove() {
	node.scene.removeBodies(node)
	node.detach()
}

// Index is where the node is among its parent's children.
func (node *Node) Index() int {
	if node.parent == nil {
		return 0
	}
	for index, sibling := range node.parent.children {
		if sibling == node {
			return index
		}
	}
	return 0
}

func (node *Node) detach() {
	parent := node.parent
	if parent == nil {
		return
	}
	children := parent.children[:0]
	for _, child := range parent.children {
		if child != node {
			children = append(children, child)
		}
	}
	clear(parent.children[len(children):])
	parent.children = children
	parent.rebuildRenderChildren()
	node.parent = nil
}

// rebuildRenderChildren puts the render node's children back in order: the node's own geometry
// and helper, then its children's.
func (node *Node) rebuildRenderChildren() {
	node.render.Children = node.render.Children[:0]
	if node.moving {
		markMoving(node.geometry, true) // a node not marked leaves its model's own flags as they are
	}
	if node.geometry != nil {
		node.render.Children = append(node.render.Children, node.geometry)
	}
	if node.helper != nil {
		node.render.Children = append(node.render.Children, node.helper)
	}
	for _, child := range node.children {
		node.render.Children = append(node.render.Children, child.render)
	}
}

// contains tells whether another node hangs, at any depth, from this one.
func (node *Node) contains(other *Node) bool {
	for each := other; each != nil; each = each.parent {
		if each == node {
			return true
		}
	}
	return false
}

// World is where the node is in the world, all its parents' transforms applied.
func (node *Node) World() space.Transform {
	if node.parent == nil {
		return node.Transform
	}
	return node.parent.World().Place(node.Transform)
}

// worldMatrix is World as go-renderer's matrix, multiplied the way go-renderer walks its nodes.
func (node *Node) worldMatrix() graphics.Matrix {
	if node.parent == nil {
		return node.Transform.Matrix()
	}
	return node.parent.worldMatrix().Multiply(node.Transform.Matrix())
}

// Shape is the shape the node is ("cube", "sphere", "plane", "cylinder"), or "".
func (node *Node) Shape() string { return node.shape }

// Colour is the colour of the node's shape, "#rrggbb".
func (node *Node) Colour() string { return node.colour }

// Texture is the built-in texture the node's shape wears, or "".
func (node *Node) Texture() string { return node.texture }

// TextureScale is how many times the shape's texture repeats a metre (0 is 1).
func (node *Node) TextureScale() float64 { return node.textureScale }

// SetTexture gives the node's shape a built-in texture ("" for none) repeated scale times a metre.
func (node *Node) SetTexture(name string, scale float64) {
	node.texture, node.textureScale = name, scale
	node.rebuild()
}

// Model is the model the node shows, by its path in the project, or "".
func (node *Node) Model() string { return node.model }

// Sprite is the picture the node shows in a 2D scene, by its path in the project, or "".
func (node *Node) Sprite() string { return node.sprite }

// SetSprite makes the node show a picture of the project's, or a sprite's file, in a 2D scene.
func (node *Node) SetSprite(path string) error {
	node.sprite = path
	if path == "" {
		return nil
	}
	if scenefile.IsSpriteFile(path) {
		if _, err := node.scene.assets.spriteFile(path); err != nil {
			return err
		}
	}
	_, err := node.scene.assets.image(node.picture())
	return err
}

// Camera is the node's camera settings, if it is a camera.
func (node *Node) Camera() (*scenefile.Camera, bool) { return node.camera, node.camera != nil }

// SetShape makes the node a shape of a colour ("" for no shape).
func (node *Node) SetShape(shape, colour string) {
	node.shape, node.colour, node.model = shape, colour, ""
	node.rebuild()
}

// SetModel makes the node show a model of the project's.
func (node *Node) SetModel(path string) error {
	node.shape, node.model = "", path
	return node.rebuildModel()
}

// SetCamera makes the node a camera, or no longer one with nil.
func (node *Node) SetCamera(camera *scenefile.Camera) {
	node.camera = camera
	node.rebuild()
}

// SetMesh makes the node show a mesh the game's code built (a vehicle from its design), instead
// of its shape or model; another call replaces it, and nil goes back to the shape or model. The
// mesh lives only while the game plays: scene files keep shapes and models.
func (node *Node) SetMesh(mesh *scene.Mesh) {
	node.codeMesh = mesh
	node.rebuild()
}

// Mesh is the mesh the game's code gave the node, if any.
func (node *Node) Mesh() (*scene.Mesh, bool) { return node.codeMesh, node.codeMesh != nil }

// SetMoving says the game moves or turns the node every frame. Go Renderer then lights it on the
// GPU each frame; otherwise the 3DS works its light out again for every way it is turned.
func (node *Node) SetMoving(moving bool) {
	node.moving = moving
	markMoving(node.geometry, moving)
}

// Moving tells whether the node is marked as moving every frame.
func (node *Node) Moving() bool { return node.moving }

// markMoving sets Go Renderer's moving flag on a render node and all it holds.
func markMoving(render *scene.Node, moving bool) {
	if render == nil {
		return
	}
	render.Moving = moving
	for _, child := range render.Children {
		markMoving(child, moving)
	}
}

// rebuild makes the node's geometry again after what it shows changed.
func (node *Node) rebuild() {
	node.geometry, node.helper = nil, nil
	if node.scene.Is2D() {
		node.rebuildRenderChildren() // a 2D scene is drawn flat, by Draw2D
		return
	}
	if node.codeMesh != nil {
		node.geometry = &scene.Node{Name: node.codeMesh.Name, Transform: graphics.Identity(), Mesh: node.codeMesh}
		node.rebuildRenderChildren()
		return
	}
	if node.shape != "" {
		mesh := shapes.Mesh(node.shape, node.scene.material(shapeLook{node.colour, node.texture, node.textureScale}))
		if mesh != nil {
			node.geometry = &scene.Node{Name: node.shape, Transform: graphics.Identity(), Mesh: mesh}
			node.fitTexture(true)
		}
	}
	if node.model != "" {
		node.rebuildModel()
		return
	}
	node.helper = node.scene.helperFor(node)
	node.rebuildRenderChildren()
}

func (node *Node) rebuildModel() error {
	node.geometry, node.helper = nil, nil
	defer node.rebuildRenderChildren()
	if node.model == "" || node.scene.Is2D() {
		return nil
	}
	model, err := node.scene.assets.model(node.model)
	if err != nil {
		node.helper = node.scene.helperFor(node)
		return err
	}
	node.geometry = &scene.Node{Name: node.model, Transform: graphics.Identity()}
	for _, part := range model.Nodes {
		node.geometry.Children = append(node.geometry.Children, cloneRenderNode(part))
	}
	node.modelBase = node.model
	return nil
}

// cloneRenderNode copies a model's node tree, sharing its meshes, so one model can stand in many
// places.
func cloneRenderNode(original *scene.Node) *scene.Node {
	copied := &scene.Node{Name: original.Name, Transform: original.Transform, Mesh: original.Mesh, Instances: original.Instances, Hidden: original.Hidden, Tags: original.Tags}
	for _, child := range original.Children {
		copied.Children = append(copied.Children, cloneRenderNode(child))
	}
	return copied
}

// Bounds is the box, in the world, around what the node itself shows (not its children); empty
// for a node that shows nothing.
func (node *Node) Bounds() space.Box {
	if node.scene.Is2D() {
		return node.flatBounds()
	}
	box := space.EmptyBox()
	world := node.worldMatrix()
	for _, own := range []*scene.Node{node.geometry, node.helper} {
		if own != nil {
			box = box.Union(renderBounds(own, world))
		}
	}
	return box
}

// renderBounds is the world box around a render node's meshes, under a matrix.
func renderBounds(node *scene.Node, parent graphics.Matrix) space.Box {
	box := space.EmptyBox()
	if node.Hidden {
		return box
	}
	world := parent.Multiply(node.Transform)
	if node.Mesh != nil && node.Mesh.VertexCount() > 0 {
		minimum, maximum := node.Mesh.Bounds()
		local := space.Box{Minimum: space.NewVector(float64(minimum[0]), float64(minimum[1]), float64(minimum[2])),
			Maximum: space.NewVector(float64(maximum[0]), float64(maximum[1]), float64(maximum[2]))}
		for _, corner := range local.Corners() {
			point := world.Apply(graphics.Vector{float32(corner.X), float32(corner.Y), float32(corner.Z), 1})
			box = box.Grow(space.NewVector(float64(point[0]), float64(point[1]), float64(point[2])))
		}
	}
	for _, child := range node.Children {
		box = box.Union(renderBounds(child, world))
	}
	return box
}

// fitTexture lays a textured shape's UVs in metres of the node as it is scaled, so its texture
// repeats as often a metre whatever the scale; again only when the scale changed (or when made),
// into the same memory. A shape without a texture keeps no UVs.
func (node *Node) fitTexture(made bool) {
	if node.geometry == nil || node.shape == "" || node.codeMesh != nil {
		return
	}
	mesh := node.geometry.Mesh
	if node.texture == "" {
		mesh.UVs = nil
		return
	}
	scale := node.Transform.Scale
	size := [3]float64{math.Abs(scale.X), math.Abs(scale.Y), math.Abs(scale.Z)}
	if !made && size == node.textureSize {
		return
	}
	shapes.StretchUVs(mesh, node.shape, size)
	node.textureSize, mesh.Dirty = size, !made
}

// sync copies the node's transform and visibility to what go-renderer draws.
func (node *Node) sync() {
	node.fitTexture(false)
	node.render.Transform = node.Transform.Matrix()
	node.render.Hidden = !node.Visible
	if node.helper != nil {
		node.helper.Hidden = !node.scene.Editing
	}
	for _, child := range node.children {
		child.sync()
	}
}
