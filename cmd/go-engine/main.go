// Command go-engine opens the editor on a game's project, or makes a new project and opens it:
//
//	go run ./cmd/go-engine new ~/games/harbour "Harbour lights"
//	go run ./cmd/go-engine new -open=false ~/games/harbour "Harbour lights"
//	go run ./cmd/go-engine ~/games/harbour
//	go run ./cmd/go-engine -capture editor.png ~/games/harbour
//	go run ./cmd/go-engine -update ~/games/harbour
//
// The editor runs inside the project's own game, so it plays the game's code in its view: this
// command builds the game with the editor in it (into build/editor) and starts it. The game's
// code is written in any editor with Go's language server; `go run .` in the project plays it.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/project"
)

func main() {
	// the editor's flags, declared here too so the folder can be told from their values; they are
	// passed on to the editor as they were given
	flags := flag.NewFlagSet("go-engine", flag.ExitOnError)
	var options window.Options
	options.RegisterFlags(flags)
	flags.String("select", "", "choose a node by its path when the editor opens (for captures)")
	flags.Bool("play", false, "play the scene in the view when the editor opens (for captures)")
	flags.Bool("3ds", false, "show the game played at the 3DS's screens (for captures)")
	flags.String("workspace", "", "the workspace to open in: project, art, level, code, play or build (for captures)")
	update := flags.Bool("update", false, "update the project to go-engine's latest version, and do not open the editor")
	version := flags.String("engine", project.LatestEngine, "with -update: the go-engine version to update to (a branch, a tag or a commit)")
	open := flags.Bool("open", true, "open the editor on the project; false only makes it (new) or opens nothing, as on a server with no display")
	flags.Usage = func() {
		output := flags.Output()
		fmt.Fprintln(output, "usage: go-engine [flags] [project folder]")
		fmt.Fprintln(output, "       go-engine new [-open=false] <folder> [name]")
		fmt.Fprintln(output, "       go-engine -update [-engine version] [project folder]")
		flags.PrintDefaults()
	}
	passed, positional := parseAnywhere(flags, os.Args[1:])
	opened, err := projectFrom(positional)
	switch {
	case err != nil:
	case *update:
		err = updateEngine(opened, *version)
	case !*open:
		fmt.Printf("%s is ready in %s\n", opened.Settings.Name, opened.Folder)
	default:
		err = edit(opened, withoutFlag(passed, "open"))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// parseAnywhere parses flags before, between and after the other arguments (Go's flag stops at
// the first argument that is not one), and returns the flags as given and the other arguments.
func parseAnywhere(flags *flag.FlagSet, arguments []string) (passed, positional []string) {
	for {
		flags.Parse(arguments) // ExitOnError: a wrong flag ends the program with its usage
		rest := flags.Args()
		passed = append(passed, arguments[:len(arguments)-len(rest)]...)
		if len(rest) == 0 {
			return passed, positional
		}
		positional, arguments = append(positional, rest[0]), rest[1:]
	}
}

// withoutFlag is the flags as given, without a boolean flag of this name, which only this command
// reads: the editor would refuse it.
func withoutFlag(passed []string, name string) []string {
	var kept []string
	for _, given := range passed {
		bare, _, _ := strings.Cut(strings.TrimLeft(given, "-"), "=")
		if strings.HasPrefix(given, "-") && bare == name {
			continue
		}
		kept = append(kept, given)
	}
	return kept
}

// missingFolder is go-engine new without its folder.
var missingFolder = failure.Define("go-engine.missing-folder", "usage: go-engine new <folder> [name]")

// projectFrom opens the project the arguments name, making it first for "new".
func projectFrom(arguments []string) (*project.Project, error) {
	if len(arguments) > 0 && arguments[0] == "new" {
		if len(arguments) < 2 {
			return nil, missingFolder
		}
		folder := arguments[1]
		name := filepath.Base(folder)
		if len(arguments) > 2 {
			name = arguments[2]
		}
		fmt.Printf("Making %s in %s…\n", name, folder)
		return project.Create(folder, name)
	}
	folder := "."
	if len(arguments) > 0 {
		folder = arguments[0]
	}
	opened, err := project.Open(folder)
	if err != nil {
		return nil, failure.Wrap(project.NotAProject, folder, err)
	}
	return opened, nil
}

// updateEngine moves the project to go-engine at a version, and says from which to which.
func updateEngine(opened *project.Project, version string) error {
	before, _, err := opened.EngineVersion()
	if err != nil {
		return err
	}
	fmt.Printf("Updating %s to go-engine@%s…\n", opened.Settings.Name, version)
	after, err := opened.UpdateEngine(version)
	if err != nil {
		return err
	}
	if after == before {
		fmt.Printf("Already at the latest: %s\n", after)
		return nil
	}
	fmt.Printf("go-engine %s → %s\n", before, after)
	return nil
}

// edit builds the project's game with the editor in it and starts it, in the project's folder.
func edit(opened *project.Project, flags []string) error {
	fmt.Printf("Go Engine %s · building %s with the editor…\n", engine.Version(), opened.Settings.Name)
	if err := opened.BuildEditor(); err != nil {
		return err
	}
	command := exec.Command(opened.EditorProgram(), flags...)
	command.Dir = opened.Folder
	command.Env = append(os.Environ(), engine.EditorVariable+"=1")
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}
