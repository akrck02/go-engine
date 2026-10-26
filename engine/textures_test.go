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
