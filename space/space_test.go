package space

import (
	"math"
	"testing"

	"github.com/akrck02/go-renderer/graphics"
)

const tolerance = 1e-9

func TestAPositiveTurnAroundUpTakesRightTowardsForward(test *testing.T) {
	turned := RotationAround(Up, 90).Rotate(Right)
	if !turned.Near(Forward, tolerance) {
		test.Fatalf("Right turned 90° around Up is %v", turned)
	}
}

func TestEulerAnglesComeBackAsGiven(test *testing.T) {
	for _, angles := range []Angles{{0, 0, 0}, {30, 0, 0}, {0, 45, 0}, {0, 0, -60}, {20, 130, -15}, {-80, -170, 40}} {
		back := angles.Rotation().Angles()
		if math.Abs(back.X-angles.X) > 1e-6 || math.Abs(back.Y-angles.Y) > 1e-6 || math.Abs(back.Z-angles.Z) > 1e-6 {
			test.Errorf("%v came back as %v", angles, back)
		}
	}
	// straight down: all of the turn is yaw
	if back := (Angles{-90, 30, 0}).Rotation().Angles(); math.Abs(back.X+90) > 1e-4 || math.Abs(back.Y-30) > 1e-4 {
		test.Errorf("looking down came back as %v", back)
	}
}

func TestYawIsAroundTheWorldsVertical(test *testing.T) {
	// a camera tilted down and turned: its right stays level
	right := (Angles{-30, 60, 0}).Rotation().Rotate(Right)
	if math.Abs(right.Y) > tolerance {
		test.Fatalf("the right of a yawed and pitched camera tilts: %v", right)
	}
}

func TestATransformMatchesGoRenderersMatrix(test *testing.T) {
	transform := Transform{Position: NewVector(1, 2, 3), Rotation: (Angles{10, 20, 30}).Rotation(), Scale: NewVector(2, 3, 4)}
	local := NewVector(0.5, -1, 2)
	want := transform.ToParent(local)
	got := transform.Matrix().Apply(graphics.Vec4{float32(local.X), float32(local.Y), float32(local.Z), 1})
	if !want.Near(NewVector(float64(got[0]), float64(got[1]), float64(got[2])), 1e-4) {
		test.Fatalf("the matrix puts the point at %v, the transform at %v", got, want)
	}
}

func TestAChildIsPlacedInsideItsParent(test *testing.T) {
	parent := Transform{Position: NewVector(10, 0, 0), Rotation: RotationAround(Up, 90), Scale: NewVector(2, 2, 2)}
	child := At(NewVector(1, 0, 0))
	world := parent.Place(child)
	if !world.Position.Near(NewVector(10, 0, -2), tolerance) || !world.Scale.Near(NewVector(2, 2, 2), tolerance) {
		test.Fatalf("the child is at %v, scaled %v", world.Position, world.Scale)
	}
}

func TestLookAtPointsForwardAtTheTarget(test *testing.T) {
	for _, target := range []Vector{NewVector(5, 0, 0), NewVector(0, 3, -4), NewVector(-2, -2, 7), NewVector(0, -10, 0)} {
		transform := At(NewVector(0, 0, 0))
		transform.LookAt(target)
		if !transform.Forward().Near(target.Normalized(), 1e-9) {
			test.Errorf("looking at %v, forward is %v", target, transform.Forward())
		}
	}
}

func TestSlerpGoesHalfWay(test *testing.T) {
	half := NoRotation.TurnToward(RotationAround(Up, 90), 0.5)
	if angles := half.Angles(); math.Abs(angles.Y-45) > 1e-9 {
		test.Fatalf("half way is %v", angles)
	}
}

func TestARayHitsABox(test *testing.T) {
	box := Box{Minimum: NewVector(-1, -1, -1), Maximum: NewVector(1, 1, 1)}
	if distance, hit := box.Hit(NewVector(0, 0, 10), Forward); !hit || math.Abs(distance-9) > tolerance {
		test.Fatalf("hit %v at %v", hit, distance)
	}
	if _, hit := box.Hit(NewVector(0, 5, 10), Forward); hit {
		test.Fatal("a ray above the box hit it")
	}
	if _, hit := box.Hit(NewVector(0, 0, 10), Back); hit {
		test.Fatal("a ray going away hit the box behind it")
	}
	turned := box.Transformed(Transform{Position: NewVector(5, 0, 0), Rotation: RotationAround(Up, 45), Scale: One})
	if math.Abs(turned.Size().X-2*math.Sqrt2) > 1e-9 {
		test.Fatalf("a turned box is %v wide", turned.Size().X)
	}
}

func TestLocalUndoesPoint(test *testing.T) {
	transform := Transform{Position: NewVector(3, -1, 2), Rotation: (Angles{15, -40, 5}).Rotation(), Scale: NewVector(2, 0.5, 3)}
	point := NewVector(0.3, 7, -2)
	if back := transform.FromParent(transform.ToParent(point)); !back.Near(point, 1e-9) {
		test.Fatalf("%v came back as %v", point, back)
	}
}
