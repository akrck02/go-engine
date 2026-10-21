package engine

import (
	"github.com/akrck02/go-engine/space"
	"testing"

	"github.com/akrck02/go-renderer/scene"
)

// triangle is a mesh of one triangle, as a game's code builds one.
func triangle(name string) *scene.Mesh {
	return &scene.Mesh{Name: name, Positions: []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, Normals: []float32{0, 0, 1, 0, 0, 1, 0, 0, 1},
		Material: &scene.Material{Kind: scene.KindUnlit, BaseColor: [4]float32{1, 1, 1, 1}}}
}

// drawnMeshes are the meshes go-renderer draws for a node itself (not its children).
func drawnMeshes(node *Node) []*scene.Mesh {
	var meshes []*scene.Mesh
	if node.geometry != nil {
		meshes = append(meshes, node.geometry.Mesh)
	}
	return meshes
}

func TestAMeshSetFromCodeIsDrawnAndReplacingItFreesTheOldOne(test *testing.T) {
	spinner, _ := load(test).Find("Spinner")
	first, second := triangle("first"), triangle("second")
	spinner.SetMesh(first)
	if meshes := drawnMeshes(spinner); len(meshes) != 1 || meshes[0] != first {
		test.Fatalf("after SetMesh the node draws %v", meshes)
	}
	spinner.SetMesh(second)
	meshes := drawnMeshes(spinner)
	if len(meshes) != 1 || meshes[0] != second {
		test.Fatalf("after replacing the mesh the node draws %v", meshes)
	}
	if found, _ := spinner.Mesh(); found != second {
		test.Fatalf("the node's mesh is %v", found)
	}
	spinner.SetMesh(nil)
	if meshes := drawnMeshes(spinner); len(meshes) != 1 || meshes[0] == first || meshes[0] == second {
		test.Fatal("without a mesh from code the node does not go back to its shape")
	}
}

func TestAMovingNodeIsLitOnTheGPUWhateverItShows(test *testing.T) {
	tower, _ := load(test).Find("Tower")
	tower.SetMoving(true)
	if !allMoving(tower.geometry) {
		test.Fatal("the model of a moving node is not marked moving")
	}
	tower.SetMesh(triangle("hull"))
	if !allMoving(tower.geometry) || !tower.Moving() {
		test.Fatal("a mesh given to a moving node is not marked moving")
	}
	tower.SetMoving(false)
	if tower.geometry.Moving {
		test.Fatal("a node no longer moving stays marked")
	}
}

// allMoving tells whether a render node and all it holds are marked moving.
func allMoving(render *scene.Node) bool {
	if render == nil || !render.Moving {
		return false
	}
	for _, child := range render.Children {
		if !allMoving(child) {
			return false
		}
	}
	return true
}

func TestACamerasFarPlaneIsTheOneSet(test *testing.T) {
	eye, _ := load(test).Find("Eye")
	if view := eye.View(); view.Far != DefaultFar || view.Near != DefaultNear {
		test.Fatalf("a camera without distances sees from %v to %v", view.Near, view.Far)
	}
	settings, _ := eye.Camera()
	settings.Near, settings.Far = 1, 7000
	if view := eye.View(); view.Far != 7000 || view.Near != 1 {
		test.Fatalf("a camera set to see from 1 to 7000 sees from %v to %v", view.Near, view.Far)
	}
}

func TestACameraLooksAtItsFocus(test *testing.T) {
	eye, _ := load(test).Find("Eye")
	distance := func() float64 {
		view := eye.View()
		position, target := view.Position, view.Target
		return space.NewVector(float64(target[0]-position[0]), float64(target[1]-position[1]), float64(target[2]-position[2])).Length()
	}
	if seen := distance(); seen < DefaultFocus-0.01 || seen > DefaultFocus+0.01 {
		test.Fatalf("a camera without a focus looks %v metres ahead", seen)
	}
	settings, _ := eye.Camera()
	settings.Focus = 15
	if seen := distance(); seen < 14.99 || seen > 15.01 {
		test.Fatalf("a camera focused at 15 metres looks %v metres ahead", seen)
	}
}
