# 📐 Style guide

go-engine's code must be easy to follow for a junior developer, or for anyone who can read code.
These rules hold everywhere in the repository, and in the code the editor writes into new
projects. A test (`internal/style`) checks the ones a program can check.

## 1. English, always

Code, comments, names, error codes, commit messages and documentation are in English.

## 2. Names say what things are

- **Whole words.** No acronyms and no abbreviations: `position`, not `pos`; `index`, not `i`;
  `context`, not `ctx`; `message`, not `msg`; `button`, not `btn`.
- **Exceptions**, because they are what everyone calls them:

  | Name | Why |
  |---|---|
  | `err` | the error a call returned, as all Go code names it |
  | `x`, `y`, `z` | the axes' own names, for a coordinate |
  | `ui` | the name of the go-ui package when it is imported |

- **Short English words are words**, not abbreviations: `at`, `by`, `to`, `up`, `is` are fine.
- **Tests** call their `*testing.T` `test`, not `t`.
- **No mathematical shorthand** a junior would have to look up: `Vector`, not `Vec3`; `Rotation`,
  not `Quat`; `Toward`, not `Lerp`.
- **Say the unit when there is one:** `seconds`, `pixelsPerMetre`, `degrees`.
- **Booleans read as a question:** `visible`, `isCamera`, `found`, `playing`.
- **Functions are verbs, or what they give back:** `LoadScene`, `Bounds`, `MainCamera`.
- **Don't repeat the package:** `space.Transform`, not `space.SpaceTransform`.

```go
// No
func upd(n *Node, dt float64) { n.T.Pos.X += n.V * dt }

// Yes
func moveForward(node *Node, seconds float64) {
	node.Transform.Position.X += node.Speed * seconds
}
```

## 3. One function, one job

- A function does **one thing**, and its name says which.
- If a function needs a comment to split it into parts, split it into functions instead.
- As a guide, a function fits on a screen: **about 40 lines at most**. A longer one is usually
  doing two jobs.
- A function that **finds** something does not also **change** it.

```go
// No: reads, checks and builds in one place
func Load(files fs.FS, file string) (*Scene, error) { /* 80 lines */ }

// Yes: each step has a name
func Load(files fs.FS, file string) (*Scene, error) {
	data, err := fs.ReadFile(files, file)
	if err != nil {
		return nil, failure.Wrap(SceneNotFound, file, err)
	}
	description, err := scenefile.Read(data)
	if err != nil {
		return nil, failure.Wrap(SceneUnreadable, file, err)
	}
	return Build(files, description)
}
```

## 4. Early returns

Handle the errors and the special cases first, and **return**. The normal path then runs down the
left margin, without `else` and without nesting.

```go
// No
func (node *Node) Shape() string {
	if node != nil {
		if node.shape != "" {
			return node.shape
		} else {
			return "none"
		}
	}
	return ""
}

// Yes
func (node *Node) Shape() string {
	if node == nil {
		return ""
	}
	if node.shape == "" {
		return "none"
	}
	return node.shape
}
```

## 5. Errors have codes

Every error go-engine returns carries a **code** from the `failure` package. A program checks the
code, never the error's text.

- **Each package declares its codes** as variables, near the top of an `errors.go` file:

  ```go
  // Error codes of the engine package.
  var (
  	SceneNotFound   = failure.Define("engine.scene-not-found", "the scene file does not exist")
  	SceneUnreadable = failure.Define("engine.scene-unreadable", "the scene file is not a scene")
  )
  ```

- **Names of codes:** the package, a dot, and the error in a few words joined by dashes.
- **Return the bare code** when there is nothing to add: it costs nothing.
- **Wrap** when there is detail worth knowing (which file, which name, the error underneath):
  `failure.Wrap(SceneNotFound, file, err)`.
- **Check** with `failure.Is(err, engine.SceneNotFound)` (or `errors.Is`, which does the same).
- **Performance:** a bare code is a pointer to a value made once, so returning it allocates
  nothing and comparing it compares two pointers. `Wrap` allocates once, so it belongs in
  loading, saving and building, **never in a frame's update**. Code that runs every frame does not
  make errors.

## 6. Packages and modules

A **package** and a **module** divide different things:

| | Package | Module |
|---|---|---|
| What it is | A folder of Go code with a name (`space`, `engine`) | A set of packages with a `go.mod`, versioned and published together |
| What dividing is for | **Meaning**: each concept in its place | **Publishing**: things that change at different paces |
| A unit of | Compiling and importing | Versions, releases and dependencies |

### Dividing packages, by meaning

1. **A package is a concept you can name.** Its name is a good prefix for what is in it:
   `space.Vector`, `animation.Tween`, `scenefile.Node`. If no such name fits, the boundary is
   wrong.
2. **The two-package test.** If using one almost always means importing the other, they are one
   package.
3. **Distinct ideas apart, even small ones.** A package of two files is fine if it is an idea of
   its own (`failure`).
4. **No `util`, `common`, `helpers`, `misc`, `api`, `types`.** They mean nothing, and gather
   dependencies.
5. **Dependencies point one way.** The basic concepts know nothing of the ones built on them:

   ```
   failure, space             the base: they import nothing of ours
     ↑
   animation, scenefile, shapes
     ↑
   engine                     what a game carries; builds for the 3DS
     ↑
   project                    desktop tools
     ↑
   editor  →  cmd/go-engine
   ```

   When something below needs something above (`engine` importing `editor`), the concept is in
   the wrong package. The editor's hook is the other way round for this reason: the editor
   registers itself with the engine, which never imports it.
6. **What a game carries stays portable.** The packages a game imports build with TinyGo for the
   3DS: no editor, no `go/parser`, no `os/exec` in them. What a package imports decides what goes
   into a `.3dsx`.
7. **What only go-engine uses goes in `internal/`.** It can change without breaking anyone's game.
8. **Inside a package, divide into files, not into more packages:** `node.go`, `scene.go`,
   `errors.go`, each a part of the concept.

**Package names** follow Go's: lower case, one word (two words written together when needed, as
`filepath` or `scenefile`), short and clear, with no abbreviation (our rule goes further than
Go's, which allows familiar ones as `fmt`), not the name of a common variable, and not repeated
in what the package holds (`http.Server`, not `http.HTTPServer`).

### Dividing modules, by publishing

Go's advice is **one repository, one module**, unless there is a strong reason: *"For all but
power users, you probably want to adopt the usual convention that one repo = one module"* (Russ
Cox, in Go's modules FAQ). More modules cost work: `go test ./...` no longer tests everything,
`replace` directives must be kept between them, and adding, removing or versioning one takes care.

A part becomes a module of its own when:

- it is **published and versioned apart**, used by other projects at their own pace;
- it has **very different or heavy dependencies** that those who use only the other part should
  not have to take;
- **different people** look after it.

Not for size alone: Go downloads through its module proxy, and a program only links the packages
it imports, whatever module they are in.

**Applied to the author's libraries:** go-renderer, go-ui, go-sound, tinygo-3ds and go-engine are
separate modules because each is published and reused on its own (go-ui serves without go-engine,
go-sound without go-renderer). **go-engine stays one module**: a game for the 3DS links `engine`
and what it imports, never `editor`, so splitting would add `replace` directives and versions to
keep in step for nothing. That would change if another project wanted only `space` and `animation`
without go-ui and GLFW, or if the editor came to be released at a pace of its own.

### The packages today

| Package | Concept | Who uses it |
|---|---|---|
| `failure` | Error codes | Every package |
| `space` | Positions, rotations, transforms, boxes | Games and the engine |
| `animation` | Values that move over time | Games |
| `scenefile` | The scene file format | The engine and the editor |
| `shapes` | Meshes of simple shapes | The engine |
| `engine` | The running game: nodes, behaviours, input, cameras | Games |
| `project` | A game's folder: creating, building, reading its code | The editor (desktop only) |
| `editor` | The editor's window | The game while it is being made (desktop only) |
| `internal/build` | Running Go and TinyGo | `project` |
| `internal/notices` | Third-party notices | `project` |
| `internal/style` | The tests of this guide | The repository |

Sources: [Package names](https://go.dev/blog/package-names) (the Go blog),
[Organizing a Go module](https://go.dev/doc/modules/layout) (go.dev),
[Go modules FAQ: should I have multiple modules in a single repository?](https://go.dev/wiki/Modules),
and [Google's Go style guide, best practices](https://google.github.io/styleguide/go/best-practices)
(package size, errors).

## 7. Comments

- A comment says **why**, or what a reader cannot see in the code. It does not repeat the code.
- Every exported name has a comment that starts with the name: `// Bounds is the box…`.
- A package starts with a comment saying what it is for, with an example when it helps.

## 8. Files

- A file holds one part of its package: `node.go`, `scene.go`, `errors.go`.
- No file of thousands of lines, and no swarm of tiny files either.

## 9. Tests

- Every change comes with tests, and `go vet ./...` and `go test ./...` pass before a commit.
- A test's name is a sentence about what must be true: `TestAChildIsPlacedInsideItsParent`.

## 10. Performance in a frame

What runs every frame (a behaviour's `Update`, drawing, input) must not allocate when it can
avoid it: reuse slices and lists, keep texts that do not change, and make no errors. See
[3ds.md in go-ui](https://github.com/akrck02/go-ui/blob/feature/3ds/docs/3ds.md) for why it
matters most on the console.
