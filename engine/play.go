package engine

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/space"
)

// Context is what a behaviour is given each time it runs.
type Context struct {
	Node  *Node   // the node the behaviour is attached to
	Scene *Scene  // the scene it is in
	Delta float64 // seconds since the last frame
	Time  float64 // seconds since the scene started
	Input *Input
}

// Find finds a node of the scene by its path, as "Car/Wheel".
func (context *Context) Find(path string) (*Node, bool) { return context.Scene.Find(path) }

// Start makes the scene's behaviours and runs their Start, once: the engine does it before the
// first frame. A behaviour that is not registered is reported and left out; the game goes on.
func (owner *Scene) Start() []error {
	if owner.started {
		return nil
	}
	owner.started = true
	defer owner.startPhysics()
	defer owner.startAnimations(owner.Root)
	var problems []error
	owner.Root.Walk(func(node *Node) bool {
		for _, description := range node.Behaviours {
			behaviour, err := makeBehaviour(description.Name, description.Fields)
			if err != nil {
				problems = append(problems, err)
				continue
			}
			node.behaviours = append(node.behaviours, behaviour)
			owner.pendingStarts = append(owner.pendingStarts, startOf{node, behaviour})
		}
		return true
	})
	return problems
}

// Update moves the scene on by some seconds: behaviours that were just made start, every
// behaviour updates, the bodies of a 2D scene move, the tweens move, and what go-renderer draws
// follows. A frame allocates nothing it can avoid.
func (owner *Scene) Update(seconds float64, keys *input.State) {
	owner.beginFrame(seconds, keys)
	owner.advance(seconds)
}

// beginFrame moves the scene's clock on and sets the frame's context.
func (owner *Scene) beginFrame(seconds float64, keys *input.State) {
	owner.Start()
	owner.clock += seconds
	owner.input.state = keys
	owner.context = Context{Scene: owner, Delta: seconds, Time: owner.clock, Input: &owner.input}
}

// advance runs the frame's behaviours, physics and animations, and syncs what is drawn.
func (owner *Scene) advance(seconds float64) {
	owner.startPending()
	owner.updateBehaviours()
	owner.stepPhysics(seconds)
	owner.advanceAnimations(owner.Root, seconds)
	owner.followTargets(seconds)
	owner.Animate.Update(seconds)
	owner.Sync()
}

// startPending runs Start on the behaviours made since the last frame.
func (owner *Scene) startPending() {
	starting := owner.pendingStarts
	owner.pendingStarts = nil
	for _, pending := range starting {
		if starter, found := pending.behaviour.(Starter); found {
			owner.context.Node = pending.node
			starter.Start(&owner.context)
		}
	}
}

// updateBehaviours runs Update on every behaviour, node by node in the scene's order.
func (owner *Scene) updateBehaviours() {
	owner.updating = owner.collectBehaving(owner.updating[:0], owner.Root)
	for _, node := range owner.updating {
		owner.context.Node = node
		for _, behaviour := range node.behaviours {
			if updater, found := behaviour.(Updater); found {
				updater.Update(&owner.context)
			}
		}
	}
}

// collectBehaving lists the nodes with behaviours, a node before its children.
func (owner *Scene) collectBehaving(nodes []*Node, node *Node) []*Node {
	if len(node.behaviours) > 0 {
		nodes = append(nodes, node)
	}
	for _, child := range node.children {
		nodes = owner.collectBehaving(nodes, child)
	}
	return nodes
}

// Input is the keyboard, the mouse and the controls, as the game reads them.
type Input struct {
	state   *input.State
	console ctr.Reading // the 3DS's controls this frame; empty elsewhere (controls.go)
	// lastKind is the kind of input the player used last (hints.go)
	lastKind InputKind
}

// Down tells whether a key is held; Pressed whether it went down this frame.
func (keys *Input) Down(key Key) bool    { return keys.state != nil && keys.state.Down(key) }
func (keys *Input) Pressed(key Key) bool { return keys.state != nil && keys.state.Pressed(key) }

// Axis is -1 while the first key is held, 1 while the second is, 0 with neither or both.
func (keys *Input) Axis(negative, positive Key) float64 {
	value := 0.0
	if keys.Down(negative) {
		value--
	}
	if keys.Down(positive) {
		value++
	}
	return value
}

// Move is the direction the arrows (or WASD, or a 3DS's cross and circle pad) point, on the
// ground: right is +X, up is Forward (-Z). Its length is at most 1.
func (keys *Input) Move() space.Vector {
	x := keys.Axis(KeyLeft, KeyRight) + keys.Axis(KeyA, KeyD)
	forward := keys.Axis(KeyDown, KeyUp) + keys.Axis(KeyS, KeyW)
	if keys.state != nil {
		x += keys.state.StickX
		forward += keys.state.StickY // up on the stick is forward
	}
	direction := space.NewVector(x, 0, -forward)
	if direction.Length() > 1 {
		direction = direction.Normalized()
	}
	return direction
}

// Pointer is where the mouse (or a finger) is, in window pixels, and whether it is pressed.
func (keys *Input) Pointer() (x, y float64, pressed bool) {
	if keys.state == nil {
		return 0, 0, false
	}
	return keys.state.CursorX, keys.state.CursorY, keys.state.Button(input.MouseLeft)
}

// State is go-renderer's input, for what Input does not cover.
func (keys *Input) State() *input.State { return keys.state }

// Key is a key of the keyboard (a 3DS's buttons stand for some: the cross is the arrows, A is
// Enter, B is Escape).
type Key = input.Key

// Keys.
const (
	KeyA         = input.KeyA
	KeyB         = input.KeyB
	KeyC         = input.KeyC
	KeyD         = input.KeyD
	KeyE         = input.KeyE
	KeyF         = input.KeyF
	KeyG         = input.KeyG
	KeyH         = input.KeyH
	KeyI         = input.KeyI
	KeyJ         = input.KeyJ
	KeyK         = input.KeyK
	KeyL         = input.KeyL
	KeyM         = input.KeyM
	KeyN         = input.KeyN
	KeyO         = input.KeyO
	KeyP         = input.KeyP
	KeyQ         = input.KeyQ
	KeyR         = input.KeyR
	KeyS         = input.KeyS
	KeyT         = input.KeyT
	KeyU         = input.KeyU
	KeyV         = input.KeyV
	KeyW         = input.KeyW
	KeyX         = input.KeyX
	KeyY         = input.KeyY
	KeyZ         = input.KeyZ
	KeyUp        = input.KeyUp
	KeyDown      = input.KeyDown
	KeyLeft      = input.KeyLeft
	KeyRight     = input.KeyRight
	KeySpace     = input.KeySpace
	KeyEnter     = input.KeyEnter
	KeyEscape    = input.KeyEscape
	KeyShift     = input.KeyShift
	KeyControl   = input.KeyControl
	KeyTab       = input.KeyTab
	KeyBackspace = input.KeyBackspace
)
