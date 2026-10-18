# 3. Undo, several nodes, sharp pixels: an editor to trust

> Status: done · Author's approval: proposal ☑ plan ☑ (2026-10-18, "haz todas pero además haz
> que aparezca la versión en la ventana de editor y logs": approved together)

## Why

Proposal 2 left four things out on purpose: undo, choosing several nodes, pixel art drawn sharp,
and a 3D Level like the 2D one. Without undo one wrong click loses work; one node at a time makes
ten coins ten jobs; pixel art drawn blurred misleads about how the game looks. And nobody can tell
which Go Engine an editor or a game was built with.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | The editor's window, its top bar and the logs of the editor, the game and the go-engine command say which Go Engine they run: its release and commit, and whether it is a local copy | test `the version says which Go Engine it is`; test `a local Go Engine gives its commit to the build` |
| R2 | Every change to the scene can be undone and redone (Ctrl+Z, Ctrl+Shift+Z, and buttons); a drag or a stretch of typing is one step | tests `a change is undone and redone`, `a drag is one step to undo` |
| R3 | Several nodes can be chosen (Shift in the tree and the view); dragging moves them together, Delete and Duplicate act on all | tests `shift adds a node to the choice and takes it out`, `deleting and duplicating act on every node chosen`, `dragging moves every node chosen` |
| R4 | The mouse in the 2D view does what the tools say: drags, right-button erasing, the hand, fill | tests that press, move and let go the buttons over the view |
| R5 | Pixel art is drawn sharp: pictures of a 2D scene are scaled to the nearest pixel, on the desktop and the 3DS | a test of go-renderer's sampling; captures |
| R6 | Level's tool rail is the same in 3D: Select and Hand, the others greyed | capture |

## Out of scope

- Undoing changes to sprite and tileset files (Art saves them itself).
- Choosing several nodes by dragging a box.

## Design

**Version** (done first): `engine.Version()` reads the program's build information; for a local
copy (a `go.work` or a folder `replace`) go-engine passes the commit with `-ldflags -X`.

**Undo**: the editor keeps the scene's description after each change (a list and a place in it);
undo builds the scene again from the one before, keeping the chosen nodes by path. Changes within
half a second of each other are one step, so a drag or typing a number is undone at once.

**Several nodes**: `Editor.Chosen` holds the nodes besides `Selected`, the one the inspector shows.

**Sharp pixels**: go-renderer's 2D pictures get a filter setting; a 2D scene draws its sprites
and tiles with the nearest pixel. That change belongs to go-renderer, on its `feature/3ds`
branch, committed only with the author's say.

## Tasks

- [x] T1 — the version, in the editor's window, its top bar and the logs · `engine/`, `project/`,
  `editor/`, `cmd/` · R1 · tests: `the version says which Go Engine it is`, `a local Go Engine
  gives its commit to the build`
- [x] T2 — undo and redo · `editor/` · R2 · tests: `a change is undone and redone`, `a drag is one
  step to undo`
- [x] T3 — several nodes chosen · `editor/` · R3 · tests: `shift adds a node to the choice and
  takes it out`, `deleting and duplicating act on every node chosen`, `dragging moves every node
  chosen`
- [x] T4 — the 2D view's mouse, tested · `editor/` · R4
- [x] T5 — Level's tool rail in 3D · `editor/` · R6
- [x] T6 — sharp pixel art · go-renderer `draw2d`, `opengl`, `ctr`; `engine/` · R5

## Result

**Built.** The version of Go Engine in the editor's title and top bar and in the logs of the
editor, the game and the go-engine command, with the commit of a local copy and whether it has
changes; undo and redo of every change to the scene, a drag or a number typed being one step;
several nodes chosen with Shift, dragged, deleted and duplicated together; tests that press, move
and let go the mouse over the 2D view; Level's tool rail in 3D (Select and Hand); pixel art drawn
sharp, through go-renderer's `draw2d.Image.Sharp` (feature/3ds 3b08cd9, committed with the
author's say), which the engine sets on every sprite, frame and tile.

**Checked.** Every task's tests; sharp pixels compared before and after in captures, on the desktop
and in the 3DS emulator.

**Not done.** Metal draws sharp pictures blended, as before: it is hand-written and was not
touched. Changes to sprite and tileset files are not undone (Art saves them itself). Choosing
nodes with a box was out of scope.

**For a person.** Undo after real edits, Shift-clicks and drags with a real mouse.
