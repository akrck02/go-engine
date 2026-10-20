# 🗂️ The project's files

Every file the editor writes is JSON a person can read, and kind to version control. This page
lists each key; the editor sets them all, so you only need it to read a file or fix one by hand.

## `engine.json`

| Key | Type | What it is |
|---|---|---|
| `name` | text | The game's name |
| `startScene` | text | The scene the game opens, as `scenes/main.scene.json` |
| `codeEditor` | text | The command that opens the project, as `zed` or `code`; empty tries `$VISUAL` and `$EDITOR` |
| `androidId` | text | The Android application id, as `io.example.forest` |
| `consoleAddress` | text | The 3DS's address to send the game to; empty looks for it on the network |
| `sceneMenu` | yes/no | Lists the scenes on the 3DS's bottom screen |

## A scene: `scenes/*.scene.json`

| Key | Type | What it is |
|---|---|---|
| `name` | text | The scene's name |
| `mode` | `3d` or `2d` | 3D by default |
| `sky` | colour | The background, `#rrggbb`; empty is a daylight sky |
| `pixelsPerMetre` | number | In 2D, a sprite's pixels in a metre; 0 is 32 |
| `ambient` | ambient | The light that reaches everything (below) |
| `nodes` | nodes | The top of the tree of nodes |

## A node

| Key | Type | What it is |
|---|---|---|
| `name` | text | Unique among its siblings, with no slash |
| `position`, `rotation`, `scale` | 3 numbers | Relative to its parent: metres, degrees (X tilts, Y turns, Z rolls), times its size |
| `hidden` | yes/no | Hidden, and its children with it |
| `shape`, `colour` | text | `cube`, `sphere`, `plane` or `cylinder`, in a colour |
| `model` | text | A model, as `assets/models/house.glb` |
| `sprite` | text | In 2D, a picture or a sprite file |
| `animation` | text | The animation its sprite starts with |
| `sheet` | sheet | A sheet cut on the node itself (a sprite file is better) |
| `tiles` | tile map | In 2D, a map of tiles |
| `camera` | camera | Makes it a camera |
| `body` | body | In 2D, a body in the physics |
| `light` | light | Makes it a light |
| `behaviours` | list | Each `{"name": "Player", "fields": {"speed": 3.2}}` |
| `children` | nodes | Its children |

| Part | Keys |
|---|---|
| camera | `projection` (`perspective`, `orthographic`, `isometric`), `fieldOfView` (degrees; 0 is 45), `height` and `width` (metres seen), `main`, `follow` (a node's path), `smoothing` (seconds), `bounds` (lowest X, lowest Y, highest X, highest Y), `ceiling` |
| body | `kind` (`solid`, `moving`, `trigger`, `pushable`, `platform`), `size` (width and height; 0 is the node's own), `gravityScale` (1 if absent), `pull` (a trigger's gravity, across and up), `ignoresZones` |
| light | `kind` (`point`, `sprite`), `colour`, `intensity` (1 if absent, at most 2), `inner` and `outer` (metres; outer is 3 if absent) |
| ambient | `colour`, `intensity` |
| tiles | `tileset` (a tileset file, or a picture), `tileSize` (for a picture), `columns`, `rows`, `tiles` (a number per cell, row by row from the top; -1 is empty), `solid` (for a picture) |

## A sprite: `*.sprite.json`

| Key | What it is |
|---|---|
| `picture` | The sheet's picture |
| `columns`, `rows` | Frames across and down |
| `animations` | Each with a `name`, its `frames`, `framesPerSecond` and whether it should `loop` |
| `start` | The animation playing first |

## A tileset: `*.tileset.json`

| Key | What it is |
|---|---|
| `picture` | The tiles' picture |
| `tileSize` | Pixels a tile, a side |
| `solid` | The tile numbers bodies stand on, counted row by row from the top-left, from 0 |

---

[← The reference](0-packages.md)
