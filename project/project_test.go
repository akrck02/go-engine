package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

func TestANewProjectIsAGoModuleWithAScene(test *testing.T) {
	folder := filepath.Join(test.TempDir(), "harbour")
	created, err := Create(folder, "Harbour Lights")
	if err != nil {
		test.Fatal(err)
	}
	for _, file := range []string{"go.mod", "main.go", "spin.go", "engine.json", "scenes/main.scene.json", ".vscode/launch.json", ".zed/tasks.json"} {
		if _, err := os.Stat(created.Path(file)); err != nil {
			test.Errorf("%s is missing", file)
		}
	}
	module, _ := os.ReadFile(created.Path("go.mod"))
	if !strings.Contains(string(module), "module game/harbourlights") {
		test.Errorf("go.mod:\n%s", module)
	}
	opened, err := Open(folder)
	if err != nil || opened.Settings.Name != "Harbour Lights" {
		test.Fatalf("the project did not open: %v", err)
	}
	scene, err := engine.Load(opened.Files(), opened.Settings.StartScene)
	if err != nil {
		test.Fatal(err)
	}
	if _, found := scene.MainCamera(); !found {
		test.Fatal("the first scene has no camera")
	}
	if _, err := Create(folder, "Again"); err == nil {
		test.Fatal("a project was made over another")
	}
}

func TestModelsAreImportedAndScenesAdded(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	source := filepath.Join(test.TempDir(), "Boat.glb")
	os.WriteFile(source, []byte("glTF"), 0o644)
	imported, err := created.Import(source)
	if err != nil || imported != "assets/models/Boat.glb" {
		test.Fatalf("imported as %q: %v", imported, err)
	}
	if models := created.Models(); len(models) != 1 || models[0] != imported {
		test.Fatalf("models: %v", models)
	}
	if _, err := created.Import(filepath.Join(test.TempDir(), "photo.png")); err == nil {
		test.Fatal("a picture was imported as a model")
	}
	file, err := created.NewScene("Second level", scenefile.Mode2D)
	if read, _ := created.ReadScene(file); err != nil || file != "scenes/second-level.scene.json" || !read.Is2D() {
		test.Fatalf("new scene %q: %v", file, err)
	}
	if scenes := created.Scenes(); len(scenes) != 2 {
		test.Fatalf("scenes: %v", scenes)
	}
}

func TestTerminalEditorsAreKnown(test *testing.T) {
	if !isTerminalEditor("/usr/bin/nvim") || isTerminalEditor("zed") {
		test.Fatal("nvim runs in a terminal, zed does not")
	}
}

func TestBehavioursAreFoundInTheProjectsCode(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
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

// Bell only hears touches.
type Bell struct{}

func (bell *Bell) Touched(context *engine.Context, other *engine.Node) {}

// NotABehaviour has no Start nor Update.
type NotABehaviour struct{ Value int }
`), 0o644)
	found := created.Behaviours()
	if len(found) != 3 || found[0].Name != "Bell" || found[1].Name != "Drift" || found[2].Name != "Spin" {
		test.Fatalf("behaviours: %+v", found)
	}
	drift := found[1]
	if drift.Doc != "Drift moves a node slowly." || len(drift.Fields) != 2 || drift.Fields[0] != (Field{"speed", "float64", "metres a second"}) || drift.Fields[1].Name != "Label" {
		test.Fatalf("drift: %+v", drift)
	}
}

func TestTheAssetsTreeIsTheAssetsFolder(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	os.MkdirAll(created.Path("assets/sprites/hero"), 0o755)
	os.WriteFile(created.Path("assets/sprites/hero/idle.png"), nil, 0o644)
	os.WriteFile(created.Path("assets/readme.txt"), nil, 0o644)
	os.WriteFile(created.Path("assets/models/Tree.glb"), nil, 0o644)
	tree := created.AssetTree()
	if len(tree.Children) != 3 || tree.Children[0].Name != "models" || tree.Children[1].Name != "sprites" || tree.Children[2].Name != "readme.txt" {
		test.Fatalf("tree: %+v", tree.Children)
	}
	if models := tree.Children[0].Children; len(models) != 1 || models[0].Path != "assets/models/Tree.glb" {
		test.Fatalf("models: %+v (the .gitkeep is hidden)", models)
	}
	if KindOf("a/b/IDLE.PNG") != ImageAsset || KindOf("x.glb") != ModelAsset || KindOf("x.txt") != OtherAsset {
		test.Fatal("kinds")
	}
}

func TestAnAssetsUsesAreFoundInEveryScene(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	flat, _ := created.NewScene("Coast", scenefile.Mode2D)
	scene, _ := created.ReadScene(flat)
	lamp := scenefile.NewNode("Lamp")
	lamp.Sprite = "assets/lamp.sprite.json"
	harbour := scenefile.NewNode("Harbour")
	harbour.Children = []scenefile.Node{lamp}
	scene.Nodes = append(scene.Nodes, harbour)
	created.WriteScene(flat, scene)
	sprite := &scenefile.Sprite{Picture: "assets/lamp.png", SpriteSheet: scenefile.SpriteSheet{Columns: 1, Rows: 1}}
	data, _ := sprite.Marshal()
	os.WriteFile(created.Path("assets/lamp.sprite.json"), data, 0o644)
	uses := created.UsesOf("assets/lamp.sprite.json")
	if len(uses) != 1 || uses[0].File != flat || uses[0].Node != "Harbour/Lamp" {
		test.Fatalf("the sprite is used in %+v", uses)
	}
	if uses := created.UsesOf("assets/lamp.png"); len(uses) != 1 || uses[0].File != "assets/lamp.sprite.json" || uses[0].Node != "" {
		test.Fatalf("the picture is used in %+v", uses)
	}
}

func TestSettingsAreSaved(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	created.Settings.Name, created.Settings.CodeEditor = "Harbour Lights", "nvim"
	if err := created.SaveSettings(); err != nil {
		test.Fatal(err)
	}
	opened, err := Open(created.Folder)
	if err != nil || opened.Settings.Name != "Harbour Lights" || opened.Settings.CodeEditor != "nvim" {
		test.Fatalf("the settings came back as %+v, %v", opened.Settings, err)
	}
}

func TestALocalGoEngineGivesItsCommitToTheBuild(test *testing.T) {
	created, err := Create(filepath.Join(test.TempDir(), "game"), "Game")
	if err != nil {
		test.Fatal(err)
	}
	flags := created.EngineFlags()
	if len(flags) != 1 || !strings.Contains(flags[0], "engine.localRevision=") {
		test.Fatalf("a project that replaces Go Engine with this folder builds with %v", flags)
	}
}
