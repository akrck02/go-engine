# 13. io-world's Unity levels, made again

> Status: proposed, after 9 to 12 · Author's approval: proposal ☐ plan ☐

## Why

io-world began in Unity, in December 2022. Its scene `Demoivania` had more than the Godot level
proposal 7 made again: a menu with music and lights, a long level with levels 2 and 3 started,
buttons, a cube, rails, magnets, liquids and a second character. They went away when the game
moved to Godot.

Made again in Go Engine, they are the next test of the engine. They need prefabs (proposal 9),
richer tile maps (10), sprites made of tiles (11) and sound (12). And they bring two pieces of
their own.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A **liquid**: a trigger that slows what moves in it and lets it float up | tests `a body slows in a liquid`, `a light body floats up in a liquid` |
| R2 | **Two characters**: a behaviour hands control from one to the other, and the camera follows the one in control | test `the camera follows the character in control` |
| R3 | io-world's start menu, with its music and lights, opens the game | capture on the 3DS |
| R4 | The long level, with levels 2 and 3 as far as Unity had them, plays on the 3DS | the author on the console |

## Out of scope

- What Unity's version never finished: levels 2 and 3 end where they ended.
- Changing io-world's design. Its open choices are written down, not decided here.

## Design

**Liquids** (`physics`): a trigger gains a `drag`, a share of speed lost each second, and a
`lift`, a pull upwards, the opposite of gravity. Both are settings beside `pull` (proposal 7), so
the inspector shows them with it.

**Two characters** live in the game, not the engine: a `Swap` behaviour moves the camera's
`follow` and which node reads the input. The engine only needs `follow` to change while playing,
which the test of R2 checks.

**io-world**: a script reads `Demoivania.unity` from io-world's history once, as proposal 7's did
for Godot. Its pieces become placed scenes (9), its bricks a tileset (10), and its music Ogg (12).

## Tasks

To be planned once proposals 9 to 12 are built: they change what this one needs.

## Result

Not built yet.
