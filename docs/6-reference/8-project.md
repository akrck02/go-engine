# project

```go
import "github.com/akrck02/go-engine/project"
```

Package project is a game made with go-engine, as a folder: a Go module whose
main runs the engine, the game's own Go code beside it, its scenes in scenes/
and what it imports in assets/. It is an ordinary Go module, so any editor with
Go's language server (Zed, Neovim, Visual Studio Code, GoLand…) edits the code,
and `go run .` plays it.

### Functions

| Function | |
|---|---|
| `func RunInTerminal(words []string) error` | RunInTerminal runs a command in a new terminal window, left open until the person presses Enter: its questions (a password) and what it says are theirs to answer and read. |

### Constants and variables

```go
const (
	ScenesFolder = "scenes"
	AssetsFolder = "assets"
	ModelsFolder = "assets/models"
	SceneSuffix  = ".scene.json"
)
```

Folders in a project.

```go
const EditorFile = "editor.go"
```

EditorFile is the file that installs the editor in a project's game.

```go
const LatestEngine = "main"
```

LatestEngine is the branch go-engine's latest version is taken from.

```go
const ReleaseWorkflowFile = ".github/workflows/release.yml"
```

ReleaseWorkflowFile is where the workflow goes in the game's folder.

```go
var (
	NotAProject                 = failure.Define("project.not-a-project", "the folder is not a go-engine project; make one with: go-engine new <folder>")
	FolderNotEmpty              = failure.Define("project.folder-not-empty", "a new project needs an empty folder")
	SceneExists                 = failure.Define("project.scene-exists", "a scene of that name exists already")
	ImportFormat                = failure.Define("project.import-format", "only .glb models can be imported for now (export glTF as binary .glb)")
	ImportFailed                = failure.Define("project.import-failed", "the file could not be copied into the project")
	CodeEditorNotFound          = failure.Define("project.code-editor-not-found", "no code editor found: set codeEditor in engine.json (\"code\", \"zed\", \"nvim\"…) or $VISUAL")
	TerminalNotFound            = failure.Define("project.terminal-not-found", "the code editor runs in a terminal, and no terminal was found")
	SettingsNotSaved            = failure.Define("project.settings-not-saved", "engine.json could not be written")
	ApplicationIDWrong          = failure.Define("project.application-id-wrong", "an application id is two words or more joined by dots, each starting with a letter, as io.example.game")
	TargetCannot                = failure.Define("project.target-cannot", "the target cannot do that")
	EngineNotRead               = failure.Define("project.engine-not-read", "the version of go-engine the project requires could not be read from its go.mod")
	ReleaseNeedsPublishedEngine = failure.Define("project.release-needs-published-engine", "the game's go.mod uses go-engine from a folder of this computer, which GitHub's machines do not have: update it first (go-engine -update)")
	EngineNotUpdated            = failure.Define("project.engine-not-updated", "go-engine could not be updated: check the network, and that the version exists")
	NoAndroidDevice             = failure.Define("project.no-android-device", "adb sees no Android device: plug a phone in with USB debugging on, or start the emulator")
)
```

Error codes of the project package.

```go
var Targets = []Target{Desktop, Nintendo3DS, Android}
```

Targets are every target, in the order the editor shows them.

### AssetEntry

AssetEntry is a file or a folder of the project's assets, as the editor's
tree shows them.

```go
type AssetEntry struct {
	Name     string
	Path     string // in the project, as "assets/models/house.glb"
	Folder   bool
	Children []AssetEntry
}
```

### AssetKind

AssetKind is what a file of the assets is, by its extension.

```go
type AssetKind int
```

```go
const (
	OtherAsset AssetKind = iota
	ModelAsset           // .glb
	ImageAsset           // .png, .jpg: a sprite
	SoundAsset           // .wav, .ogg, .mp3
)
```

| Function or method | |
|---|---|
| `func KindOf(file string) AssetKind` | KindOf tells what an asset file is. |

### Behaviour

Behaviour is a behaviour the project's code defines: a struct type with one
of the methods the engine calls (behaviourMethods), as the editor offers it.

```go
type Behaviour struct {
	Name    string
	Doc     string // the type's comment, first sentence
	File    string // where it is defined, in the project
	Fields  []Field
	Methods []string // the methods the engine calls that it has: Start, Update, Touched
}
```

### Built

Built is what a build made: the program, and the notices of what it carries.

```go
type Built struct{ Program, Notices string }
```

### Field

Field is a field the scene can set: an exported field, by its JSON name.

```go
type Field struct {
	Name    string // as the scene names it: the JSON name, or the field's own
	Type    string // as written in the code: float64, string, bool, space.Vector…
	Comment string
}
```

### Project

Project is a game's folder.

```go
type Project struct {
	Folder   string
	Settings engine.Settings
	// Tools are the places of the tools that build for consoles and phones, kept on this
	// computer for every project.
	Tools build.Tools
	// ToolsUnread is why the places kept could not be read, when they could not: the builds look
	// for the tools as if none were kept, and saving writes them anew.
	ToolsUnread error
}
```

| Function or method | |
|---|---|
| `func Create(folder, name string) (*Project, error)` | Create makes a new project in a folder (which must be empty or not exist): its settings, a scene with a ground, a cube and a camera, a main and a first behaviour, and the code editors' settings. |
| `func Open(folder string) (*Project, error)` | Open opens a project's folder. |
| `func (project *Project) ApplicationID() string` | ApplicationID is the game's application id, as io.example.game, which each target turns into what it needs (Android's package name): the settings' ApplicationID, or one made from its name. |
| `func (project *Project) AssetTree() AssetEntry` | AssetTree reads the project's assets folder as a tree: folders first, then files, by name; hidden files (as .gitkeep) left out. |
| `func (project *Project) Behaviours() []Behaviour` | Behaviours reads the project's Go files (not compiling them) and finds its behaviours, so the editor can offer them while the code is being written in another editor. |
| `func (project *Project) BehavioursUsed() map[string][]Use` | BehavioursUsed are the behaviours the project's scenes give their nodes, by name, with the nodes that have each. |
| `func (project *Project) Build(target Target) (Built, error)` | Build builds the game for a target into build/: a program, a .3dsx, an .apk. |
| `func (project *Project) BuildEditor() error` | BuildEditor builds the game with the editor in it, for go-engine to start. |
| `func (project *Project) CodeChangedSince(built time.Time) bool` | CodeChangedSince tells whether a Go file of the project changed after a time: the game the editor runs in is older than its code. |
| `func (project *Project) EditorProgram() string` | EditorProgram is where the game built with the editor in it goes. |
| `func (project *Project) EngineFlags() []string` | EngineFlags tell a build which commit a local copy of Go Engine is at (one the project's go.mod replaces with a folder, or a go.work uses), so the program can say it (engine.Version): a local copy's build information does not. |
| `func (project *Project) EngineVersion() (version, localCopy string, err error)` | EngineVersion is the version of go-engine the project's go.mod requires, and the folder that stands in for it when go.mod replaces it with a local copy (empty when none does). |
| `func (project *Project) EnsureEditor() (bool, error)` | EnsureEditor adds editor.go to a project made before it existed; it tells whether it did. |
| `func (project *Project) Files() fs.FS` | Files are the project's files, as the engine reads them. |
| `func (project *Project) Import(source string) (string, error)` | Import copies a file into the project's assets, and returns its path in the project. |
| `func (project *Project) Models() []string` | Models are the project's models, as "assets/models/tree.glb". |
| `func (project *Project) NewScene(name string, mode scenefile.Mode) (string, error)` | NewScene adds a scene, 3D or 2D, with a ground and a camera, and returns its file. |
| `func (project *Project) OpenCode() (string, error)` | OpenCode opens the project's folder in the code editor: the settings' CodeEditor, $VISUAL, $EDITOR, or the first of Zed, Visual Studio Code and GoLand found. |
| `func (project *Project) Path(file string) string` | Path is where a file of the project is on disk, from its path in the project. |
| `func (project *Project) PlayInEmulator(target Target) (*exec.Cmd, error)` | PlayInEmulator builds the game for a target and returns the command that opens it in an emulator of the target, not started. |
| `func (project *Project) PlayOn(target Target) (*exec.Cmd, error)` | PlayOn builds the game for a target and plays it there: in a window on this computer, sent to a 3DS's Homebrew Launcher, installed and started on the Android device adb sees. |
| `func (project *Project) ReadScene(file string) (*scenefile.Scene, error)` | ReadScene reads a scene file of the project. |
| `func (project *Project) SaveSettings() error` | SaveSettings writes the project's settings, when its Android application id has the form one has. |
| `func (project *Project) Scenes() []string` | Scenes are the project's scene files, as "scenes/main.scene.json". |
| `func (project *Project) StopOn(target Target) error` | StopOn stops the game where PlayOn played it, when stopping its command is not enough: on an Android device, where the game goes on after its log is closed. |
| `func (project *Project) Tidy() error` | Tidy brings the project's go.mod up to date with its code and with the go-engine it builds against, as `go mod tidy` does: an engine of a newer version may need other versions of its libraries. |
| `func (project *Project) UpdateEngine(version string) (string, error)` | UpdateEngine requires go-engine at a version (a branch as LatestEngine, a tag or a commit) with its replaces, and tidies go.mod; it returns the version required now. |
| `func (project *Project) UsesOf(asset string) []Use` | UsesOf finds where an asset is used: the nodes of every scene that show it (as a sprite, a model or a tile map's tileset), and the sprites and tilesets cut from it. |
| `func (project *Project) WriteReleaseWorkflow() (string, error)` | WriteReleaseWorkflow writes the workflow that builds the game for Windows, macOS and Linux on each version tag (v1.0.0), and the notices it attaches. |
| `func (project *Project) WriteScene(file string, scene *scenefile.Scene) error` | WriteScene writes a scene file of the project. |

### Target

Target is a place a game runs.

```go
type Target int
```

```go
const (
	// Desktop is this computer.
	Desktop Target = iota
	// Nintendo3DS is the Nintendo 3DS, through the Homebrew Launcher.
	Nintendo3DS
	// Android is a phone or a tablet with Android.
	Android
)
```

| Function or method | |
|---|---|
| `func (target Target) String() string` | String is the target's name, or "an unknown target" for a number that is none. |

### Use

Use is a place an asset is used: a node of a scene, by its path, or a
sprite's or tileset's file that cuts a picture (Node is then empty).

```go
type Use struct {
	File string // the scene, or the sprite's or tileset's file
	Node string // the node's path in the scene, as "Harbour/Lamp"
}
```

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
