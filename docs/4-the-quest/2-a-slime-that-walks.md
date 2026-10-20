# 🎭 Level 2: a slime that walks

You bring in the hero, a slime, give it a body that falls and stands, and write your first
behaviour: the code that walks and jumps when the player presses the arrows and Space.

#### The idea

The slime is two nodes, as a puppet is a body and a costume. **Player** is the body: the physics
moves it, and your code tells it where to go. Its child **Sprite** is the costume: the picture
that follows the body and plays its animations. Keeping them apart lets the costume sit a little
higher than the body, or change, without the physics noticing.

## The slime's animations

Open the forest project of level 1. In **Project**, choose `slime.sprite.json` (it came with the
pictures you copied) and press **Open in Art**. A **sprite file** cuts a picture into frames, here
5 across and 2 down, and names runs of them: **idle** is frame 0, **walk** frames 0 to 3 at 3 a
second, looping. The other frames are spare. Choose an animation to watch it under **Preview**;
**Play it when the scene starts** would make it the one the slime starts with.

To make one of your own, choose any picture in Project and press **Make it a sprite or a tileset**.

## Something to stand on

Back in **Level**, choose the **Ground**, the green box the new scene came with (a flat shape: a cube
is drawn as a square in 2D). In its inspector, under **Body**, press **Give it a body**. Leave it
**Solid**: it never moves, and what falls lands on it.

## The body

1. Press **+** above the tree and choose **Empty**. Name the new node *Player*.
2. Give it a body, and choose **Moving**. Under the kinds, set its size: width 0.94, height 0.63
   metres, the slime's own shape.
3. Set its **gravity scale** to 5.75: it falls with 5.75 times Earth's pull. Games make their
   heroes fall much faster than real things do; it feels better under the thumbs.
4. An empty node shows only as a small dot: set its position in the inspector, X 8 and Y 0, above
   the ground.

## The costume

1. With *Player* chosen, go to **Project**, choose `slime.sprite.json` and press **Place in the
   scene**. A new node lands after *Player*, and is chosen.
2. Over the tree, press **Make it a child of the node above** (the third arrow; its tooltip says
   so). Name it *Sprite*. A child's position is from its parent, and both are centred on their
   position. The slime's picture is a little taller than its body, so set the sprite's Y to 0.19:
   its feet then sit on the body's bottom.

## The code

Add a file `player.go` beside `main.go`:

```go
package main

import "github.com/akrck02/go-engine/engine"

// Player walks with the arrows (or the 3DS's cross), jumps with Space (or A) while it stands on
// the ground, and plays its sprite's animations.
type Player struct {
	Speed float64 `json:"speed"` // metres a second, walking
	Jump  float64 `json:"jump"`  // metres a second, upwards, when it jumps
}

// Update runs every frame.
func (player *Player) Update(context *engine.Context) {
	body, found := context.Node.Body()
	if !found {
		return
	}
	direction := context.Input.Move().X
	body.Velocity.X = direction * player.Speed
	if context.Input.Pressed(engine.KeySpace) && body.OnFloor() {
		body.Velocity.Y = player.Jump
	}
	player.animate(context.Node, direction)
}

// animate plays "walk" while the player walks and "idle" while it stands, facing the way it goes.
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
	sprite.FaceLeft(direction > 0) // the slime is drawn facing left
}
```

The code never moves the slime itself. It sets the body's **velocity**, how fast it goes, and the
physics moves it, stopping it at the ground. `Move()` is -1, 0 or 1 across, from the arrows, the
3DS's cross or a phone's buttons. Space is also the 3DS's A. `OnFloor()` keeps the slime from
jumping in the air. `node.Find("Sprite")` finds the child named *Sprite*, and `found` says whether
it is there. The slime's picture faces left: `FaceLeft(true)` mirrors it, so walking right
(`direction > 0`) shows it facing right.

Then register it in `main.go`, whose line `engine.Register(Spin{})` becomes:

```go
engine.Register(Spin{}, Player{})
```

In **Code** (key 4), press **Rebuild**. It saves the scene, builds the game again, and reopens the
editor where you were.

## The behaviour

Choose *Player*. Under **Behaviours**, press **Add** and choose **Player**. Its fields appear: set
**speed** to 3.2 and **jump** to 18.75.

## Checkpoint

Press **Play**. The slime falls onto the ground; the arrows walk it, Space jumps.

<p align="center"><img alt="Playing: the slime in the air over the green ground, in front of the snowy hills and the dark forest" src="../images/quest/2-play.png" width="680"></p>

The slime should turn the way it walks, and wobble while it walks. If it does not move, check in
**Code** that *Player* is listed: the scene names the behaviour by its Go type's name. If it falls
through the ground, check that the ground's body is **Solid**; **Bodies**, over the view, outlines
every body.

## Bonus: run

Make the slime run twice as fast while Shift, or X (the keyboard's, or the 3DS's button), is held.

<details>
<summary>🔓 Show the solution</summary>

`Down` tells whether a key is held now, where `Pressed` tells whether it was just pressed. These
lines take the place of `body.Velocity.X = direction * player.Speed`:

```go
speed := player.Speed
if context.Input.Down(engine.KeyShift) || context.Input.Down(engine.KeyX) {
	speed *= 2
}
body.Velocity.X = direction * speed
```

</details>

**🎭 Achievement unlocked: Puppeteer.** You gave a node a body and moved it with your own code.

---

[← Level 1](1-a-forest-of-your-own.md) · [Level 3 →](3-ground-to-stand-on.md)
