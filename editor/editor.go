// Package editor is go-engine's editor: a window over a project's scene, in 3D, where its nodes
// are added, chosen with the mouse, moved, turned and scaled, given models from the project's
// assets and the behaviours its Go code defines, and the game is played and built. This file holds
// what the editor does; view.go draws its window with go-ui and viewport.go its 3D view.
package editor

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
)

// Editor is a project open in the editor, and the scene shown.
type Editor struct {
	Project   *project.Project
	SceneFile string
	Scene     *engine.Scene
	Selected  *engine.Node
	Saved     bool

	Message string
	Failed  bool

	// Tool is what the left button does in the 2D view (tools.go); Brush is the tile of the
	// tileset that painting and filling put down.
	Tool  Tool
	Brush int

	workspace Workspace // the workspace shown
	before    Workspace // the one shown before it, for leaving Play

	// Paused stops the game played in the view; StepFrame moves it on one frame. Target is the
	// shape of the screen it is shown at.
	Paused   bool
	stepOnce bool
	Target   PlayTarget
	playLog  playLog
	unlisten func() // stops keeping the log, while playing

	behaviours []project.Behaviour // what the project's code defines, read again on demand

	// playingHere is the scene played in the view: a copy of the one edited, made again from its
	// description, with the game's behaviours running.
	playingHere *engine.Scene

	builtAt       time.Time // when the game the editor runs in was built
	codeCheckedAt time.Time
	codeChanged   bool

	running  bool      // building or playing the game
	playing  *exec.Cmd // the game, while it plays
	stopping bool      // Stop was asked: the game's end is no failure
	output   *outputTail
	events   chan func(editor *Editor)
	// Wake asks the window for a frame, when background work has something to show.
	Wake func()
}

// Open opens a project at its start scene.
func Open(opened *project.Project) (*Editor, error) {
	editor := &Editor{Project: opened, events: make(chan func(editor *Editor), 16), workspace: LevelWorkspace, before: LevelWorkspace}
	if err := editor.OpenScene(opened.Settings.StartScene); err != nil {
		return nil, err
	}
	editor.ReadBehaviours()
	return editor, nil
}

// OpenScene shows another scene of the project.
func (editor *Editor) OpenScene(file string) error {
	description, err := editor.Project.ReadScene(file)
	if err != nil {
		return err
	}
	built, err := engine.Build(editor.Project.Files(), description)
	if built == nil {
		return err
	}
	built.Editing = true
	built.Sync()
	editor.Scene, editor.SceneFile, editor.Selected, editor.Saved = built, file, nil, true
	if err != nil {
		editor.say(err.Error(), true)
	} else {
		editor.say("Opened "+file, false)
	}
	return nil
}

// Save writes the scene shown to its file.
func (editor *Editor) Save() error {
	if err := editor.Project.WriteScene(editor.SceneFile, editor.Scene.Describe()); err != nil {
		editor.say(err.Error(), true)
		return err
	}
	editor.Saved = true
	editor.say("Saved "+editor.SceneFile, false)
	return nil
}

// say puts a message on the status line, on one line.
func (editor *Editor) say(message string, failed bool) {
	editor.Message, editor.Failed = strings.Join(strings.Fields(message), " "), failed
}

// Changed marks the scene as changed since it was saved.
func (editor *Editor) Changed() { editor.Saved = false }

// Select chooses a node (nil for none).
func (editor *Editor) Select(node *engine.Node) { editor.Selected = node }

// Kinds of node the editor adds.
var Kinds = []string{"Empty", "Cube", "Sphere", "Plane", "Cylinder", "Camera"}

// SetMode makes the scene shown 3D or 2D.
func (editor *Editor) SetMode(mode scenefile.Mode) {
	if editor.Scene.Mode == mode || (mode == scenefile.Mode3D && !editor.Scene.Is2D()) {
		return
	}
	editor.Scene.SetMode(mode)
	if mode == scenefile.Mode2D && editor.Scene.PixelsPerMetre == 0 {
		editor.Scene.PixelsPerMetre = scenefile.DefaultPixelsPerMetre
	}
	editor.Changed()
	if mode == scenefile.Mode2D {
		editor.say("The scene is 2D: sprites and flat shapes, seen from the front; models are not shown", false)
		return
	}
	editor.say("The scene is 3D", false)
}

// Kinds2D are the kinds of node the editor adds to a 2D scene.
var Kinds2D = []string{"Empty", "Square", "Circle", "Tile map", "Camera"}

// KindsHere are the kinds of node the scene shown takes.
func (editor *Editor) KindsHere() []string {
	if editor.Scene.Is2D() {
		return Kinds2D
	}
	return Kinds
}

// Add adds a node of a kind ("Cube", "Camera"…) under the selected node's parent, after it (or at
// the end of the scene), and selects it.
func (editor *Editor) Add(kind string) *engine.Node {
	description := scenefile.NewNode(kind)
	if editor.Scene.Is2D() {
		switch kind {
		case "Square":
			description.Shape, description.Colour = "cube", "#c8c8c8"
		case "Circle":
			description.Shape, description.Colour = "sphere", "#c8c8c8"
		case "Camera":
			description.Camera = &scenefile.Camera{Projection: scenefile.Orthographic, Height: 10}
		case "Tile map":
			description.Tiles = newTileMap()
		}
		return editor.place(description)
	}
	switch kind {
	case "Empty":
	case "Camera":
		description.Position, description.Rotation = [3]float64{0, 3, 7}, [3]float64{-20, 0, 0}
		description.Camera = &scenefile.Camera{Projection: scenefile.Perspective, FieldOfView: 50}
	default:
		description.Shape, description.Colour = strings.ToLower(kind), "#c8c8c8"
		if kind != "Plane" {
			description.Position[1] = 0.5 // standing on the ground, not half in it
		}
	}
	return editor.place(description)
}

// AddModel adds a node showing one of the project's models, and selects it.
func (editor *Editor) AddModel(model string) *engine.Node {
	base := model[strings.LastIndex(model, "/")+1:]
	base = strings.TrimSuffix(base, base[strings.LastIndex(base, "."):])
	description := scenefile.NewNode(base)
	description.Model = model
	return editor.place(description)
}

// AddSprite adds a node showing one of the project's pictures, in a 2D scene, and selects it.
func (editor *Editor) AddSprite(picture string) *engine.Node {
	base := picture[strings.LastIndex(picture, "/")+1:]
	base = strings.TrimSuffix(base, base[strings.LastIndex(base, "."):])
	description := scenefile.NewNode(base)
	description.Sprite = picture
	return editor.place(description)
}

// PickFlat finds the node drawn on top at a point of a 2D scene: a sprite or a shape there, or
// an empty node or a camera within a share of a metre of it.
func (editor *Editor) PickFlat(point space.Vector, reach float64) (*engine.Node, bool) {
	var found *engine.Node
	editor.Scene.Root.Walk(func(node *engine.Node) bool {
		if node == editor.Scene.Root {
			return true
		}
		if !node.Visible {
			return false
		}
		box := node.Bounds()
		if box.IsEmpty() {
			position := node.World().Position
			if math.Abs(position.X-point.X) <= reach && math.Abs(position.Y-point.Y) <= reach && (found == nil || found.Bounds().IsEmpty()) {
				found = node
			}
			return true
		}
		if point.X >= box.Minimum.X && point.X <= box.Maximum.X && point.Y >= box.Minimum.Y && point.Y <= box.Maximum.Y {
			if found == nil || found.Bounds().IsEmpty() || node.World().Position.Z >= found.World().Position.Z {
				found = node
			}
		}
		return true
	})
	return found, found != nil
}

func (editor *Editor) place(description scenefile.Node) *engine.Node {
	parent, at := editor.Scene.Root, len(editor.Scene.Root.Children())
	if selected := editor.Selected; selected != nil {
		parent, at = selected.Parent(), selected.Index()+1
	}
	description.Name = uniqueName(parent, description.Name)
	node, err := editor.Scene.NewNode(description)
	if err != nil {
		editor.say(err.Error(), true)
	}
	parent.Insert(node, at)
	editor.Scene.Sync()
	editor.Select(node)
	editor.Changed()
	return node
}

// Duplicate copies the selected node, and what hangs from it, beside it.
func (editor *Editor) Duplicate() {
	selected := editor.Selected
	if selected == nil {
		return
	}
	description := selected.Describe()
	description.Name = uniqueName(selected.Parent(), description.Name)
	node, _ := editor.Scene.NewNode(description)
	selected.Parent().Insert(node, selected.Index()+1)
	editor.Scene.Sync()
	editor.Select(node)
	editor.Changed()
}

// Delete removes the selected node and what hangs from it.
func (editor *Editor) Delete() {
	selected := editor.Selected
	if selected == nil {
		return
	}
	parent, index := selected.Parent(), selected.Index()
	selected.Remove()
	editor.Selected = nil
	if siblings := parent.Children(); len(siblings) > 0 {
		editor.Select(siblings[min(index, len(siblings)-1)])
	} else if parent != editor.Scene.Root {
		editor.Select(parent)
	}
	editor.Changed()
}

// MoveUp and MoveDown move the selected node among its siblings.
func (editor *Editor) MoveUp()   { editor.moveAmongSiblings(-1) }
func (editor *Editor) MoveDown() { editor.moveAmongSiblings(1) }

func (editor *Editor) moveAmongSiblings(by int) {
	selected := editor.Selected
	if selected == nil {
		return
	}
	to := selected.Index() + by
	if to < 0 || to >= len(selected.Parent().Children()) {
		return
	}
	selected.Parent().Insert(selected, to)
	editor.Changed()
}

// Nest hangs the selected node from the sibling before it, keeping where it is in the world.
func (editor *Editor) Nest() {
	selected := editor.Selected
	if selected == nil || selected.Index() == 0 {
		return
	}
	editor.reparent(selected, selected.Parent().Children()[selected.Index()-1])
}

// Unnest hangs the selected node from its grandparent, after its parent, keeping where it is.
func (editor *Editor) Unnest() {
	selected := editor.Selected
	if selected == nil || selected.Parent() == editor.Scene.Root {
		return
	}
	parent := selected.Parent()
	editor.reparent(selected, parent.Parent())
	parent.Parent().Insert(selected, parent.Index()+1)
}

// reparent hangs a node from another, keeping where it is in the world.
func (editor *Editor) reparent(node, parent *engine.Node) {
	world := node.World()
	node.Name = uniqueNameExcept(parent, node.Name, node)
	parent.Add(node)
	parentWorld := parent.World()
	if parent == editor.Scene.Root {
		parentWorld = space.NewTransform()
	}
	node.Transform = space.Transform{
		Position: parentWorld.FromParent(world.Position),
		Rotation: world.Rotation.Then(parentWorld.Rotation.Inverse()).Normalized(),
		Scale:    space.NewVector(divide(world.Scale.X, parentWorld.Scale.X), divide(world.Scale.Y, parentWorld.Scale.Y), divide(world.Scale.Z, parentWorld.Scale.Z)),
	}
	editor.Scene.Sync()
	editor.Changed()
}

func divide(value, by float64) float64 {
	if by == 0 {
		return value
	}
	return value / by
}

// Rename gives the selected node a name, if no sibling has it; it tells whether it could.
func (editor *Editor) Rename(name string) bool {
	selected := editor.Selected
	name = strings.TrimSpace(name)
	if selected == nil || name == "" || strings.Contains(name, "/") {
		return false
	}
	for _, sibling := range selected.Parent().Children() {
		if sibling != selected && sibling.Name == name {
			return false
		}
	}
	selected.Name = name
	editor.Changed()
	return true
}

// uniqueName is a name for a new node under a parent: the name itself, or with a number.
func uniqueName(parent *engine.Node, name string) string { return uniqueNameExcept(parent, name, nil) }

func uniqueNameExcept(parent *engine.Node, name string, except *engine.Node) string {
	taken := map[string]bool{}
	for _, child := range parent.Children() {
		if child != except {
			taken[child.Name] = true
		}
	}
	if !taken[name] {
		return name
	}
	base := strings.TrimRight(name, " 0123456789")
	for number := 2; ; number++ {
		candidate := base + " " + strconv.Itoa(number)
		if !taken[candidate] {
			return candidate
		}
	}
}

// Pick finds the node a ray from the camera meets first: the one whose box it enters nearest.
func (editor *Editor) Pick(origin, direction space.Vector) (*engine.Node, bool) {
	var nearest *engine.Node
	best := 1e300
	editor.Scene.Root.Walk(func(node *engine.Node) bool {
		if node == editor.Scene.Root {
			return true
		}
		if !node.Visible {
			return false
		}
		box := node.Bounds()
		if box.IsEmpty() {
			return true
		}
		if distance, hit := box.Hit(origin, direction); hit && distance < best {
			nearest, best = node, distance
		}
		return true
	})
	return nearest, nearest != nil
}

// MoveTo puts a node at a point of the world, wherever it hangs from.
func (editor *Editor) MoveTo(node *engine.Node, world space.Vector) {
	parent := node.Parent()
	if parent == nil || parent == editor.Scene.Root {
		node.Transform.Position = world
	} else {
		node.Transform.Position = parent.World().FromParent(world)
	}
	editor.Scene.Sync()
	editor.Changed()
}

// Import copies a model into the project and places it in the scene.
func (editor *Editor) Import(source string) {
	model, err := editor.Project.Import(source)
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	editor.Scene.Forget(model)
	editor.AddModel(model)
	editor.say("Imported "+model, false)
}

// Behaviours are the behaviours the project's code defines, as last read.
func (editor *Editor) Behaviours() []project.Behaviour { return editor.behaviours }

// ReadBehaviours reads the project's code again for its behaviours: after it was edited.
func (editor *Editor) ReadBehaviours() { editor.behaviours = editor.Project.Behaviours() }

// Behaviour finds what the code says of a behaviour.
func (editor *Editor) FindBehaviour(name string) (project.Behaviour, bool) {
	for _, definition := range editor.behaviours {
		if definition.Name == name {
			return definition, true
		}
	}
	return project.Behaviour{}, false
}

// AttachBehaviour gives the selected node a behaviour of the project's, its fields empty.
func (editor *Editor) AttachBehaviour(name string) {
	if editor.Selected == nil || name == "" {
		return
	}
	editor.Selected.Behaviours = append(editor.Selected.Behaviours, scenefile.Behaviour{Name: name, Fields: map[string]any{}})
	editor.Changed()
}

// DetachBehaviour takes a behaviour off the selected node.
func (editor *Editor) DetachBehaviour(index int) {
	if editor.Selected == nil || index < 0 || index >= len(editor.Selected.Behaviours) {
		return
	}
	behaviours := editor.Selected.Behaviours
	editor.Selected.Behaviours = append(behaviours[:index], behaviours[index+1:]...)
	editor.Changed()
}

// FieldValue reads a field as typed in the inspector: a number, true or false, or else text.
func FieldValue(typed string) any {
	typed = strings.TrimSpace(typed)
	if number, err := strconv.ParseFloat(typed, 64); err == nil {
		return number
	}
	if typed == "true" || typed == "false" {
		return typed == "true"
	}
	return typed
}

// FieldText writes a field's value for the inspector.
func FieldText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64)
	case string:
		return typed
	}
	return fmt.Sprint(value)
}

// PlayHere plays the scene in the view, with the behaviours compiled into the game the editor runs
// in; the scene edited stays as it was. StopHere goes back to editing.
func (editor *Editor) PlayHere() {
	if editor.unlisten == nil {
		editor.unlisten = editor.listenToLog()
	}
	built, err := engine.Build(editor.Project.Files(), editor.Scene.Describe())
	if built == nil {
		editor.say(err.Error(), true)
		return
	}
	problems := built.Start()
	editor.playingHere = built
	switch {
	case len(problems) > 0:
		editor.say(problems[0].Error(), true)
	case err != nil:
		editor.say(err.Error(), true)
	case len(engine.Registered()) == 0:
		editor.say("Playing without the project's code: open the project with go-engine to play its behaviours", true)
	default:
		editor.say("Playing in the view: Escape or Stop goes back to editing", false)
	}
}

// StopHere stops the play in the view.
func (editor *Editor) StopHere() {
	if editor.playingHere == nil {
		return
	}
	editor.playingHere, editor.Paused = nil, false
	if editor.unlisten != nil {
		editor.unlisten()
		editor.unlisten = nil
	}
	editor.say("Back to editing", false)
}

// AdvancePlay moves the scene played in the view on a frame, and goes on to the scene the game
// asked for (a level started again, the next one); one that cannot be read is reported and the
// view stays where it is. While paused it stands still, except for one frame after StepFrame.
func (editor *Editor) AdvancePlay(seconds float64, keys *input.State) {
	played := editor.playingHere
	if played == nil || editor.Paused && !editor.stepOnce {
		return
	}
	if editor.stepOnce {
		editor.stepOnce, seconds = false, oneFrame
	}
	played.Update(seconds, keys)
	next, replaced, err := played.TakeReplacement()
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	if replaced {
		editor.playingHere = next
	}
}

// PlayingHere is the scene played in the view, while it is.
func (editor *Editor) PlayingHere() (*engine.Scene, bool) {
	return editor.playingHere, editor.playingHere != nil
}

// PlayInWindow saves the scene, builds the game and runs it in a window of its own, in the
// background: what happens arrives
// through Events. Stop ends it.
func (editor *Editor) PlayInWindow() {
	if editor.running {
		return
	}
	if !editor.Saved && editor.Save() != nil {
		return
	}
	editor.running = true
	editor.output = &outputTail{}
	editor.say("Building the game…", false)
	go func() {
		command, err := editor.Project.PlayCommand()
		if err != nil {
			editor.send(func(editor *Editor) { editor.running = false; editor.say("The game did not build: "+err.Error(), true) })
			return
		}
		command.Stdout, command.Stderr = editor.output, editor.output
		if err := command.Start(); err != nil {
			editor.send(func(editor *Editor) { editor.running = false; editor.say(err.Error(), true) })
			return
		}
		editor.send(func(editor *Editor) { editor.playing = command; editor.say("Playing", false) })
		err = command.Wait()
		editor.send(func(editor *Editor) {
			editor.running, editor.playing = false, nil
			if editor.stopping {
				editor.stopping = false
				editor.say("Stopped the game", false)
				return
			}
			if last := editor.output.Last(); err != nil && last != "" {
				editor.say("The game ended: "+last, true)
				return
			}
			editor.say("The game ended", false)
		})
	}()
}

// Stop ends the game that is playing.
func (editor *Editor) Stop() {
	if editor.playing != nil && editor.playing.Process != nil {
		editor.stopping = true
		editor.playing.Process.Kill()
	}
}

// Running tells whether the game is building or playing.
func (editor *Editor) Running() bool { return editor.running }

// Run runs a task in the background (a build), saying when it starts and how it ends.
func (editor *Editor) Run(starting string, task func() (string, error)) {
	if editor.running {
		return
	}
	editor.running = true
	editor.say(starting, false)
	go func() {
		done, err := task()
		editor.send(func(editor *Editor) {
			editor.running = false
			if err != nil {
				editor.say(err.Error(), true)
				return
			}
			editor.say(done, false)
		})
	}()
}

// send hands a change to the editor's goroutine: the window's, which applies it in TakeEvents.
func (editor *Editor) send(change func(editor *Editor)) {
	editor.events <- change
	if editor.Wake != nil {
		editor.Wake()
	}
}

// TakeEvents applies what background work sent since the last frame.
func (editor *Editor) TakeEvents() {
	for {
		select {
		case change := <-editor.events:
			change(editor)
		default:
			return
		}
	}
}

// outputTail keeps the last lines a program printed.
type outputTail struct {
	mutex sync.Mutex
	lines []string
}

func (tail *outputTail) Write(data []byte) (int, error) {
	tail.mutex.Lock()
	defer tail.mutex.Unlock()
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(line) != "" {
			tail.lines = append(tail.lines, line)
		}
	}
	if len(tail.lines) > 200 {
		tail.lines = tail.lines[len(tail.lines)-200:]
	}
	return len(data), nil
}

// Last is the last line printed.
func (tail *outputTail) Last() string {
	tail.mutex.Lock()
	defer tail.mutex.Unlock()
	if len(tail.lines) == 0 {
		return ""
	}
	return tail.lines[len(tail.lines)-1]
}
