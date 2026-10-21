package main

import (
	"math"

	"github.com/akrck02/go-renderer/scene"
)

// Meshes the behaviours build in code, as a game builds what no model file holds. Each vertex has
// a colour, which the material's colour multiplies: the beam fades along its length that way.

// sidesOfARound is how many sides make a circle round enough at this size.
const sidesOfARound = 24

// meshBuilder collects a mesh's triangles, one vertex after another.
type meshBuilder struct{ mesh scene.Mesh }

// add puts a vertex with its normal and colour.
func (builder *meshBuilder) add(position, normal [3]float32, colour [4]float32) {
	builder.mesh.Positions = append(builder.mesh.Positions, position[:]...)
	builder.mesh.Normals = append(builder.mesh.Normals, normal[:]...)
	builder.mesh.Colors = append(builder.mesh.Colors, colour[:]...)
}

// rimPoint is a point of a circle of a radius round the X axis, at a distance along it.
func rimPoint(distance, radius float32, index int) [3]float32 {
	angle := 2 * math.Pi * float64(index) / sidesOfARound
	return [3]float32{distance, radius * float32(math.Cos(angle)), radius * float32(math.Sin(angle))}
}

// beamCone is a cone along +X from its tip, bright at the tip and gone at its open end.
func beamCone(length, radius float32, material *scene.Material) *scene.Mesh {
	var builder meshBuilder
	bright, gone := [4]float32{1, 1, 1, 1}, [4]float32{1, 1, 1, 0}
	for index := range sidesOfARound {
		first, second := rimPoint(length, radius, index), rimPoint(length, radius, index+1)
		builder.add([3]float32{}, [3]float32{-1, 0, 0}, bright)
		builder.add(first, [3]float32{0, first[1] / radius, first[2] / radius}, gone)
		builder.add(second, [3]float32{0, second[1] / radius, second[2] / radius}, gone)
	}
	builder.mesh.Name, builder.mesh.Material = "beam", material
	return &builder.mesh
}

// flatDisc is a round, flat, facing up, with its centre at the origin.
func flatDisc(radius float32, material *scene.Material) *scene.Mesh {
	var builder meshBuilder
	white, up := [4]float32{1, 1, 1, 1}, [3]float32{0, 1, 0}
	for index := range sidesOfARound {
		first, second := rimPoint(0, radius, index), rimPoint(0, radius, index+1)
		builder.add([3]float32{}, up, white)
		builder.add([3]float32{first[1], 0, first[2]}, up, white)
		builder.add([3]float32{second[1], 0, second[2]}, up, white)
	}
	builder.mesh.Name, builder.mesh.Material = "disc", material
	return &builder.mesh
}

// diamond is two square pyramids base to base, a metre tall.
func diamond(material *scene.Material) *scene.Mesh {
	var builder meshBuilder
	white := [4]float32{1, 1, 1, 1}
	corners := [4][3]float32{{0.35, 0, 0}, {0, 0, 0.35}, {-0.35, 0, 0}, {0, 0, -0.35}}
	for _, tip := range [][3]float32{{0, 0.5, 0}, {0, -0.5, 0}} {
		for index := range corners {
			next := corners[(index+1)%len(corners)]
			builder.add(tip, tip, white)
			builder.add(corners[index], corners[index], white)
			builder.add(next, next, white)
		}
	}
	builder.mesh.Name, builder.mesh.Material = "diamond", material
	return &builder.mesh
}
