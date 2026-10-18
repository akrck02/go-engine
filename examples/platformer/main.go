// Platformer, a demo made with go-engine: a slime walks and jumps through three forest levels,
// collecting coins, avoiding spikes and water, and reaching the sign at each level's end.
// Play it with `go run .` (arrows or the 3DS's cross move, Space or A jumps); edit it with
// `go run ../../cmd/go-engine .` from this folder.
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
	register()
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}

// register registers each behaviour the scenes use, by its type's name.
func register() {
	engine.Register(Player{}, Coin{}, Hazard{}, Flag{})
}
