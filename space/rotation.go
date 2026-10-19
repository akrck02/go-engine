package space

import "math"

// Rotation is how a thing is turned. Make one with RotationAround, Angles.Rotation or
// RotationLookingAt; the zero value is not a rotation, NoRotation is.
//
// Inside, it is a unit quaternion: a real part and three imaginary ones, which turn vectors without
// the gimbal lock of angles and blend smoothly from one to another.
type Rotation struct {
	imaginaryX, imaginaryY, imaginaryZ, real float64
}

// NoRotation leaves things as they are.
var NoRotation = Rotation{real: 1}

// Angles are a rotation as three angles in degrees, the way an inspector shows them: X tilts
// (pitch), Y turns (yaw) and Z rolls. They apply roll first, then pitch, then yaw, so the yaw is
// always around the world's vertical, as a character or a camera turns.
type Angles struct{ X, Y, Z float64 }

// RotationAround is a turn of some degrees around an axis.
func RotationAround(axis Vector, degrees float64) Rotation {
	axis = axis.Normalized()
	half := Radians(degrees) / 2
	sine := math.Sin(half)
	return Rotation{imaginaryX: axis.X * sine, imaginaryY: axis.Y * sine, imaginaryZ: axis.Z * sine, real: math.Cos(half)}
}

// Rotation is the rotation the angles describe.
func (angles Angles) Rotation() Rotation {
	roll := RotationAround(Back, angles.Z)
	pitch := RotationAround(Right, angles.X)
	yaw := RotationAround(Up, angles.Y)
	return roll.Then(pitch).Then(yaw)
}

// Angles are the rotation as angles in degrees, each between -180 and 180 (the pitch between -90
// and 90). They are read from where the rotation takes the three axes.
func (rotation Rotation) Angles() Angles {
	turnedRight, turnedUp, turnedBack := rotation.Rotate(Right), rotation.Rotate(Up), rotation.Rotate(Back)
	pitch := math.Asin(math.Max(-1, math.Min(1, -turnedBack.Y)))
	if math.Abs(turnedBack.Y) > 0.9999 {
		return straightUpOrDown(pitch, turnedRight)
	}
	yaw := math.Atan2(turnedBack.X, turnedBack.Z)
	roll := math.Atan2(turnedRight.Y, turnedUp.Y)
	return Angles{Degrees(pitch), Degrees(yaw), Degrees(roll)}
}

// straightUpOrDown are the angles of a rotation that looks straight up or down: there the yaw and
// the roll turn about the same axis, so all of it is taken as yaw.
func straightUpOrDown(pitch float64, turnedRight Vector) Angles {
	yaw := math.Atan2(-turnedRight.Z, turnedRight.X)
	return Angles{Degrees(pitch), Degrees(yaw), 0}
}

// Then is the rotation that turns by this one first, then by next.
func (rotation Rotation) Then(next Rotation) Rotation {
	first, second := rotation, next
	return Rotation{
		imaginaryX: second.real*first.imaginaryX + second.imaginaryX*first.real + second.imaginaryY*first.imaginaryZ - second.imaginaryZ*first.imaginaryY,
		imaginaryY: second.real*first.imaginaryY - second.imaginaryX*first.imaginaryZ + second.imaginaryY*first.real + second.imaginaryZ*first.imaginaryX,
		imaginaryZ: second.real*first.imaginaryZ + second.imaginaryX*first.imaginaryY - second.imaginaryY*first.imaginaryX + second.imaginaryZ*first.real,
		real:       second.real*first.real - second.imaginaryX*first.imaginaryX - second.imaginaryY*first.imaginaryY - second.imaginaryZ*first.imaginaryZ,
	}
}

// Normalized is the rotation kept to a length of 1, as small errors pile up after many turns.
func (rotation Rotation) Normalized() Rotation {
	length := math.Sqrt(rotation.imaginaryX*rotation.imaginaryX + rotation.imaginaryY*rotation.imaginaryY +
		rotation.imaginaryZ*rotation.imaginaryZ + rotation.real*rotation.real)
	if length == 0 {
		return NoRotation
	}
	return Rotation{rotation.imaginaryX / length, rotation.imaginaryY / length, rotation.imaginaryZ / length, rotation.real / length}
}

// Inverse is the rotation that undoes this one.
func (rotation Rotation) Inverse() Rotation {
	normal := rotation.Normalized()
	return Rotation{-normal.imaginaryX, -normal.imaginaryY, -normal.imaginaryZ, normal.real}
}

// Rotate turns a vector by the rotation.
func (rotation Rotation) Rotate(vector Vector) Vector {
	normal := rotation.Normalized()
	axis := Vector{normal.imaginaryX, normal.imaginaryY, normal.imaginaryZ}
	twice := axis.Cross(vector).Scale(2)
	return vector.Add(twice.Scale(normal.real)).Add(axis.Cross(twice))
}

// Parts are the quaternion's numbers, as go-renderer wants them: the imaginary parts, then the
// real one.
func (rotation Rotation) Parts() (imaginaryX, imaginaryY, imaginaryZ, real float64) {
	return rotation.imaginaryX, rotation.imaginaryY, rotation.imaginaryZ, rotation.real
}

// IsZero tells whether the rotation is the zero value, which is no rotation at all.
func (rotation Rotation) IsZero() bool { return rotation == Rotation{} }

// TurnToward is the rotation a share of the way from this one to a target, turning at an even pace
// the shorter way round: 0 is this rotation, 1 the target.
func (rotation Rotation) TurnToward(target Rotation, share float64) Rotation {
	from, to := rotation.Normalized(), target.Normalized()
	cosine := from.dot(to)
	if cosine < 0 { // the same rotation has two quaternions; turn toward the nearer
		to, cosine = to.negated(), -cosine
	}
	if cosine > 0.9995 {
		return from.blendStraight(to, share)
	}
	angle := math.Acos(cosine)
	sine := math.Sin(angle)
	fromShare, toShare := math.Sin((1-share)*angle)/sine, math.Sin(share*angle)/sine
	return from.scaled(fromShare).plus(to.scaled(toShare))
}

func (rotation Rotation) dot(other Rotation) float64 {
	return rotation.imaginaryX*other.imaginaryX + rotation.imaginaryY*other.imaginaryY +
		rotation.imaginaryZ*other.imaginaryZ + rotation.real*other.real
}

func (rotation Rotation) negated() Rotation { return rotation.scaled(-1) }

func (rotation Rotation) scaled(factor float64) Rotation {
	return Rotation{rotation.imaginaryX * factor, rotation.imaginaryY * factor, rotation.imaginaryZ * factor, rotation.real * factor}
}

func (rotation Rotation) plus(other Rotation) Rotation {
	return Rotation{rotation.imaginaryX + other.imaginaryX, rotation.imaginaryY + other.imaginaryY,
		rotation.imaginaryZ + other.imaginaryZ, rotation.real + other.real}
}

// blendStraight blends two rotations almost the same: straight between their numbers, which is
// as good and avoids dividing by the sine of a tiny angle.
func (rotation Rotation) blendStraight(target Rotation, share float64) Rotation {
	return rotation.scaled(1 - share).plus(target.scaled(share)).Normalized()
}

// RotationLookingAt is the rotation that makes Forward point along a direction, keeping Up as near
// the given up as it can.
func RotationLookingAt(direction, up Vector) Rotation {
	forward := direction.Normalized()
	if forward == Zero {
		return NoRotation
	}
	right := forward.Cross(up).Normalized()
	if right == Zero { // looking straight along up: any right will do
		right = sideways(forward)
	}
	trueUp := right.Cross(forward)
	return rotationOfAxes(right, trueUp, forward.Scale(-1))
}

// sideways is a direction square to another.
func sideways(direction Vector) Vector {
	right := direction.Cross(Back).Normalized()
	if right == Zero {
		return Right
	}
	return right
}

// rotationOfAxes is the rotation that takes Right, Up and Back to three given directions, square
// to each other.
func rotationOfAxes(right, up, back Vector) Rotation {
	trace := right.X + up.Y + back.Z
	switch {
	case trace > 0:
		scale := math.Sqrt(trace+1) * 2
		return Rotation{(up.Z - back.Y) / scale, (back.X - right.Z) / scale, (right.Y - up.X) / scale, scale / 4}.Normalized()
	case right.X > up.Y && right.X > back.Z:
		scale := math.Sqrt(1+right.X-up.Y-back.Z) * 2
		return Rotation{scale / 4, (up.X + right.Y) / scale, (back.X + right.Z) / scale, (up.Z - back.Y) / scale}.Normalized()
	case up.Y > back.Z:
		scale := math.Sqrt(1+up.Y-right.X-back.Z) * 2
		return Rotation{(up.X + right.Y) / scale, scale / 4, (back.Y + up.Z) / scale, (back.X - right.Z) / scale}.Normalized()
	default:
		scale := math.Sqrt(1+back.Z-right.X-up.Y) * 2
		return Rotation{(back.X + right.Z) / scale, (back.Y + up.Z) / scale, scale / 4, (right.Y - up.X) / scale}.Normalized()
	}
}
