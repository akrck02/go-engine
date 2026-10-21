# 10. Richer tile maps: several tilesets, turned tiles, tiles solid in part

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

Making io-world (proposal 7) found three limits of tile maps:

- A tile map takes one tileset. Godot's tile set mixes pictures, so one level became two tile
  maps, painted apart.
- Tiles cannot be mirrored or turned. io-world's flipped tiles are drawn unflipped, and an artist
  draws every corner four times.
- A solid tile is solid all over. io-world's attractor tiles are solid only in part, here whole
  squares.

Tiled and Godot both solve the first two the same way, and the third with a shape per tile.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A tile map can use several tilesets; the palette shows each, and any cell can hold a tile of any of them | test `a map paints tiles of two tilesets`; capture |
| R2 | A painted tile can be mirrored across, mirrored up and down, or turned a quarter, with keys while painting | test `a turned tile is drawn turned`; capture |
| R3 | A tileset's tile can be solid in part: a box inside it, in pixels, set by dragging in Art | test `a half-solid tile stops a body at its half` |
| R4 | Old tile maps open and play as before | the platformer's and Io World's tests |
| R5 | On the 3DS and on Android, turned tiles cost no more a frame than plain ones | the allocation test; Io World in Azahar |

## Out of scope

- Slopes and shapes other than boxes: the physics only has boxes (see proposal 1).
- Animated tiles (water that ripples). A sprite does it for now.
- Tiles painted by rules (autotiling): a later proposal.

## Design

**Several tilesets**, as Tiled does it. A tile map lists `tilesets`, in order. Tile numbers run
through them: the first tileset's tiles come first, then the second's. A map with one `tileset`
keeps working (R4). The palette has a tab per tileset.

**Turned tiles**: a tile map gains `turns`, a number per cell from 0 to 7. Its three bits are
mirror across, mirror up and down, and swap the axes, as Tiled's. Absent means none. When a map
loads, each turned tile used gets its own picture, made once, as mirrored sprites are. Drawing
stays one picture per cell (R5). In Level, **X**, **Y** and **Z** turn the tile being painted.

**Solid in part**: a tileset's `solid` stays a list of whole tiles. A new `boxes` maps a tile
number to a box in pixels inside it. Merging rows into one body only joins whole tiles; a part
box is a body of its own. Art draws the box over the tile, and dragging changes it. A turned
tile turns its box too.

**Alternatives not taken**:

- A layer per tileset (several tile maps under one node). It is what io-world does now, and
  painting means switching nodes all the time.
- Turn bits inside the tile number, as Tiled's files do. That is compact, but unreadable in a
  scene file a person opens.

## Tasks

- [ ] T1 — `scenefile`: `tilesets`, `turns` and a tileset's `boxes` · R1–R4 · test: `a map with
  two tilesets and turned tiles comes back as written`
- [ ] T2 — `engine`: tile numbers across tilesets; turned pictures made once · R1, R2, R5 · tests
  of R1, R2
- [ ] T3 — `engine`: part boxes as bodies, turned with their tile · R3 · test of R3
- [ ] T4 — editor: palette tabs, X/Y/Z while painting · R1, R2 · captures
- [ ] T5 — Art: drag a tile's box · R3 · capture
- [ ] T6 — Io World: one tile map, turned and part-solid tiles from Godot's level · R4, R5

## Result

Not built yet.
