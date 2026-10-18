# go-engine

A game engine in Go with an editor (go-ui over go-renderer): games are Go modules whose scenes the
editor arranges and whose code uses the engine's standard library; they build for the desktop,
Android and the Nintendo 3DS. Start with `docs/README.md`.

## Commands

```bash
go vet ./...
go test ./...                                   # no display needed
go run ./cmd/go-engine new <folder> [name]      # a new project
go run ./cmd/go-engine [folder]                 # the editor; needs a display and OpenGL 4.1
go run ./cmd/go-engine -capture editor.png <folder>   # one frame to PNG
go run ./cmd/grid-maker [game.json]             # the grid prototype
```

## Rules

**Follow `docs/style-guide.md`**: code a junior can follow. In short: English always; whole-word
names (only `err`, `x`, `y`, `z`, `ui` are short); one job per function, about 40 lines at most;
early returns, no `else` after a return; every error has a code from `failure`; packages divided
by meaning, one module. `internal/style` tests what a program can check.

- **Generic only.** Never put names, places or settings from a user's own project in code,
  comments, tests, examples, docs or commit messages. The sample is made up (a lighthouse keeper).
- **Tests with every change.** `go vet ./...` and `go test ./...` must pass before committing.
- **What games use stays portable**: `engine`, `space`, `anim`, `scenes` and `shapes` build under
  TinyGo for the 3DS; no cgo, no `os` file system (games read their embedded files).
- **Use the author's libraries.** Drawing, windows and input go through go-renderer and go-ui;
  what they lack is added there (on a `feature/` branch, with the author's permission), not here.
- **Docs live in `docs/`**, updated in the same commit as the behaviour they describe.
- **AI disclosure.** Commits written with an AI assistant end with a `Co-Authored-By: Claude …`
  line, and the README's AI disclosure stays true.
- **Git.** AI work goes on `feature/` branches; `main` is for code the author has reviewed. Never
  force-push, rewrite history or delete branches without permission. Do not change the global git
  config.
