# space

```go
import "github.com/akrck02/go-engine/space"
```

Package space is where things are: positions, directions, rotations and sizes
in a game's world. Its Transform places a thing: where it is, how it is turned,
and how large it is.

The world follows glTF and go-renderer. One unit is one metre and Y points up.
The axes are right-handed: with X to the right and Y up, Z points towards
whoever looks. A camera with no rotation looks along -Z (Forward). Angles
are in degrees wherever a person reads or types them. A positive turn is
counter-clockwise, looking down the axis towards the origin: a positive turn
around Up takes Right (+X) towards Forward (-Z).

### Functions

| Function | |
|---|---|
| `func Degrees(radians float64) float64` | Degrees turns radians into degrees. |
| `func Radians(degrees float64) float64` | Radians turns degrees into radians. |

### Constants and variables

```go
const (
	Metre      = 1.0
	Centimetre = 0.01
	Kilometre  = 1000.0
)
```

Units: a length of 1 is a metre.

```go
var (
	Zero    = Vector{}
	One     = Vector{1, 1, 1}
	Right   = Vector{1, 0, 0}
	Up      = Vector{0, 1, 0}
	Back    = Vector{0, 0, 1} // towards whoever looks at the world from the front
	Forward = Vector{0, 0, -1}
)
```

Directions and points that have names.

```go
var NoRotation = Rotation{real: 1}
```

NoRotation leaves things as they are.

### Angles

Angles are a rotation as three angles in degrees, the way an inspector shows
them: X tilts (pitch), Y turns (yaw) and Z rolls. They apply roll first,
then pitch, then yaw, so the yaw is always around the world's vertical,
as a character or a camera turns.

```go
type Angles struct{ X, Y, Z float64 }
```

| Function or method | |
|---|---|
| `func (angles Angles) Rotation() Rotation` | Rotation is the rotation the angles describe. |

### Box

Box is a box lined up with the axes, from its lowest corner to its highest.

```go
type Box struct{ Minimum, Maximum Vector }
```

| Function or method | |
|---|---|
| `func EmptyBox() Box` | EmptyBox holds nothing; growing it by a point makes a box of that point. |
| `func (box Box) Centre() Vector` | Centre is the box's middle. |
| `func (box Box) Contains(point Vector) bool` | Contains tells whether a point is in the box, its faces included. |
| `func (box Box) Corners() [8]Vector` | Corners are the box's eight corners. |
| `func (box Box) Grow(point Vector) Box` | Grow is the box that holds this one and a point. |
| `func (box Box) Hit(origin, direction Vector) (float64, bool)` | Hit is how far along a ray (from an origin, along a direction of length 1) it first enters the box, if it does. |
| `func (box Box) IsEmpty() bool` | IsEmpty tells whether the box holds nothing. |
| `func (box Box) Size() Vector` | Size is the box's extent along each axis. |
| `func (box Box) Transformed(transform Transform) Box` | Transformed is the box, lined up with the axes again, that holds this box once transformed. |
| `func (box Box) Union(other Box) Box` | Union is the box that holds both. |

### Rotation

Rotation is how a thing is turned. Make one with RotationAround,
Angles.Rotation or RotationLookingAt; the zero value is not a rotation,
NoRotation is.
Inside, it is a unit quaternion: a real part and three imaginary ones,
which turn vectors without the gimbal lock of angles and blend smoothly from
one to another.

```go
type Rotation struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func RotationAround(axis Vector, degrees float64) Rotation` | RotationAround is a turn of some degrees around an axis. |
| `func RotationLookingAt(direction, up Vector) Rotation` | RotationLookingAt is the rotation that makes Forward point along a direction, keeping Up as near the given up as it can. |
| `func (rotation Rotation) Angles() Angles` | Angles are the rotation as angles in degrees, each between -180 and 180 (the pitch between -90 and 90). |
| `func (rotation Rotation) Inverse() Rotation` | Inverse is the rotation that undoes this one. |
| `func (rotation Rotation) IsZero() bool` | IsZero tells whether the rotation is the zero value, which is no rotation at all. |
| `func (rotation Rotation) Normalized() Rotation` | Normalized is the rotation kept to a length of 1, as small errors pile up after many turns. |
| `func (rotation Rotation) Parts() (imaginaryX, imaginaryY, imaginaryZ, real float64)` | Parts are the quaternion's numbers, as go-renderer wants them: the imaginary parts, then the real one. |
| `func (rotation Rotation) Rotate(vector Vector) Vector` | Rotate turns a vector by the rotation. |
| `func (rotation Rotation) Then(next Rotation) Rotation` | Then is the rotation that turns by this one first, then by next. |
| `func (rotation Rotation) TurnToward(target Rotation, share float64) Rotation` | TurnToward is the rotation a share of the way from this one to a target, turning at an even pace the shorter way round: 0 is this rotation, 1 the target. |

### Transform

Transform places a thing relative to its parent: it is scaled first,
then turned, then moved.

```go
type Transform struct {
	Position Vector
	Rotation Rotation
	Scale    Vector
}
```

| Function or method | |
|---|---|
| `func At(position Vector) Transform` | At is a transform at a position, unturned, at its own size. |
| `func NewTransform() Transform` | NewTransform is a thing at the origin, unturned, at its own size. |
| `func (transform Transform) Angles() Angles` | Angles are the rotation as angles in degrees. |
| `func (transform Transform) DirectionToParent(local Vector) Vector` | DirectionToParent is where a direction in the thing's own space points in its parent's (turned, not scaled). |
| `func (transform Transform) Forward() Vector` | Forward is the way the thing faces, in its parent's space. |
| `func (transform Transform) FromParent(point Vector) Vector` | FromParent is where a point given in the parent's space is in the thing's own space: ToParent undone. |
| `func (transform *Transform) LookAt(target Vector)` | LookAt turns the thing so its Forward points at a target, keeping its Up towards the sky. |
| `func (transform Transform) Matrix() graphics.Matrix` | Matrix is the transform as go-renderer draws it. |
| `func (transform *Transform) Move(offset Vector)` | Move moves the thing by an offset in its parent's space. |
| `func (transform Transform) Place(child Transform) Transform` | Place is where a child is, placed by its own transform inside the thing this transform places: the child's transform in the thing's parent's space. |
| `func (transform Transform) Right() Vector` | Right is the thing's right, in its parent's space. |
| `func (transform *Transform) SetAngles(angles Angles)` | SetAngles sets the rotation from angles in degrees. |
| `func (transform Transform) ToParent(local Vector) Vector` | ToParent is where a point given in the thing's own space is in its parent's space. |
| `func (transform *Transform) Turn(axis Vector, degrees float64)` | Turn turns the thing around an axis of its parent's space by some degrees. |
| `func (transform Transform) Up() Vector` | Up is the thing's up, in its parent's space. |

### Vector

Vector is a point or a direction in the world, in metres.

```go
type Vector struct{ X, Y, Z float64 }
```

| Function or method | |
|---|---|
| `func NewVector(x, y, z float64) Vector` | NewVector makes a vector from its three coordinates. |
| `func (vector Vector) Add(other Vector) Vector` | Add is the vector moved by another. |
| `func (vector Vector) Cross(other Vector) Vector` | Cross is the cross product: a direction square to both, by the right-hand rule. |
| `func (vector Vector) Distance(other Vector) float64` | Distance is how far apart two points are. |
| `func (vector Vector) Dot(other Vector) float64` | Dot is the dot product: how much two directions point the same way, times their lengths. |
| `func (vector Vector) Length() float64` | Length is how long the vector is. |
| `func (vector Vector) MultiplyEach(other Vector) Vector` | MultiplyEach multiplies each axis by the other vector's: a size scaled per axis. |
| `func (vector Vector) Near(other Vector, tolerance float64) bool` | Near tells whether two vectors are the same within a tolerance on each axis. |
| `func (vector Vector) Normalized() Vector` | Normalized is the vector's direction with a length of 1 (Zero stays Zero). |
| `func (vector Vector) Scale(factor float64) Vector` | Scale is the vector times a number. |
| `func (vector Vector) Subtract(other Vector) Vector` | Subtract is the vector from other to this one. |
| `func (vector Vector) Toward(other Vector, share float64) Vector` | Toward is the point a share of the way from this one to another: 0 is this point, 1 the other. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
