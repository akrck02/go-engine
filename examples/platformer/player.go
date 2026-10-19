package main

import "github.com/akrck02/go-engine/engine"

// Player walks with the arrows (or the 3DS's cross and circle pad), jumps with Space, the up
// arrow or the 3DS's A while it stands on the ground, and plays its sprite's animations: "walk"
// while it walks, "idle" otherwise. Its picture is its child "Sprite", a little above its body.
type Player struct {
	Speed float64 `json:"speed"` // metres a second, walking
	Jump  float64 `json:"jump"`  // metres a second, upwards, when it jumps

	Coins int `json:"-"` // the coins collected in this level
}

// Update runs every frame.
func (player *Player) Update(context *engine.Context) {
	body, found := context.Node.Body()
	if !found {
		return
	}
	direction := context.Input.Move().X
	body.Velocity.X = direction * player.Speed
	jumping := context.Input.Pressed(engine.KeySpace) || context.Input.Pressed(engine.KeyUp)
	if jumping && body.OnFloor() {
		body.Velocity.Y = player.Jump
	}
	player.animate(context.Node, direction)
}

// animate plays the walk while the player walks, and turns it the way it goes.
func (player *Player) animate(node *engine.Node, direction float64) {
	sprite, found := node.Find("Sprite")
	if !found {
		return
	}
	if direction == 0 {
		sprite.Play("idle")
		return
	}
	sprite.Play("walk")
	sprite.FaceLeft(direction > 0) // the slime is drawn facing left: walking right mirrors it
}

// collect counts a coin the player touched.
func (player *Player) collect() { player.Coins++ }

// isPlayer tells whether a node is the player.
func isPlayer(node *engine.Node) bool {
	_, found := engine.Behaviour[Player](node)
	return found
}
