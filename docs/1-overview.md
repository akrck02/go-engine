<p align="center"><img alt="" src="images/logo.svg" width="120"></p>

# Welcome to Go Engine 👋

> 🚧 **Go Engine is in development.** What these pages describe can change from one version to
> the next: names, scene files and the editor's layout included.

**Go Engine makes games in Go.** You arrange a game's scenes in its editor, and you write what moves
in them in Go. One game runs on a computer, on an Android phone and on a Nintendo 3DS.

<p align="center"><img alt="The editor's Level workspace: a scene with a ground, a coloured cube and a camera; the cube chosen, with its transform, shape, light and Spin behaviour in the inspector on the right" src="images/editor-level.png" width="760"></p>

A game is an ordinary Go module. Its scenes are files the editor writes, and its code is plain Go,
written in any editor you like (Zed, Neovim, Visual Studio Code, GoLand…). The engine's small
standard library gives that code the scene's nodes, the player's buttons and the time.

## Is it for me?

- ✅ You write Go, or want to, and want to **make a small game**: 2D, 3D or isometric.
- ✅ You want it to run on a **handheld console**: the 3DS, through homebrew.
- ✅ You want to **learn** how a game engine works inside: every part is readable Go.
- ❌ You need a finished engine to ship a commercial game: this one is young and changes often.

## Where to start

| If you… | Read | Time |
|---|---|---|
| want to **try** it | [Get started](2-getting-started.md): install it and play a first scene | 15 min |
| want to **understand** how it thinks | [Core ideas](3-core-ideas.md): games, scenes, nodes, behaviours | 15 min |
| want to **learn** step by step | 🗺️ [The quest](4-the-quest/0-the-quest.md): a forest platformer in six levels | 2 h |
| want to do **one thing** | [The guides](5-guides/1-a-2d-level-with-tiles.md): tiles, sprites, physics, lights, the 3DS | — |
| are **looking up** a name | [The reference](6-reference/0-packages.md): every package, and the scene files | — |
| are **stuck** | [Troubleshooting](7-troubleshooting.md): errors by their code | — |
| want to **change** the engine | [Contributing](8-contributing/1-contributing.md): the code's layout, tests, style | — |

## What it can do

| | |
|---|---|
| Scenes | 2D, 3D and isometric; shapes, `.glb` models, sprites with animations, tile maps |
| Physics | 2D bodies: solid, moving, pushable, platforms, triggers and gravity zones |
| Lights | Point lights in 2D and 3D, sprite lights, an ambient light |
| Code | Behaviours: Go types with `Start` and `Update`, played inside the editor |
| Platforms | Computers (tried on Linux; macOS and Windows still to try), Android and the Nintendo 3DS, built from the editor |

**Not yet:** sound, a game's own interface (menus, scores on screen), bodies that turn or slopes,
and textures from models. The list lives in [Core ideas](3-core-ideas.md#not-yet).

## How it is built

Go Engine stands on the author's other libraries: [Go Renderer](https://github.com/akrck02/go-renderer)
draws, [Go UI](https://github.com/akrck02/go-ui) is the editor's window, and
[TinyGo for the 3DS](https://github.com/akrck02/tinygo-3ds) builds for the console. Every design
decision has a proposal in [proposals](proposals/).

## 🤖 AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction.** Every commit made this way carries a
`Co-Authored-By: Claude` line.

**Status: pending thorough review.** The code compiles and its tests pass, but I have not yet
reviewed it line by line. Treat it as a working draft, not as code I vouch for.

---

[Get started →](2-getting-started.md)
