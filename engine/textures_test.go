package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/scene"

	"github.com/akrck02/go-engine/scenefile"
)

// pavedSquare is a scene of one cube that wears the built-in cobbles twice a metre.
func pavedSquare() *scenefile.Scene {
	square := scenefile.NewNode("Square")
	square.Shape, square.Colour, square.Texture, square.TextureScale = "cube", "#a0a0a0", "cobbles", 2
	return &scenefile.Scene{Name: "Main", Nodes: []scenefile.Node{square}}
}

func TestAShapesTextureReachesTheWorldsMaterial(test *testing.T) {
	built, err := Build(nil, pavedSquare())
	if err != nil {
		test.Fatal(err)
	}
	material := built.Root.children[0].geometry.Mesh.Material
	cobbles, _ := scene.BuiltInTexture("cobbles")
	if material.Texture != cobbles || material.TextureScale != [2]float32{2, 2} {
		test.Errorf("the square's material has texture %p (cobbles %p) and scale %v", material.Texture, cobbles, material.TextureScale)
	}
}

func TestASavedSceneKeepsItsShapesTexture(test *testing.T) {
	built, err := Build(nil, pavedSquare())
	if err != nil {
		test.Fatal(err)
	}
	built.Root.children[0].SetTexture("panels", 0.5)
	data, err := built.Describe().Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := scenefile.Read(data)
	if err != nil {
		test.Fatal(err)
	}
	if node := read.Nodes[0]; node.Texture != "panels" || node.TextureScale != 0.5 {
		test.Errorf("the saved square wears %q %v times a metre, wanted panels 0.5", node.Texture, node.TextureScale)
	}
}

func TestAScaledNodesTextureIsNotStretched(test *testing.T) {
	description := pavedSquare()
	description.Nodes[0].Shape, description.Nodes[0].Scale = "cylinder", [3]float64{2, 12, 2}
	built, err := Build(nil, description)
	if err != nil {
		test.Fatal(err)
	}
	tower := built.Root.children[0]
	built.Sync()
	if tallest := highestV(tower.geometry.Mesh); tallest < 11.99 || tallest > 12.01 {
		test.Errorf("a 12 m tower's UVs reach %v metres up, wanted 12", tallest)
	}
	tower.Transform.Scale.Y = 6
	built.Sync()
	if tallest := highestV(tower.geometry.Mesh); tallest < 5.99 || tallest > 6.01 || !tower.geometry.Mesh.Dirty {
		test.Errorf("after shrinking to 6 m the UVs reach %v metres (sent again: %v)", tallest, tower.geometry.Mesh.Dirty)
	}
	allocations := testing.AllocsPerRun(10, built.Sync)
	if allocations != 0 {
		test.Errorf("a sync with nothing changed allocates %v times", allocations)
	}
}

func highestV(mesh *scene.Mesh) float32 {
	var highest float32
	for index := 1; index < len(mesh.UVs); index += 2 {
		highest = max(highest, mesh.UVs[index])
	}
	return highest
}
