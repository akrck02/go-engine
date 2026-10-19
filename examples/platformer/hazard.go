package main

import "github.com/akrck02/go-engine/engine"

// Hazard starts the level again when the player touches it: spikes, water.
type Hazard struct{}

// Touched runs when a moving body begins touching the hazard.
func (hazard *Hazard) Touched(context *engine.Context, other *engine.Node) {
	if isPlayer(other) {
		context.RestartScene()
	}
}
