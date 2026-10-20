// Package space is where things are: positions, directions, rotations and sizes in a game's world.
// Its Transform places a thing: where it is, how it is turned, and how large it is.
//
// The world follows glTF and go-renderer. One unit is one metre and Y points up. The axes are
// right-handed: with X to the right and Y up, Z points towards whoever looks. A camera with no
// rotation looks along -Z (Forward). Angles are in degrees wherever a person reads or types them.
// A positive turn is counter-clockwise, looking down the axis towards the origin: a positive turn
// around Up takes Right (+X) towards Forward (-Z).
package space

import "math"

// Units: a length of 1 is a metre.
const (
	Metre      = 1.0
	Centimetre = 0.01
	Kilometre  = 1000.0
)

// Vector is a point or a direction in the world, in metres.
type Vector struct{ X, Y, Z float64 }

// Directions and points that have names.
var (
	Zero    = Vector{}
	One     = Vector{1, 1, 1}
	Right   = Vector{1, 0, 0}
	Up      = Vector{0, 1, 0}
	Back    = Vector{0, 0, 1} // towards whoever looks at the world from the front
	Forward = Vector{0, 0, -1}
)

// NewVector makes a vector from its three coordinates.
func NewVector(x, y, z float64) Vector { return Vector{x, y, z} }

// Add is the vector moved by another.
func (vector Vector) Add(other Vector) Vector {
	return Vector{vector.X + other.X, vector.Y + other.Y, vector.Z + other.Z}
}

// Subtract is the vector from other to this one.
func (vector Vector) Subtract(other Vector) Vector {
	return Vector{vector.X - other.X, vector.Y - other.Y, vector.Z - other.Z}
}

// Scale is the vector times a number.
func (vector Vector) Scale(factor float64) Vector {
	return Vector{vector.X * factor, vector.Y * factor, vector.Z * factor}
}

// MultiplyEach multiplies each axis by the other vector's: a size scaled per axis.
func (vector Vector) MultiplyEach(other Vector) Vector {
	return Vector{vector.X * other.X, vector.Y * other.Y, vector.Z * other.Z}
}

// Dot is the dot product: how much two directions point the same way, times their lengths.
func (vector Vector) Dot(other Vector) float64 {
	return vector.X*other.X + vector.Y*other.Y + vector.Z*other.Z
}

// Cross is the cross product: a direction square to both, by the right-hand rule.
func (vector Vector) Cross(other Vector) Vector {
	return Vector{
		vector.Y*other.Z - vector.Z*other.Y,
		vector.Z*other.X - vector.X*other.Z,
		vector.X*other.Y - vector.Y*other.X,
	}
}

// Length is how long the vector is.
func (vector Vector) Length() float64 { return math.Sqrt(vector.Dot(vector)) }

// Distance is how far apart two points are.
func (vector Vector) Distance(other Vector) float64 { return vector.Subtract(other).Length() }

// Normalized is the vector's direction with a length of 1 (Zero stays Zero).
func (vector Vector) Normalized() Vector {
	length := vector.Length()
	if length == 0 {
		return vector
	}
	return vector.Scale(1 / length)
}

// Toward is the point a share of the way from this one to another: 0 is this point, 1 the other.
func (vector Vector) Toward(other Vector, share float64) Vector {
	return vector.Add(other.Subtract(vector).Scale(share))
}

// Near tells whether two vectors are the same within a tolerance on each axis.
func (vector Vector) Near(other Vector, tolerance float64) bool {
	return math.Abs(vector.X-other.X) <= tolerance &&
		math.Abs(vector.Y-other.Y) <= tolerance &&
		math.Abs(vector.Z-other.Z) <= tolerance
}

// Radians turns degrees into radians.
func Radians(degrees float64) float64 { return degrees * math.Pi / 180 }

// Degrees turns radians into degrees.
func Degrees(radians float64) float64 { return radians * 180 / math.Pi }
