package space

import "math"

// Quat is a rotation, as a unit quaternion. The zero value is not a rotation: start from Identity,
// from Euler or from AxisAngle.
type Quat struct{ X, Y, Z, W float64 }

// Identity is no rotation at all.
var Identity = Quat{W: 1}

// Euler is a rotation as three angles in degrees, the way an inspector shows it: X tilts (pitch),
// Y turns (yaw) and Z rolls. They apply roll first, then pitch, then yaw, so the yaw is always
// around the world's vertical, as a character or a camera turns.
type Euler struct{ X, Y, Z float64 }

// AxisAngle is a rotation of some degrees around an axis.
func AxisAngle(axis Vec3, degrees float64) Quat {
	axis = axis.Normalized()
	half := Radians(degrees) / 2
	sine := math.Sin(half)
	return Quat{axis.X * sine, axis.Y * sine, axis.Z * sine, math.Cos(half)}
}

// Rotation is the rotation the angles describe.
func (angles Euler) Rotation() Quat {
	yaw := AxisAngle(Up, angles.Y)
	pitch := AxisAngle(Right, angles.X)
	roll := AxisAngle(Back, angles.Z)
	return yaw.Mul(pitch).Mul(roll)
}

// Euler is the rotation as angles in degrees, each between -180 and 180 (pitch between -90 and
// 90).
func (rotation Quat) Euler() Euler {
	q := rotation.Normalized()
	// the rotation matrix's entries that the yaw-pitch-roll order leaves simple
	m12 := 2 * (q.Y*q.Z - q.W*q.X)
	m02 := 2 * (q.X*q.Z + q.W*q.Y)
	m22 := 1 - 2*(q.X*q.X+q.Y*q.Y)
	m10 := 2 * (q.X*q.Y + q.W*q.Z)
	m11 := 1 - 2*(q.X*q.X+q.Z*q.Z)
	pitch := math.Asin(math.Max(-1, math.Min(1, -m12)))
	if math.Abs(m12) > 0.9999 {
		// looking straight up or down: the yaw and the roll turn about the same axis, so all of
		// it is yaw
		m00 := 1 - 2*(q.Y*q.Y+q.Z*q.Z)
		m20 := 2 * (q.X*q.Z - q.W*q.Y)
		return Euler{Degrees(pitch), Degrees(math.Atan2(-m20, m00)), 0}
	}
	return Euler{Degrees(pitch), Degrees(math.Atan2(m02, m22)), Degrees(math.Atan2(m10, m11))}
}

// Mul is the rotation that turns by other first, then by rotation.
func (rotation Quat) Mul(other Quat) Quat {
	a, b := rotation, other
	return Quat{
		a.W*b.X + a.X*b.W + a.Y*b.Z - a.Z*b.Y,
		a.W*b.Y - a.X*b.Z + a.Y*b.W + a.Z*b.X,
		a.W*b.Z + a.X*b.Y - a.Y*b.X + a.Z*b.W,
		a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z,
	}
}

// Normalized is the rotation with a length of 1, as small errors pile up after many turns.
func (rotation Quat) Normalized() Quat {
	length := math.Sqrt(rotation.X*rotation.X + rotation.Y*rotation.Y + rotation.Z*rotation.Z + rotation.W*rotation.W)
	if length == 0 {
		return Identity
	}
	return Quat{rotation.X / length, rotation.Y / length, rotation.Z / length, rotation.W / length}
}

// Inverse is the rotation that undoes this one.
func (rotation Quat) Inverse() Quat {
	q := rotation.Normalized()
	return Quat{-q.X, -q.Y, -q.Z, q.W}
}

// Apply turns a vector by the rotation.
func (rotation Quat) Apply(v Vec3) Vec3 {
	q := rotation.Normalized()
	axis := Vec3{q.X, q.Y, q.Z}
	twice := axis.Cross(v).Scale(2)
	return v.Add(twice.Scale(q.W)).Add(axis.Cross(twice))
}

// Slerp is the rotation a share of the way from rotation to other, turning at an even pace the
// shorter way round.
func (rotation Quat) Slerp(other Quat, share float64) Quat {
	a, b := rotation.Normalized(), other.Normalized()
	cosine := a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
	if cosine < 0 { // the same rotation has two quaternions; take the nearer
		b, cosine = Quat{-b.X, -b.Y, -b.Z, -b.W}, -cosine
	}
	if cosine > 0.9995 {
		return Quat{a.X + (b.X-a.X)*share, a.Y + (b.Y-a.Y)*share, a.Z + (b.Z-a.Z)*share, a.W + (b.W-a.W)*share}.Normalized()
	}
	angle := math.Acos(cosine)
	sine := math.Sin(angle)
	from, to := math.Sin((1-share)*angle)/sine, math.Sin(share*angle)/sine
	return Quat{a.X*from + b.X*to, a.Y*from + b.Y*to, a.Z*from + b.Z*to, a.W*from + b.W*to}
}

// LookRotation is the rotation that makes Forward point along a direction, keeping Up as near
// the given up as it can.
func LookRotation(direction, up Vec3) Quat {
	forward := direction.Normalized()
	if forward == Zero {
		return Identity
	}
	right := forward.Cross(up).Normalized()
	if right == Zero { // looking straight along up: any right will do
		right = forward.Cross(Back).Normalized()
		if right == Zero {
			right = Right
		}
	}
	trueUp := right.Cross(forward)
	back := forward.Scale(-1)
	// the matrix with columns right, up, back, as a quaternion
	m00, m11, m22 := right.X, trueUp.Y, back.Z
	trace := m00 + m11 + m22
	switch {
	case trace > 0:
		s := math.Sqrt(trace+1) * 2
		return Quat{(trueUp.Z - back.Y) / s, (back.X - right.Z) / s, (right.Y - trueUp.X) / s, s / 4}.Normalized()
	case m00 > m11 && m00 > m22:
		s := math.Sqrt(1+m00-m11-m22) * 2
		return Quat{s / 4, (trueUp.X + right.Y) / s, (back.X + right.Z) / s, (trueUp.Z - back.Y) / s}.Normalized()
	case m11 > m22:
		s := math.Sqrt(1+m11-m00-m22) * 2
		return Quat{(trueUp.X + right.Y) / s, s / 4, (back.Y + trueUp.Z) / s, (back.X - right.Z) / s}.Normalized()
	default:
		s := math.Sqrt(1+m22-m00-m11) * 2
		return Quat{(back.X + right.Z) / s, (back.Y + trueUp.Z) / s, s / 4, (right.Y - trueUp.X) / s}.Normalized()
	}
}
