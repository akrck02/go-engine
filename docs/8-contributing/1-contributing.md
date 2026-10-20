# 🛠️ Contributing

How the engine's code is laid out, how to build and test it, and how a change is proposed. Read the
[style guide](2-style-guide.md) next: it says how the code is written.

## The code

The repository is one Go module. Its packages are divided by meaning, and depend one way: the
game's library never imports the editor's.

| Package | What it holds |
|---|---|
| `space` | Vectors, rotations, transforms and boxes: the maths, in metres |
| `animation` | Tweens |
| `physics` | The 2D physics: bodies, gravity, zones, platforms |
| `scenefile` | The files: scenes, sprites, tilesets, read, written and checked |
| `shapes` | The shapes' meshes |
| `failure` | Error codes |
| `engine` | The running game: nodes, scenes, behaviours, input, drawing, `Run` |
| `project` | A project's folder: create, import, build, its behaviours read from its code |
| `editor` | The editor, which a game installs in itself while it is made |
| `cmd/go-engine` | The `go-engine` command: `new`, open, `-update` |
| `internal/build` | Building for each platform, and finding its tools |
| `internal/install` | Installing the tools from the editor |
| `internal/notices` | The licences a built game carries |
| `internal/style` | Tests of the style guide's rules a program can check |
| `examples/platformer` | The quest's game, a module of its own |

What a game imports (`engine`, `space`, `animation`, `physics`, `scenefile`, `shapes`) builds with
TinyGo for the 3DS: no cgo, and no files read from a disk. Drawing, windows and input go through
[Go Renderer](https://github.com/akrck02/go-renderer) and [Go UI](https://github.com/akrck02/go-ui);
what they lack is added there, not here.

## Build and test

```bash
go vet ./...
go test ./...                                   # no display needed
go run ./cmd/go-engine new /tmp/try "Try"        # a project to try changes on
go run ./cmd/go-engine /tmp/try                  # the editor: a display and OpenGL 4.1
go run ./cmd/go-engine -capture shot.png /tmp/try   # one frame of the editor to a picture
```

`-capture` takes the editor's flags too: `-select Node`, `-play`, `-workspace build`, `-3ds`,
`-dark`, `-width` and `-height`. The pictures in these pages were made with it.

After changing a public name or an error code, write the reference again:

```bash
python3 docs/6-reference/generate.py
```

## Proposing a change

A small fix is a change, a test and a commit. Anything larger goes through a **proposal**, one file
in [proposals](../proposals/): why, what (each requirement with a check), what is out of scope,
the design, the tasks, and at the end the result. The author approves the proposal, then the plan,
before anything is built. The proposals there are the record of why the engine is the way it is.

Every change in behaviour comes with its test, and its docs in the same commit. Work goes on a
`feature/…` branch; `main` holds what the author has reviewed.

## Commits

`area: what changed`, in the imperative: `physics: carry what stands on a platform`. Commits
written with an AI assistant end with a `Co-Authored-By:` line naming it.

---

[← Troubleshooting](../7-troubleshooting.md) · [Style guide →](2-style-guide.md)
