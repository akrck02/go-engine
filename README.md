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
> review. See the [AI disclosure](docs/README.md#-ai-disclosure).

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
| [A project](docs/README.md#a-project) | What a game is made of: scenes, assets, code |
| [The editor](docs/README.md#the-editor) | Its workspaces, from Project to Build |
| [2D](docs/README.md#2d), [bodies](docs/README.md#bodies) and [lights](docs/README.md#lights) | Sprites, tiles, physics, and lighting in 2D and 3D |
| [The standard library](docs/README.md#the-standard-library) | Behaviours: the code that moves a game |
| [Proposals](docs/proposals/) | Why each part is the way it is |
| [Style guide](docs/style-guide.md) | How the code is written |

## Licence

MIT, in [LICENSE](LICENSE). A built game carries other software too; its notices are written
next to it.

## AI disclosure

Written by an AI assistant (Claude, by Anthropic) under the author's direction, pending thorough
review; details in [docs](docs/README.md#-ai-disclosure).
