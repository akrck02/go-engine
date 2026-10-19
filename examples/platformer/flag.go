package main

import "github.com/akrck02/go-engine/engine"

// Flag ends the level when the player touches it, and goes on to the next one.
type Flag struct {
	Next string `json:"next"` // the next level's scene file, as "scenes/forest-2.scene.json"
}

// Touched runs when a moving body begins touching the flag.
func (flag *Flag) Touched(context *engine.Context, other *engine.Node) {
	if isPlayer(other) {
		context.ChangeScene(flag.Next)
	}
}
