// Package shapes makes the meshes of simple shapes for go-renderer to draw. Each is a metre across
// and centred on the origin: the cube, the sphere, the plane (flat, facing up) and the cylinder
// (standing up).
package shapes

import (
	"math"

	"github.com/akrck02/go-renderer/scene"
)

// Mesh makes a shape's mesh by name, with a material of a colour, and its UVs for a shape a metre
// across (see StretchUVs); nil for an unknown name.
func Mesh(name string, material *scene.Material) *scene.Mesh {
	var mesh *scene.Mesh
	switch name {
	case "cube":
		mesh = Cube()
	case "sphere":
		mesh = Sphere(sphereSlices, sphereRings)
	case "plane":
		mesh = Plane()
	case "cylinder":
		mesh = Cylinder(cylinderSides)
	default:
		return nil
	}
	mesh.Name, mesh.Material = name, material
	StretchUVs(mesh, name, [3]float64{1, 1, 1})
	return mesh
}

// builder gathers a mesh's vertices, white so the material gives the colour.
type builder struct{ mesh scene.Mesh }

func (build *builder) vertex(x, y, z, normalX, normalY, normalZ float64) uint32 {
	index := uint32(build.mesh.VertexCount())
	build.mesh.Positions = append(build.mesh.Positions, float32(x), float32(y), float32(z))
	build.mesh.Normals = append(build.mesh.Normals, float32(normalX), float32(normalY), float32(normalZ))
	build.mesh.Colors = append(build.mesh.Colors, 1, 1, 1, 1)
	return index
}

func (build *builder) triangle(first, second, third uint32) {
	build.mesh.Indices = append(build.mesh.Indices, first, second, third)
}

// quad adds a face from four corners, counter-clockwise seen from where it faces.
func (build *builder) quad(corners [4][3]float64, normal [3]float64) {
	var indices [4]uint32
	for index, corner := range corners {
		indices[index] = build.vertex(corner[0], corner[1], corner[2], normal[0], normal[1], normal[2])
	}
	build.triangle(indices[0], indices[1], indices[2])
	build.triangle(indices[0], indices[2], indices[3])
}

// Cube is a cube a metre a side.
func Cube() *scene.Mesh {
	var build builder
	const half = 0.5
	faces := []struct {
		normal  [3]float64
		corners [4][3]float64
	}{
		{[3]float64{1, 0, 0}, [4][3]float64{{half, -half, half}, {half, -half, -half}, {half, half, -half}, {half, half, half}}},
		{[3]float64{-1, 0, 0}, [4][3]float64{{-half, -half, -half}, {-half, -half, half}, {-half, half, half}, {-half, half, -half}}},
		{[3]float64{0, 1, 0}, [4][3]float64{{-half, half, half}, {half, half, half}, {half, half, -half}, {-half, half, -half}}},
		{[3]float64{0, -1, 0}, [4][3]float64{{-half, -half, -half}, {half, -half, -half}, {half, -half, half}, {-half, -half, half}}},
		{[3]float64{0, 0, 1}, [4][3]float64{{-half, -half, half}, {half, -half, half}, {half, half, half}, {-half, half, half}}},
		{[3]float64{0, 0, -1}, [4][3]float64{{half, -half, -half}, {-half, -half, -half}, {-half, half, -half}, {half, half, -half}}},
	}
	for _, face := range faces {
		build.quad(face.corners, face.normal)
	}
	return &build.mesh
}

// Plane is a square a metre a side, lying flat and facing up.
func Plane() *scene.Mesh {
	var build builder
	const half = 0.5
	build.quad([4][3]float64{{-half, 0, half}, {half, 0, half}, {half, 0, -half}, {-half, 0, -half}}, [3]float64{0, 1, 0})
	return &build.mesh
}

// Sphere is a ball a metre across, of some slices around and rings from top to bottom.
func Sphere(slices, rings int) *scene.Mesh {
	var build builder
	for ring := 0; ring <= rings; ring++ {
		polar := math.Pi * float64(ring) / float64(rings)
		for slice := 0; slice <= slices; slice++ {
			around := 2 * math.Pi * float64(slice) / float64(slices)
			x, y, z := math.Sin(polar)*math.Cos(around), math.Cos(polar), -math.Sin(polar)*math.Sin(around)
			build.vertex(x/2, y/2, z/2, x, y, z)
		}
	}
	row := uint32(slices + 1)
	for ring := uint32(0); ring < uint32(rings); ring++ {
		for slice := uint32(0); slice < uint32(slices); slice++ {
			top, bottom := ring*row+slice, (ring+1)*row+slice
			build.triangle(top, bottom, bottom+1)
			build.triangle(top, bottom+1, top+1)
		}
	}
	return &build.mesh
}

// Cylinder is a cylinder a metre across and a metre high, standing up, of some sides.
func Cylinder(sides int) *scene.Mesh {
	var build builder
	const half = 0.5
	for side := 0; side < sides; side++ {
		from := 2 * math.Pi * float64(side) / float64(sides)
		to := 2 * math.Pi * float64(side+1) / float64(sides)
		fromX, fromZ := math.Cos(from)/2, -math.Sin(from)/2
		toX, toZ := math.Cos(to)/2, -math.Sin(to)/2
		middle := (from + to) / 2
		build.quad([4][3]float64{{fromX, -half, fromZ}, {toX, -half, toZ}, {toX, half, toZ}, {fromX, half, fromZ}}, [3]float64{math.Cos(middle), 0, -math.Sin(middle)})
		top := [3]uint32{build.vertex(0, half, 0, 0, 1, 0), build.vertex(fromX, half, fromZ, 0, 1, 0), build.vertex(toX, half, toZ, 0, 1, 0)}
		build.triangle(top[0], top[1], top[2])
		bottom := [3]uint32{build.vertex(0, -half, 0, 0, -1, 0), build.vertex(toX, -half, toZ, 0, -1, 0), build.vertex(fromX, -half, fromZ, 0, -1, 0)}
		build.triangle(bottom[0], bottom[1], bottom[2])
	}
	return &build.mesh
}
