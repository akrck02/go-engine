# 7. Pushing, riding and pulling: io-world on the 3DS

> Status: done (the console check is the author's) · Author's approval: proposal ☑ plan ☑ (2026-10-19, "sigue": together; the io-world repository private)

## Why

[io-world](https://github.com/akrck02/io-world) is a small puzzle-platformer made with Godot 3.
It has one level of 100-pixel tiles and a player of three pictures (body, tail, eyes). It has
jump pads, a magnet the player pushes and stands on, a platform on a rail, and a magnetic field
that pulls the other way.

Made again from nothing in Go Engine and played on the 3DS, it answers two questions. Can a person
make a game from scratch with the editor? And what does the engine still lack?

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

- `Pushable` moves as `Moving` does. A moving body that runs into it along X moves it, by as much
  as it is not blocked; a chain of boxes moves together. Along Y it stops the body, which stands
  on it. Solids stop it as they stop any moving body.
- `Platform` moves by its velocity before the others each step. Bodies whose feet rest on it move
  with it, and a body in its way is moved out of it, as a pushable is.
- A trigger's `Gravity`, when set, replaces the world's for the bodies inside it during that step.
  A body knows which way gravity pulled it, and `OnFloor` means "stopped going that way".

The order of a step: platforms, then moving and pushable bodies, then triggers.

**Scene files**: `BodyKind` gains `"pushable"` and `"platform"`. `Body` gains `gravity: [x, y]`,
read for triggers.

**Editor**: Level's body inspector offers the five kinds, and a gravity field for triggers. The
body overlay draws a platform's direction and a zone's arrow.

**io-world**: a new repository, made with `go-engine new`. The art comes from io-world's PNGs
through Art: sprites, and a tileset of 100-pixel tiles. The level is drawn in Level, or converted
once from `Level1.tscn` by a script in the repository. Its behaviours in Go:

- `Player`: run, jump, reset.
- `Jumper`: a push along its direction.
- `Rail`: a platform going back and forth.
- `Field`: a gravity zone the player can turn.
- The cameras.

It is built with tinygo-3ds from Build's Install, and sent with 3dslink.

**Branches**: the engine's part on `feature/physics`, from `feature/engine`, where the other
session's work is. io-world goes in its own repository.

## Tasks

- [x] T1 — pushable bodies · `physics`, `scenefile`, `engine` · R1 · tests of R1
- [x] T2 — platforms · `physics`, `scenefile`, `engine` · R2 · tests of R2
- [x] T3 — gravity zones · `physics`, `scenefile`, `engine` · R3 · tests of R3
- [x] T4 — switching cameras, tested · `engine` · R4
- [x] T5 — the kinds and the gravity in Level · `editor` · R5 · capture
- [x] T6 — io-world: project, art, level, behaviours · new repository · R6 · captures
- [x] T7 — io-world on the 3DS, and the list of what was hard · R6, R7
- [x] T8 — asked while building: a menu of the scenes on the 3DS's bottom screen (`sceneMenu`), and
  `SetBottomScreen` for a game to draw it its own way · `engine`, `editor` · tests `the menu lists
  the scenes by their names`, `a game's own bottom screen is given its scenes`

## Result

**Built.**

- Two new kinds of body, `pushable` and `platform`, and gravity zones (a trigger's `pull`). They
  are in the physics, the scene files and Level's inspector, each drawn its own way.
- A body's floor is the side its gravity pulls it to.
- Pushing goes across that gravity, in a row of at most eight, and never round a ring.
- A platform carries the stacks standing on it, pushes what is in its way, and stops against what
  it cannot move.
- Triggers notice pushable bodies too.
- On the 3DS, a menu of the scenes on the bottom screen (`sceneMenu`), or the game's own
  (`SetBottomScreen`).

Io World has io-world's first level, and its dev branch's version: a button that starts a rail
that waits, and a cube. A menu of levels is made of the game's own pieces. It was sent to the
author's 3DS by FTP. It now lives in [io-world](https://github.com/akrck02/io-world)'s own
repository, in its `IoWorldGo/` folder.

**Checked.** Every requirement's tests pass: the physics', the engine's allocations with the new
kinds, and the editor's. Io World's own tests cover landing, running, pads, the magnet, the rail,
the field, the button and the cube. The editor was captured, and the game in Azahar, with the menu
changing the level by a touch.

**Not done.** Bodies that turn, slopes, tiles' own shapes and flipped tiles. io-world's attractor
tiles are partial polygons, here whole squares; flipped tiles are drawn unflipped. The original's
two fixed screens became one camera that follows Io. At 400 pixels, a whole screen's width made
the tiles 14 pixels.

**What was hard, making it** (for the next proposals):

1. **The editor was not where it was made.** The level came from a script
   (`tools/convert_level.py`), and the behaviours from a code editor. The editor opened, played and
   changed it. Drawing a 90-by-37 level of 100-pixel tiles by hand, and five jump pads, would take
   an afternoon: there are no prefabs to place again and again.
2. **Prefabs.** io-world's jump pad, rail and button are small scenes placed many times. Here each
   is written out in full. Instancing a scene in another is the biggest missing piece.
3. **Sprites made of tiles.** The pad and the rail are rows of a tile picture. Art cannot compose
   them, so the script did, into new pictures.
4. **One tileset per tile map.** Godot's tile set mixes pictures: one level became two tile maps.
5. **Touching needs overlapping.** A pad's trigger, level with the ground it is on, never notices
   what stands on it: it has to reach above it. Easy to get wrong, and silent.
6. **`go-engine new` opens the editor**, so it needs a display. Making a project on a server
   needed `project.Create` from Go.
7. **Choices the original left open.** The field pulled left in the level and up in its prefab,
   and worked in neither. The button only printed. Here the field pulls up (the inspector changes
   it), and the button starts the rail, as io-world's first, Unity version did.

**The original's other scenes.** io-world began in Unity, with `Demoivania` and `MainMenu`.
`Demoivania` (December 2022) had a menu with music and lights, and a long level with levels 2 and
3 started. It had buttons, a cube, rails, magnets, liquids and another character, with about 710
brick sprites and 30 C# scripts. They went away when it moved to Godot, which only ever had
`Level1`. They could come here too, as a later proposal.

**For a person.** Play both levels on the console: the controls, the rail and the field's feel,
and the menu under a thumb.
