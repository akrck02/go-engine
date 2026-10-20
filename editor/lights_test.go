package editor

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
)

func TestALightSetInTheInspectorIsSaved(test *testing.T) {
	editor := openFlat(test)
	editor.GiveLight(scenefile.PointLight)
	light, _ := editor.Selected.Light()
	light.Outer = 5
	editor.GiveLight(scenefile.SpriteLight)
	if editor.Saved {
		test.Fatal("making a light did not change the scene")
	}
	square := saveAndOpenAgain(test, editor, editor.Selected.Path())
	saved, found := square.Light()
	if !found || saved.Kind != scenefile.SpriteLight || saved.Outer != 5 || saved.Colour != newLightColour {
		test.Fatalf("the light came back as %+v", saved)
	}
	editor.Select(square)
	editor.TakeLight()
	if _, found := square.Light(); found {
		test.Fatal("the light was not taken")
	}
}

func TestANewLightIsAWarmLampInBothDimensions(test *testing.T) {
	flat := openFlat(test)
	lamp2D, _ := flat.Add("Light").Light()
	deep := open(test)
	lampNode := deep.Add("Light")
	lamp3D, _ := lampNode.Light()
	if lamp2D == nil || lamp3D == nil || lamp2D.Kind != scenefile.PointLight || lamp3D.Colour != newLightColour {
		test.Fatalf("new lights are %+v and %+v", lamp2D, lamp3D)
	}
	if lampNode.Transform.Position.Y != newLampHeight || len(deep.Scene.World().Lights) != 1 {
		test.Errorf("a new 3D lamp hangs at %v and the world has %d lights", lampNode.Transform.Position.Y, len(deep.Scene.World().Lights))
	}
}

func TestTheSceneAmbientIsSaved(test *testing.T) {
	editor := openFlat(test)
	intensity := 0.3
	editor.SetAmbient(&scenefile.Ambient{Colour: "#102040", Intensity: &intensity})
	if editor.Save() != nil || editor.OpenScene(editor.SceneFile) != nil {
		test.Fatal("the scene was not saved and opened again")
	}
	if ambient := editor.Scene.Ambient; ambient == nil || ambient.Strength() != 0.3 || ambient.Colour != "#102040" {
		test.Fatalf("the ambient came back as %+v", ambient)
	}
}

func TestTheViewCanBeShownUnlit(test *testing.T) {
	editor := openFlat(test)
	editor.Add("Light")
	view := newFlatView()
	lit, unlit := draw2d.List{}, draw2d.List{}
	area := draw2d.Rect{Width: 400, Height: 300}
	view.paint(editor, area, &lit)
	view.showLit = false
	view.paint(editor, area, &unlit)
	if !hasLighting(&lit) || hasLighting(&unlit) {
		test.Fatalf("lit view lights: %v, unlit view lights: %v", hasLighting(&lit), hasLighting(&unlit))
	}
}

func hasLighting(list *draw2d.List) bool {
	for _, command := range list.Commands {
		if command.Kind == draw2d.LightingKind {
			return true
		}
	}
	return false
}
