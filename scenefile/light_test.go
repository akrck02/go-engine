package scenefile

import "testing"

func TestALitSceneComesBackAsWritten(test *testing.T) {
	scene := flatScene()
	dim, strong := 0.4, 1.5
	scene.Ambient = &Ambient{Colour: "#203050", Intensity: &dim}
	scene.Nodes[2].Light = &Light{Kind: PointLight, Colour: "#ffd27a", Intensity: &strong, Inner: 0.5, Outer: 3.5}
	data, err := scene.Marshal()
	if err != nil {
		test.Fatal(err)
	}
	read, err := Read(data)
	if err != nil {
		test.Fatal(err)
	}
	light := read.Nodes[2].Light
	if read.Ambient.Colour != "#203050" || read.Ambient.Strength() != 0.4 {
		test.Errorf("the ambient came back as %+v", read.Ambient)
	}
	if inner, outer := light.Radii(); light.Kind != PointLight || light.Strength() != 1.5 || inner != 0.5 || outer != 3.5 {
		test.Errorf("the light came back as %+v", light)
	}
}

func TestWhatALightLeavesUnsaidHasItsDefaults(test *testing.T) {
	light := &Light{Kind: PointLight}
	if inner, outer := light.Radii(); inner != 0 || outer != DefaultLightOuter || light.Strength() != DefaultIntensity {
		test.Errorf("an empty light is %v to %v at %v", inner, outer, light.Strength())
	}
	tooStrong := 5.0
	if (&Ambient{Intensity: &tooStrong}).Strength() != MostIntensity {
		test.Error("an intensity above the most is not held at the most")
	}
}

func TestACopyOfALightIsIndependent(test *testing.T) {
	strong := 1.5
	light := &Light{Kind: PointLight, Intensity: &strong}
	ambient := &Ambient{Intensity: &strong}
	lightCopy, ambientCopy := light.Copy(), ambient.Copy()
	*lightCopy.Intensity, *ambientCopy.Intensity = 0.1, 0.2
	if light.Strength() != 1.5 || ambient.Strength() != 1.5 {
		test.Error("changing a copy changed the original")
	}
}
