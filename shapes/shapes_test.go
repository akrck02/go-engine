package shapes

import (
	"math"
	"testing"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"
)

func TestEveryShapeIsAMetreAcrossAndFacesOutwards(test *testing.T) {
	for _, name := range []string{"cube", "sphere", "plane", "cylinder"} {
		mesh := Mesh(name, &scene.Material{})
		minimum, maximum := mesh.Bounds()
		if math.Abs(float64(maximum[0]-minimum[0])-1) > 1e-6 || math.Abs(float64(maximum[2]-minimum[2])-1) > 0.01 {
			test.Errorf("%s spans %v to %v", name, minimum, maximum)
		}
		// each triangle, wound counter-clockwise, faces the way its first vertex's normal says
		mesh.ForEachTriangle(func(first, second, third uint32) {
			corner, next, last := mesh.Vertex(first), mesh.Vertex(second), mesh.Vertex(third)
			normal := cross(subtract(next, corner), subtract(last, corner))
			given := graphics.Vector{mesh.Normals[3*first], mesh.Normals[3*first+1], mesh.Normals[3*first+2]}
			if length := normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2]; length > 1e-12 && dot(normal, given) <= 0 {
				test.Errorf("%s has a triangle facing inwards", name)
			}
		})
	}
	if Mesh("teapot", nil) != nil {
		test.Error("an unknown shape made a mesh")
	}
}

func subtract(from, other graphics.Vector) graphics.Vector {
	return graphics.Vector{from[0] - other[0], from[1] - other[1], from[2] - other[2]}
}

func cross(first, second graphics.Vector) graphics.Vector {
	return graphics.Vector{first[1]*second[2] - first[2]*second[1], first[2]*second[0] - first[0]*second[2], first[0]*second[1] - first[1]*second[0]}
}

func dot(first, second graphics.Vector) float32 {
	return first[0]*second[0] + first[1]*second[1] + first[2]*second[2]
}
