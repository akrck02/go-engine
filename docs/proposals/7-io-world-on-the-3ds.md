# 7. Pushing, riding and pulling: io-world on the 3DS

> Status: in progress · Author's approval: proposal ☑ plan ☑ (2026-10-19, "sigue": together; the io-world repository private)

## Why

[io-world](https://github.com/akrck02/io-world) is a small puzzle-platformer made with Godot 3:
one level of 100-pixel tiles, a player of three pictures (body, tail, eyes), jump pads, a magnet
the player pushes and stands on, a platform on a rail, and a magnetic field that pulls the other
way. Made again from nothing in Go Engine and played on the 3DS, it answers two questions: can a
person make a game from scratch with the editor, and what does the engine still lack?

The engine's physics has three kinds of body: solids that never move, moving bodies that only
meet solids, and triggers. Three of io-world's mechanics need more, and any puzzle-platformer
will want the same:

- **Something to push**: a box the player pushes along the floor and stands on.
- **Something to ride**: a platform that moves and carries what stands on it.
- **Somewhere gravity turns**: an area where bodies fall another way.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A **pushable** body falls like a moving one, stops against solids, holds up a moving body that stands on it, and is pushed by one that walks into it, until a solid stops both | tests `a body pushes a box until a wall stops both`, `a body stands on a box` |
| R2 | A **platform** body moves at its velocity, ignores gravity and solids, carries the bodies standing on it, and pushes those in its way | tests `a platform carries what stands on it`, `a platform pushes a body in its way` |
| R3 | A trigger with a **gravity** pulls the moving and pushable bodies inside it that way instead of the world's; "on the floor" is then the side it pulls them to | tests `a body falls the way its zone pulls`, `a body stands on the floor its gravity pulls it to` |
| R4 | A behaviour changes the camera the scene is seen through by marking another one Main | test `marking another camera Main changes the view` |
| R5 | The three kinds and the gravity are chosen in Level's inspector, and drawn over the view when bodies are shown | capture |
| R6 | io-world, made with the editor (Art, Level, Code), plays on the 3DS: walk, jump, jump pads, the magnet, the rail and the field | the author on the console; captures from Azahar |
| R7 | What was hard while making it is written down, to become the next proposals | the list in this proposal's Result |

## Out of scope

- Bodies that turn, slopes, and shapes other than boxes.
- io-world's sounds, menus and levels beyond its first.
- Converting Godot scenes in the editor: a one-off script, kept in the game's own repository, as
  the platformer demo's was.

## Design

**Physics** (`physics`): two new kinds beside `Solid`, `Moving` and `Trigger`.

- `Pushable` moves as `Moving` does. A moving body that runs into it along X moves it by as much
  as it is not blocked (a chain of boxes moves together); along Y it stops the body, which stands
  on it. Solids stop it as they stop any moving body.
- `Platform` moves by its velocity before the others each step. Bodies whose feet rest on it move
  with it, and a body in its way is moved out of it, as a pushable is.
- A trigger's `Gravity`, when set, replaces the world's for the bodies inside it during that step.
  A body knows which way gravity pulled it, and `OnFloor` means "stopped going that way".

The order of a step: platforms, then moving and pushable bodies, then triggers.

**Scene files**: `BodyKind` gains `"pushable"` and `"platform"`; `Body` gains `gravity: [x, y]`,
read for triggers.

**Editor**: Level's body inspector offers the five kinds, and a gravity field for triggers; the
body overlay draws a platform's direction and a zone's arrow.

**io-world**: a new private repository, `akrck02/go-engine-io-world` (as the platformer's), made
with `go-engine new`. The art comes from io-world's PNGs through Art (sprites and a tileset of
100-pixel tiles); the level is drawn in Level, or converted once from `Level1.tscn` by a script in
the repository. Its behaviours in Go: `Player` (run, jump, reset), `Jumper` (a push along its
direction), `Rail` (a platform going back and forth), `Field` (a gravity zone the player can turn)
and the cameras. Built with tinygo-3ds from Build's Install and sent with 3dslink.

**Branches**: the engine's part on `feature/physics`, from `feature/engine` (where the other
session's work is); io-world in its own repository.

## Tasks

- [x] T1 — pushable bodies · `physics`, `scenefile`, `engine` · R1 · tests of R1
- [x] T2 — platforms · `physics`, `scenefile`, `engine` · R2 · tests of R2
- [ ] T3 — gravity zones · `physics`, `scenefile`, `engine` · R3 · tests of R3
- [ ] T4 — switching cameras, tested · `engine` · R4
- [ ] T5 — the kinds and the gravity in Level · `editor` · R5 · capture
- [ ] T6 — io-world: project, art, level, behaviours · new repository · R6 · captures
- [ ] T7 — io-world on the 3DS, and the list of what was hard · R6, R7
