# 2. Editor workspaces: one task at a time

> Status: building · Author's approval: proposal ☑ plan ☑ (2026-10-18, "sigue hasta tener todo
> implementado": approved together, while away)

## Why

The editor is one screen that does six jobs at once: with a coin chosen, the inspector asks for
its place in the level, the cut and timing of its picture, and its code, while the top bar offers
testing, building and the project. A level designer scrolls past animation settings on every
coin; an artist has no screen of their own; a solo developer is offered every job at every
moment. DaVinci Resolve answers this with pages, one screen per stage of the work. The study
behind this change: "Editor workspaces" (a private page of the author's).

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A bar along the bottom switches between six workspaces (Project, Art, Level, Code, Play, Build), each with an icon and a word, and the keys 1–6 when no field is being typed in; the open scene and the chosen node stay the same across them | test `switching workspaces keeps the scene and the choice`; capture of each workspace |
| R2 | Pictures become sprites and tilesets of their own: a `.sprite.json` (picture, frames, animations) and a `.tileset.json` (picture, tile size, solid tiles) that nodes and tile maps point to, so ten coins share one spin; scenes written before still open | tests `a node plays the animations of the sprite it points to`, `a tile map takes its solid tiles from its tileset`, `a sheet of its own wins over its sprite's file` |
| R3 | Art: the project's sprites and tilesets in a list, the picture large with its cut drawn over it, frames chosen by clicking, animations in a strip with a preview, tiles made solid by clicking; a picture becomes a sprite or a tileset with one action | tests `a picture becomes a sprite with its frames`, `clicked frames make an animation`, `a clicked tile turns solid and back`, `an edited sprite is saved and read back` |
| R4 | Level: a tool rail beside the view (Select, Paint, Erase, Fill, Hand) with keys V, B, E, G, H; bodies, tile cells and camera bounds drawn on demand; the palette shows the tileset as drawn; the inspector only places (2D transform, sprite and first animation, body, behaviour values) | tests `the paint tool paints and the select tool chooses`, `fill paints the region of the tile clicked`, `the palette keeps the tileset's columns`; captures |
| R5 | Play: the game at the shape of a target (16:9 or the 3DS's top screen), with pause, a step of one frame and a restart; the chosen node watched live (position, velocity, on the floor, animation, behaviour fields); the game's log under it | tests `a paused game stands still and steps one frame`, `restart starts the scene as written`, `the watch shows a behaviour's fields`; captures |
| R6 | Project: the assets as a grid of thumbnails with folders and filters, a preview that says where each is used, scenes as items that open in Level, and the project's settings | tests `an asset's uses are found in every scene`, `settings are saved`; capture |
| R7 | Code: the behaviours of the project with their comment, fields and methods, the nodes that use each, behaviours used but missing from the code, and Rebuild and Open in code editor with their output | tests `a behaviour's uses are listed by scene`, `a missing behaviour is reported`; capture |
| R8 | Build: one card per target (this computer, the 3DS) with its last result, size and time, its actions and its own log | test `a build's result is kept per target`; capture |
| R9 | Icons say what they do: one meaning per icon, a word beside any icon not known by everyone | test `no icon stands for two things` |
| R10 | The demo uses the new assets: its coins share one sprite and its levels one tileset | the demo's tests pass; capture of its coin in Art |

## Out of scope

- A code editor inside the editor (Zed is the project's code editor).
- Undo and redo, choosing several nodes, sharp pixel art (a go-renderer filter): findings of the
  first study that belong to their own changes.
- 3D scenes in Art (models are not cut or animated here); 3D keeps today's Level view.
- Documentation chapters: v0.0.1 is not documented yet.

## Design

**Workspaces in the editor.** `Editor.Workspace` (a small type with the six values) and
`SetWorkspace`; entering Play plays the scene, leaving it stops. `View.Build` draws the top bar
(project, scene, saved state), the chosen workspace's panels, and the workspace bar with the status
line. Each workspace is its own file of panels (`workspace_project.go`, `workspace_art.go`, …);
what is shared (tree, inspector sections) stays where it is. Code says "workspace" because
"mode" already means 2D or 3D.

**Sprites and tilesets as assets** (`scenefile`). `Sprite{Picture, Columns, Rows, Animations,
Start}` in `*.sprite.json` and `Tileset{Picture, TileSize, Solid}` in `*.tileset.json`, read and
checked like scenes. A node's `sprite` may name a picture (as today) or a sprite file; a tile map's
`tileset` may name a picture (with its own tile size and solid tiles, as today) or a tileset file.
The engine resolves them when a scene is built, once, through its asset cache; an inline `sheet`
still works and wins, so old scenes play. Alternatives not taken: moving every sheet into files at
once (breaks scenes) and per-scene resources (the coins of two levels would still disagree).

**Art** edits these files through `editor/art.go` (no interface in it, so it is tested): make a
sprite or tileset from a picture, set the cut, add, rename and remove animations, add a clicked
frame, toggle a solid tile, save. Its panels: list, canvas (the picture with its cut, frames
numbered), cut and collision on the right, the animation strip with a preview underneath.

**Level** gets `Tool` (select, paint, erase, fill, hand) instead of the paint checkbox, an overlay
switch for bodies and camera bounds, the tile map's cell grid while it is chosen, a flood fill, and
a palette in the tileset's own columns. The inspector loses the sheet and animations (they are in
Art: it shows the sprite, a "Starts with" choice and a button to open it in Art), the tile size and
solid tiles (in the tileset), Go types in behaviour fields, and the 3D/2D tab (moved to Project's
scene settings).

**Play** uses the editor's play-in-the-view, with `Pause`, `StepFrame` and `Restart`, a target
shape, a watch of the chosen node (exported behaviour fields read by reflection), and a log fed by
an `slog` handler the editor installs while playing.

**Project**, **Code** and **Build** are lists and details over what the editor already knows:
`Project.AssetTree`, `Project.Behaviours`, the scenes' descriptions (to find uses), and the builds
it runs (kept per target).

**Icons** are named constants in `editor/icons.go`, one per meaning, with a test that no two
meanings share an icon.

## Tasks

- [x] T1 — sprite and tileset files: read, check, write; engine resolves them; old inline sheets
  still work · `scenefile/`, `engine/` · R2 · tests: `a node plays the animations of the sprite it
  points to`, `a tile map takes its solid tiles from its tileset`, `a sheet of its own wins over
  its sprite's file`
- [x] T2 — workspaces and their bar; top bar trimmed; keys 1–6; icons as constants · `editor/` ·
  R1, R9 · tests: `switching workspaces keeps the scene and the choice`, `no icon stands for two
  things`
- [x] T3 — Level: tool rail, fill, overlays (bodies, camera bounds, cell grid), palette as drawn,
  inspector that only places · `editor/` · R4 · tests: `the paint tool paints and the select tool
  chooses`, `fill paints the region of the tile clicked`, `the palette keeps the tileset's columns`
- [x] T4 — Play: pause, step, restart, target shape, watch, log · `editor/`, `engine/` · R5 ·
  tests: `a paused game stands still and steps one frame`, `restart starts the scene as written`,
  `the watch shows a behaviour's fields`
- [ ] T5 — Art: sprites and tilesets, canvas with the cut, animation strip and preview, solid
  tiles · `editor/` · R3 · tests: `a picture becomes a sprite with its frames`, `clicked frames
  make an animation`, `a clicked tile turns solid and back`, `an edited sprite is saved and read
  back`
- [ ] T6 — Project: thumbnails, folders, filters, preview with uses, scenes, settings · `editor/`,
  `project/` · R6 · tests: `an asset's uses are found in every scene`, `settings are saved`
- [ ] T7 — Code: behaviours, fields, uses, missing ones, rebuild and open with output · `editor/`
  · R7 · tests: `a behaviour's uses are listed by scene`, `a missing behaviour is reported`
- [ ] T8 — Build: target cards, results kept per target, logs · `editor/`, `project/` · R8 · test:
  `a build's result is kept per target`
- [ ] T9 — the demo on sprites and a tileset; captures of every workspace · `examples/platformer/`
  · R10 · the demo's tests

## Result

Not built yet.
