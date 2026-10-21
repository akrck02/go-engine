// Night watch, a sample of go-engine for a game that plays the same on every device: it reads
// actions, not keys, and shows its surfaces (the lamp's heading, the keeper's log) where
// engine.json places them on each target, with no file of its own for any device. The keeper turns
// the lamp (arrows, the circle pad or a phone's stick) and marks in the log a ship where the lamp
// points (Space, A, a phone's button). Play it with `go run .`; edit it with
// `go run ../../cmd/go-engine .`.
package main

import (
	"embed"
	"log"

	"github.com/akrck02/go-engine/engine"
)

// files are the project's settings and scenes, carried inside the game.
//
//go:embed engine.json scenes
var files embed.FS

func main() {
	engine.Register(Lamp{})
	engine.SetSurface("Lamp", showHeading)
	engine.SetSurface("Log", showLog)
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}
