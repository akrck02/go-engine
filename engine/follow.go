package engine

import (
	"math"

	"github.com/akrck02/go-engine/space"
)

// A camera of a 2D scene may follow a node (scenefile.Camera's Follow): every frame, after the
// bodies move, it comes nearer to the node, as fast as its Smoothing says, and then its view is
// kept inside the level's bounds. How wide the view is depends on the screen's shape, which the
// scene remembers from the last time it was drawn.

// defaultAspect is the screen's width over its height before the scene is first drawn: a wide
// screen's.
const defaultAspect = 16.0 / 9

// followTargets moves the cameras that follow a node.
func (owner *Scene) followTargets(seconds float64) {
	if !owner.Is2D() {
		return
	}
	owner.followFrom(owner.Root, seconds)
}

func (owner *Scene) followFrom(node *Node, seconds float64) {
	if node.camera != nil && node.camera.Follow != "" {
		owner.follow(node, seconds)
	}
	for _, child := range node.children {
		owner.followFrom(child, seconds)
	}
}

// follow brings a camera nearer to its target, and keeps its view inside the bounds.
func (owner *Scene) follow(camera *Node, seconds float64) {
	target, found := owner.Find(camera.camera.Follow)
	if !found {
		return
	}
	position := camera.World().Position
	wanted := target.World().Position
	share := catchUpShare(seconds, camera.camera.SmoothingSeconds)
	position.X += (wanted.X - position.X) * share
	position.Y += (wanted.Y - position.Y) * share
	if bounds, bounded := owner.cameraBounds(camera); bounded {
		position = owner.keepViewInside(position, camera, bounds)
	}
	owner.placeAt(camera, position)
}

// catchUpShare is the share of the way to its target a camera goes in some seconds: all of it
// with no smoothing, and with one, about two thirds of it in the smoothing's seconds, at an even
// pace whatever the frame rate.
func catchUpShare(seconds, smoothing float64) float64 {
	if smoothing <= 0 {
		return 1
	}
	return 1 - math.Exp(-seconds/smoothing)
}

// CameraBounds is the box a camera's view stays inside while it follows: its own, or the tile
// maps' together, with no top (an infinite one) unless the camera has a ceiling.
func (owner *Scene) CameraBounds(camera *Node) (space.Box, bool) {
	if camera.camera == nil {
		return space.EmptyBox(), false
	}
	return owner.cameraBounds(camera)
}

// cameraBounds is CameraBounds for a node known to be a camera.
func (owner *Scene) cameraBounds(camera *Node) (space.Box, bool) {
	if written := camera.camera.Bounds; written != nil {
		return space.Box{Minimum: space.NewVector(written[0], written[1], 0), Maximum: space.NewVector(written[2], written[3], 0)}, true
	}
	box := tileMapsBounds(owner.Root, space.EmptyBox())
	if box.IsEmpty() {
		return box, false
	}
	if !camera.camera.Ceiling {
		box.Maximum.Y = math.Inf(1)
	}
	return box, true
}

// tileMapsBounds grows a box by the tile maps under a node.
func tileMapsBounds(node *Node, box space.Box) space.Box {
	if node.tiles != nil {
		box = box.Union(node.tileBounds())
	}
	for _, child := range node.children {
		box = tileMapsBounds(child, box)
	}
	return box
}

// keepViewInside moves a camera's centre so its view stays inside the bounds; along an axis where
// the bounds are smaller than the view, it centres the view on them.
func (owner *Scene) keepViewInside(position space.Vector, camera *Node, bounds space.Box) space.Vector {
	height := owner.heightSeen(camera)
	halfWidth, halfHeight := height*owner.aspect()/2, height/2
	position.X = keepInside(position.X, halfWidth, bounds.Minimum.X, bounds.Maximum.X)
	position.Y = keepInside(position.Y, halfHeight, bounds.Minimum.Y, bounds.Maximum.Y)
	return position
}

// keepInside keeps a view's centre so that half its size either side stays between two edges.
func keepInside(centre, half, lowest, highest float64) float64 {
	if highest-lowest <= 2*half {
		return (lowest + highest) / 2
	}
	return math.Min(math.Max(centre, lowest+half), highest-half)
}

// aspect is the screen's width over its height, as last drawn.
func (owner *Scene) aspect() float64 {
	if owner.drawnAspect <= 0 {
		return defaultAspect
	}
	return owner.drawnAspect
}
