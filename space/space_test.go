package space

import (
	"math"
	"testing"

	"github.com/akrck02/go-renderer/graphics"
)

const tolerance = 1e-9

func TestAPositiveTurnAroundUpTakesRightTowardsForward(t *testing.T) {
	turned := AxisAngle(Up, 90).Apply(Right)
	if !turned.Near(Forward, tolerance) {
		t.Fatalf("Right turned 90° around Up is %v", turned)
	}
}

func TestEulerAnglesComeBackAsGiven(t *testing.T) {
	for _, angles := range []Euler{{0, 0, 0}, {30, 0, 0}, {0, 45, 0}, {0, 0, -60}, {20, 130, -15}, {-80, -170, 40}} {
		back := angles.Rotation().Euler()
		if math.Abs(back.X-angles.X) > 1e-6 || math.Abs(back.Y-angles.Y) > 1e-6 || math.Abs(back.Z-angles.Z) > 1e-6 {
			t.Errorf("%v came back as %v", angles, back)
		}
	}
	// straight down: all of the turn is yaw
	if back := (Euler{-90, 30, 0}).Rotation().Euler(); math.Abs(back.X+90) > 1e-4 || math.Abs(back.Y-30) > 1e-4 {
		t.Errorf("looking down came back as %v", back)
	}
}

func TestYawIsAroundTheWorldsVertical(t *testing.T) {
	// a camera tilted down and turned: its right stays level
	right := (Euler{-30, 60, 0}).Rotation().Apply(Right)
	if math.Abs(right.Y) > tolerance {
		t.Fatalf("the right of a yawed and pitched camera tilts: %v", right)
	}
}

func TestATransformMatchesGoRenderersMatrix(t *testing.T) {
	transform := Transform{Position: V(1, 2, 3), Rotation: (Euler{10, 20, 30}).Rotation(), Scale: V(2, 3, 4)}
	local := V(0.5, -1, 2)
	want := transform.Point(local)
	got := transform.Matrix().Apply(graphics.Vec4{float32(local.X), float32(local.Y), float32(local.Z), 1})
	if !want.Near(V(float64(got[0]), float64(got[1]), float64(got[2])), 1e-4) {
		t.Fatalf("the matrix puts the point at %v, the transform at %v", got, want)
	}
}

func TestAChildIsPlacedInsideItsParent(t *testing.T) {
	parent := Transform{Position: V(10, 0, 0), Rotation: AxisAngle(Up, 90), Scale: V(2, 2, 2)}
	child := At(V(1, 0, 0))
	world := parent.Then(child)
	if !world.Position.Near(V(10, 0, -2), tolerance) || !world.Scale.Near(V(2, 2, 2), tolerance) {
		t.Fatalf("the child is at %v, scaled %v", world.Position, world.Scale)
	}
}

func TestLookAtPointsForwardAtTheTarget(t *testing.T) {
	for _, target := range []Vec3{V(5, 0, 0), V(0, 3, -4), V(-2, -2, 7), V(0, -10, 0)} {
		transform := At(V(0, 0, 0))
		transform.LookAt(target)
		if !transform.Forward().Near(target.Normalized(), 1e-9) {
			t.Errorf("looking at %v, forward is %v", target, transform.Forward())
		}
	}
}

func TestSlerpGoesHalfWay(t *testing.T) {
	half := Identity.Slerp(AxisAngle(Up, 90), 0.5)
	if angles := half.Euler(); math.Abs(angles.Y-45) > 1e-9 {
		t.Fatalf("half way is %v", angles)
	}
}

func TestARayHitsABox(t *testing.T) {
	box := Box{Minimum: V(-1, -1, -1), Maximum: V(1, 1, 1)}
	if distance, hit := box.Hit(V(0, 0, 10), Forward); !hit || math.Abs(distance-9) > tolerance {
		t.Fatalf("hit %v at %v", hit, distance)
	}
	if _, hit := box.Hit(V(0, 5, 10), Forward); hit {
		t.Fatal("a ray above the box hit it")
	}
	if _, hit := box.Hit(V(0, 0, 10), Back); hit {
		t.Fatal("a ray going away hit the box behind it")
	}
	turned := box.Transformed(Transform{Position: V(5, 0, 0), Rotation: AxisAngle(Up, 45), Scale: One})
	if math.Abs(turned.Size().X-2*math.Sqrt2) > 1e-9 {
		t.Fatalf("a turned box is %v wide", turned.Size().X)
	}
}

func TestLocalUndoesPoint(t *testing.T) {
	transform := Transform{Position: V(3, -1, 2), Rotation: (Euler{15, -40, 5}).Rotation(), Scale: V(2, 0.5, 3)}
	point := V(0.3, 7, -2)
	if back := transform.Local(transform.Point(point)); !back.Near(point, 1e-9) {
		t.Fatalf("%v came back as %v", point, back)
	}
}
