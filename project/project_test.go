package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
)

func TestANewProjectIsAGoModuleWithAScene(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "harbour")
	created, err := Create(folder, "Harbour Lights")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"go.mod", "main.go", "spin.go", "engine.json", "scenes/main.scene.json", ".vscode/launch.json", ".zed/tasks.json"} {
		if _, err := os.Stat(created.Path(file)); err != nil {
			t.Errorf("%s is missing", file)
		}
	}
	module, _ := os.ReadFile(created.Path("go.mod"))
	if !strings.Contains(string(module), "module game/harbourlights") {
		t.Errorf("go.mod:\n%s", module)
	}
	opened, err := Open(folder)
	if err != nil || opened.Settings.Name != "Harbour Lights" {
		t.Fatalf("the project did not open: %v", err)
	}
	scene, err := engine.Load(opened.Files(), opened.Settings.StartScene)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := scene.MainCamera(); !found {
		t.Fatal("the first scene has no camera")
	}
	if _, err := Create(folder, "Again"); err == nil {
		t.Fatal("a project was made over another")
	}
}

func TestModelsAreImportedAndScenesAdded(t *testing.T) {
	created, err := Create(filepath.Join(t.TempDir(), "game"), "Game")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "Boat.glb")
	os.WriteFile(source, []byte("glTF"), 0o644)
	imported, err := created.Import(source)
	if err != nil || imported != "assets/models/Boat.glb" {
		t.Fatalf("imported as %q: %v", imported, err)
	}
	if models := created.Models(); len(models) != 1 || models[0] != imported {
		t.Fatalf("models: %v", models)
	}
	if _, err := created.Import(filepath.Join(t.TempDir(), "photo.png")); err == nil {
		t.Fatal("a picture was imported as a model")
	}
	file, err := created.NewScene("Second level")
	if err != nil || file != "scenes/second-level.scene.json" {
		t.Fatalf("new scene %q: %v", file, err)
	}
	if scenes := created.Scenes(); len(scenes) != 2 {
		t.Fatalf("scenes: %v", scenes)
	}
}

func TestTerminalEditorsAreKnown(t *testing.T) {
	if !isTerminalEditor("/usr/bin/nvim") || isTerminalEditor("zed") {
		t.Fatal("nvim runs in a terminal, zed does not")
	}
}

func TestBehavioursAreFoundInTheProjectsCode(t *testing.T) {
	created, err := Create(filepath.Join(t.TempDir(), "game"), "Game")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(created.Path("drift.go"), []byte(`package main

import "github.com/akrck02/go-engine/engine"

// Drift moves a node slowly. It never stops.
type Drift struct {
	Speed  float64 `+"`json:\"speed\"`"+` // metres a second
	Label  string
	hidden int
	Skip   bool `+"`json:\"-\"`"+`
}

func (drift *Drift) Start(context *engine.Context) {}

// NotABehaviour has no Start nor Update.
type NotABehaviour struct{ Value int }
`), 0o644)
	found := created.Behaviours()
	if len(found) != 2 || found[0].Name != "Drift" || found[1].Name != "Spin" {
		t.Fatalf("behaviours: %+v", found)
	}
	drift := found[0]
	if drift.Doc != "Drift moves a node slowly." || len(drift.Fields) != 2 || drift.Fields[0] != (FieldInfo{"speed", "float64", "metres a second"}) || drift.Fields[1].Name != "Label" {
		t.Fatalf("drift: %+v", drift)
	}
}
