# go-engine

A game maker: an editor (go-ui over go-renderer) that makes grid games and exports them for the
desktop and the Nintendo 3DS. Start with `docs/README.md`.

## Commands

```bash
go vet ./...
go test ./...                                   # no display needed
go run ./cmd/go-engine [game.json]              # the editor; needs a display and OpenGL 4.1
go run ./cmd/go-engine -capture editor.png      # one frame to PNG
go run ./cmd/play [game.json]
```

## Rules

- **Generic only.** Never put names, places or settings from a user's own project in code,
  comments, tests, examples, docs or commit messages. The sample is made up (a lighthouse keeper).
- **Readable code.** Full-word names, small functions, early returns, comments that say *why*.
- **Tests with every change.** `go vet ./...` and `go test ./...` must pass before committing.
- **The game package stays portable**: pure Go over `draw2d`, no window, no cgo, so it runs in
  tests, on the desktop and under TinyGo on the 3DS.
- **Use the author's libraries.** Drawing, windows and input go through go-renderer and go-ui;
  what they lack is added there (on a `feature/` branch, with the author's permission), not here.
- **Docs live in `docs/`**, updated in the same commit as the behaviour they describe.
- **AI disclosure.** Commits written with an AI assistant end with a `Co-Authored-By: Claude …`
  line, and the README's AI disclosure stays true.
- **Git.** AI work goes on `feature/` branches; `main` is for code the author has reviewed. Never
  force-push, rewrite history or delete branches without permission. Do not change the global git
  config.
