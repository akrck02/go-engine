package main

import "github.com/akrck02/go-engine/engine"

// Coin is collected when the player touches it: it counts for the player and leaves the level.
type Coin struct{}

// Touched runs when a moving body begins touching the coin.
func (coin *Coin) Touched(context *engine.Context, other *engine.Node) {
	player, found := engine.Behaviour[Player](other)
	if !found {
		return
	}
	player.collect()
	context.Node.Remove()
}
