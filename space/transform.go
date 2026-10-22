package space

import "github.com/akrck02/go-renderer/graphics"

// Transform places a thing relative to its parent: it is scaled first, then turned, then moved.
type Transform struct {
	Position Vector
	Rotation Rotation
	Scale    Vector
}

// NewTransform is a thing at the origin, unturned, at its own size.
func NewTransform() Transform { return Transform{Rotation: NoRotation, Scale: One} }

// At is a transform at a position, unturned, at its own size.
func At(position Vector) Transform {
	transform := NewTransform()
	transform.Position = position
	return transform
}

// Matrix is the transform as go-renderer draws it.
func (transform Transform) Matrix() graphics.Matrix {
	imaginaryX, imaginaryY, imaginaryZ, real := transform.rotation().Parts()
	return graphics.FromTranslationRotationScale(
		graphics.Vector{float32(transform.Position.X), float32(transform.Position.Y), float32(transform.Position.Z), 1},
		graphics.Vector{float32(imaginaryX), float32(imaginaryY), float32(imaginaryZ), float32(real)},
		graphics.Vector{float32(transform.Scale.X), float32(transform.Scale.Y), float32(transform.Scale.Z), 0},
	)
}

// rotation is the transform's rotation, NoRotation for a zero one (a transform made as a literal).
func (transform Transform) rotation() Rotation {
	if transform.Rotation.IsZero() {
		return NoRotation
	}
	return transform.Rotation
}

// ToParent is where a point given in the thing's own space is in its parent's space.
func (transform Transform) ToParent(local Vector) Vector {
	return transform.rotation().Rotate(local.MultiplyEach(transform.Scale)).Add(transform.Position)
}

// FromParent is where a point given in the parent's space is in the thing's own space: ToParent
// undone.
func (transform Transform) FromParent(point Vector) Vector {
	turned := transform.rotation().Inverse().Rotate(point.Subtract(transform.Position))
	return Vector{divide(turned.X, transform.Scale.X), divide(turned.Y, transform.Scale.Y), divide(turned.Z, transform.Scale.Z)}
}

// divide is a division that gives 0 instead of failing when the divisor is 0: a thing scaled to
// nothing on an axis.
func divide(value, divisor float64) float64 {
	if divisor == 0 {
		return 0
	}
	return value / divisor
}

// DirectionToParent is where a direction in the thing's own space points in its parent's
// (turned, not scaled).
func (transform Transform) DirectionToParent(local Vector) Vector {
	return transform.rotation().Rotate(local)
}

// Forward is the way the thing faces, in its parent's space.
func (transform Transform) Forward() Vector { return transform.DirectionToParent(Forward) }

// Right is the thing's right, in its parent's space.
func (transform Transform) Right() Vector { return transform.DirectionToParent(Right) }

// Up is the thing's up, in its parent's space.
func (transform Transform) Up() Vector { return transform.DirectionToParent(Up) }

// Angles are the rotation as angles in degrees.
func (transform Transform) Angles() Angles { return transform.rotation().Angles() }

// SetAngles sets the rotation from angles in degrees.
func (transform *Transform) SetAngles(angles Angles) { transform.Rotation = angles.Rotation() }

// Turn turns the thing around an axis of its parent's space by some degrees.
func (transform *Transform) Turn(axis Vector, degrees float64) {
	transform.Rotation = transform.rotation().Then(RotationAround(axis, degrees)).Normalized()
}

// Move moves the thing by an offset in its parent's space.
func (transform *Transform) Move(offset Vector) { transform.Position = transform.Position.Add(offset) }

// LookAt turns the thing so its Forward points at a target, keeping its Up towards the sky.
func (transform *Transform) LookAt(target Vector) {
	transform.Rotation = RotationLookingAt(target.Subtract(transform.Position), Up)
}

// Place is where a child is, placed by its own transform inside the thing this transform places:
// the child's transform in the thing's parent's space. A parent scaled unevenly with a turned
// child would shear it; Place keeps the scale per axis instead, as most engines do.
func (transform Transform) Place(child Transform) Transform {
	return Transform{
		Position: transform.ToParent(child.Position),
		Rotation: child.rotation().Then(transform.rotation()).Normalized(),
		Scale:    transform.Scale.MultiplyEach(child.Scale),
	}
}
