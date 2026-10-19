package space

import "math"

// Box is a box lined up with the axes, from its lowest corner to its highest.
type Box struct{ Minimum, Maximum Vector }

// EmptyBox holds nothing; growing it by a point makes a box of that point.
func EmptyBox() Box {
	return Box{Minimum: Vector{math.Inf(1), math.Inf(1), math.Inf(1)}, Maximum: Vector{math.Inf(-1), math.Inf(-1), math.Inf(-1)}}
}

// IsEmpty tells whether the box holds nothing.
func (box Box) IsEmpty() bool { return box.Minimum.X > box.Maximum.X }

// Grow is the box that holds this one and a point.
func (box Box) Grow(point Vector) Box {
	return Box{
		Minimum: Vector{min(box.Minimum.X, point.X), min(box.Minimum.Y, point.Y), min(box.Minimum.Z, point.Z)},
		Maximum: Vector{max(box.Maximum.X, point.X), max(box.Maximum.Y, point.Y), max(box.Maximum.Z, point.Z)},
	}
}

// Union is the box that holds both.
func (box Box) Union(other Box) Box {
	if other.IsEmpty() {
		return box
	}
	return box.Grow(other.Minimum).Grow(other.Maximum)
}

// Centre is the box's middle.
func (box Box) Centre() Vector { return box.Minimum.Add(box.Maximum).Scale(0.5) }

// Size is the box's extent along each axis.
func (box Box) Size() Vector { return box.Maximum.Subtract(box.Minimum) }

// Contains tells whether a point is in the box, its faces included.
func (box Box) Contains(point Vector) bool {
	return point.X >= box.Minimum.X && point.X <= box.Maximum.X &&
		point.Y >= box.Minimum.Y && point.Y <= box.Maximum.Y &&
		point.Z >= box.Minimum.Z && point.Z <= box.Maximum.Z
}

// Corners are the box's eight corners.
func (box Box) Corners() [8]Vector {
	low, high := box.Minimum, box.Maximum
	return [8]Vector{
		{low.X, low.Y, low.Z}, {high.X, low.Y, low.Z}, {low.X, high.Y, low.Z}, {high.X, high.Y, low.Z},
		{low.X, low.Y, high.Z}, {high.X, low.Y, high.Z}, {low.X, high.Y, high.Z}, {high.X, high.Y, high.Z},
	}
}

// Transformed is the box, lined up with the axes again, that holds this box once transformed.
func (box Box) Transformed(transform Transform) Box {
	if box.IsEmpty() {
		return box
	}
	result := EmptyBox()
	for _, corner := range box.Corners() {
		result = result.Grow(transform.ToParent(corner))
	}
	return result
}

// Hit is how far along a ray (from an origin, along a direction of length 1) it first enters the
// box, if it does.
func (box Box) Hit(origin, direction Vector) (float64, bool) {
	nearest, farthest := math.Inf(-1), math.Inf(1)
	for _, axis := range []func(Vector) float64{alongX, alongY, alongZ} {
		enter, leave, crosses := box.slab(axis, origin, direction)
		if !crosses {
			return 0, false
		}
		nearest, farthest = max(nearest, enter), min(farthest, leave)
		if nearest > farthest || farthest < 0 {
			return 0, false
		}
	}
	return max(nearest, 0), true
}

// slab is where a ray enters and leaves the space between the box's two faces across one axis;
// a ray parallel to them crosses it only if it starts between them.
func (box Box) slab(axis func(Vector) float64, origin, direction Vector) (enter, leave float64, crosses bool) {
	start, step := axis(origin), axis(direction)
	low, high := axis(box.Minimum), axis(box.Maximum)
	if step == 0 {
		return math.Inf(-1), math.Inf(1), start >= low && start <= high
	}
	enter, leave = (low-start)/step, (high-start)/step
	if enter > leave {
		enter, leave = leave, enter
	}
	return enter, leave, true
}

func alongX(vector Vector) float64 { return vector.X }
func alongY(vector Vector) float64 { return vector.Y }
func alongZ(vector Vector) float64 { return vector.Z }
