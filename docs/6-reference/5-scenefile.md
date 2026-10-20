# scenefile

```go
import "github.com/akrck02/go-engine/scenefile"
```

Package scenefile is how a scene is kept: a JSON file of nodes, each with its
transform (position in metres, rotation in degrees, scale), what it shows (a
shape, a model from the project's assets, a camera) and the behaviours the
game's own Go code gives it. The editor writes these files; the engine reads
them.

### Functions

| Function | |
|---|---|
| `func IsSpriteFile(path string) bool` | IsSpriteFile tells whether a path names a sprite's file. |
| `func IsTilesetFile(path string) bool` | IsTilesetFile tells whether a path names a tileset's file. |

### Constants and variables

```go
const (
	SpriteSuffix  = ".sprite.json"
	TilesetSuffix = ".tileset.json"
)
```

The files of sprites and tilesets end with these.

```go
const (
	DefaultIntensity  = 1.0
	DefaultLightOuter = 3.0 // metres
	// MostIntensity is the strongest a light or an ambient can be: twice as bright.
	MostIntensity = 2.0
)
```

The defaults of what a light or an ambient leaves unsaid.

```go
const DefaultPixelsPerMetre = 32
```

DefaultPixelsPerMetre is a 2D scene's when it does not say.

```go
const EmptyTile = -1
```

EmptyTile is a cell with no tile.

```go
var (
	Unreadable      = failure.Define("scenefile.unreadable", "the file is not a scene's JSON")
	NodeWithoutName = failure.Define("scenefile.node-without-name", "a node has no name")
	SlashInName     = failure.Define("scenefile.slash-in-name", "a node's name has a slash, which separates the names of a path")
	SameName        = failure.Define("scenefile.same-name", "two nodes under the same parent have the same name")
	UnknownShape    = failure.Define("scenefile.unknown-shape", "a node is a shape that does not exist")

	UnknownBodyKind   = failure.Define("scenefile.unknown-body-kind", "a body's kind is not one of BodyKinds")
	TileSizeWrong     = failure.Define("scenefile.tile-size-wrong", "a tile map's tiles have no size")
	TileCountWrong    = failure.Define("scenefile.tile-count-wrong", "a tile map's tiles don't fill its columns and rows")
	SheetSizeWrong    = failure.Define("scenefile.sheet-size-wrong", "a sprite sheet has no columns or no rows")
	FrameOutsideSheet = failure.Define("scenefile.frame-outside-sheet", "an animation shows a frame the sheet does not have")
	SameAnimationName = failure.Define("scenefile.same-animation-name", "two animations of a sheet have the same name")
	UnknownAnimation  = failure.Define("scenefile.unknown-animation", "a sheet starts with an animation it does not have")
	NoPicture         = failure.Define("scenefile.no-picture", "a sprite or a tileset names no picture")
	UnknownLightKind  = failure.Define("scenefile.unknown-light-kind", "a light's kind is not point or sprite")
	LightRadiiWrong   = failure.Define("scenefile.light-radii-wrong", "a light's radii are below zero, or the inner one is beyond the outer")
)
```

Error codes of the scenefile package.

```go
var BodyKinds = []BodyKind{Solid, Moving, Trigger, Pushable, Platform}
```

BodyKinds are the kinds of body, in the order the editor offers them.

```go
var IsometricAngles = [3]float64{-35.264389682754654, 45, 0}
```

IsometricAngles are the rotation, in degrees, of an isometric camera: the three axes look the same length on screen.

```go
var LightKinds = []LightKind{PointLight, SpriteLight}
```

LightKinds are the kinds of light, in the order the editor offers them.

```go
var Projections = []Projection{Perspective, Orthographic, Isometric}
```

Projections are the projections in the order the editor offers them.

```go
var Shapes = []string{"cube", "sphere", "plane", "cylinder"}
```

Shapes are the shapes a node may be, each a metre across.

### Ambient

Ambient is the light that reaches everything: daylight is white at 1,
a dim room a dark blue at 0.4. A 3D scene takes its intensity only, as the
strength of its sky light.

```go
type Ambient struct {
	Colour string `json:"colour,omitempty"` // "#rrggbb"; empty is white
	// Intensity is how strong it is: 1 (the default) leaves the scene's colours as they are, less
	// darkens them, up to 2 brightens them.
	Intensity *float64 `json:"intensity,omitempty"`
}
```

| Function or method | |
|---|---|
| `func (ambient *Ambient) Copy() *Ambient` | Copy is an ambient that shares nothing with this one; nil for none. |
| `func (ambient *Ambient) Strength() float64` | Strength is the ambient's intensity, or 1. |

### Animation

Animation is a run of frames of a sheet, shown at a pace.

```go
type Animation struct {
	Name            string  `json:"name"`
	Frames          []int   `json:"frames"`
	FramesPerSecond float64 `json:"framesPerSecond"`
	Loop            bool    `json:"loop,omitempty"` // starts again after its last frame; otherwise stays on it
}
```

### Behaviour

Behaviour is a piece of the game's own Go code attached to a node, by the
name it was registered with, and the values of its fields.

```go
type Behaviour struct {
	Name   string         `json:"name"`
	Fields map[string]any `json:"fields,omitempty"`
}
```

### Body

Body gives a node a body in the physics of a 2D scene.

```go
type Body struct {
	Kind BodyKind   `json:"kind"`
	Size [2]float64 `json:"size,omitempty"` // width and height in metres; zero is the node's own sprite or shape
	// GravityScale is how much gravity pulls a moving body: 1 (the default) fully, 0 not at all.
	GravityScale *float64 `json:"gravityScale,omitempty"`
	// Pull, on a trigger, is the gravity of the area it covers, in metres a second each second
	// across and up: the falling bodies inside are pulled this way instead of the world's. Nil
	// leaves the world's.
	Pull *[2]float64 `json:"pull,omitempty"`
	// IgnoresZones, on a moving or pushable body, keeps the world's gravity for it inside zones too.
	IgnoresZones bool `json:"ignoresZones,omitempty"`
}
```

| Function or method | |
|---|---|
| `func (body *Body) Copy() *Body` | Copy is an independent copy of the body: changing one leaves the other alone. |
| `func (body *Body) Gravity() float64` | Gravity is how much gravity pulls the body: its GravityScale, or 1. |

### BodyKind

BodyKind is what a body does: as physics.Kind, by name.

```go
type BodyKind string
```

```go
const (
	// Solid never moves; moving bodies stop against it.
	Solid BodyKind = "solid"
	// Moving falls, moves and is stopped by solids.
	Moving BodyKind = "moving"
	// Trigger stops nothing; it notices the moving and pushable bodies that touch it.
	Trigger BodyKind = "trigger"
	// Pushable falls and moves as a moving body, which stands on it and pushes it walking into it.
	Pushable BodyKind = "pushable"
	// Platform moves at the velocity its behaviours give it, carrying what stands on it.
	Platform BodyKind = "platform"
)
```

### Camera

Camera makes a node a camera; the scene is seen through the first one marked
Main (or the first camera, if none is).

```go
type Camera struct {
	Projection  Projection `json:"projection"`
	FieldOfView float64    `json:"fieldOfView,omitempty"` // degrees, top to bottom, in perspective; 0 is 45
	Height      float64    `json:"height,omitempty"`      // metres seen top to bottom, orthographic and isometric; 0 is 10
	// Width is the metres a 2D camera sees across, at least: a screen narrower than Height and
	// Width make (a phone held upright) sees more height instead of losing the sides. 0 is no least.
	Width float64 `json:"width,omitempty"`
	Main  bool    `json:"main,omitempty"`

	// Follow is the path of a node the camera follows, in a 2D scene ("Keeper", "Harbour/Keeper").
	Follow string `json:"follow,omitempty"`
	// SmoothingSeconds is about how many seconds the camera takes to catch up with what it follows; 0
	// keeps it right on it.
	SmoothingSeconds float64 `json:"smoothing,omitempty"`
	// Bounds keep a following camera's view inside a box of the world: its lowest x and y, then its
	// highest. Without them, the scene's tile maps make the box, open above unless Ceiling.
	Bounds *[4]float64 `json:"bounds,omitempty"`
	// Ceiling makes the top of the tile maps a limit too, as a cave's; without it, the sky above
	// them stays open and the camera rises with what it follows.
	Ceiling bool `json:"ceiling,omitempty"`
}
```

### Light

Light makes a node a light: a point light, in 2D or 3D, or in 2D a sprite
light.

```go
type Light struct {
	Kind      LightKind `json:"kind"`
	Colour    string    `json:"colour,omitempty"`    // "#rrggbb"; empty is white
	Intensity *float64  `json:"intensity,omitempty"` // as the ambient's; nil is 1
	// Inner and Outer are a point light's radii in metres: it is full within Inner and fades to
	// nothing at Outer.
	Inner float64 `json:"inner,omitempty"`
	Outer float64 `json:"outer,omitempty"`
}
```

| Function or method | |
|---|---|
| `func (light *Light) Copy() *Light` | Copy is a light that shares nothing with this one; nil for none. |
| `func (light *Light) Radii() (inner, outer float64)` | Radii are a point light's inner and outer radii in metres, the outer one 3 when it has none. |
| `func (light *Light) Strength() float64` | Strength is the light's intensity, or 1. |

### LightKind

LightKind is what a light shines from.

```go
type LightKind string
```

```go
const (
	// PointLight shines from the node's position all around, as a lamp.
	PointLight LightKind = "point"
	// SpriteLight shines through the node's sprite: its colours light where it is, and it is not
	// drawn itself. A glowing sign, a window.
	SpriteLight LightKind = "sprite"
)
```

### Mode

Mode is whether a scene is 3D or 2D.

```go
type Mode string
```

```go
const (
	Mode3D Mode = "3d"
	Mode2D Mode = "2d"
)
```

### Node

Node is a thing in a scene, and its children, which move with it.

```go
type Node struct {
	Name     string     `json:"name"`
	Position [3]float64 `json:"position"`
	Rotation [3]float64 `json:"rotation"` // degrees: pitch (X), yaw (Y), roll (Z), as space.Angles
	Scale    [3]float64 `json:"scale"`
	Hidden   bool       `json:"hidden,omitempty"`

	Shape  string  `json:"shape,omitempty"`  // one of Shapes
	Model  string  `json:"model,omitempty"`  // a model in the project, as "assets/models/tree.glb"
	Sprite string  `json:"sprite,omitempty"` // in a 2D scene, a picture or a sprite's file, as "assets/sprites/hero.sprite.json"
	Colour string  `json:"colour,omitempty"` // of a shape, "#rrggbb"
	Camera *Camera `json:"camera,omitempty"`

	Body  *Body        `json:"body,omitempty"`  // in a 2D scene, a body in its physics
	Sheet *SpriteSheet `json:"sheet,omitempty"` // in a 2D scene, the sprite cut into animated frames
	// Animation is the animation of its sprite the node starts with, in a 2D scene; empty is the
	// sprite's own first one. Nodes that share a sprite can start differently.
	Animation string   `json:"animation,omitempty"`
	Tiles     *TileMap `json:"tiles,omitempty"` // in a 2D scene, a map of tiles
	Light     *Light   `json:"light,omitempty"` // a light: a point light in 2D or 3D, a sprite light in 2D

	Behaviours []Behaviour `json:"behaviours,omitempty"`
	Children   []Node      `json:"children,omitempty"`
}
```

| Function or method | |
|---|---|
| `func NewNode(name string) Node` | NewNode is a node at the origin, unturned, at its own size. |

### Projection

Projection is how a camera flattens the world.

```go
type Projection string
```

```go
const (
	Perspective  Projection = "perspective"  // far things look smaller: most 3D games
	Orthographic Projection = "orthographic" // no perspective: plans, side views, 2D in 3D
	Isometric    Projection = "isometric"    // orthographic, from the classic angle (looking down 35.264° from 45° round)
)
```

### Scene

Scene is a scene file.

```go
type Scene struct {
	Name string `json:"name"`
	// Mode is "3d" (the default: shapes and models in perspective, orthographic or isometric) or
	// "2d" (sprites and flat shapes, seen from the front by an orthographic camera).
	Mode Mode   `json:"mode,omitempty"`
	Sky  string `json:"sky,omitempty"` // the background colour, "#rrggbb"; empty is a daylight sky
	// PixelsPerMetre is, in a 2D scene, how many of a sprite's pixels make a metre: a sprite 32
	// pixels high stands a metre high at 32. 0 is 32.
	PixelsPerMetre float64 `json:"pixelsPerMetre,omitempty"`
	// Ambient is the light that reaches everything: in a 2D scene its colour and intensity (nil is
	// white at 1 when the scene has lights, and no lighting at all when it has none), in a 3D scene
	// how strong its sky light is (nil is the sky's own).
	Ambient *Ambient `json:"ambient,omitempty"`
	Nodes   []Node   `json:"nodes"`
}
```

| Function or method | |
|---|---|
| `func Read(data []byte) (*Scene, error)` | Read reads a scene file and checks it. |
| `func (scene *Scene) Check() error` | Check tells what is wrong with a scene: a node without a name, a slash in a name (names make paths, as "car/wheel"), two siblings with one name, or a shape that does not exist. |
| `func (scene *Scene) Is2D() bool` | Is2D tells whether the scene is 2D. |
| `func (scene *Scene) Marshal() ([]byte, error)` | Marshal writes the scene as indented JSON. |

### Sprite

Sprite is a picture cut into frames, with the animations made of them.

```go
type Sprite struct {
	Picture string `json:"picture"` // a picture in the project, as "assets/sprites/lamp.png"
	SpriteSheet
}
```

| Function or method | |
|---|---|
| `func ReadSprite(data []byte) (*Sprite, error)` | ReadSprite reads a sprite's file and checks it. |
| `func (sprite *Sprite) Check() error` | Check tells what is wrong with a sprite, if anything. |
| `func (sprite *Sprite) Marshal() ([]byte, error)` | Marshal writes the sprite as its file keeps it. |

### SpriteSheet

SpriteSheet cuts a node's sprite into frames of equal size, in rows read
left to right and top to bottom (frame 0 is the top-left one), and names
animations made of them.

```go
type SpriteSheet struct {
	Columns    int         `json:"columns"`
	Rows       int         `json:"rows"`
	Animations []Animation `json:"animations,omitempty"`
	Start      string      `json:"start,omitempty"` // the animation playing when the scene starts
}
```

| Function or method | |
|---|---|
| `func (sheet *SpriteSheet) Copy() *SpriteSheet` | Copy is an independent copy of the sheet and its animations. |
| `func (sheet *SpriteSheet) Find(name string) (Animation, bool)` | Find finds an animation by its name. |
| `func (sheet *SpriteSheet) FrameCount() int` | FrameCount is how many frames the sheet has. |

### TileMap

TileMap is a grid of tiles cut from a tileset. Its node's position is the
map's top-left corner; rows go down from there. A tile is TileSize pixels
a side, which the scene's pixels per metre make metres. The tileset is a
tileset's file, which brings its tile size and solid tiles, or a picture,
with the map's own.

```go
type TileMap struct {
	Tileset  string `json:"tileset"`            // a tileset's file, or a picture cut into tiles in rows
	TileSize int    `json:"tileSize,omitempty"` // pixels a side, for a picture
	Columns  int    `json:"columns"`
	Rows     int    `json:"rows"`
	Tiles    []int  `json:"tiles"`           // a tile number per cell, row by row; EmptyTile for none
	Solid    []int  `json:"solid,omitempty"` // the tile numbers that bodies stand on, for a picture
}
```

| Function or method | |
|---|---|
| `func (tiles *TileMap) Copy() *TileMap` | Copy is an independent copy of the tile map. |
| `func (tiles *TileMap) IsSolid(tile int) bool` | IsSolid tells whether a tile number is solid. |
| `func (tiles *TileMap) TileAt(column, row int) int` | TileAt is the tile in a cell, or EmptyTile outside the map. |

### Tileset

Tileset is a picture cut into square tiles, in rows, and which of them
bodies stand on.

```go
type Tileset struct {
	Picture  string `json:"picture"`
	TileSize int    `json:"tileSize"`        // pixels a side
	Solid    []int  `json:"solid,omitempty"` // the tile numbers that bodies stand on
}
```

| Function or method | |
|---|---|
| `func ReadTileset(data []byte) (*Tileset, error)` | ReadTileset reads a tileset's file and checks it. |
| `func (tileset *Tileset) Check() error` | Check tells what is wrong with a tileset, if anything. |
| `func (tileset *Tileset) IsSolid(tile int) bool` | IsSolid tells whether bodies stand on a tile of the tileset. |
| `func (tileset *Tileset) Marshal() ([]byte, error)` | Marshal writes the tileset as its file keeps it. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
