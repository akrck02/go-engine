# 1. A platformer demo that drives the engine

> Status: done · Author's approval: proposal ☑ plan ☑

## Why

Go Engine (v0.0.1) can place things in a scene and run code on them, but nothing yet makes a
game feel like one: things don't fall, collide, animate or get followed by the camera. A small
platformer needs all of that, so building one tells us what the engine is missing and proves each
piece works, on the desktop and on the 3DS. The author's own `godot-generic-platformer` is the
reference for how it should play.

## What

The engine gains generic 2D pieces any game can use; the demo is a project made with the editor
that uses them, with only its own behaviours as game code.

| Id | Requirement | Check |
|---|---|---|
| R1 | A body falls with gravity and stands on solid ground; it never passes through a solid, at any frame rate | test `a falling body lands on the ground and stays there`; test `a fast body does not pass through a thin wall` |
| R2 | A body slides along walls and floors instead of stopping dead, and knows whether it stands on the floor | test `a body pushed into a wall slides up it`; test `a body on the ground is on the floor` |
| R3 | A level is painted in the editor from a tileset picture onto a grid, and its solid tiles collide | in the editor, paint tiles and play: the player stands on them; test `solid tiles stop a body` |
| R4 | A sprite plays named animations from a sprite sheet (idle, walk, jump), at a set speed, and faces left or right | test `an animation shows each frame for its time and loops`; the demo's player walks and turns |
| R5 | The camera follows a node smoothly and stays inside the level's bounds | test `the camera follows its target and stops at the level's edge` |
| R6 | Triggers: something happens when the player touches a node (a coin is collected, spikes restart the level, a flag goes to the next level) | test `touching a trigger is reported once`; the demo's coins, spikes and flag |
| R7 | The demo plays the same on the desktop and the 3DS: arrows or the cross move, Space, the up arrow, A or up on the cross jumps | play it in the editor, in a window, and on the 3DS emulator; the author plays it on the console |
| R8 | Everything the demo needs from the engine is generic: no part of the engine knows it is a platformer | review: the engine's new code names no coin, spike or player |

## Out of scope

- Slopes, one-way platforms, moving platforms and ladders (later, if the demo grows).
- Enemies with their own behaviour, shooting, bounce pads.
- Sound and music (Go Sound comes in a later change), menus, a game interface, saving.
- 3D physics.
- Documentation of the new pieces: v0.0.1 is not documented yet.

## Decisions

- **Art**: the sprites and tilesets of `godot-generic-platformer`, as the author allowed; the
  demo says where they come from.
- **Where**: `examples/platformer/`, a project with its own `go.mod`, beside the engine, so each
  change to the engine is played with it in the same commit.
- **Scope**: one world, the forest, in three short levels. Lava and snow come later, from the same
  pieces.

## Design

Four generic pieces join the engine; the demo only adds its behaviours, scenes and art.

```
space ← physics ← engine ← editor
```

**1. `physics`, a new package**: boxes that move and collide in 2D, knowing nothing of nodes.

- A `Body` has a box (its size in metres), a velocity, a gravity scale and a kind: *solid*
  (never moves), *moving* (falls and is pushed back by solids) or *trigger* (only notices what
  touches it).
- `World.Step(seconds)` adds gravity, then moves each moving body one axis at a time, in steps no
  longer than half its smallest side, so a fast body cannot jump through a thin wall (R1). A body
  that hits a solid stops on that axis only and keeps the other, which is sliding (R2); hitting a
  floor sets `OnFloor`.
- Overlaps of triggers are reported once, when they begin (R6).
- Alternatives not taken: a physics library (Box2D ports, Chipmunk). They bring rotation, joints
  and a dependency the demo doesn't need; boxes lined up with the axes are what platformers use,
  and small enough to learn from.

**2. Tiles, in `engine` and `scenefile`**: a node may be a tile map: a tileset picture, the size
of a tile in pixels, a grid of tile numbers, and which tiles are solid (R3).

- The tileset is cut into one picture per tile when it loads, so `draw2d` draws whole pictures
  and Go Renderer needs no change.
- Solid tiles become solid bodies, rows of them merged into one box, so a level of a thousand
  tiles is a few dozen boxes.
- The editor gains a tile tool in the 2D view: the tileset in the inspector to choose from, drag
  to paint, right-drag to erase.

**3. Animated sprites, in `engine` and `scenefile`**: a sprite may be a sheet of equal frames, with
named animations (frames, frames per second, looping), and may face left (R4).

- Frames are cut from the sheet when it loads; facing left uses a mirrored copy made once, since
  `draw2d` draws pictures upright.
- A behaviour calls `node.Play("walk")` and `node.FaceLeft(true)`.

**4. Bodies, triggers, the camera and scenes, in `engine`**:

- A node may have a body (`"body": {"kind": "moving", "size": [0.8, 0.6]}` in the scene file);
  the engine keeps the physics world in step with the nodes every frame: behaviours update, the
  world steps, nodes take their bodies' positions.
- A behaviour with `Touched(context, other *Node)` hears when its node's trigger begins touching
  another body (R6).
- A camera may follow a node, with a smoothing time, kept inside the level's bounds: the tile
  maps' extent, or bounds of its own (R5). The tile maps bound the sides and the bottom; the sky
  above stays open, so the camera rises with a jump, unless the camera has a ceiling (a cave).
- `context.RestartScene()` and `context.ChangeScene(file)` let a game restart a level or go to the
  next, in a window, on the 3DS and in the editor's view (R6, R7).

**The demo**, `examples/platformer/`: a project like any other (its `go.mod` points at the engine
beside it). Behaviours `Player` (walk, jump with Space or A, animations, facing), `Coin`
(collected when touched), `Hazard` (spikes and water: restart the level) and `Flag` (next level);
three forest levels: the first two converted from the reference's `World.tscn` and `World2.tscn`
(its third is a copy of the second), the third joining them into one wider than the screen. No behaviour of the demo reaches into the engine's insides
(R8).

## Tasks

- [x] T1 — `physics`: bodies, gravity, moving one axis at a time in short steps, `OnFloor` ·
  `physics/` · R1, R2 · tests: `a falling body lands on the ground and stays there`,
  `a fast body does not pass through a thin wall`, `a body pushed into a wall slides up it`,
  `a body on the ground is on the floor`
- [x] T2 — `physics`: triggers that report when an overlap begins, once · `physics/` · R6 ·
  test: `touching a trigger is reported once`
- [x] T3 — scene files: bodies, tile maps and sprite sheets with animations, read and written ·
  `scenefile/` · R3, R4 · test: `a scene with bodies, tiles and animations comes back as written`
- [x] T4 — engine: nodes with bodies stepped every frame, `Touched` for behaviours · `engine/` ·
  R1, R2, R6 · test: `a node with a body falls onto a solid node`
- [x] T5 — engine: tile maps drawn in 2D, solid tiles as merged solid bodies · `engine/` · R3 ·
  test: `solid tiles stop a body`, `a row of solid tiles is one box`
- [x] T6 — engine: animated sprites, `Play`, `FaceLeft` · `engine/` · R4 · test:
  `an animation shows each frame for its time and loops`
- [x] T7 — engine: a camera that follows a node within bounds · `engine/` · R5 · test:
  `the camera follows its target and stops at the level's edge`
- [x] T8 — engine: `RestartScene` and `ChangeScene`, in the window and in the editor's view ·
  `engine/`, `editor/` · R6, R7 · test: `changing scene starts the new one`
- [x] T9 — editor: inspector sections for bodies, sprite sheets and cameras that follow · `editor/` ·
  R3–R5 · test: `a body set in the inspector is saved`
- [x] T10 — editor: the tile tool (choose, paint, erase) in the 2D view · `editor/` · R3 · test:
  `painting a tile and erasing it` (and played by hand)
- [x] T11 — the demo: project, art, `Player`, `Coin`, `Hazard`, `Flag`, three levels ·
  `examples/platformer/` · R1–R8 · played in the editor, in a window and on the 3DS emulator
- [x] T12 — check on the 3DS: build the demo with TinyGo and play it in the emulator · R7
- [x] T13 — the author plays it on the console · R7 (2026-10-18: "funciona en ds")

## Result

**Built.** The `physics` package (boxes that fall, collide, slide, and triggers that report a
touch once); bodies, tile maps with merged solid rows, animated sprites that face either way, a
camera that follows within bounds (the sky open above the tile maps unless the camera has a
ceiling), and restarting or changing scenes, in `engine`; inspector sections for bodies, sheets and
following cameras and a tile tool in the editor; and the demo in `examples/platformer/`: three
forest levels (two converted from the reference's Godot levels, the third joining them), with
`Player`, `Coin`, `Hazard` and `Flag` as its only code.

**Checked.** Every requirement's test passes (`go test ./...` here and in the demo). The demo was
played by a script in the editor's view and drawn after some seconds; on the 3DS it was built
with TinyGo and played in Azahar at 60 frames a second: walking, jumping over the spikes with A,
the camera rising with the jump, a restart on the spikes. That found a real bug: go-ui hands the
scene a fresh copy of its input each frame, so a button was never "pressed"; the engine now keeps
the game's keys itself (`engine/console.go`).

**Review.** The reviewer's findings were fixed: a node moved or added while playing keeps or gets
its body; a body with no size takes short steps; a scene made by hand cannot restart and says so
(`engine.scene-not-restartable`); a missing picture is tried once and costs nothing after; a
restart shares the pictures already read (`TakeReplacement`, renamed because it acts); typed
frames are checked against the sheet before a run is counted out; removing an animation forgets
the texts typed for the others; the first animation can be unticked; a missing tileset says why;
remove buttons have tooltips; the engine's own code and tests name no coin, spike or player.
Not changed, on purpose: `DefaultGravity` keeps its name (its unit, metres a second each second,
is in its comment); two long editor functions (`followMouse`, `nodeInspector`) are rewritten by
proposal 2's Level task instead of split twice.

**For a person.** Play the demo on the console (T13): the jump's height, the camera and the
frame rate on the real machine.
