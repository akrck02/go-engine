# 9. Prefabs: a scene placed inside another

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

io-world's jump pad, rail and button are small scenes placed many times. In Go Engine each one is
written out in full, node by node, wherever it stands. Changing how a pad looks means changing
every pad. Proposal 7 found this the biggest missing piece, and any game with enemies, doors or
coins will hit it.

Godot calls them instanced scenes, Unity prefabs. We take Godot's idea, since our scenes are
already files: a node may *be* another scene.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A node can name a scene file; it then holds that scene's nodes, under its own name and transform | test `a placed scene brings its nodes under the node` |
| R2 | Changing the placed scene's file changes every place it is used, the next time the scene opens | test `every place shows the scene file as it is now` |
| R3 | A place can set the fields of its top node's behaviours, so two pads push with different strengths | test `a place's own fields win over the scene file's` |
| R4 | A scene that places itself, directly or through others, is refused with its code, not loaded forever | test `a scene that places itself says so` |
| R5 | In the editor: Project's **Place in the scene** places a scene file; Level shows the placed nodes greyed, and **Open** opens their file | captures; test `placing a scene file adds a node that names it` |
| R6 | **Make it local** turns a place into ordinary nodes, copied from the file, to change one alone | test `making a place local copies its nodes` |
| R7 | Placed scenes play the same on the desktop, the 3DS and Android, and cost nothing a frame beyond their nodes | the allocation test; Io World's pads as placed scenes on the 3DS |

## Out of scope

- Changing a placed node deep inside a place (only the top node's behaviour fields, R3). Godot
  allows any property; that can come later, once a game needs it.
- Placing scenes while the game runs (spawning). A behaviour can already add nodes in code.
- 3D models made of scenes: this is for any scene, but only 2D is tried here.

## Design

**Scene files** (`scenefile`): `Node` gains `scene`, a scene file in the project, and `fields`, a
map from behaviour name to field values. A node with `scene` has no children of its own in the
file. Reading a scene stays as it is: placing happens when the engine builds it.

**Building** (`engine`): `Build` reads each placed file once per build, through the asset cache,
and adds its nodes under the place. The place's name and transform are its own; the file's top
node's transform is applied under it. A stack of the files being built finds a scene that places
itself (`engine.scene-places-itself`). The place's `fields` are set after the file's.

Each place remembers its file (`Node.Placed`), so the editor can grey its nodes and save the
place, not its copied nodes.

**Editor**: Project's **Place in the scene** accepts scene files. Level draws a placed node's
children greyed, with a mark in the tree, and the inspector shows **Open** and **Make it local**.
The top node's behaviours show their fields, which write into the place's `fields`.

**Io World** turns its pads, rails and buttons into placed scenes, as the check of R7.

**Alternatives not taken**:

- Copying the file's nodes into the scene when placed (a template). It is simpler, but R2 is the
  whole point.
- Overrides of any property by node path, as Godot does. They are powerful but hard to show in an
  inspector, and no game here needs them yet.

## Tasks

- [ ] T1 — `scenefile`: `scene` and `fields` on a node, read, checked and written · R1, R3 ·
  test: `a scene with a placed scene comes back as written`
- [ ] T2 — `engine`: placed scenes built under their node, fields set, a loop refused · R1–R4 ·
  tests of R1–R4
- [ ] T3 — `engine`: placed scenes on every platform, without allocating a frame · R7 · the
  allocation test
- [ ] T4 — editor: place a scene file, grey its nodes, Open, the top node's fields · R5 · test and
  captures
- [ ] T5 — editor: Make it local · R6 · test `making a place local copies its nodes`
- [ ] T6 — Io World: pads, rails and buttons as placed scenes, played on the 3DS · R7
- [ ] T7 — docs: a guide, "Scenes inside scenes", and the quest's coins as one placed scene

## Result

Not built yet.
