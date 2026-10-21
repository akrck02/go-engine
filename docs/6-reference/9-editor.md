# editor

```go
import "github.com/akrck02/go-engine/editor"
```

Package editor is go-engine's editor: a window over a project's scene, in 3D,
where its nodes are added, chosen with the mouse, moved, turned and scaled,
given models from the project's assets and the behaviours its Go code defines,
and the game is played and built. This file holds what the editor does; view.go
draws its window with go-ui and viewport.go its 3D view.

### Functions

| Function | |
|---|---|
| `func FieldText(value any) string` | FieldText writes a field's value for the inspector. |
| `func FieldValue(typed string) any` | FieldValue reads a field as typed in the inspector: a number, true or false, or else text. |
| `func Install()` | Install makes the game open the editor when go-engine starts it to edit it. |
| `func ReadFrames(written string, count int) ([]int, error)` | ReadFrames reads frames written as numbers and runs, separated by commas, of a sheet of a count of frames: "0-3, 6" is 0, 1, 2, 3 and 6. |
| `func RunHere() error` | RunHere opens the editor on the project in the folder the game runs in, with the window's flags from the command line. |

### Constants and variables

```go
var (
	RestartFailed  = failure.Define("editor.restart-failed", "the editor could not start again with the project's new code")
	ToolsInUse     = failure.Define("editor.tools-in-use", "the tools cannot change while the game builds or plays: wait for it to end, or stop it")
	NoLicence      = failure.Define("editor.no-licence", "the tool has no licence to accept before it is installed")
	FramesWrong    = failure.Define("editor.frames-wrong", "an animation's frames are not numbers and runs of the sheet's frames, as 0-3, 6")
	ArtNotFound    = failure.Define("editor.art-not-found", "the sprite or tileset is not a file of the project's")
	ArtNotSaved    = failure.Define("editor.art-not-saved", "the sprite or tileset could not be written")
	ArtWrongKind   = failure.Define("editor.art-wrong-kind", "the file is not a sprite's or a tileset's")
	ArtUnsaved     = failure.Define("editor.art-unsaved", "the sprite or tileset open in Art has changes not saved: save it first")
	NoOwnSheet     = failure.Define("editor.no-own-sheet", "the node does not cut its picture into frames itself")
	UnsavedChanges = failure.Define("editor.unsaved-changes", "the scene has changes not saved: save it first (Ctrl+S)")
)
```

Error codes of the editor package.

```go
var Kinds = []string{"Empty", "Cube", "Sphere", "Plane", "Cylinder", "Camera", "Light"}
```

Kinds of node the editor adds.

```go
var Kinds2D = []string{"Empty", "Square", "Circle", "Tile map", "Camera", "Light"}
```

Kinds2D are the kinds of node the editor adds to a 2D scene.

```go
var Tools = []Tool{SelectTool, PaintTool, EraseTool, FillTool, HandTool}
```

Tools are the tools in the order of the rail.

```go
var Workspaces = []Workspace{ProjectWorkspace, ArtWorkspace, LevelWorkspace, CodeWorkspace, PlayWorkspace, BuildWorkspace}
```

Workspaces are the workspaces in the order of the bar.

### BuildResult

BuildResult is what the last build for a target made, or why it did not.

```go
type BuildResult struct {
	Program string
	Bytes   int64
	At      time.Time
	Err     error
	Log     []string
}
```

### Editor

Editor is a project open in the editor, and the scene shown.

```go
type Editor struct {
	Project   *project.Project
	SceneFile string
	Scene     *engine.Scene
	Selected  *engine.Node   // the node chosen, which the inspector shows
	Chosen    []*engine.Node // other nodes chosen with it (Shift), which act with it
	Saved     bool

	// Tool is what the left button does in the 2D view (tools.go); Brush is the tile of the
	// tileset that painting and filling put down.
	Tool  Tool
	Brush int

	// Paused stops the game played in the view; StepFrame moves it on one frame. Target is the
	// shape of the screen it is shown at.
	Paused bool

	Watched string // the path of the node of the scene playing that Play's watch shows
	Target  PlayTarget

	// Wake asks the window for a frame, when background work has something to show.
	Wake func()
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func Open(opened *project.Project) (*Editor, error)` | Open opens a project at its start scene. |
| `func (editor *Editor) AcceptLicence()` | AcceptLicence installs the tool whose licence the person read (Licence), with that licence accepted, and sets its place. |
| `func (editor *Editor) Add(kind string) *engine.Node` | Add adds a node of a kind ("Cube", "Camera"…) under the selected node's parent, after it (or at the end of the scene), and selects it. |
| `func (editor *Editor) AddAnimation() int` | AddAnimation adds an animation of the first frame to the sheet being edited, with a name of its own, and tells where it is in the sheet's list (-1 without a sheet). |
| `func (editor *Editor) AddFrame(index, frame int)` | AddFrame puts a frame of the sheet at the end of an animation: clicking frames on the picture makes an animation of them, in the order clicked. |
| `func (editor *Editor) AddModel(model string) *engine.Node` | AddModel adds a node showing one of the project's models, and selects it. |
| `func (editor *Editor) AddSprite(picture string) *engine.Node` | AddSprite adds a node showing one of the project's pictures, in a 2D scene, and selects it. |
| `func (editor *Editor) AdvancePlay(seconds float64, keys *input.State)` | AdvancePlay moves the scene played in the view on a frame, and goes on to the scene the game asked for (a level started again, the next one); one that cannot be read is reported and the view stays where it is. |
| `func (editor *Editor) AllChosen() []*engine.Node` | AllChosen are the nodes chosen, the inspector's first. |
| `func (editor *Editor) ArtOpen() (file string, isSprite, isTileset bool)` | ArtOpen is the file open in Art, and whether it is a sprite or a tileset. |
| `func (editor *Editor) ArtSaved() bool` | ArtSaved tells whether the file open in Art has no changes left to save. |
| `func (editor *Editor) AttachBehaviour(name string)` | AttachBehaviour gives the selected node a behaviour of the project's, its fields empty. |
| `func (editor *Editor) BehaviourUses(name string) []project.Use` | BehaviourUses are the nodes of the project's saved scenes that have a behaviour. |
| `func (editor *Editor) Behaviours() []project.Behaviour` | Behaviours are the behaviours the project's code defines, as last read. |
| `func (editor *Editor) BuildFor(target project.Target)` | BuildFor builds the game for a target in the background; its result is kept when it ends. |
| `func (editor *Editor) CanRedo() bool` |  |
| `func (editor *Editor) CanUndo() bool` | CanUndo and CanRedo tell whether there is a step to go back or forward to. |
| `func (editor *Editor) Changed()` | Changed marks the scene as changed since it was saved, and keeps it for undo. |
| `func (editor *Editor) ClearFrames(index int)` | ClearFrames empties an animation, to click its frames again. |
| `func (editor *Editor) CloseArt() error` | CloseArt closes the file open in Art, which must have no changes left to save. |
| `func (editor *Editor) CloseNotice(index int)` | CloseNotice lets go of a notice, by its place among Notices. |
| `func (editor *Editor) CodeChanged() bool` | CodeChanged tells whether the project's code changed since the game the editor runs in was built: Reload brings the changes in. |
| `func (editor *Editor) DeclineLicence()` | DeclineLicence leaves the tool whose licence waits uninstalled. |
| `func (editor *Editor) Delete()` | Delete removes the nodes chosen and what hangs from them; the node after the inspector's is chosen next. |
| `func (editor *Editor) DetachBehaviour(index int)` | DetachBehaviour takes a behaviour off the selected node. |
| `func (editor *Editor) Duplicate()` | Duplicate copies the nodes chosen, and what hangs from them, each beside its original; the copies are chosen. |
| `func (editor *Editor) FillTileAt(point space.Vector, tile int) bool` | FillTileAt paints a tile (scenefile.EmptyTile erases) over the region of the chosen tile map that holds the same tile as the cell at a point, cell by cell to the sides and up and down; it tells whether the map changed. |
| `func (editor *Editor) FindBehaviour(name string) (project.Behaviour, bool)` | Behaviour finds what the code says of a behaviour. |
| `func (editor *Editor) Followable(camera *engine.Node) []string` | Followable are the paths of the nodes a camera may follow: all but itself. |
| `func (editor *Editor) GiveBody(kind scenefile.BodyKind)` | GiveBody gives the chosen node a body of a kind, or changes the kind of the one it has. |
| `func (editor *Editor) GiveLight(kind scenefile.LightKind)` | GiveLight makes the chosen node a light of a kind, or changes the kind of the one it has. |
| `func (editor *Editor) Import(source string)` | Import copies a model into the project and places it in the scene. |
| `func (editor *Editor) ImportModel(source string)` | ImportModel copies a model into the project's assets, without placing it: the Project grid shows it, to place in a scene. |
| `func (editor *Editor) InstallEmulator()` | InstallEmulator installs Azahar's latest release for this system and sets the 3DS emulator's place. |
| `func (editor *Editor) InstallTinyGo()` | InstallTinyGo installs tinygo-3ds's release for this system and sets TinyGo's place. |
| `func (editor *Editor) Installing() (tool string, done, total int64, busy bool)` | Installing tells which tool is being installed, and how much of its download has arrived (total is 0 when unknown). |
| `func (editor *Editor) IsChosen(node *engine.Node) bool` | IsChosen tells whether a node is among those chosen. |
| `func (editor *Editor) KindsHere() []string` | KindsHere are the kinds of node the scene shown takes. |
| `func (editor *Editor) LastBuild(target project.Target) (BuildResult, bool)` | LastBuild is a target's last build in this session. |
| `func (editor *Editor) LeavePlay()` | LeavePlay goes back from Play to the workspace it was entered from. |
| `func (editor *Editor) Licence() (PendingLicence, bool)` | Licence is the licence waiting for the person to accept it or not, if one is. |
| `func (editor *Editor) LightChanged()` | LightChanged keeps a change to a light or the ambient, and shows it at once. |
| `func (editor *Editor) LookForTools(tools build.Tools)` | LookForTools looks for every tool where some places say (the ones typed, before saving them). |
| `func (editor *Editor) MakeSprite(picture string, columns, rows int) (string, error)` | MakeSprite makes a picture of the project a sprite of columns and rows of frames, in a file beside it ("lamp.png" makes "lamp.sprite.json"), and opens it; a sprite already made is opened. |
| `func (editor *Editor) MakeSpriteOfNode() (string, error)` | MakeSpriteOfNode moves the chosen node's own sheet into a sprite's file beside its picture, makes the node show that file, and opens it in Art: a scene written before sprites had files of their own is brought up to date one node at a time. |
| `func (editor *Editor) MakeTileset(picture string, tileSize int) (string, error)` | MakeTileset makes a picture of the project a tileset of square tiles of a size, in pixels, in a file beside it, and opens it; a tileset already made is opened. |
| `func (editor *Editor) MissingBehaviours() []MissingBehaviour` | MissingBehaviours are the behaviours the project's saved scenes use that its code does not define, by name: a typo in a scene, or code not written yet. |
| `func (editor *Editor) MoveDown()` |  |
| `func (editor *Editor) MoveTo(node *engine.Node, world space.Vector)` | MoveTo puts a node at a point of the world, wherever it hangs from. |
| `func (editor *Editor) MoveUp()` | MoveUp and MoveDown move the selected node among its siblings. |
| `func (editor *Editor) Nest()` | Nest hangs the selected node from the sibling before it, keeping where it is in the world. |
| `func (editor *Editor) Notices() []Notice` | Notices are what the editor said that is still shown, the oldest first. |
| `func (editor *Editor) OpenArt(file string) error` | OpenArt opens a sprite's or a tileset's file of the project in Art; the file open before must have no changes left to save. |
| `func (editor *Editor) OpenScene(file string) error` | OpenScene shows another scene of the project. |
| `func (editor *Editor) PaintTileAt(point space.Vector, tile int) bool` | PaintTileAt puts a tile (scenefile.EmptyTile erases) in the chosen tile map's cell at a point of the world; it tells whether the map changed. |
| `func (editor *Editor) PaintingHere() bool` | PaintingHere tells whether the 2D view paints tiles now: a tile tool is taken and a tile map is chosen. |
| `func (editor *Editor) PaletteColumns() int` | PaletteColumns is how many tiles across the chosen tile map's tileset is, as it was drawn: the palette keeps that shape, so corners and edges stay beside each other. |
| `func (editor *Editor) Pick(origin, direction space.Vector) (*engine.Node, bool)` | Pick finds the node a ray from the camera meets first: the one whose box it enters nearest. |
| `func (editor *Editor) PickFlat(point space.Vector, reach float64) (*engine.Node, bool)` | PickFlat finds the node drawn on top at a point of a 2D scene: a sprite or a shape there, or an empty node or a camera within a share of a metre of it. |
| `func (editor *Editor) PlayHere()` | PlayHere plays the scene in the view, with the behaviours compiled into the game the editor runs in; the scene edited stays as it was. |
| `func (editor *Editor) PlayInEmulator(target project.Target)` | PlayInEmulator builds the game for a target and opens it in the target's emulator; Stop closes it. |
| `func (editor *Editor) PlayLog() []string` | PlayLog is what the game and the engine said while the scene played in the view, oldest first. |
| `func (editor *Editor) PlayOn(target project.Target)` | PlayInWindow saves the scene, builds the game and runs it in a window of its own, in the PlayOn builds the game for a target and plays it there (project.Project.PlayOn): in a window, on the device, or sent to the 3DS. |
| `func (editor *Editor) PlayingHere() (*engine.Scene, bool)` | PlayingHere is the scene played in the view, while it is. |
| `func (editor *Editor) PlayingOn() (project.Target, bool)` | PlayingOn tells where the game plays now, when it does. |
| `func (editor *Editor) ReadBehaviours()` | ReadBehaviours reads the project's code again for its behaviours: after it was edited. |
| `func (editor *Editor) ReadLicence(tool string)` | ReadLicence reads a tool's licence in the background, for the person to read and accept before the tool is installed (Licence). |
| `func (editor *Editor) Redo()` | Redo takes the scene forward one step, after an undo. |
| `func (editor *Editor) Reload()` | Reload saves the scene, builds the game again with the project's code as it is now, and starts the editor again in it, on the same scene and node. |
| `func (editor *Editor) RemoveAnimation(index int)` | RemoveAnimation takes an animation from the sheet being edited; if it was the one the scene starts with, the first one left starts it. |
| `func (editor *Editor) RemoveFrameAt(index, place int)` | RemoveFrameAt takes the frame at a place out of an animation. |
| `func (editor *Editor) Rename(name string) bool` | Rename gives the selected node a name, if no sibling has it; it tells whether it could. |
| `func (editor *Editor) RenameAnimation(index int, name string) bool` | RenameAnimation names an animation of the sheet being edited; it refuses an empty name or one another animation has, and tells whether it took it. |
| `func (editor *Editor) ResizeSheet(columns, rows int)` | ResizeSheet cuts the picture of the sheet being edited into columns and rows of frames, at least one of each: the sprite open in Art, or else the chosen node's own sheet, made if it has none. |
| `func (editor *Editor) ResizeTileMap(columns, rows int)` | ResizeTileMap gives the chosen tile map columns and rows, at least one of each. |
| `func (editor *Editor) RestartPlay()` | RestartPlay plays the scene again from how it is written in the editor. |
| `func (editor *Editor) Run(starting string, task func() (string, error))` | Run runs a task in the background (a build), saying when it starts and how it ends. |
| `func (editor *Editor) Running() bool` | Running tells whether the game is building or playing. |
| `func (editor *Editor) Save() error` | Save writes the scene shown to its file. |
| `func (editor *Editor) SaveArt() error` | SaveArt writes the file open in Art; the scene shows it changed at once. |
| `func (editor *Editor) SaveTools(tools build.Tools) error` | SaveTools keeps the tools' places on this computer, for the builds and every project, and looks for them again. |
| `func (editor *Editor) Select(node *engine.Node)` | Select chooses a node (nil for none), and lets go of any other chosen with it. |
| `func (editor *Editor) SetAmbient(ambient *scenefile.Ambient)` | SetAmbient gives the scene an ambient light of its own, or takes it away (nil). |
| `func (editor *Editor) SetAnimationFrames(index int, written string) error` | SetAnimationFrames sets the frames an animation shows, written as numbers and runs ("0-3, 6"). |
| `func (editor *Editor) SetArtTileSize(pixels int)` | SetArtTileSize sets how many pixels a side the tiles of the tileset open in Art are. |
| `func (editor *Editor) SetGravityScale(scale float64)` | SetGravityScale sets how much gravity pulls the chosen node's body: 1 fully, 0 not at all. |
| `func (editor *Editor) SetMode(mode scenefile.Mode)` | SetMode makes the scene shown 3D or 2D. |
| `func (editor *Editor) SetSprite(path string) error` | SetSprite makes the chosen node show a picture, or a sprite's file, of the project's. |
| `func (editor *Editor) SetStartsWith(name string)` | SetStartsWith makes the chosen node start with an animation of its sprite; "" starts with the sprite's own first one. |
| `func (editor *Editor) SetTileSize(pixels int)` | SetTileSize sets how many pixels a side the chosen tile map's tiles are, at least one. |
| `func (editor *Editor) SetTileset(picture string)` | SetTileset sets the picture the chosen tile map's tiles are cut from. |
| `func (editor *Editor) SetZonePull(pull *[2]float64)` | SetZonePull gives the chosen trigger a gravity of its own, across and up, or the world's (nil). |
| `func (editor *Editor) ShowWorkspace(workspace Workspace)` | ShowWorkspace changes to a workspace. |
| `func (editor *Editor) ShownWorkspace() Workspace` | ShownWorkspace is the workspace the editor shows. |
| `func (editor *Editor) StartWith(index int)` | StartWith makes an animation the one that plays when the scene starts. |
| `func (editor *Editor) StepFrame()` | StepFrame moves a paused game on by one frame, on the next frame the view draws. |
| `func (editor *Editor) Stop()` | Stop ends the game that is playing. |
| `func (editor *Editor) StopHere()` | StopHere stops the play in the view. |
| `func (editor *Editor) TakeBody()` | TakeBody takes the chosen node's body away. |
| `func (editor *Editor) TakeEvents()` | TakeEvents applies what background work sent since the last frame. |
| `func (editor *Editor) TakeLight()` | TakeLight makes the chosen node a light no more. |
| `func (editor *Editor) TakeSheet()` | TakeSheet shows the chosen node's sprite whole again. |
| `func (editor *Editor) ToggleChoice(node *engine.Node)` | ToggleChoice adds a node to those chosen, or takes it out if it was; taking out the inspector's node gives the inspector the next one. |
| `func (editor *Editor) ToggleSolid(tile int)` | ToggleSolid makes a tile of the chosen map's tileset solid, for bodies to stand on, or not. |
| `func (editor *Editor) ToggleSolidTile(tile int)` | ToggleSolidTile makes a tile of the tileset open in Art solid, or not. |
| `func (editor *Editor) ToolsFound() []ToolFound` | ToolsFound are the tools and where they were found, in the order of buildToolParts: where the tools kept say, until LookForTools looks where others say. |
| `func (editor *Editor) Undo()` | Undo takes the scene back one step. |
| `func (editor *Editor) Unnest()` | Unnest hangs the selected node from its grandparent, after its parent, keeping where it is. |
| `func (editor *Editor) Watch() []WatchLine` | Watch is what the chosen node of the scene played is doing: where it is, its body's velocity and whether it stands on the floor, the animation it plays, and its behaviours' exported fields. |
| `func (editor *Editor) WriteReleaseWorkflow()` | WriteReleaseWorkflow writes the workflow that builds the game for Windows, macOS and Linux on GitHub at each version tag, and says where it went or why it did not. |

### MissingBehaviour

MissingBehaviour is a behaviour the scenes use that the project's code does
not define.

```go
type MissingBehaviour struct {
	Name string
	Uses []project.Use
}
```

### Notice

Notice is something the editor said, and when.

```go
type Notice struct {
	Text   string
	Failed bool
	At     time.Time
}
```

### PendingLicence

PendingLicence is a tool's licence, read and waiting for the person to
accept it or not.

```go
type PendingLicence struct {
	Tool, Text string
}
```

### PlayTarget

PlayTarget is the shape of the screen the game is shown at while it plays in
the view.

```go
type PlayTarget int
```

```go
const (
	// WideScreen is a computer's screen, sixteen wide for nine high.
	WideScreen PlayTarget = iota
	// Nintendo3DSScreens are the Nintendo 3DS's two screens: the top one, 400 × 240 pixels, and
	// the second under it.
	Nintendo3DSScreens
)
```

| Function or method | |
|---|---|
| `func (target PlayTarget) Aspect() float64` | Aspect is the target's width over its height. |

### Tool

Tool is what the left button does in the 2D view.

```go
type Tool int
```

```go
const (
	// SelectTool chooses a node, and drags it.
	SelectTool Tool = iota
	// PaintTool paints the chosen tile map with the brush.
	PaintTool
	// EraseTool takes tiles off the chosen tile map.
	EraseTool
	// FillTool paints a region of the same tile, as far as it goes.
	FillTool
	// HandTool moves the view.
	HandTool
)
```

| Function or method | |
|---|---|
| `func ToolsFor(flat bool) []Tool` | ToolsFor are the tools of a scene: all of them in 2D, choosing and moving the view in 3D. |
| `func (tool Tool) Does() string` | Does says what the tool does, for its tooltip. |
| `func (tool Tool) Icon() string` | Icon is the tool's icon on the rail. |
| `func (tool Tool) Key() input.Key` | Key is the key that takes the tool. |
| `func (tool Tool) Letter() string` | Letter is the tool's key, as written on it. |
| `func (tool Tool) Name() string` | Name is the tool's word. |

### ToolFound

ToolFound is a tool, what needs it, where it was found, and why it was not
or why it will not do (found, but missing parts).

```go
type ToolFound struct {
	Name, Needs string
	Found       string
	Err         error
}
```

### View

View is the editor's window and what it remembers between frames.

```go
type View struct {
	Editor *Editor

	// DarkByDefault starts the editor dark when no mode was kept (the window's -dark).
	DarkByDefault bool

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func NewView(editor *Editor) *View` | NewView shows an editor. |
| `func (view *View) Build(context *ui.Context)` | Build describes the window for a frame. |
| `func (view *View) Options(options window.Options) window.Options` | Options are the window's, with the step that shows the scene and reads what the interface leaves of the mouse and the keys. |

### ViewMode

ViewMode is how the editor's camera shows the scene.

```go
type ViewMode int
```

```go
const (
	ViewPerspective ViewMode = iota
	ViewOrthographic
	ViewIsometric
)
```

### WatchLine

WatchLine is one thing the chosen node is doing while the game plays: what,
and its value.

```go
type WatchLine struct{ Name, Value string }
```

### Workspace

Workspace is one of the editor's screens.

```go
type Workspace int
```

```go
const (
	// ProjectWorkspace gathers the project's assets, scenes and settings.
	ProjectWorkspace Workspace = iota
	// ArtWorkspace makes pictures into sprites and tilesets.
	ArtWorkspace
	// LevelWorkspace places things in a scene and paints its tiles.
	LevelWorkspace
	// CodeWorkspace shows the behaviours the project's Go code defines.
	CodeWorkspace
	// PlayWorkspace plays the scene, to try it.
	PlayWorkspace
	// BuildWorkspace makes the game for each place it runs.
	BuildWorkspace
)
```

| Function or method | |
|---|---|
| `func WorkspaceNamed(name string) (Workspace, bool)` | WorkspaceNamed finds a workspace by its word, in any case: "art" is Art. |
| `func (workspace Workspace) Icon() string` | Icon is the workspace's icon on the bar. |
| `func (workspace Workspace) Key() input.Key` | Key is the key that shows the workspace. |
| `func (workspace Workspace) Name() string` | Name is the workspace's word on the bar. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
