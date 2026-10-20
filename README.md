<p align="center"><img alt="" src="docs/images/logo.svg" width="120"></p>

<div align="center">

# Go Engine

**A game engine in Go, with an editor**, for 2D, 3D and isometric games on the desktop, Android
phones and the Nintendo 3DS. A game is an ordinary Go module: the editor arranges its scenes, and
its code moves them through a small standard library.

[![Go](https://img.shields.io/github/go-mod/go-version/akrck02/go-engine)](go.mod)
![Platforms](https://img.shields.io/badge/platforms-desktop%20%7C%20Android%20%7C%20Nintendo%203DS-informational)
![Status](https://img.shields.io/badge/status-in%20development-orange)
[![Licence](https://img.shields.io/github/license/akrck02/go-engine)](LICENSE)

</div>

> **Note:** written with an AI assistant under the author's direction, and pending thorough
> review. See the [AI disclosure](docs/1-overview.md#-ai-disclosure).

<p align="center">
  <img alt="The editor's Level workspace on a 2D street at night: a street lamp chosen, its reach drawn over the view, a window glowing on a house, and the inspector's Light section" src="docs/images/editor-lights.png" width="760">
</p>

**Status: in development.** Scenes in 2D and 3D, sprites, tiles and models, physics, lights,
behaviours in Go played in the editor, and builds for the desktop, Android and the 3DS all work.
Names, scene files and the editor can still change from one version to the next.

## Install

Go 1.25 or newer, a C compiler and OpenGL 4.1 (the editor's window):

```bash
go install github.com/akrck02/go-engine/cmd/go-engine@latest
```

## Make a game

```bash
go-engine new ~/games/harbour "Harbour lights"   # a project with a sample scene
go-engine ~/games/harbour                        # open it in the editor
```

The editor's **Build** workspace builds the game for the desktop, Android and the 3DS, and
installs the tools each one needs.

## Learn more

| | |
|---|---|
| [Overview](docs/1-overview.md) | What it is, its status, and where to start |
| [Get started](docs/2-getting-started.md) | Install it, and play a first scene in fifteen minutes |
| [Core ideas](docs/3-core-ideas.md) | Games, scenes, nodes, behaviours, physics and light |
| 🗺️ [The quest](docs/4-the-quest/0-the-quest.md) | A forest platformer, made in six levels |
| [Guides](docs/5-guides/1-a-2d-level-with-tiles.md) | Tiles, sprites, physics, lights, cameras, the 3DS, Android |
| [Reference](docs/6-reference/0-packages.md) | Every package, and the project's files |
| [Troubleshooting](docs/7-troubleshooting.md) | Common problems, and every error code |
| [Contributing](CONTRIBUTING.md) | The code's layout, tests, proposals and style |

## Licence

MIT, in [LICENSE](LICENSE). A built game carries other software too; its notices are written
next to it.

## AI disclosure

Written by an AI assistant (Claude, by Anthropic) under the author's direction, pending thorough
review; details in [the overview](docs/1-overview.md#-ai-disclosure).
