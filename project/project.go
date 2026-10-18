// Package project is a game made with go-engine, as a folder: a Go module whose main runs the
// engine, the game's own Go code beside it, its scenes in scenes/ and what it imports in
// assets/. It is an ordinary Go module, so any editor with Go's language server (Zed, Neovim,
// Visual Studio Code, GoLand…) edits the code, and `go run .` plays it.
package project

import (
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/internal/build"
	"github.com/akrck02/go-engine/internal/notices"
	"github.com/akrck02/go-engine/scenefile"
)

// Project is a game's folder.
type Project struct {
	Folder   string
	Settings engine.Settings
	// Tools are the places of the tools that build for consoles and phones, kept on this
	// computer for every project.
	Tools build.Tools
}

// Folders in a project.
const (
	ScenesFolder = "scenes"
	AssetsFolder = "assets"
	ModelsFolder = "assets/models"
	SceneSuffix  = ".scene.json"
)

// Open opens a project's folder.
func Open(folder string) (*Project, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}
	settings, err := engine.ReadSettings(os.DirFS(folder))
	if err != nil {
		return nil, err
	}
	return &Project{Folder: folder, Settings: settings, Tools: keptTools()}, nil
}

// keptTools are the tools' places kept on this computer; when they cannot be read, none, and the
// builds look for the tools as if none were kept.
func keptTools() build.Tools {
	tools, err := build.ReadTools()
	if err != nil {
		slog.Warn("the tools' places kept on this computer are left out", "error", err)
	}
	return tools
}

// Files are the project's files, as the engine reads them.
func (project *Project) Files() fs.FS { return os.DirFS(project.Folder) }

// Path is where a file of the project is on disk, from its path in the project.
func (project *Project) Path(file string) string {
	return filepath.Join(project.Folder, filepath.FromSlash(file))
}

// Scenes are the project's scene files, as "scenes/main.scene.json".
func (project *Project) Scenes() []string {
	return project.filesUnder(ScenesFolder, func(name string) bool { return strings.HasSuffix(name, SceneSuffix) })
}

// Models are the project's models, as "assets/models/tree.glb".
func (project *Project) Models() []string {
	return project.filesUnder(AssetsFolder, func(name string) bool { return strings.EqualFold(path.Ext(name), ".glb") })
}

func (project *Project) filesUnder(folder string, wanted func(name string) bool) []string {
	var found []string
	fs.WalkDir(project.Files(), folder, func(file string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && wanted(entry.Name()) {
			found = append(found, file)
		}
		return nil
	})
	sort.Strings(found)
	return found
}

// ReadScene reads a scene file of the project.
func (project *Project) ReadScene(file string) (*scenefile.Scene, error) {
	data, err := os.ReadFile(project.Path(file))
	if err != nil {
		return nil, err
	}
	return scenefile.Read(data)
}

// WriteScene writes a scene file of the project.
func (project *Project) WriteScene(file string, scene *scenefile.Scene) error {
	if err := scene.Check(); err != nil {
		return err
	}
	data, err := scene.Marshal()
	if err != nil {
		return err
	}
	return writeFile(project.Path(file), data)
}

// NewScene adds a scene, 3D or 2D, with a ground and a camera, and returns its file.
func (project *Project) NewScene(name string, mode scenefile.Mode) (string, error) {
	file := path.Join(ScenesFolder, build.FileName(name)+SceneSuffix)
	if build.FileExists(project.Path(file)) {
		return "", failure.Wrap(SceneExists, file, nil)
	}
	if mode == scenefile.Mode2D {
		return file, project.WriteScene(file, starterScene2D(name))
	}
	return file, project.WriteScene(file, starterScene(name))
}

// Import copies a file into the project's assets, and returns its path in the project. Models go
// to assets/models.
func (project *Project) Import(source string) (string, error) {
	extension := strings.ToLower(filepath.Ext(source))
	if extension != ".glb" {
		return "", failure.Wrap(ImportFormat, filepath.Base(source), nil)
	}
	file := path.Join(ModelsFolder, filepath.Base(source))
	input, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(project.Path(file)), 0o755); err != nil {
		return "", err
	}
	output, err := os.Create(project.Path(file))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return "", err
	}
	return file, output.Close()
}

// SaveSettings writes the project's settings, when its Android application id has the form one
// has.
func (project *Project) SaveSettings() error {
	if asked := project.Settings.AndroidID; asked != "" && !applicationIDForm.MatchString(asked) {
		return failure.Wrap(ApplicationIDWrong, asked, nil)
	}
	data, err := json.MarshalIndent(project.Settings, "", "\t")
	if err == nil {
		err = writeFile(project.Path(engine.SettingsFile), append(data, '\n'))
	}
	if err != nil {
		return failure.Wrap(SettingsNotSaved, "", err)
	}
	return nil
}

// PlayCommand builds the game for this computer, as BuildDesktop does, and returns the command
// that runs it from the project's folder, not started: the editor shows its output and can stop
// it (a program started through go run would outlive its stopping).
func (project *Project) PlayCommand() (*exec.Cmd, error) {
	built, err := project.BuildDesktop()
	if err != nil {
		return nil, err
	}
	command := exec.Command(built.Program)
	command.Dir = project.Folder
	return command, nil
}

// Tidy brings the project's go.mod up to date with its code and with the go-engine it builds
// against, as `go mod tidy` does: an engine of a newer version may need other versions of its
// libraries. A game built for Android has go-ui's replace written again first (keepUIReplace).
func (project *Project) Tidy() error {
	module, err := project.readModule()
	if err != nil {
		return err
	}
	if module.hasTool(androidTool) {
		if err := project.keepUIReplace(module); err != nil {
			return err
		}
	}
	return build.Run(project.Folder, nil, build.GoTool(), "mod", "tidy")
}

// EditorProgram is where the game built with the editor in it goes.
func (project *Project) EditorProgram() string {
	return project.Path(path.Join("build", "editor", build.FileName(project.Settings.Name)+"-editor"+build.ExecutableSuffix()))
}

// BuildEditor builds the game with the editor in it, for go-engine to start.
func (project *Project) BuildEditor() error {
	if _, err := project.EnsureEditor(); err != nil {
		return err
	}
	if err := project.Tidy(); err != nil {
		return err
	}
	arguments := append([]string{"build"}, project.EngineFlags()...)
	return build.Run(project.Folder, nil, build.GoTool(), append(arguments, "-o", project.EditorProgram(), ".")...)
}

// EngineFlags tell a build which commit a local copy of Go Engine is at (one the project's go.mod
// replaces with a folder, or a go.work uses), so the program can say it (engine.Version): a
// local copy's build information does not. A project that requires a version needs none.
func (project *Project) EngineFlags() []string {
	listed, err := build.Output(project.Folder, nil, build.GoTool(), "list", "-m", "-f", engineListing, engine.Module)
	if err != nil {
		return nil
	}
	folder, local, _ := strings.Cut(strings.TrimSpace(listed), "|")
	if local != "local" || folder == "" {
		return nil
	}
	revision, err := build.Output(folder, nil, "git", "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return nil // not a git copy: the program says "local"
	}
	revision = strings.TrimSpace(revision)
	if changes, err := build.Output(folder, nil, "git", "status", "--porcelain"); err == nil && strings.TrimSpace(changes) != "" {
		revision += "-modified" // the copy has changes no commit holds
	}
	return []string{"-ldflags=-X " + engine.Module + "/engine.localRevision=" + revision}
}

// engineListing is how go list says where Go Engine is, and whether it is a local copy: one with
// no version of its own (a go.work), or replaced by a folder.
const engineListing = "{{.Dir}}|{{if not .Version}}local{{else if .Replace}}{{if not .Replace.Version}}local{{end}}{{end}}"

// CodeChangedSince tells whether a Go file of the project changed after a time: the game the
// editor runs in is older than its code.
func (project *Project) CodeChangedSince(built time.Time) bool {
	changed := false
	fs.WalkDir(project.Files(), ".", func(file string, entry fs.DirEntry, err error) error {
		if err != nil || changed {
			return fs.SkipAll
		}
		if entry.IsDir() && (file == "build" || strings.HasPrefix(entry.Name(), ".")) && file != "." {
			return fs.SkipDir
		}
		if strings.HasSuffix(file, ".go") {
			if details, err := entry.Info(); err == nil && details.ModTime().After(built) {
				changed = true
			}
		}
		return nil
	})
	return changed
}

// Built is what a build made: the program, and the notices of what it carries.
type Built struct{ Program, Notices string }

// BuildDesktop builds the game for this computer into build/.
func (project *Project) BuildDesktop() (Built, error) {
	output := project.Path(path.Join("build", build.FileName(project.Settings.Name)+build.ExecutableSuffix()))
	if err := project.Tidy(); err != nil {
		return Built{}, err
	}
	if err := build.Run(project.Folder, nil, build.GoTool(), append(append([]string{"build", "-tags", "release"}, project.EngineFlags()...), "-o", output, ".")...); err != nil {
		return Built{}, err
	}
	noticesFile, err := writeNotices(project.Path("build"), notices.Desktop)
	return Built{Program: output, Notices: noticesFile}, err
}

// Build3DS builds the game for the Nintendo 3DS into build/, with TinyGo.
func (project *Project) Build3DS() (Built, error) {
	tinygo, err := project.Tools.FindTinyGo()
	if err != nil {
		return Built{}, err
	}
	output := project.Path(path.Join("build", build.FileName(project.Settings.Name)+".3dsx"))
	if err := project.Tidy(); err != nil {
		return Built{}, err
	}
	environment := []string{"GOFLAGS=-buildvcs=false", build.GoOnPath()}
	if err := build.Run(project.Folder, environment, tinygo, append(append([]string{"build", "-target=nintendo3ds"}, project.EngineFlags()...), "-o", output, ".")...); err != nil {
		return Built{}, err
	}
	noticesFile, err := writeNotices(project.Path("build"), notices.Console)
	return Built{Program: output, Notices: noticesFile}, err
}

func writeNotices(folder string, notices []byte) (string, error) {
	file := filepath.Join(folder, "THIRD_PARTY_NOTICES.md")
	return file, writeFile(file, notices)
}

// OpenCode opens the project's folder in the code editor: the settings' CodeEditor, $VISUAL,
// $EDITOR, or the first of Zed, Visual Studio Code and GoLand found. A terminal editor (Neovim,
// Vim, Helix…) opens in a new terminal window.
func (project *Project) OpenCode() (string, error) {
	editor := project.Settings.CodeEditor
	for _, candidate := range []string{os.Getenv("VISUAL"), os.Getenv("EDITOR"), "zed", "zeditor", "code", "codium", "goland"} {
		if editor != "" {
			break
		}
		if candidate == "" {
			continue
		}
		if _, err := exec.LookPath(strings.Fields(candidate)[0]); err == nil {
			editor = candidate
		}
	}
	if editor == "" {
		return "", CodeEditorNotFound
	}
	words := strings.Fields(editor)
	if isTerminalEditor(words[0]) {
		terminal, found := findTerminal()
		if !found {
			return "", failure.Wrap(TerminalNotFound, words[0]+", "+project.Folder, nil)
		}
		words = append([]string{terminal, "-e"}, words...)
	}
	command := exec.Command(words[0], append(words[1:], project.Folder)...)
	command.Dir = project.Folder
	if err := command.Start(); err != nil {
		return "", err
	}
	go command.Wait()
	return strings.Join(words, " "), nil
}

func isTerminalEditor(program string) bool {
	switch filepath.Base(program) {
	case "nvim", "vim", "vi", "hx", "helix", "emacs", "nano", "micro", "kak":
		return true
	}
	return false
}

func findTerminal() (string, bool) {
	for _, terminal := range []string{os.Getenv("TERMINAL"), "x-terminal-emulator", "kitty", "alacritty", "wezterm", "gnome-terminal", "konsole", "xterm"} {
		if terminal == "" {
			continue
		}
		if found, err := exec.LookPath(terminal); err == nil {
			return found, true
		}
	}
	return "", false
}

func writeFile(file string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}
