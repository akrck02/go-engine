package space

import "github.com/akrck02/go-renderer/graphics"

// Transform places a thing relative to its parent: it is scaled first, then turned, then moved.
type Transform struct {
	Position Vec3
	Rotation Quat
	Scale    Vec3
}

// NewTransform is a thing at the origin, unturned, at its own size.
func NewTransform() Transform { return Transform{Rotation: Identity, Scale: One} }

// At is a transform at a position.
func At(position Vec3) Transform {
	transform := NewTransform()
	transform.Position = position
	return transform
}

// Matrix is the transform as go-renderer draws it.
func (transform Transform) Matrix() graphics.Mat4 {
	rotation := transform.rotation()
	return graphics.FromTranslationRotationScale(
		graphics.Vec4{float32(transform.Position.X), float32(transform.Position.Y), float32(transform.Position.Z), 1},
		graphics.Vec4{float32(rotation.X), float32(rotation.Y), float32(rotation.Z), float32(rotation.W)},
		graphics.Vec4{float32(transform.Scale.X), float32(transform.Scale.Y), float32(transform.Scale.Z), 0},
	)
}

// rotation is the transform's rotation, Identity for a zero one (a transform made as a literal).
func (transform Transform) rotation() Quat {
	if transform.Rotation == (Quat{}) {
		return Identity
	}
	return transform.Rotation
}

// Point is where a point given in the thing's own space is in its parent's.
func (transform Transform) Point(local Vec3) Vec3 {
	return transform.rotation().Apply(local.Mul(transform.Scale)).Add(transform.Position)
}

// Direction is where a direction in the thing's own space points in its parent's (unscaled).
func (transform Transform) Direction(local Vec3) Vec3 { return transform.rotation().Apply(local) }

// Forward, Right and Up are the thing's own directions, in its parent's space.
func (transform Transform) Forward() Vec3 { return transform.Direction(Forward) }
func (transform Transform) Right() Vec3   { return transform.Direction(Right) }
func (transform Transform) Up() Vec3      { return transform.Direction(Up) }

// Euler is the rotation as angles in degrees; SetEuler sets it from them.
func (transform Transform) Euler() Euler           { return transform.rotation().Euler() }
func (transform *Transform) SetEuler(angles Euler) { transform.Rotation = angles.Rotation() }

// Turn turns the thing around an axis of its parent's space by some degrees.
func (transform *Transform) Turn(axis Vec3, degrees float64) {
	transform.Rotation = AxisAngle(axis, degrees).Mul(transform.rotation()).Normalized()
}

// Move moves the thing by an offset in its parent's space.
func (transform *Transform) Move(offset Vec3) { transform.Position = transform.Position.Add(offset) }

// LookAt turns the thing so its Forward points at a target, keeping its Up towards the sky.
func (transform *Transform) LookAt(target Vec3) {
	transform.Rotation = LookRotation(target.Sub(transform.Position), Up)
}

// Then is the transform of a child placed by child inside a thing placed by transform: where the
// child is in the thing's parent's space. A parent scaled unevenly and a turned child do not make
// a transform (they shear); Then keeps the scale per axis, as most engines do.
func (transform Transform) Then(child Transform) Transform {
	return Transform{
		Position: transform.Point(child.Position),
		Rotation: transform.rotation().Mul(child.rotation()).Normalized(),
		Scale:    transform.Scale.Mul(child.Scale),
	}
}

// Box is a box lined up with the axes, from its lowest corner to its highest.
type Box struct{ Minimum, Maximum Vec3 }

// EmptyBox holds nothing; growing it by a point makes a box of that point.
func EmptyBox() Box {
	huge := 1e300
	return Box{Minimum: Vec3{huge, huge, huge}, Maximum: Vec3{-huge, -huge, -huge}}
}

// IsEmpty tells whether the box holds nothing.
func (box Box) IsEmpty() bool { return box.Minimum.X > box.Maximum.X }

// Grow is the box that holds this one and a point.
func (box Box) Grow(point Vec3) Box {
	return Box{
		Minimum: Vec3{min(box.Minimum.X, point.X), min(box.Minimum.Y, point.Y), min(box.Minimum.Z, point.Z)},
		Maximum: Vec3{max(box.Maximum.X, point.X), max(box.Maximum.Y, point.Y), max(box.Maximum.Z, point.Z)},
	}
}

// Union is the box that holds both.
func (box Box) Union(other Box) Box {
	if other.IsEmpty() {
		return box
	}
	return box.Grow(other.Minimum).Grow(other.Maximum)
}

// Centre and Size are the box's middle and its extent along each axis.
func (box Box) Centre() Vec3 { return box.Minimum.Add(box.Maximum).Scale(0.5) }
func (box Box) Size() Vec3   { return box.Maximum.Sub(box.Minimum) }

// Corners are the box's eight corners.
func (box Box) Corners() [8]Vec3 {
	low, high := box.Minimum, box.Maximum
	return [8]Vec3{
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
		result = result.Grow(transform.Point(corner))
	}
	return result
}

// Hit is how far along a ray (from an origin, along a direction of length 1) it first enters the
// box, if it does.
func (box Box) Hit(origin, direction Vec3) (float64, bool) {
	near, far := -1e300, 1e300
	for axis := 0; axis < 3; axis++ {
		start, step := component(origin, axis), component(direction, axis)
		low, high := component(box.Minimum, axis), component(box.Maximum, axis)
		if step == 0 {
			if start < low || start > high {
				return 0, false
			}
			continue
		}
		enter, leave := (low-start)/step, (high-start)/step
		if enter > leave {
			enter, leave = leave, enter
		}
		near, far = max(near, enter), min(far, leave)
		if near > far || far < 0 {
			return 0, false
		}
	}
	return max(near, 0), true
}

func component(v Vec3, axis int) float64 {
	switch axis {
	case 0:
		return v.X
	case 1:
		return v.Y
	}
	return v.Z
}

// Local is where a point given in the parent's space is in the thing's own space: Point undone.
func (transform Transform) Local(point Vec3) Vec3 {
	turned := transform.rotation().Inverse().Apply(point.Sub(transform.Position))
	scale := transform.Scale
	return Vec3{divide(turned.X, scale.X), divide(turned.Y, scale.Y), divide(turned.Z, scale.Z)}
}

func divide(value, by float64) float64 {
	if by == 0 {
		return 0
	}
	return value / by
}
