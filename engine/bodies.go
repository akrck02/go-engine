package engine

import (
	"github.com/akrck02/go-engine/physics"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// A 2D scene's nodes with a body (scenefile.Body) move by its physics while the game plays: every
// frame, after the behaviours update, each body takes its node's position (so a behaviour may
// move a node by hand), the physics world steps, and each node takes its body's position.

// DefaultGravity is how fast a 2D scene's gravity pulls down, in metres a second each second:
// Earth's. A game changes it with Scene.Physics().SetGravity.
const DefaultGravity = 9.81

// Toucher is a behaviour that hears when its node's body begins touching another's: a trigger and
// the moving body that enters it. Both nodes' behaviours hear it, each with the other node.
type Toucher interface {
	Touched(context *Context, other *Node)
}

// Body is the node's body in the scene's physics, while the game plays: its velocity, whether it
// stands on the floor. A node without a body, or a scene not playing, has none.
func (node *Node) Body() (*physics.Body, bool) { return node.physicsBody, node.physicsBody != nil }

// BodySettings is the body the node is given in the scene file, to edit in place; a node without
// one has none.
func (node *Node) BodySettings() (*scenefile.Body, bool) { return node.body, node.body != nil }

// SetBodySettings gives the node a body in its scene's physics, or takes it away (nil). The body
// is made when the scene starts.
func (node *Node) SetBodySettings(body *scenefile.Body) { node.body = body }

// Physics is the scene's physics world while it plays; nil for a 3D scene, or one not started.
func (owner *Scene) Physics() *physics.World { return owner.physics }

// startPhysics makes the physics world and the bodies of the nodes that have one.
func (owner *Scene) startPhysics() {
	if !owner.Is2D() {
		return
	}
	owner.physics = physics.NewWorld(space.NewVector(0, -DefaultGravity, 0))
	owner.bodyOwners = map[*physics.Body]*Node{}
	owner.addBodies(owner.Root)
}

// addBodies gives a node, and what hangs from it, the bodies they lack.
func (owner *Scene) addBodies(node *Node) {
	if node.body != nil && node.physicsBody == nil {
		owner.addBody(node)
	}
	if node.tiles != nil && node.tileBodies == nil {
		owner.addTileBodies(node)
	}
	for _, child := range node.children {
		owner.addBodies(child)
	}
}

// addBody gives a node its body: its kind, its size (the one written, or the node's own sprite or
// shape) and its place.
func (owner *Scene) addBody(node *Node) {
	size := node.bodySize()
	body := &physics.Body{
		Kind: physicsKinds[node.body.Kind], Position: node.World().Position, Size: size,
		GravityScale: node.body.Gravity(),
	}
	node.physicsBody = body
	owner.bodyOwners[body] = node
	owner.bodied = append(owner.bodied, node)
	owner.physics.Add(body)
}

// bodySize is the width and height of the node's body: the size written, or the node's own.
func (node *Node) bodySize() space.Vector {
	size := space.NewVector(node.body.Size[0], node.body.Size[1], 0)
	if size.X == 0 || size.Y == 0 {
		width, height, _ := node.flatSize()
		size = space.NewVector(width, height, 0)
	}
	return size
}

// BodyBox is the box of the world the node's body takes, from its settings, playing or not: what
// the editor outlines. A node without a body has none.
func (node *Node) BodyBox() (space.Box, bool) {
	if node.body == nil {
		return space.EmptyBox(), false
	}
	half := node.bodySize().Scale(0.5)
	centre := node.World().Position
	return space.Box{Minimum: centre.Subtract(half), Maximum: centre.Add(half)}, true
}

// physicsKinds are the scene file's kinds of body, as the physics names them.
var physicsKinds = map[scenefile.BodyKind]physics.Kind{
	scenefile.Solid: physics.Solid, scenefile.Moving: physics.Moving, scenefile.Trigger: physics.Trigger,
	scenefile.Pushable: physics.Pushable, scenefile.Platform: physics.Platform,
}

// removeBodies takes a node's body, and those of what hangs from it, out of the physics.
func (owner *Scene) removeBodies(node *Node) {
	if owner.physics == nil {
		return
	}
	for _, tile := range node.tileBodies {
		owner.physics.Remove(tile)
		delete(owner.bodyOwners, tile)
	}
	node.tileBodies = nil
	if node.physicsBody != nil {
		owner.physics.Remove(node.physicsBody)
		delete(owner.bodyOwners, node.physicsBody)
		owner.bodied = withoutNode(owner.bodied, node)
		node.physicsBody = nil
	}
	for _, child := range node.children {
		owner.removeBodies(child)
	}
}

func withoutNode(nodes []*Node, removed *Node) []*Node {
	kept := nodes[:0]
	for _, node := range nodes {
		if node != removed {
			kept = append(kept, node)
		}
	}
	clear(nodes[len(kept):])
	return kept
}

// stepPhysics moves the bodies a frame and tells the behaviours what began touching.
func (owner *Scene) stepPhysics(seconds float64) {
	if owner.physics == nil {
		return
	}
	for _, node := range owner.bodied {
		node.physicsBody.Position = node.World().Position
	}
	owner.physics.Step(seconds)
	for _, node := range owner.bodied {
		if node.physicsBody.Kind.Moves() {
			owner.placeAt(node, node.physicsBody.Position)
		}
	}
	for _, touch := range owner.physics.Begun() {
		trigger, other := owner.bodyOwners[touch.Trigger], owner.bodyOwners[touch.Other]
		owner.tellTouched(trigger, other)
		owner.tellTouched(other, trigger)
	}
}

// placeAt puts a node at a point of the world, wherever it hangs from.
func (owner *Scene) placeAt(node *Node, world space.Vector) {
	if node.parent == nil || node.parent == owner.Root {
		node.Transform.Position = world
		return
	}
	node.Transform.Position = node.parent.World().FromParent(world)
}

// tellTouched tells a node's behaviours that it began touching another.
func (owner *Scene) tellTouched(node, other *Node) {
	if node == nil || other == nil {
		return
	}
	owner.context.Node = node
	for _, behaviour := range node.behaviours {
		if toucher, found := behaviour.(Toucher); found {
			toucher.Touched(&owner.context, other)
		}
	}
}
