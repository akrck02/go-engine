package shapes

import (
	"math"

	"github.com/akrck02/go-renderer/scene"
)

// The shapes' UVs are in metres of the shape as it stands in the world, so a texture repeated
// once a metre shows square repeats whatever the node's scale: a cube's face from 0 to its width
// and height, a cylinder's side around its circumference and up its height, a sphere by longitude
// and latitude (around its equator, and pole to pole along a meridian), a plane by x and z.

// sphereSlices and sphereRings are the sphere Mesh makes.
const sphereSlices, sphereRings = 24, 16

// cylinderSides is the cylinder Mesh makes.
const cylinderSides = 24

// verticesPerCylinderSide are the vertices one side of a cylinder adds: its quad, and a triangle
// of each cap.
const verticesPerCylinderSide = 10

// StretchUVs writes a shape's UVs for a node of this size (its scale, the shape being a metre
// across), in metres, into the mesh's UVs, reusing their memory: a node whose scale changes calls
// it again without allocating. The mesh must be the one Mesh made for that shape's name.
func StretchUVs(mesh *scene.Mesh, name string, size [3]float64) {
	vertices := mesh.VertexCount()
	if cap(mesh.UVs) < 2*vertices {
		mesh.UVs = make([]float32, 2*vertices)
	}
	mesh.UVs = mesh.UVs[:2*vertices]
	for vertex := range vertices {
		across, down := uvOf(mesh, name, vertex, size)
		mesh.UVs[2*vertex], mesh.UVs[2*vertex+1] = float32(across), float32(down)
	}
}

// uvOf is one vertex's UV, in metres.
func uvOf(mesh *scene.Mesh, name string, vertex int, size [3]float64) (float64, float64) {
	switch name {
	case "sphere":
		return sphereUV(vertex, size)
	case "cylinder":
		return cylinderUV(mesh, vertex, size)
	}
	return faceUV(mesh, vertex, size)
}

// faceUV lays a flat face's texture along the two axes it spans, from its corner: a cube's faces
// and the plane.
func faceUV(mesh *scene.Mesh, vertex int, size [3]float64) (float64, float64) {
	position, normal := mesh.Positions[3*vertex:3*vertex+3], mesh.Normals[3*vertex:3*vertex+3]
	along := func(axis int) float64 { return (float64(position[axis]) + 0.5) * size[axis] }
	switch {
	case normal[0] != 0:
		return along(2), 1 - along(1) // v grows downwards, as a picture's rows do
	case normal[1] != 0:
		return along(0), along(2)
	}
	return along(0), 1 - along(1)
}

// cylinderUV wraps the side around the circumference and up the height; the caps lie flat.
func cylinderUV(mesh *scene.Mesh, vertex int, size [3]float64) (float64, float64) {
	if mesh.Normals[3*vertex+1] != 0 {
		return faceUV(mesh, vertex, size)
	}
	side, corner := vertex/verticesPerCylinderSide, vertex%verticesPerCylinderSide
	step := side
	if corner == 1 || corner == 2 { // the quad's second edge: the next step around
		step++
	}
	circumference := math.Pi * (size[0] + size[2]) / 2
	height := (float64(mesh.Positions[3*vertex+1]) + 0.5) * size[1]
	return circumference * float64(step) / cylinderSides, size[1] - height
}

// sphereUV is a vertex's longitude around the equator and latitude down a meridian, in metres.
func sphereUV(vertex int, size [3]float64) (float64, float64) {
	ring, slice := vertex/(sphereSlices+1), vertex%(sphereSlices+1)
	equator := math.Pi * (size[0] + size[2]) / 2
	meridian := math.Pi * size[1] / 2
	return equator * float64(slice) / sphereSlices, meridian * float64(ring) / sphereRings
}
