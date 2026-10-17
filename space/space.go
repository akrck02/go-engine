// Package space is where things are: positions, directions, rotations and sizes in a game's world,
// and the transform that places a thing (where it is, how it is turned, how large it is).
//
// The world follows glTF and go-renderer: one unit is one metre, Y points up, and the axes are
// right-handed, so with X to the right and Y up, Z points towards whoever looks. A camera with no
// rotation looks along -Z (Forward). Angles are in degrees wherever a person reads or types them,
// and turning by a positive angle is counter-clockwise when looking down the axis towards the
// origin: a positive turn around Y takes Right (+X) towards Forward (-Z).
package space

import "math"

// Units: a length of 1 is a metre.
const (
	Metre      = 1.0
	Centimetre = 0.01
	Kilometre  = 1000.0
)

// Vec3 is a point or a direction in the world, in metres.
type Vec3 struct{ X, Y, Z float64 }

// Directions and points that have names.
var (
	Zero    = Vec3{}
	One     = Vec3{1, 1, 1}
	Right   = Vec3{1, 0, 0}
	Up      = Vec3{0, 1, 0}
	Back    = Vec3{0, 0, 1} // towards whoever looks at the world from the front
	Forward = Vec3{0, 0, -1}
)

// V makes a vector.
func V(x, y, z float64) Vec3 { return Vec3{x, y, z} }

func (v Vec3) Add(other Vec3) Vec3       { return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z} }
func (v Vec3) Sub(other Vec3) Vec3       { return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z} }
func (v Vec3) Scale(factor float64) Vec3 { return Vec3{v.X * factor, v.Y * factor, v.Z * factor} }

// Mul multiplies each axis by the other's: a size scaled per axis.
func (v Vec3) Mul(other Vec3) Vec3 { return Vec3{v.X * other.X, v.Y * other.Y, v.Z * other.Z} }

func (v Vec3) Dot(other Vec3) float64 { return v.X*other.X + v.Y*other.Y + v.Z*other.Z }

func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{v.Y*other.Z - v.Z*other.Y, v.Z*other.X - v.X*other.Z, v.X*other.Y - v.Y*other.X}
}

func (v Vec3) Length() float64 { return math.Sqrt(v.Dot(v)) }

// Distance is how far apart two points are.
func (v Vec3) Distance(other Vec3) float64 { return v.Sub(other).Length() }

// Normalized is the direction of v with a length of 1 (Zero stays Zero).
func (v Vec3) Normalized() Vec3 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return v.Scale(1 / length)
}

// Lerp is the point a share of the way from v to other (0 is v, 1 is other).
func (v Vec3) Lerp(other Vec3, share float64) Vec3 {
	return v.Add(other.Sub(v).Scale(share))
}

// Near tells whether two vectors are the same within a tolerance.
func (v Vec3) Near(other Vec3, tolerance float64) bool {
	return math.Abs(v.X-other.X) <= tolerance && math.Abs(v.Y-other.Y) <= tolerance && math.Abs(v.Z-other.Z) <= tolerance
}

// Radians turns degrees into radians; Degrees the other way.
func Radians(degrees float64) float64 { return degrees * math.Pi / 180 }
func Degrees(radians float64) float64 { return radians * 180 / math.Pi }
