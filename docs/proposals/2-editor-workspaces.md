# 2. Editor workspaces: one task at a time

> Status: done · Author's approval: proposal ☑ plan ☑ (2026-10-18, "sigue hasta tener todo
> implementado": approved together, while away)

## Why

The editor is one screen that does six jobs at once. With a coin chosen, the inspector asks for
its place in the level, the cut and timing of its picture, and its code. The top bar offers
testing, building and the project. A level designer scrolls past animation settings on every
coin. An artist has no screen of their own. A solo developer is offered every job at every
moment.

DaVinci Resolve answers this with pages, one screen per stage of the work. The study behind this
change is "Editor workspaces", a private page of the author's.

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
- Undo and redo, choosing several nodes, sharp pixel art (a go-renderer filter). The first study
  found them; they belong to their own changes.
- 3D scenes in Art (models are not cut or animated here). 3D keeps today's Level view.
- Documentation chapters: v0.0.1 is not documented yet.

## Design

**Workspaces in the editor.** `Editor.Workspace` is a small type with the six values, set with
`SetWorkspace`. Entering Play plays the scene, and leaving it stops. `View.Build` draws three
things: the top bar (project, scene, saved state), the chosen workspace's panels, and the
workspace bar with the status line.

Each workspace is its own file of panels (`workspace_project.go`, `workspace_art.go`, …). What is
shared (tree, inspector sections) stays where it is. Code says "workspace" because "mode" already
means 2D or 3D.

**Sprites and tilesets as assets** (`scenefile`). They are read and checked like scenes:

- `Sprite{Picture, Columns, Rows, Animations, Start}` in `*.sprite.json`.
- `Tileset{Picture, TileSize, Solid}` in `*.tileset.json`.

A node's `sprite` may name a picture (as today) or a sprite file. A tile map's `tileset` may name
a picture (with its own tile size and solid tiles, as today) or a tileset file. The engine
resolves them once, when a scene is built, through its asset cache. An inline `sheet` still works
and wins, so old scenes play.

Alternatives not taken:

- Moving every sheet into files at once: it breaks scenes.
- Resources per scene: the coins of two levels would still disagree.

**Art** edits these files through `editor/art.go`. It has no interface in it, so it is tested. It
makes a sprite or tileset from a picture, sets the cut, and adds, renames and removes animations.
It adds a clicked frame, toggles a solid tile, and saves. Its panels:

- the list;
- the canvas: the picture with its cut, frames numbered;
- the cut and collision, on the right;
- the animation strip with a preview, underneath.

**Level** gets a `Tool` (select, paint, erase, fill, hand) instead of the paint checkbox. It also
gets an overlay switch for bodies and camera bounds, and the tile map's cell grid while it is
chosen. A flood fill and a palette in the tileset's own columns join them.

The inspector loses what moved elsewhere:

- the sheet and animations, now in Art (it shows the sprite, a "Starts with" choice and a button
  to open it in Art);
- the tile size and solid tiles, now in the tileset;
- Go types in behaviour fields;
- the 3D/2D tab, moved to Project's scene settings.

**Play** uses the editor's play-in-the-view, with `Pause`, `StepFrame` and `Restart`. It adds a
target shape and a watch of the chosen node, whose exported behaviour fields are read by
reflection. A log is fed by an `slog` handler the editor installs while playing.

**Project**, **Code** and **Build** are lists and details over what the editor already knows:
`Project.AssetTree`, `Project.Behaviours`, the scenes' descriptions (to find uses), and the builds
it runs (kept per target).

**Icons** are named constants in `editor/icons.go`, one per meaning. A test checks that no two
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
- [x] T5 — Art: sprites and tilesets, canvas with the cut, animation strip and preview, solid
  tiles · `editor/` · R3 · tests: `a picture becomes a sprite with its frames`, `clicked frames
  make an animation`, `a clicked tile turns solid and back`, `an edited sprite is saved and read
  back`
- [x] T6 — Project: thumbnails, folders, filters, preview with uses, scenes, settings · `editor/`,
  `project/` · R6 · tests: `an asset's uses are found in every scene`, `settings are saved`
- [x] T7 — Code: behaviours, fields, uses, missing ones, rebuild and open with output · `editor/`
  · R7 · tests: `a behaviour's uses are listed by scene`, `a missing behaviour is reported`
- [x] T8 — Build: target cards, results kept per target, logs · `editor/`, `project/` · R8 · test:
  `a build's result is kept per target`
- [x] T9 — the demo on sprites and a tileset; captures of every workspace · `examples/platformer/`
  · R10 · the demo's tests

## Result

**Built.** Sprites and tilesets are files of their own. The engine reads them once, and nodes and
tile maps point to them. A node may name the animation it starts with, so nodes that share a
sprite start differently.

The editor is six workspaces along a bottom bar (keys 1 to 6, in Play too):

- Level: a tool rail, fill, overlays, a palette in the tileset's columns, and an inspector that
  only places (with "Starts with").
- Art: the picture large with its cut numbered, frames clicked into animations, a preview, and
  solid tiles.
- Play: pause, a step, a restart, a screen's shape, a live watch of the scene playing, and the log.
- Project: thumbnails, filters, uses and settings.
- Code: the behaviours, their fields and methods, their uses, the missing ones, and the output.
- Build: a card and a log per target.

Icons are constants with one meaning each. The demo's coins share one sprite file, and its levels
one tileset file. It plays on the 3DS from them.

**Checked.** Every task's tests pass, with the order shuffled too. Each workspace was captured
with the demo open, on the author's private progress page (docs are not written for v0.0.1). The
demo was played in Azahar.

**Review.** The reviewer's findings were fixed:

- The play log gives the log package back, and is only taken when a game starts.
- Opening a scene or an Art file over unsaved changes is refused.
- A shared sprite is never written over, and builds save first.
- Only Level and Play take the view's input.
- Play's tree is the scene playing.
- Open in Art opens that node's sprite.
- A sprite file chosen drops the node's own sheet.
- The screen's shape is offered for 2D games only.
- Rebuilding's errors show in Code.
- Code reads the scenes at most every two seconds.
- Sprites, tilesets and clearing have icons of their own.
- New error codes have tests.

**Not done.** Undo and redo, choosing several nodes, and sharp pixel art (a go-renderer filter),
as the proposal said. In a 3D scene, Level keeps its old view.

**For a person.** Use the mouse: painting and filling, dragging in Select, clicking frames and
tiles in Art, the palette. The tests cover the model and one simulated click. Nobody has painted
a level by hand yet.
