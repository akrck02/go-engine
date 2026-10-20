# 💡 Core ideas

A handful of ideas explain how everything in Go Engine behaves. Read this page once, and the
editor, the scene files and the code will rarely surprise you.

## A game is a Go module

There is no special language and no project format to learn. A game is a folder with a `go.mod`,
like any Go program. Its scenes and pictures are files in it, and `main.go` carries them inside the
game (Go's `embed`), so the finished program needs nothing beside it.

Why a plain module? Because then everything Go already has works: your code editor finds the
engine's functions and their comments, `go test` tests the game, and `go build` builds it.

## The editor lives inside your game

`go-engine` does not run your code itself. It **builds your game with the editor in it**
(`editor.go`) and starts that. So the editor knows your behaviours because they are compiled into
it, and **Play** plays your real code, frame by frame, inside its view.

The price is small: after changing the code, **Rebuild** (in Code) builds the game again and reopens
the editor where you were. Builds for players leave the editor out (`-tags release`), and so does
the 3DS.

## Scenes are trees of nodes

A **scene** is a file, `scenes/*.scene.json`, that the editor writes and you can read. It holds a
tree of **nodes**, as a family tree holds people. A node is a thing in the scene, and what it is
comes from its parts:

| Part | Makes the node… |
|---|---|
| a shape or a model | something solid in 3D: a cube, a house from a `.glb` |
| a sprite | a picture in 2D, perhaps animated |
| a tile map | a 2D level painted with tiles |
| a camera | the eye the scene is seen through |
| a body | something the 2D physics moves or stops |
| a light | a lamp, or a glowing sign |
| behaviours | something your code moves |

Each node has a **transform**: where it is, how it is turned and how large it is, **relative to its
parent**. Move a ship and its sailors move with it. `node.World()` says where a node is in the world.

## Behaviours are your code

A **behaviour** is a Go type with methods the engine calls:

| Method | When |
|---|---|
| `Start(context)` | Once, before the scene's first frame |
| `Update(context)` | Every frame; `context.Delta` is the seconds since the last one |
| `Touched(context, other)` | When a moving body begins touching the node's trigger |

The `context` holds the node the behaviour is on, the scene, the time and the player's input.
`main.go` registers each behaviour once, `engine.Register(Spin{}, Player{})`, and the scene names
it. Its exported fields are its settings, filled from the scene by their JSON names, so one
behaviour can drive many nodes, each set differently.

```go
// Bob moves its node up and down.
type Bob struct {
	Height float64 `json:"height"` // metres
}

func (bob *Bob) Update(context *engine.Context) {
	context.Node.Transform.Position.Y = bob.Height * math.Sin(context.Time*2)
}
```

A behaviour finds another with `engine.Behaviour[Player](node)`, and a node with
`context.Find("Car/Wheel")`. `context.ChangeScene(file)` goes to another scene, and
`context.RestartScene()` starts this one again.

## Everything is in metres

| | |
|---|---|
| Unit | One unit is a metre, in 2D and 3D alike |
| Axes | Y up, X to the right, Z towards whoever looks (right-handed, as glTF) |
| Forward | A node with no turn faces -Z, as a camera does |
| Turns | Degrees: X tilts, Y turns, Z rolls; counter-clockwise looking down the axis |
| Scale | Times the node's own size; shapes are a metre across |

The editor's grid has one-metre cells, darker every five; X is red, Y green and Z blue.

## 2D is 3D seen from the front

A **2D scene** uses the same metres, seen from the front by a camera without perspective. A
sprite is as large as its pixels at the scene's **pixels per metre**: at 32, a sprite 32 pixels high
stands a metre high. Z only says what is drawn in front: the larger, the nearer. Isometric is 3D
seen from a fixed angle, so the three axes look the same length.

## Physics moves bodies, in 2D

A 2D node with a **body** is moved by the scene's physics while the game plays. Its kind says how:

| Kind | What it does |
|---|---|
| Solid | Never moves; others stand on it and stop against it (ground, walls) |
| Moving | Falls, moves at its velocity, stops against solids (the player) |
| Pushable | Falls; a moving body stands on it, or pushes it by walking into it (a box) |
| Platform | Moves as its behaviours say, carrying what stands on it (a lift) |
| Trigger | Stops nothing; its behaviours hear what touches it (a coin, a door) |

A trigger can have a gravity of its own, a **pull**: what falls inside it is pulled that way.
"On the floor" is wherever a body's gravity pulls it, so a body pulled up stands on a ceiling.

## Light multiplies colours

A scene can be lit. Its **ambient** is the light that reaches everything: white at 1 leaves the
colours as they are, a dark blue at 0.4 is a night. Lights on nodes add to it where they shine,
like lamps in a dark room, and can brighten up to twice. A scene with neither is drawn unlit, at no
cost. In 3D, lights are lamps that shine on the shapes around them, besides the sun.

## Not yet

- Sound, and a game's own interface (menus, a score on screen).
- Bodies that turn, slopes, and shapes other than boxes.
- Textures from models: they are drawn with their colours.
- Moving and turning things with handles in the view (the inspector and dragging do it now).

---

[← Get started](2-getting-started.md) · [The quest →](4-the-quest/0-the-quest.md)
