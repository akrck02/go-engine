package engine

import (
	"testing"
	"testing/fstest"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// litScene is a 2D scene of eight pixels a metre: a shape, a lamp two metres right of it, and a
// sign that shines through its sprite.
func litScene(test *testing.T, ambient *scenefile.Ambient) *Scene {
	strong := 1.5
	description := &scenefile.Scene{Name: "Lit", Mode: scenefile.Mode2D, PixelsPerMetre: 8, Ambient: ambient, Nodes: []scenefile.Node{
		{Name: "Crate", Shape: "cube", Scale: [3]float64{1, 1, 1}},
		{Name: "Lamp", Position: [3]float64{2, 0, 0}, Scale: [3]float64{1, 1, 1},
			Light: &scenefile.Light{Kind: scenefile.PointLight, Colour: "#ff8000", Intensity: &strong, Inner: 0.5, Outer: 2}},
		{Name: "Sign", Sprite: "assets/sign.png", Position: [3]float64{-2, 0, 0}, Scale: [3]float64{1, 1, 1},
			Light: &scenefile.Light{Kind: scenefile.SpriteLight}},
	}}
	built, err := Build(fstest.MapFS{"assets/sign.png": {Data: sheetPicture()}}, description)
	if err != nil {
		test.Fatal(err)
	}
	return built
}

// kinds lists the kinds of a list's commands, in order.
func kinds(list *draw2d.List) []draw2d.Kind {
	var found []draw2d.Kind
	for _, command := range list.Commands {
		found = append(found, command.Kind)
	}
	return found
}

func TestAnUnlitSceneDrawsNoLight(test *testing.T) {
	scene := walkerScene(test)
	var list draw2d.List
	scene.DrawContent2D(&list, draw2d.Rectangle{Width: 80, Height: 80}, View2D{Height: 10})
	for _, kind := range kinds(&list) {
		if kind == draw2d.LightingKind || kind == draw2d.PointLightKind {
			test.Fatalf("a scene with no lights drew %v", kinds(&list))
		}
	}
}

func TestLightsAreDrawnAfterTheSceneAndLightItOverItsAmbient(test *testing.T) {
	dim := 0.4
	scene := litScene(test, &scenefile.Ambient{Colour: "#204060", Intensity: &dim})
	var list draw2d.List
	area := draw2d.Rectangle{Width: 80, Height: 80} // ten metres high: eight pixels a metre
	scene.DrawContent2D(&list, area, View2D{Height: 10})
	want := []draw2d.Kind{draw2d.RectangleKind, draw2d.PointLightKind, draw2d.SpriteLightKind, draw2d.LightingKind}
	got := kinds(&list)
	if len(got) != len(want) {
		test.Fatalf("drew %v, wanted %v: the shape, the lamp, the sign's light and the lighting", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			test.Fatalf("drew %v, wanted %v", got, want)
		}
	}
	lamp := list.Commands[1]
	if centreX := lamp.Area.X + lamp.Area.Width/2; centreX != 56 || lamp.Area.Width != 32 || lamp.Radius != 4 {
		test.Errorf("the lamp is at %+v, radius %v; wanted its centre 2 m right of the middle and its radii in metres", lamp.Area, lamp.Radius)
	}
	if lamp.Fill[3] != 1.5 || lamp.Fill[0] != 1 {
		test.Errorf("the lamp's colour is %v; wanted orange at its intensity", lamp.Fill)
	}
	scheme := draw2d.MustHex("#204060")
	if ambient := list.Commands[3].Fill; ambient[3] != 0.4 || ambient[0] != scheme[0] || ambient[2] != scheme[2] {
		test.Errorf("the ambient is %v; wanted the scene's at 0.4", ambient)
	}
}

func TestLightsWithNoAmbientLightTheSceneByDaylight(test *testing.T) {
	scene := litScene(test, nil)
	var list draw2d.List
	scene.DrawContent2D(&list, draw2d.Rectangle{Width: 80, Height: 80}, View2D{Height: 10})
	last := list.Commands[len(list.Commands)-1]
	if last.Kind != draw2d.LightingKind || last.Fill != (draw2d.Color{1, 1, 1, 1}) {
		test.Fatalf("the last command is %+v; wanted a lighting in plain daylight", last)
	}
}

func TestAHiddenLightDoesNotShine(test *testing.T) {
	scene := litScene(test, nil)
	lamp, _ := scene.Find("Lamp")
	lamp.Visible = false
	var list draw2d.List
	scene.DrawContent2D(&list, draw2d.Rectangle{Width: 80, Height: 80}, View2D{Height: 10})
	for _, kind := range kinds(&list) {
		if kind == draw2d.PointLightKind {
			test.Fatal("a hidden lamp still shone")
		}
	}
}

func TestASceneKeepsItsLightsWhenWrittenBack(test *testing.T) {
	dim := 0.4
	described := litScene(test, &scenefile.Ambient{Intensity: &dim}).Describe()
	if described.Ambient == nil || described.Ambient.Strength() != 0.4 {
		test.Errorf("the ambient came back as %+v", described.Ambient)
	}
	if light := described.Nodes[1].Light; light == nil || light.Kind != scenefile.PointLight || light.Strength() != 1.5 {
		test.Errorf("the lamp came back as %+v", light)
	}
}

func TestALitFrameAllocatesNothing(test *testing.T) {
	scene := litScene(test, nil)
	var list draw2d.List
	area := draw2d.Rectangle{Width: 80, Height: 80}
	scene.DrawContent2D(&list, area, View2D{Height: 10})
	allocations := testing.AllocsPerRun(100, func() {
		list.Reset()
		scene.DrawContent2D(&list, area, View2D{Height: 10})
	})
	if allocations != 0 {
		test.Fatalf("a lit frame allocates %v times", allocations)
	}
}

// litScene3D is a 3D scene with a lamp scaled to twice its size and a sprite light, which 3D leaves out.
func litScene3D(test *testing.T, ambient *scenefile.Ambient) *Scene {
	description := &scenefile.Scene{Name: "Lit 3D", Ambient: ambient, Nodes: []scenefile.Node{
		{Name: "Lamp", Position: [3]float64{1, 2, 3}, Scale: [3]float64{2, 2, 2},
			Light: &scenefile.Light{Kind: scenefile.PointLight, Colour: "#808080", Inner: 0.5, Outer: 2}},
		{Name: "Sign", Scale: [3]float64{1, 1, 1}, Light: &scenefile.Light{Kind: scenefile.SpriteLight}},
	}}
	built, err := Build(fstest.MapFS{}, description)
	if err != nil {
		test.Fatal(err)
	}
	return built
}

func TestA3DLampLightsTheWorldWhereItsNodeIs(test *testing.T) {
	scene := litScene3D(test, nil)
	lights := scene.World().Lights
	if len(lights) != 1 {
		test.Fatalf("the world has %d lights; wanted the lamp only (a sprite light is 2D)", len(lights))
	}
	lamp := lights[0]
	if lamp.Position[0] != 1 || lamp.Position[2] != 3 || lamp.Outer != 4 || lamp.Inner != 1 {
		test.Errorf("the lamp is %+v; wanted it at its node, its radii twice for its scale", lamp)
	}
	if grey := lamp.Color[0]; grey < 0.2 || grey > 0.23 || lamp.Color[3] != 1 {
		test.Errorf("the lamp's colour is %v; wanted #808080 in linear light (about 0.22), at 1", lamp.Color)
	}
	node, _ := scene.Find("Lamp")
	node.Transform.Position.X = 5
	scene.Sync()
	if scene.World().Lights[0].Position[0] != 5 {
		test.Error("a lamp that moved still lights from where it was")
	}
}

func TestA3DAmbientScalesTheSkyLight(test *testing.T) {
	dim := 0.5
	daylight := litScene3D(test, nil).World().Environment.Ambient
	dimmed := litScene3D(test, &scenefile.Ambient{Intensity: &dim}).World().Environment.Ambient
	if dimmed != daylight/2 {
		test.Fatalf("an ambient at 0.5 makes the sky light %v; wanted half of %v", dimmed, daylight)
	}
}

func TestSyncingLampsAllocatesNothing(test *testing.T) {
	scene := litScene3D(test, nil)
	allocations := testing.AllocsPerRun(100, scene.Sync)
	if allocations != 0 {
		test.Fatalf("syncing a scene with lamps allocates %v times", allocations)
	}
}

func TestALightOutOfSightIsNotDrawn(test *testing.T) {
	scene := litScene(test, nil)
	lamp, _ := scene.Find("Lamp")
	lamp.Transform.Position.X = 100
	var list draw2d.List
	scene.DrawContent2D(&list, draw2d.Rectangle{Width: 80, Height: 80}, View2D{Height: 10})
	for _, kind := range kinds(&list) {
		if kind == draw2d.PointLightKind {
			test.Fatal("a lamp a hundred metres out of sight was drawn")
		}
	}
}

func TestUnlitASpriteLightShowsItsSprite(test *testing.T) {
	scene := litScene(test, nil)
	var list draw2d.List
	scene.DrawContent2DUnlit(&list, draw2d.Rectangle{Width: 80, Height: 80}, View2D{Height: 10})
	got := kinds(&list)
	want := []draw2d.Kind{draw2d.RectangleKind, draw2d.ImageKind}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		test.Fatalf("unlit, the scene drew %v; wanted the shape and the sign's sprite, and no light", got)
	}
}

func TestA2DLampsReachGrowsWithItsScaleAcrossAndUp(test *testing.T) {
	scene := litScene(test, nil)
	lamp, _ := scene.Find("Lamp")
	lamp.Transform.Scale = space.NewVector(2, 2, 7) // depth means nothing in 2D
	if inner, outer := lamp.LightRadii(); inner != 1 || outer != 4 {
		test.Fatalf("the lamp reaches %v to %v; wanted twice its 0.5 to 2", inner, outer)
	}
}
