package editor

import "testing"

func TestABehavioursUsesAreListedByScene(test *testing.T) {
	editor := open(test)
	uses := editor.BehaviourUses("Spin")
	if len(uses) != 1 || uses[0].File != editor.SceneFile || uses[0].Node != "Cube" {
		test.Fatalf("Spin is used in %+v", uses)
	}
	spin, _ := editor.FindBehaviour("Spin")
	if len(spin.Methods) != 1 || spin.Methods[0] != "Update" {
		test.Fatalf("Spin's methods are %v", spin.Methods)
	}
}

func TestAMissingBehaviourIsReported(test *testing.T) {
	editor := open(test)
	cube, _ := editor.Scene.Find("Cube")
	editor.Select(cube)
	editor.AttachBehaviour("Lighthouse")
	editor.Save()
	missing := editor.MissingBehaviours()
	if len(missing) != 1 || missing[0].Name != "Lighthouse" || len(missing[0].Uses) != 1 {
		test.Fatalf("the missing behaviours are %+v", missing)
	}
}
