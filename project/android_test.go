package project

import (
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
)

func TestAnApplicationIDIsMadeFromTheGamesName(test *testing.T) {
	cases := []struct{ name, wanted string }{
		{"Harbour Lights", "io.goengine.harbourlights"},
		{"3 Lamps!", "io.goengine.game3lamps"},
		{"", "io.goengine.game"},
	}
	for _, each := range cases {
		project := &Project{Settings: engine.Settings{Name: each.name}}
		if made := project.ApplicationID(); made != each.wanted {
			test.Errorf("%q gives the id %q, not %q", each.name, made, each.wanted)
		}
	}
	project := &Project{Settings: engine.Settings{Name: "Harbour Lights", AndroidID: "org.example.harbour"}}
	if kept := project.ApplicationID(); kept != "org.example.harbour" {
		test.Errorf("the id the settings keep is not used: %q", kept)
	}
}

func TestAnAndroidBuildAsksForThePackage(test *testing.T) {
	project := &Project{Folder: "/games/harbour", Settings: engine.Settings{Name: "Harbour Lights"}}
	arguments := project.androidArguments()
	joined := strings.Join(arguments, " ")
	for _, wanted := range []string{"tool go-ui-apk", "-name Harbour Lights", "-id io.goengine.harbourlights", "-abi arm64,x86_64", "-o " + project.AndroidPackage()} {
		if !strings.Contains(joined, wanted) {
			test.Errorf("the Android build is asked for with %q, without %q", joined, wanted)
		}
	}
	if !strings.HasSuffix(project.AndroidPackage(), "harbour-lights.apk") || arguments[len(arguments)-1] != "." {
		test.Errorf("the package is %q, built from %q", project.AndroidPackage(), arguments[len(arguments)-1])
	}
	if !slices.Contains(androidEnvironment("/sdk", "/ndk"), "GOFLAGS=-tags=release") {
		test.Error("an Android build carries the editor: it is not built with the release tag")
	}
}

func TestAModuleKnowsWhetherItHasTheAndroidTool(test *testing.T) {
	without := "module game\n\ngo 1.25\n"
	with := without + "\ntool " + androidTool + "\n\nreplace " + uiModule + " => " + uiModule + " v0.0.0-1\n"
	if hasLine(without, "tool "+androidTool) || !hasLine(with, "tool "+androidTool) {
		test.Error("the tool line is not found where it is, or found where it is not")
	}
	if hasLine(without, "replace "+uiModule+" ") || !hasLine(with, "replace "+uiModule+" ") {
		test.Error("go-ui's replace is not found where it is, or found where it is not")
	}
}
