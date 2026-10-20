# 11. Sprites made of tiles, in Art

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

io-world's jump pad and rail are rows of a tile picture. Art cannot put tiles together, so a
script drew them into new pictures (proposal 7). Those pictures are copies: changing the tileset
leaves them behind. An artist who draws in tiles should build bigger things from them in the
editor, and see them follow the tileset.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | Art makes a sprite from a tileset: a small grid painted with its tiles, as a tile map is | test `a sprite made of tiles has its tiles' pixels`; capture |
| R2 | Changing the tileset's picture changes every sprite made from it, the next time it loads | test `a sprite made of tiles follows its tileset` |
| R3 | Such a sprite is used as any other: animations, lights, bodies | test `a sprite made of tiles plays its animations` |
| R4 | It is put together once, when it loads, and drawn as one picture | the allocation test |

## Out of scope

- Painting pixels: Art cuts and composes pictures; drawing them is another program's job.
- Turned tiles in a sprite, until proposal 10 brings them.

## Design

A sprite file gains `tiles`, which takes the place of `picture`. It holds a tileset, a number of
columns and rows, and a tile number per cell, as a tile map's. The engine puts the picture
together when the sprite loads, through the asset cache, and the rest of the sprite (frames,
animations) cuts it as if it were a file.

Art gets **New sprite from tiles**: the tileset's palette on the right, a grid in the middle to
paint, and the usual animation strip under it. Its grid's size is the sprite's frames: a pad of
three tiles in two frames is three columns by two rows.

**Alternative not taken**: saving the composed picture as a PNG. It is simpler to load, but it
stops following the tileset (R2), which is the reason for this change.

## Tasks

- [ ] T1 — `scenefile`: a sprite's `tiles`, read, checked and written · R1 · test: `a sprite made
  of tiles comes back as written`
- [ ] T2 — `engine`: the picture put together once at load · R1, R2, R4 · tests of R1, R2
- [ ] T3 — Art: New sprite from tiles, painting its grid · R1 · capture
- [ ] T4 — Io World: the pad and the rail as sprites made of tiles · R3

## Result

Not built yet.
