# engine

```go
import "github.com/akrck02/go-engine/engine"
```

Package engine runs a game: a scene of nodes, read from the project's scene
files, drawn by go-renderer, and moved by the game's own Go code through
behaviours. A node is a thing in the world: a shape, a model, a sprite, a
camera, or an empty point to hang others from. Its Transform places it relative
to its parent, in metres and degrees (see package space).

A game's main registers its behaviours and runs the project it embeds:

### Functions

| Function | |
|---|---|
| `func BeforeBehaviours(hook func(context *Context))` | BeforeBehaviours sets code that runs once a frame, before every behaviour, with the frame's context (its Node is nil); nil removes it. |
| `func Behaviour[Wanted any](node *Node) (*Wanted, bool)` | Behaviour finds a node's behaviour of a type, while the game runs. |
| `func BetweenFrames(call func())` | BetweenFrames asks for a call to run once, at the start of the next frame, before anything is drawn: on the 3DS, the system keyboard can take both screens there. |
| `func BindActions(bindings map[string]map[string][]string, target string) []error` | BindActions reads the actions' bindings (engine.json's actions) for a target, replacing those read before. |
| `func CameraAt(world space.Transform, settings scenefile.Camera) models.Camera` | CameraAt is go-renderer's camera for a camera placed at a transform with some settings. |
| `func CheckInputs(words []string, target string) []error` | CheckInputs tells which of some words a target does not know, each by its code (InputUnknown). |
| `func DrawOverlay(list *draw2d.List, width, height float32) bool` | DrawOverlay draws the game's 2D over a 3D scene (SetOverlay) into a list, in an area of a width and a height starting at 0, 0; it tells whether the game has any. |
| `func Fullscreen() bool` | Fullscreen tells whether the game is asked to be on the whole screen. |
| `func HasInterface() bool` | HasInterface tells whether the game shows a go-ui interface of its own (SetInterface). |
| `func HeightSeen(settings *scenefile.Camera, aspect float64) float64` | HeightSeen is the metres a 2D camera sees top to bottom on a screen of a shape (its width over its height): the camera's height (10 when it has none), or more on a screen too narrow for the width it must see. |
| `func HintOf(words []string, target string, kind InputKind, layout input.PadLayout) string` | HintOf is the hint of an action bound to some words on a target, for a player using a kind of input with a pad of some layout: as ActionHint, for an editor showing a device that is not played. |
| `func InputLabel(word string, layout input.PadLayout) string` | InputLabel is what a key cap or a button says for a word of engine.json, on a pad of some layout: "Escape" is "Esc", "Pad A" is "✕" on a PlayStation pad, "Touch Jump" is "Jump". |
| `func PlaceSecondScreen(area *draw2d.Rect)` | PlaceSecondScreen shows the second screen's scene in an area of the window, in window pixels, on a device that has no second screen, with its own sky behind it: a map in a corner of a computer's window, the same scene a 3DS shows below. |
| `func PlaceSecondScreenOver(area *draw2d.Rect)` | PlaceSecondScreenOver is PlaceSecondScreen with no background: the scene floats over the game's view in its area, as a radar does. |
| `func PressedInput(state *input.State, target string) (string, bool)` | PressedInput is the word of an input pressed this frame on a target, if one was: a key, a mouse button or a pad's button (a 3DS's buttons and Android's back cannot be pressed from a computer). |
| `func Quit()` | Quit ends the game once the frame being played ends. |
| `func Register(behaviours ...any)` | Register makes behaviours known by their type's name: Register(Spin{}) is "Spin". |
| `func RegisterAs(name string, behaviour any)` | RegisterAs makes a behaviour known by a name of its own. |
| `func Registered() []string` | Registered are the names of the behaviours registered, in order. |
| `func Run(files fs.FS) error` | Run plays a project, from its files (embedded in the game, or a folder), until its window is closed: on the desktop in a window, on the Nintendo 3DS on the top screen. |
| `func RunningOn() string` | RunningOn is the name of the target the game runs on now. |
| `func SaveBindings() error` | SaveBindings keeps the actions the player changed, for the next time the game plays here. |
| `func SecondScreen() (*Scene, bool)` | SecondScreen is the scene the second screen shows (SetSecondScreen), if any: one who plays the game elsewhere than Run moves it on each frame (Scene.Update) and shows it from Scene.Camera. |
| `func SetEditor(run func() error)` | SetEditor installs the editor (the editor package does, from a project's editor.go). |
| `func SetFullscreen(on bool)` | SetFullscreen asks for the game on the whole screen, or in its window again. |
| `func SetInterface(paint func(context *ui.Context, menu Menu))` | SetInterface makes a game show an interface of its own, with go-ui, in its own look: a menu of its scenes (Menu), a map, its controls, a login form. |
| `func SetOverlay(paint func(list *draw2d.List, width, height float32))` | SetOverlay sets what the game draws over a 3D scene each frame, in the pixels of the screen it is on (the window's, or the 3DS's top screen, 400 × 240): the list is empty when it is called. |
| `func SetPointerCaptured(captured bool)` | SetPointerCaptured holds and hides the pointer, for a game that turns with the mouse, or lets it go (a pause menu): MouseMovement still reads the mouse. |
| `func SetSecondScreen(shown *Scene)` | SetSecondScreen shows a scene on the second screen, under the game's interface (SetInterface), seen from its main camera; nil leaves the screen to the interface. |
| `func SetSurface(name string, paint func(context *ui.Context))` | SetSurface makes a game show a surface of its own, by name, painted with Go UI every frame; engine.json's surfaces say where it goes on each target. |
| `func ShowSurfaces(context *ui.Context, played *Scene, scenes []MenuScene, main, second draw2d.Rect)` | ShowSurfaces builds the game's surfaces (SetSurface, SetInterface) for a scene played, offering some scenes (ScenesOf), where the target UseSurfaces chose places them: those on the main screen in an area of the window, those on the second screen in another (an empty one shows none). |
| `func TakeQuit() bool` | TakeQuit tells whether the game asked to quit since the last time, and forgets it: one who plays the game elsewhere than Run (the editor's Play) stops it then. |
| `func TouchWords(control TouchControl) []string` | TouchWords are the words a touch control is bound with: a button's name, a stick's two axes. |
| `func UseSurfaces(places map[string]map[string]SurfacePlace, target string) []error` | UseSurfaces places the game's surfaces on a target, as engine.json's surfaces say (by surface, then by target); Run does it, and the editor for the target it previews. |
| `func Version() string` | Version says which Go Engine this program was built with: its release and commit, as "v0.0.1 (ba08523)", or "v0.0.1 (77aa001, local)" for a local copy, as the editor's window and the logs show it. |

### Constants and variables

```go
const (
	ButtonLeftShoulder  = input.KeyQ
	ButtonRightShoulder = input.KeyE
	ButtonLeftTrigger   = input.KeyZ
	ButtonRightTrigger  = input.KeyC
	ButtonStart         = input.KeyEnter
)
```

The keys shoulder buttons, triggers and START are read as (a 3DS's L, R, ZL, ZR and START), and that press them on a keyboard.

```go
const (
	Windowed         = "windowed"
	FullscreenWindow = "fullscreen"
)
```

The ways a game's window can start (engine.json's window).

```go
const (
	KeyboardInputs = "Keyboard"
	MouseInputs    = "Mouse"
	PadInputs      = "Pad"
	StickInputs    = "Sticks"
	ButtonInputs   = "Buttons" // a device's own buttons (a 3DS's)
	SystemInputs   = "System"  // the system's back (Android's)
)
```

The groups of inputs, as an editor shows them.

```go
const (
	KeyA         = input.KeyA
	KeyB         = input.KeyB
	KeyC         = input.KeyC
	KeyD         = input.KeyD
	KeyE         = input.KeyE
	KeyF         = input.KeyF
	KeyG         = input.KeyG
	KeyH         = input.KeyH
	KeyI         = input.KeyI
	KeyJ         = input.KeyJ
	KeyK         = input.KeyK
	KeyL         = input.KeyL
	KeyM         = input.KeyM
	KeyN         = input.KeyN
	KeyO         = input.KeyO
	KeyP         = input.KeyP
	KeyQ         = input.KeyQ
	KeyR         = input.KeyR
	KeyS         = input.KeyS
	KeyT         = input.KeyT
	KeyU         = input.KeyU
	KeyV         = input.KeyV
	KeyW         = input.KeyW
	KeyX         = input.KeyX
	KeyY         = input.KeyY
	KeyZ         = input.KeyZ
	KeyUp        = input.KeyUp
	KeyDown      = input.KeyDown
	KeyLeft      = input.KeyLeft
	KeyRight     = input.KeyRight
	KeySpace     = input.KeySpace
	KeyEnter     = input.KeyEnter
	KeyEscape    = input.KeyEscape
	KeyShift     = input.KeyShift
	KeyControl   = input.KeyControl
	KeyTab       = input.KeyTab
	KeyBackspace = input.KeyBackspace
)
```

Keys.

```go
const (
	DefaultNear  = 0.05
	DefaultFar   = 2000
	DefaultFocus = 10 // a room's depth: most scenes look at things about this far
)
```

A camera's distances when its settings leave them at 0, in metres.

```go
const (
	OverPlace         = "over"          // over the game's view, aligned to an edge or a corner
	WindowPlace       = "window"        // a floating window over the view, opened and closed by an action
	SecondScreenPlace = "second-screen" // the second screen (a 3DS's), touchable, in a column with the rest there
	SheetPlace        = "sheet"         // a sheet from the bottom (a phone's), opened and closed by an action
	NoPlace           = "none"          // not shown on that target
)
```

The places a surface can go.

```go
const (
	DesktopTarget     = "desktop"
	Nintendo3DSTarget = "nintendo3ds"
	AndroidTarget     = "android"
	WebTarget         = "web"
)
```

The targets' names in engine.json.

```go
const DefaultGravity = 9.81
```

DefaultGravity is how fast a 2D scene's gravity pulls down, in metres a second each second: Earth's.

```go
const EditorVariable = "GO_ENGINE_EDITOR"
```

EditorVariable is set, in the environment, when go-engine starts a game to edit it.

```go
const FullscreenAction = "fullscreen"
```

FullscreenAction is the action that puts the game on the whole screen and back.

```go
const InterfaceSurface = "interface"
```

InterfaceSurface is the name of the game's interface among its surfaces (SetInterface).

```go
const Module = "github.com/akrck02/go-engine"
```

Module is Go Engine's module path.

```go
const Release = "v0.0.2"
```

Release is the version of Go Engine being made.

```go
const SettingsFile = "engine.json"
```

SettingsFile is the settings' file in a project.

```go
var (
	SettingsNotFound      = failure.Define("engine.settings-not-found", "the project has no engine.json")
	SettingsUnreadable    = failure.Define("engine.settings-unreadable", "engine.json is not the project's settings")
	NoStartScene          = failure.Define("engine.no-start-scene", "engine.json names no start scene")
	SceneNotFound         = failure.Define("engine.scene-not-found", "the scene file does not exist")
	SceneUnreadable       = failure.Define("engine.scene-unreadable", "the scene file is not a scene")
	BehaviourUnregistered = failure.Define("engine.behaviour-unregistered", "no behaviour is registered by that name (engine.Register in the game's main)")
	BehaviourFields       = failure.Define("engine.behaviour-fields", "the scene's fields do not fit the behaviour's")
	NoFiles               = failure.Define("engine.no-files", "the scene was made without the project's files")
	ModelFormat           = failure.Define("engine.model-format", "only .glb models are read for now")
	ModelNotFound         = failure.Define("engine.model-not-found", "the model's file does not exist")
	ModelUnreadable       = failure.Define("engine.model-unreadable", "the model's file is not a .glb go-renderer reads")
	PictureNotFound       = failure.Define("engine.picture-not-found", "the picture's file does not exist")
	PictureUnreadable     = failure.Define("engine.picture-unreadable", "the picture is not a PNG or a JPEG")
	SpriteNotFound        = failure.Define("engine.sprite-not-found", "the sprite's file does not exist")
	TilesetNotFound       = failure.Define("engine.tileset-not-found", "the tileset's file does not exist")
	SceneNotRestartable   = failure.Define("engine.scene-not-restartable", "the scene was not made from a description, so it cannot start again")
	InputUnknown          = failure.Define("engine.input-unknown", "an action is bound to an input the target does not have: see the words an action can be bound to")
	BindingsNotSaved      = failure.Define("engine.bindings-not-saved", "the actions the player changed could not be kept on this device")
	WindowUnknown         = failure.Define("engine.window-unknown", "engine.json's window is neither windowed nor fullscreen")
	SurfacePlaceUnknown   = failure.Define("engine.surface-place-unknown", "engine.json places a surface somewhere the engine does not know: see the places and alignments a surface can have")
)
```

Error codes of the engine package.

```go
var InputTargets = []string{DesktopTarget, Nintendo3DSTarget, AndroidTarget, WebTarget}
```

InputTargets are the targets an action can be bound on, in the order the editor shows them.

### ActionState

ActionState is an action this frame: held, pressed or let go, and its value
from -1 to 1 (an axis's, or 1 while a button holds it).

```go
type ActionState struct {
	Down, Pressed, Released bool
	Value                   float64
}
```

### Context

Context is what a behaviour is given each time it runs.

```go
type Context struct {
	Node  *Node   // the node the behaviour is attached to
	Scene *Scene  // the scene it is in
	Delta float64 // seconds since the last frame
	Time  float64 // seconds since the scene started
	Input *Input
}
```

| Function or method | |
|---|---|
| `func (context *Context) ChangeScene(file string)` | ChangeScene asks for another scene of the project, by its file, at the end of the frame. |
| `func (context *Context) Find(path string) (*Node, bool)` | Find finds a node of the scene by its path, as "Car/Wheel". |
| `func (context *Context) RestartScene()` | RestartScene asks for the scene to start again as it was written, at the end of the frame. |

### Input

Input is the keyboard, the mouse and the controls, as the game reads them.

```go
type Input struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (keys *Input) Action(name string) ActionState` | Action is an action's state this frame; an action not bound is never held. |
| `func (keys *Input) ActionHint(name string) string` | ActionHint is what to show a player for an action: the label (InputLabel) of its first input of the kind they used last ("Tab" at the keyboard, "Start" with a pad), or of its first input when none is of that kind; "" for an action not bound. |
| `func (keys *Input) ActionWords(name string) []string` | ActionWords are the words an action is bound to on the target played, as engine.json's ("Tab", "Pad Start"), rebinding included; none for an action not bound. |
| `func (keys *Input) Axis(negative, positive Key) float64` | Axis is -1 while the first key is held, 1 while the second is, 0 with neither or both. |
| `func (keys *Input) Down(key Key) bool` | Down tells whether a key is held; Pressed whether it went down this frame. |
| `func (keys *Input) Gyroscope() Turn` | Gyroscope is how fast the device turns (a 3DS's gyroscope); still without one, or in an emulator. |
| `func (keys *Input) LastInputKind() InputKind` | LastInputKind is the kind of the last input the player pressed or moved: it changes only when another kind is used, not when the player stops. |
| `func (keys *Input) MouseMovement() (x, y float64)` | MouseMovement is how far the mouse moved since the last frame, in window pixels: it keeps counting while the pointer is captured (SetPointerCaptured). |
| `func (keys *Input) Move() space.Vector` | Move is the direction the arrows (or WASD, or a 3DS's cross and circle pad) point, on the ground: right is +X, up is Forward (-Z). |
| `func (keys *Input) Pointer() (x, y float64, pressed bool)` | Pointer is where the mouse (or a finger) is, in window pixels, and whether it is pressed. |
| `func (keys *Input) Pressed(key Key) bool` |  |
| `func (keys *Input) Rebind(name string, words []string) []error` | Rebind changes an action's inputs on the target played to some words (as engine.json's), and remembers the change for SaveBindings. |
| `func (keys *Input) RightStick() (x, y float64)` | RightStick is the second stick (a New 3DS's C-stick, a game pad's right stick) from -1 to 1 on each axis, up positive; 0 without one. |
| `func (keys *Input) SecondScreenTouch() (x, y int, touching bool)` | SecondScreenTouch is where a finger or a stylus is on the second screen (a 3DS's bottom one, 320 × 240 pixels), and whether one is on it at all; never without a second screen. |
| `func (keys *Input) State() *input.State` | State is go-renderer's input, for what Input does not cover. |
| `func (keys *Input) Stick() (x, y float64)` | Stick is the main stick (a 3DS's circle pad, a game pad's left stick) from -1 to 1 on each axis, up positive. |

### InputChoice

InputChoice is a word an action can be bound with on a target, and its
group.

```go
type InputChoice struct{ Word, Group string }
```

| Function or method | |
|---|---|
| `func InputChoices(target string) []InputChoice` | InputChoices are the words a target knows, group by group, each group in a reading order. |

### InputKind

InputKind is a kind of device a player uses: a hint names the inputs of the
one in their hands.

```go
type InputKind int
```

```go
const (
	// NoInputYet is before the player used anything: a hint names an action's first input.
	NoInputYet InputKind = iota
	// KeyboardAndMouse is a computer's keys and mouse buttons.
	KeyboardAndMouse
	// GamePad is a pad's buttons and sticks, on a computer or in a browser.
	GamePad
	// TouchScreen is a phone's screen and its touch controls (and Android's back).
	TouchScreen
	// ConsoleButtons is a console's own buttons and sticks (a 3DS's).
	ConsoleButtons
)
```

### Key

Key is a key of the keyboard (a 3DS's buttons stand for some: the cross is
the arrows, A is Enter, B is Escape).

```go
type Key = input.Key
```

### Menu

Menu is what a game's own bottom screen is given: its scenes, the one
playing, and how to play another.

```go
type Menu struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (menu Menu) Play(file string)` | Play plays a scene next, by its file. |
| `func (menu Menu) Playing() *Scene` | Playing is the scene playing now: its name, and its pictures (Scene.Picture) to draw with. |
| `func (menu Menu) Scenes() []MenuScene` | Scenes are the game's scenes, in the order of their files. |

### MenuScene

MenuScene is a scene the menu offers: its file, and the name it shows.

```go
type MenuScene struct {
	File, Name string
}
```

| Function or method | |
|---|---|
| `func ScenesOf(files fs.FS) []MenuScene` | ScenesOf are the scenes in a project's files, in the order of their files, as the game's menu offers them. |

### Node

Node is a thing in a scene.

```go
type Node struct {
	Name      string
	Transform space.Transform
	Visible   bool

	// Behaviours are the game's code attached to the node, as the scene file keeps them; while the
	// game runs, behaviours holds them made.
	Behaviours []scenefile.Behaviour

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (node *Node) Add(child *Node)` | Add hangs a node from this one, taking it from where it hung before. |
| `func (node *Node) Animation() string` | Animation is the name of the animation the node plays, or "". |
| `func (node *Node) Animations() []string` | Animations are the names of the animations of the node's sprite, in order. |
| `func (node *Node) Attach(behaviour any)` | Attach gives a node a behaviour while the game runs, made by the game's code; it starts on the next frame. |
| `func (node *Node) Body() (*physics.Body, bool)` | Body is the node's body in the scene's physics, while the game plays: its velocity, whether it stands on the floor. |
| `func (node *Node) BodyBox() (space.Box, bool)` | BodyBox is the box of the world the node's body takes, from its settings, playing or not: what the editor outlines. |
| `func (node *Node) BodySettings() (*scenefile.Body, bool)` | BodySettings is the body the node is given in the scene file, to edit in place; a node without one has none. |
| `func (node *Node) Bounds() space.Box` | Bounds is the box, in the world, around what the node itself shows (not its children); empty for a node that shows nothing. |
| `func (node *Node) Camera() (*scenefile.Camera, bool)` | Camera is the node's camera settings, if it is a camera. |
| `func (node *Node) CellAt(point space.Vector) (column, row int, inside bool)` | CellAt finds the cell of the node's tile map at a point of the world; inside is false outside the map, or for a node that is not one. |
| `func (node *Node) CellBox(column, row int) space.Box` | CellBox is the box of the world one cell of the node's tile map fills. |
| `func (node *Node) Children() []*Node` | Children are the nodes that hang from this one. |
| `func (node *Node) Colour() string` | Colour is the colour of the node's shape, "#rrggbb". |
| `func (node *Node) Describe() scenefile.Node` | Describe writes a node, and its children, as a scene file keeps them. |
| `func (node *Node) FaceLeft(left bool)` | FaceLeft turns the node's sprite to face left (true) or right (false, as it was drawn). |
| `func (node *Node) FacingLeft() bool` | FacingLeft tells whether the node's sprite faces left. |
| `func (node *Node) Find(path string) (*Node, bool)` | Find finds a node under this one by its path: "Wheel", or "Car/Wheel" for a deeper one. |
| `func (node *Node) Frame() int` | Frame is the frame of the sheet the node shows now, from 0, the top-left one. |
| `func (node *Node) Index() int` | Index is where the node is among its parent's children. |
| `func (node *Node) Insert(child *Node, at int)` | Insert hangs a node from this one at a place among its children. |
| `func (node *Node) Light() (*scenefile.Light, bool)` | Light is the node's light, to edit in place; a node that is not one has none. |
| `func (node *Node) LightRadii() (inner, outer float64)` | LightRadii are a point light's radii in metres, times its node's largest scale (across and up in 2D, which has no depth to scale): a lamp scaled up reaches further. |
| `func (node *Node) MadeBehaviours() []any` | MadeBehaviours are the node's behaviours as the game runs them, made from its Behaviours: what the editor watches while a scene plays. |
| `func (node *Node) Mesh() (*scene.Mesh, bool)` | Mesh is the mesh the game's code gave the node, if any. |
| `func (node *Node) Model() string` | Model is the model the node shows, by its path in the project, or "". |
| `func (node *Node) Moving() bool` | Moving tells whether the node is marked as moving every frame. |
| `func (node *Node) NamedStart() string` | NamedStart is the animation the node itself names to start with, or "" for its sprite's own. |
| `func (node *Node) Parent() *Node` | Parent is the node this one hangs from; nil for the scene's root. |
| `func (node *Node) Path() string` | Path is the node's names from the root, as "Car/Wheel". |
| `func (node *Node) Play(name string) bool` | Play starts an animation of the node's sheet by its name. |
| `func (node *Node) Remove()` | Remove takes the node, and what hangs from it, out of its scene, and their bodies out of its physics. |
| `func (node *Node) Scene() *Scene` | Scene is the scene the node is in. |
| `func (node *Node) SetBodySettings(body *scenefile.Body)` | SetBodySettings gives the node a body in its scene's physics, or takes it away (nil). |
| `func (node *Node) SetCamera(camera *scenefile.Camera)` | SetCamera makes the node a camera, or no longer one with nil. |
| `func (node *Node) SetLight(light *scenefile.Light)` | SetLight makes the node a light, or not one (nil). |
| `func (node *Node) SetMesh(mesh *scene.Mesh)` | SetMesh makes the node show a mesh the game's code built (a vehicle from its design), instead of its shape or model; another call replaces it, and nil goes back to the shape or model. |
| `func (node *Node) SetModel(path string) error` | SetModel makes the node show a model of the project's. |
| `func (node *Node) SetMoving(moving bool)` | SetMoving says the game moves or turns the node every frame. |
| `func (node *Node) SetShape(shape, colour string)` | SetShape makes the node a shape of a colour ("" for no shape). |
| `func (node *Node) SetSheet(sheet *scenefile.SpriteSheet)` | SetSheet cuts the node's sprite into a sheet's frames, or shows it whole again (nil). |
| `func (node *Node) SetSprite(path string) error` | SetSprite makes the node show a picture of the project's, or a sprite's file, in a 2D scene. |
| `func (node *Node) SetStartsWith(name string)` | SetStartsWith makes the node start with an animation of its sprite; "" starts with the sprite's own first one. |
| `func (node *Node) SetTiles(tiles *scenefile.TileMap)` | SetTiles makes the node a tile map, or not one (nil). |
| `func (node *Node) Shape() string` | Shape is the shape the node is ("cube", "sphere", "plane", "cylinder"), or "". |
| `func (node *Node) Sheet() (*scenefile.SpriteSheet, bool)` | Sheet is the sprite sheet the node's sprite is cut into, to edit in place; a sprite shown whole has none. |
| `func (node *Node) SolidCell(column, row int) bool` | SolidCell tells whether bodies stand on a cell of the node's tile map. |
| `func (node *Node) SolidTile(tile int) bool` | SolidTile tells whether bodies stand on a tile of the node's tileset, by its number. |
| `func (node *Node) Sprite() string` | Sprite is the picture the node shows in a 2D scene, by its path in the project, or "". |
| `func (node *Node) StartsWith() string` | StartsWith is the animation the node plays when the scene starts: the one it names, or else its sprite's own first one, or "". |
| `func (node *Node) Tiles() (*scenefile.TileMap, bool)` | Tiles is the node's tile map, to edit in place; a node that is not one has none. |
| `func (node *Node) View() models.Camera` | View is what a camera node sees, as go-renderer's camera. |
| `func (node *Node) Walk(visit func(node *Node) bool)` | Walk calls a function for the node and everything under it, parents before their children; it stops going down a branch when the function returns false. |
| `func (node *Node) World() space.Transform` | World is where the node is in the world, all its parents' transforms applied. |

### Scene

Scene is a scene loaded from its file: a tree of nodes under a root,
and what go-renderer draws of it.

```go
type Scene struct {
	Name string
	Sky  string // the background colour, "#rrggbb"; empty for a daylight sky
	Root *Node
	// Mode is 3D or 2D; a 2D scene draws its sprites and shapes flat (see Draw2D).
	Mode scenefile.Mode
	// PixelsPerMetre is how many of a sprite's pixels make a metre, in a 2D scene.
	PixelsPerMetre float64
	// Ambient is the light that reaches everything in a 2D scene; nil is daylight when the scene
	// has lights (see lights.go).
	Ambient *scenefile.Ambient

	// Editing shows what only the editor shows: a marker for cameras and empty nodes.
	Editing bool
	// Animate runs the scene's tweens; the engine moves it on every frame.
	Animate animation.Player

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func Build(files fs.FS, description *scenefile.Scene) (*Scene, error)` | Build makes a scene from its description, finding its models in a project's files. |
| `func Load(files fs.FS, file string) (*Scene, error)` | Load reads a scene file from a project's files. |
| `func NewScene(files fs.FS) *Scene` | NewScene is an empty scene, whose models come from a project's files. |
| `func (owner *Scene) Camera() models.Camera` | Camera is how the scene is seen: from its main camera, or from a little above when it has none. |
| `func (owner *Scene) CameraBounds(camera *Node) (space.Box, bool)` | CameraBounds is the box a camera's view stays inside while it follows: its own, or the tile maps' together, with no top (an infinite one) unless the camera has a ceiling. |
| `func (owner *Scene) Describe() *scenefile.Scene` | Describe writes the scene back as its file keeps it. |
| `func (owner *Scene) Draw2D(list *draw2d.List, area draw2d.Rect, view View2D)` | Draw2D draws a 2D scene into an area of the screen, as a view sees it: the sky's colour, then its sprites and shapes from the furthest to the nearest, lit by its lights (lights.go). |
| `func (owner *Scene) DrawContent2D(list *draw2d.List, area draw2d.Rect, view View2D)` | DrawContent2D draws a 2D scene's sprites and shapes without its sky, lit by its lights: over something of the program's own (the editor's grid). |
| `func (owner *Scene) DrawContent2DUnlit(list *draw2d.List, area draw2d.Rect, view View2D)` | DrawContent2DUnlit draws a 2D scene's sprites and shapes as DrawContent2D does, but without its lights, and a sprite light's sprite as a sprite: what the editor shows to see the colours. |
| `func (owner *Scene) Find(path string) (*Node, bool)` | Find finds a node by its path from the root, as "Car/Wheel". |
| `func (owner *Scene) Forget(file string)` | Forget makes the scene read a model or a picture again the next time a node shows it: after the file changed. |
| `func (owner *Scene) Frames(picture string, sheet *scenefile.SpriteSheet) ([]*draw2d.Image, error)` | Frames are a picture cut into a sheet's frames, numbered from the top left, read and cut once. |
| `func (owner *Scene) Is2D() bool` | Is2D tells whether the scene is 2D. |
| `func (owner *Scene) MainCamera() (*Node, bool)` | MainCamera is the camera the scene is seen through: the first marked Main, or the first camera. |
| `func (owner *Scene) NewNode(description scenefile.Node) (*Node, error)` | NewNode makes a node of the scene from a description (and its children), not yet in the scene: Add it to a node. |
| `func (owner *Scene) Physics() *physics.World` | Physics is the scene's physics world while it plays; nil for a 3D scene, or one not started. |
| `func (owner *Scene) Picture(file string) (*draw2d.Image, error)` | Picture is a picture of the project's, read once: what the editor's Art workspace shows. |
| `func (owner *Scene) PlayFrame(seconds float64, keys *input.State)` | PlayFrame is a frame of the game's main scene: the calls asked for between frames run, the actions are read (opening and closing the surfaces they open, and the whole screen), the frame hook runs, then the scene moves on (Update). |
| `func (owner *Scene) SetAmbient(ambient *scenefile.Ambient)` | SetAmbient sets the scene's ambient light (nil for none): in 2D the light that reaches everything; in 3D how strong its sky light is, from 0 (none) to 2 (twice the default). |
| `func (owner *Scene) SetMode(mode scenefile.Mode)` | SetMode makes the scene 3D or 2D: its nodes show their shapes and models in 3D, or their sprites and flat shapes in 2D. |
| `func (owner *Scene) SetSky(colour string)` | SetSky sets the background colour ("" for a daylight sky). |
| `func (owner *Scene) SpriteFile(file string) (*scenefile.Sprite, error)` | SpriteFile is a sprite's file of the project's, read once. |
| `func (owner *Scene) Start() []error` | Start makes the scene's behaviours and runs their Start, once: the engine does it before the first frame. |
| `func (owner *Scene) Sync()` | Sync brings what go-renderer draws up to date with the nodes' transforms. |
| `func (owner *Scene) TakeReplacement() (next *Scene, replaced bool, err error)` | TakeReplacement builds and starts the scene a behaviour asked for in the last frame, and forgets the asking; replaced is false when nothing was asked. |
| `func (owner *Scene) Tileset(node *Node) ([]*draw2d.Image, error)` | Tileset is the pictures of a tile map's tileset, cut once, numbered as its tiles are. |
| `func (owner *Scene) TilesetColumns(node *Node) int` | TilesetColumns is how many tiles across a tile map's tileset picture is: the palette shows them in its rows, as they were drawn. |
| `func (owner *Scene) TilesetFile(file string) (*scenefile.Tileset, error)` | TilesetFile is a tileset's file of the project's, read once. |
| `func (owner *Scene) Time() float64` | Time is the seconds the scene has been playing. |
| `func (owner *Scene) Update(seconds float64, keys *input.State)` | Update moves the scene on by some seconds: behaviours that were just made start, every behaviour updates, the bodies of a 2D scene move, the tweens move, and what go-renderer draws follows. |
| `func (owner *Scene) View2D() View2D` | View2D is what the scene's main camera sees, or, without one, ten metres around the origin. |
| `func (owner *Scene) World() *scene.Scene` | World is what go-renderer draws, transforms brought up to date. |

### Settings

Settings are a project's, kept in its engine.json.

```go
type Settings struct {
	Name       string `json:"name"`
	StartScene string `json:"startScene"` // the scene the game opens, as "scenes/main.scene.json"
	// CodeEditor is the command the editor opens the project's code with ("code", "zed", "nvim"…);
	// empty tries $VISUAL, $EDITOR and the usual ones.
	CodeEditor string `json:"codeEditor,omitempty"`
	// ApplicationID is the game's application id, as io.example.game, which each target turns into
	// what it needs (Android's package); empty makes one from its name.
	ApplicationID string `json:"applicationId,omitempty"`
	// DeviceAddress is the address on the network of the device the game is sent to (a 3DS's);
	// empty looks for it.
	DeviceAddress string `json:"deviceAddress,omitempty"`
	// SceneMenu lists the game's scenes on a device's second screen (a 3DS's), to play any of them
	// with a touch: a level select.
	SceneMenu bool `json:"sceneMenu,omitempty"`
	// KeepStart gives a device's start button to the game (a pause, read as ButtonStart) instead of
	// ending it, where the start button ends games (a 3DS's START; its HOME menu still ends it).
	KeepStart bool `json:"keepStart,omitempty"`
	// Actions are what the player does, by name, each bound to inputs per target: "fire":
	// {"desktop": ["Space", "Pad A"], "nintendo3ds": ["A"]} (Input.Action, inputwords.go).
	Actions map[string]map[string][]string `json:"actions,omitempty"`
	// Axes are the actions that move an axis (-1 to 1) instead of pressing, by name: an action
	// with no inputs yet says so here, as its inputs cannot.
	Axes []string `json:"axes,omitempty"`
	// TouchControls are the sticks and buttons drawn on a touch screen (a phone's), as the editor
	// places them; none keeps the engine's four touch buttons (touchcontrols.go).
	TouchControls []TouchControl `json:"touchControls,omitempty"`
	// Surfaces place the game's surfaces (SetSurface) on each target, by surface and then by
	// target: "map": {"nintendo3ds": {"place": "second-screen"}, "desktop": {"place": "over",
	// "align": "top-right"}} (surfaces.go).
	Surfaces map[string]map[string]SurfacePlace `json:"surfaces,omitempty"`
	// Window says, per target with windows (desktop, web), whether the game starts "windowed" (the
	// default) or "fullscreen" (fullscreen.go).
	Window map[string]string `json:"window,omitempty"`
}
```

| Function or method | |
|---|---|
| `func ReadSettings(files fs.FS) (Settings, error)` | ReadSettings reads a project's settings. |
| `func (settings *Settings) UnmarshalJSON(data []byte) error` | UnmarshalJSON reads settings, taking the old names of two of them when the new ones are missing. |

### Starter

A behaviour is the game's own code attached to a node: a struct of the
game's, registered by name, whose exported fields the scene file fills (by
their JSON names). It does what it does through the methods it has:

```go
type Spin struct {
	Speed float64 `json:"speed"` // degrees a second
}
func (spin *Spin) Update(context *engine.Context) {
	context.Node.Transform.Turn(space.Up, spin.Speed*context.Delta)
}
```

Start, when the behaviour has it, runs once before the first Update.

```go
type Starter interface{ Start(context *Context) }
```

### SurfacePlace

SurfacePlace is where a surface goes on one target, as engine.json says.

```go
type SurfacePlace struct {
	Place string `json:"place"`
	// Align is where an over or window surface sits in the view: top-left, top, top-right, left,
	// center, right, bottom-left, bottom or bottom-right. Empty lays an over surface out on the
	// whole view, as a game's interface is, and a window in the top left corner.
	Align string `json:"align,omitempty"`
	// Width and Height are its size in pixels; 0 is its content's.
	Width  float32 `json:"width,omitempty"`
	Height float32 `json:"height,omitempty"`
	// Action opens and closes a window or a sheet; Open shows it when the game starts.
	Action string `json:"action,omitempty"`
	Open   bool   `json:"open,omitempty"`
}
```

### TouchControl

TouchControl is a stick or a button drawn on a touch screen: its name,
what it is, where its centre sits (X across the screen and Y down it, from 0
to 1) and its size (its width, as a share of the screen's shorter side).

```go
type TouchControl struct {
	Name string    `json:"name"`
	Kind TouchKind `json:"kind"`
	X    float32   `json:"x"`
	Y    float32   `json:"y"`
	Size float32   `json:"size"`
}
```

### TouchKind

TouchKind is what a touch control is.

```go
type TouchKind string
```

```go
const (
	TouchButton TouchKind = "button" // pressed while a finger is on it
	TouchStick  TouchKind = "stick"  // moves two axes as a finger drags it
)
```

The kinds of touch control.

### Toucher

Toucher is a behaviour that hears when its node's body begins touching
another's: a trigger and the moving body that enters it. Both nodes'
behaviours hear it, each with the other node.

```go
type Toucher interface {
	Touched(context *Context, other *Node)
}
```

### Turn

Turn is how fast a device turns, in degrees a second: about the axis across
its screen (Pitch), up through it (Yaw) and out of it (Roll).

```go
type Turn struct{ Roll, Pitch, Yaw float64 }
```

### Updater

A behaviour is the game's own code attached to a node: a struct of the
game's, registered by name, whose exported fields the scene file fills (by
their JSON names). It does what it does through the methods it has:

```go
type Spin struct {
	Speed float64 `json:"speed"` // degrees a second
}
func (spin *Spin) Update(context *engine.Context) {
	context.Node.Transform.Turn(space.Up, spin.Speed*context.Delta)
}
```

Start, when the behaviour has it, runs once before the first Update.

```go
type Updater interface{ Update(context *Context) }
```

### View2D

View2D is what a 2D camera sees: the point at the middle of the screen and
how many metres fit between the screen's top and bottom.

```go
type View2D struct {
	Centre space.Vector
	Height float64
}
```

| Function or method | |
|---|---|
| `func (view View2D) FromScreen(area draw2d.Rect, x, y float32) space.Vector` | FromScreen is the point of the world under a point of the screen, at Z 0. |
| `func (view View2D) MetresAcross(area draw2d.Rect) float64` | MetresAcross is how many metres an area shows from side to side. |
| `func (view View2D) ToScreen(area draw2d.Rect, point space.Vector) (x, y float32)` | ToScreen is where a point of the world falls in an area of the screen. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
