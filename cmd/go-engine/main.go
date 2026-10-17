// Command go-engine is the editor. It opens a game's project folder (by default the folder it runs
// in), or makes a new one:
//
//	go run ./cmd/go-engine new ~/games/harbour "Harbour lights"
//	go run ./cmd/go-engine ~/games/harbour
//	go run ./cmd/go-engine -capture editor.png ~/games/harbour
//
// The game's code is written in any editor with Go's language server; the project is an ordinary
// Go module, and `go run .` in it plays the game.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/editor"
	"github.com/akrck02/go-engine/project"
)

func main() {
	options := window.Options{Width: 1440, Height: 900, MinWidth: 1000, MinHeight: 600}
	options.RegisterFlags(flag.CommandLine)
	choose := flag.String("select", "", "choose a node by its path when the editor opens (for captures)")
	flag.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintln(output, "usage: go-engine [flags] [project folder]")
		fmt.Fprintln(output, "       go-engine new <folder> [name]")
		flag.PrintDefaults()
	}
	flag.Parse()
	opened, err := projectFromArguments(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if opened == nil {
		return // a project was made
	}
	edited, err := editor.Open(opened)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *choose != "" {
		if node, found := edited.Scene.Find(*choose); found {
			edited.Select(node)
		}
	}
	view := editor.NewView(edited)
	if err := window.Run(view.Options(options), view.Build); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// projectFromArguments opens the project the arguments name, or makes one for "new" (and then
// returns no project to open).
func projectFromArguments(arguments []string) (*project.Project, error) {
	if len(arguments) > 0 && arguments[0] == "new" {
		if len(arguments) < 2 {
			return nil, fmt.Errorf("usage: go-engine new <folder> [name]")
		}
		folder := arguments[1]
		name := filepath.Base(folder)
		if len(arguments) > 2 {
			name = arguments[2]
		}
		created, err := project.Create(folder, name)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Made %s in %s. Open it with: go-engine %s\n", name, created.Folder, folder)
		return nil, nil
	}
	folder := "."
	if len(arguments) > 0 {
		folder = arguments[0]
	}
	opened, err := project.Open(folder)
	if err != nil {
		return nil, fmt.Errorf("%s is not a go-engine project (%w); make one with: go-engine new <folder>", folder, err)
	}
	return opened, nil
}
