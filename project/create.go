package project

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/scenes"
)

// mainFile is the game's main: it embeds the project's files and runs them.
const mainFile = `// %s, made with go-engine. The editor keeps the scenes in scenes/ and the imported models in
// assets/; the game's own code is the Go beside this file. Play it with ` + "`go run .`" + `.
package main

import (
	"embed"
	"log"

	"github.com/akrck02/go-engine/engine"
)

// files are the project's settings, scenes and assets, carried inside the game.
//
//go:embed engine.json scenes all:assets
var files embed.FS

func main() {
	// Each behaviour the scenes use is registered here, by its type's name.
	engine.Register(Spin{})
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}
`

// spinFile is a first behaviour, to show how the game's code reaches its scenes.
const spinFile = `package main

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/space"
)

// Spin turns its node around the vertical axis. In the editor, give a node the behaviour "Spin"
// and set its speed; the field's name in the scene is its JSON name.
type Spin struct {
	Speed float64 ` + "`json:\"speed\"`" + ` // degrees a second
}

// Update runs every frame, with the seconds since the last one in context.Delta.
func (spin *Spin) Update(context *engine.Context) {
	context.Node.Transform.Turn(space.Up, spin.Speed*context.Delta)
}
`

// The code editors' settings: a task or a launch that plays the game, so it runs from the editor.
const (
	vscodeLaunch = `{
	"version": "0.2.0",
	"configurations": [
		{
			"name": "Play",
			"type": "go",
			"request": "launch",
			"mode": "auto",
			"program": "${workspaceFolder}"
		}
	]
}
`
	zedTasks = `[
	{
		"label": "Play",
		"command": "go run .",
		"use_new_terminal": false
	}
]
`
	gitignore = "/build/\n"
)

// Create makes a new project in a folder (which must be empty or not exist): its settings, a
// scene with a ground, a cube and a camera, a main and a first behaviour, and the code editors'
// settings.
func Create(folder, name string) (*Project, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}
	if entries, err := os.ReadDir(folder); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("%s is not empty", folder)
	}
	source, err := build.SourceFolder()
	if err != nil {
		return nil, err
	}
	module, err := build.ModuleFile(modulePath(name), source)
	if err != nil {
		return nil, err
	}
	created := &Project{Folder: folder, Settings: engine.Settings{Name: name, StartScene: path.Join(ScenesFolder, "main"+SceneSuffix)}}
	files := map[string]string{
		"go.mod":                 module,
		"main.go":                fmt.Sprintf(mainFile, name),
		"spin.go":                spinFile,
		".vscode/launch.json":    vscodeLaunch,
		".zed/tasks.json":        zedTasks,
		".gitignore":             gitignore,
		"assets/models/.gitkeep": "",
	}
	for file, content := range files {
		if err := writeFile(created.Path(file), []byte(content)); err != nil {
			return nil, err
		}
	}
	if sums := build.Sums(source); sums != nil {
		if err := writeFile(created.Path("go.sum"), sums); err != nil {
			return nil, err
		}
	}
	if err := created.SaveSettings(); err != nil {
		return nil, err
	}
	if err := created.WriteScene(created.Settings.StartScene, starterScene("Main")); err != nil {
		return nil, err
	}
	// tidy keeps go.sum to what the game needs; without a network it may fail, and go run tidies
	// what it must then
	build.Run(folder, nil, build.GoTool(), "mod", "tidy")
	return created, nil
}

// modulePath is the module's name, from the game's.
func modulePath(name string) string {
	return "game/" + strings.ReplaceAll(build.FileName(name), "-", "")
}

// starterScene is a new scene: a ground to stand on, a cube that spins, and a camera looking at
// them.
func starterScene(name string) *scenes.Scene {
	ground := scenes.NewNode("Ground")
	ground.Shape, ground.Colour, ground.Scale = "plane", "#7a9a6a", [3]float64{20, 1, 20}
	cube := scenes.NewNode("Cube")
	cube.Shape, cube.Colour, cube.Position = "cube", "#e07a5f", [3]float64{0, 0.5, 0}
	cube.Behaviours = []scenes.Behaviour{{Name: "Spin", Fields: map[string]any{"speed": 45}}}
	camera := scenes.NewNode("Camera")
	camera.Position, camera.Rotation = [3]float64{0, 3, 7}, [3]float64{-20, 0, 0}
	camera.Camera = &scenes.Camera{Projection: scenes.Perspective, FieldOfView: 50, Main: true}
	return &scenes.Scene{Name: name, Nodes: []scenes.Node{ground, cube, camera}}
}
