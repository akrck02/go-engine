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

// uvRange is how far a mesh's UVs reach along u and v, over the vertices that pass a test.
func uvRange(mesh *scene.Mesh, keep func(vertex int) bool) (float32, float32) {
	var lowU, highU, lowV, highV float32 = 1e9, -1e9, 1e9, -1e9
	for vertex := range mesh.VertexCount() {
		if !keep(vertex) {
			continue
		}
		across, down := mesh.UVs[2*vertex], mesh.UVs[2*vertex+1]
		lowU, highU, lowV, highV = min(lowU, across), max(highU, across), min(lowV, down), max(highV, down)
	}
	return highU - lowU, highV - lowV
}

func near(value, wanted float32) bool { return math.Abs(float64(value-wanted)) < 1e-3 }

func TestACylindersSideWrapsAroundItAndUpItsHeightInMetres(test *testing.T) {
	tower := Mesh("cylinder", &scene.Material{})
	StretchUVs(tower, "cylinder", [3]float64{2, 12, 2})
	around, up := uvRange(tower, func(vertex int) bool { return tower.Normals[3*vertex+1] == 0 })
	if !near(up, 12) || !near(around, float32(2*math.Pi)) {
		test.Errorf("a 2 × 12 m tower's side spans %v metres around and %v up, wanted %v and 12", around, up, 2*math.Pi)
	}
}

func TestEachFaceOfACubeShowsItsOwnSizeInMetres(test *testing.T) {
	box := Mesh("cube", &scene.Material{})
	StretchUVs(box, "cube", [3]float64{3, 2, 1})
	wanted := map[[3]float32][2]float32{{1, 0, 0}: {1, 2}, {0, 1, 0}: {3, 1}, {0, 0, 1}: {3, 2}}
	for normal, size := range wanted {
		across, down := uvRange(box, func(vertex int) bool {
			return [3]float32(box.Normals[3*vertex:3*vertex+3]) == normal
		})
		if !near(across, size[0]) || !near(down, size[1]) {
			test.Errorf("the face towards %v spans %v × %v, wanted %v", normal, across, down, size)
		}
	}
}

func TestStretchingUVsAgainAllocatesNothing(test *testing.T) {
	tower := Mesh("cylinder", &scene.Material{})
	allocations := testing.AllocsPerRun(10, func() { StretchUVs(tower, "cylinder", [3]float64{2, 12, 2}) })
	if allocations != 0 {
		test.Errorf("stretching UVs again allocates %v times", allocations)
	}
}
