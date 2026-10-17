package shapes

import (
	"math"
	"testing"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"
)

func TestEveryShapeIsAMetreAcrossAndFacesOutwards(t *testing.T) {
	for _, name := range []string{"cube", "sphere", "plane", "cylinder"} {
		mesh := Mesh(name, &scene.Material{})
		minimum, maximum := mesh.Bounds()
		if math.Abs(float64(maximum[0]-minimum[0])-1) > 1e-6 || math.Abs(float64(maximum[2]-minimum[2])-1) > 0.01 {
			t.Errorf("%s spans %v to %v", name, minimum, maximum)
		}
		// each triangle, wound counter-clockwise, faces the way its first vertex's normal says
		mesh.ForEachTriangle(func(first, second, third uint32) {
			a, b, c := mesh.Vertex(first), mesh.Vertex(second), mesh.Vertex(third)
			normal := cross(sub(b, a), sub(c, a))
			given := graphics.Vec4{mesh.Normals[3*first], mesh.Normals[3*first+1], mesh.Normals[3*first+2]}
			if length := normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2]; length > 1e-12 && dot(normal, given) <= 0 {
				t.Errorf("%s has a triangle facing inwards", name)
			}
		})
	}
	if Mesh("teapot", nil) != nil {
		t.Error("an unknown shape made a mesh")
	}
}

func sub(a, b graphics.Vec4) graphics.Vec4 {
	return graphics.Vec4{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}
func cross(a, b graphics.Vec4) graphics.Vec4 {
	return graphics.Vec4{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func dot(a, b graphics.Vec4) float32 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
