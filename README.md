# go-engine

> **Note:** this project was written with an AI assistant and is pending thorough review. See the
> [AI disclosure](#ai-disclosure).

A game engine in Go, with an editor, for 2D, 3D and isometric games on the desktop and the
**Nintendo 3DS**. A game is an ordinary Go module: the editor arranges its scenes and imports its
models, and the game's code, written in Zed, Neovim, Visual Studio Code or any editor that speaks
Go, moves them through a small standard library (`engine`, `space`, `anim`). It is built on the
author's own libraries: [go-renderer](https://github.com/akrck02/go-renderer) draws,
[go-ui](https://github.com/akrck02/go-ui) is the editor's window, and
[tinygo-3ds](https://github.com/akrck02/tinygo-3ds) builds for the console.

<p align="center">
  <img alt="The editor: the scene's nodes on the left, a house imported from a .glb chosen in the 3D view, the inspector on the right" src="docs/images/editor.png" width="760">
</p>

```bash
go run ./cmd/go-engine new ~/games/harbour "Harbour lights"
go run ./cmd/go-engine ~/games/harbour
go test ./...
```

**Status:** early (October 2026): editable 3D and 2D scenes, `.glb` models and sprites from the
project's assets, the coordinates and scales, behaviours in Go played in the editor's view, and
builds for the desktop and the 3DS. Everything is in
[docs/](docs/README.md).

## License

MIT, in [LICENSE](LICENSE). A built game carries other software too; its notices are written
next to it (see [internal/notices](internal/notices)).

## AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction.** Every commit made this way carries a
`Co-Authored-By: Claude` line.

**Status: pending thorough review.** The code compiles and its tests pass, but I have not yet
reviewed it line by line. Treat it as a working draft, not as code I vouch for.
