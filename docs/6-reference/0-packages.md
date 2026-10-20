# 📚 The reference

Every public package of Go Engine, generated from its doc comments by `generate.py`, and the files
a project is made of. Look names up here; learn them in the [guides](../5-guides/1-a-2d-level-with-tiles.md).

## For a game's code

| Package | What it is for |
|---|---|
| [engine](1-engine.md) | The running game: nodes, scenes, behaviours, the context, input, cameras, `Run` |
| [space](2-space.md) | Vectors, rotations, angles, transforms and boxes, in metres |
| [animation](3-animation.md) | Tweens of numbers, vectors and rotations, with easings, loops and chains |
| [physics](4-physics.md) | The 2D physics: bodies, their kinds, gravity and zones |
| [scenefile](5-scenefile.md) | The scene, sprite and tileset files: read, write, check |
| [shapes](6-shapes.md) | The meshes of the shapes, a metre across |
| [failure](7-failure.md) | Error codes, which every error the engine returns carries |

## For tools

| Package | What it is for |
|---|---|
| [project](8-project.md) | A project's folder: create, import, build, find its behaviours |
| [editor](9-editor.md) | The editor, which a game installs in itself while it is made |

## The files

[The project's files](10-files.md): `engine.json`, scenes, sprites and tilesets, key by key.

---

[← Guides](../5-guides/8-the-3ds-bottom-screen.md) · [Troubleshooting →](../7-troubleshooting.md)
