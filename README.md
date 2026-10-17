# go-engine

> **Note:** this project was written with an AI assistant and is pending thorough review. See the
> [AI disclosure](#ai-disclosure).

A small game maker in Go: an editor where a game is made scene by scene, by painting things on a
grid and saying what each kind of thing does, played on the spot and exported as a program for the
desktop or the **Nintendo 3DS**. It is built on the author's own libraries:
[go-renderer](https://github.com/akrck02/go-renderer) draws,
[go-ui](https://github.com/akrck02/go-ui) is the editor's window and the games' interface, and
[tinygo-3ds](https://github.com/akrck02/tinygo-3ds) builds for the console, as
[go-3ds](https://github.com/akrck02/go-3ds) does.

<p align="center">
  <img alt="The editor: scenes and kinds on the left, a scene of walls, crates, oil and a lantern in the middle, the kind's inspector on the right" src="docs/images/editor.png" width="760">
</p>

```bash
go run ./cmd/go-engine            # edit game.json here (a new one starts from the sample)
go run ./cmd/go-engine my.json
go run ./cmd/play my.json         # play a game
go test ./...
```

**Status: a minimum viable product** (October 2026): grid games of one player, walls, things to
push, to collect, to avoid and a goal; scenes played in order. A larger design comes next.
Everything is in [docs/](docs/README.md).

## License

MIT, in [LICENSE](LICENSE). An exported program carries other software too; its notices are
written next to it ([export/notices-desktop.md](export/notices-desktop.md),
[export/notices-3ds.md](export/notices-3ds.md)).

## AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction.** Every commit made this way carries a
`Co-Authored-By: Claude` line.

**Status: pending thorough review.** The code compiles and its tests pass, but I have not yet
reviewed it line by line. Treat it as a working draft, not as code I vouch for.
