// Command go-engine is the editor: it opens a game kept as JSON (or starts one from the sample),
// to paint its scenes, play it and export it for the desktop or the Nintendo 3DS.
//
//	go run ./cmd/go-engine                 # game.json in this folder
//	go run ./cmd/go-engine my-game.json
//	go run ./cmd/go-engine -capture editor.png
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akrck02/go-ui/window"

	"github.com/akrck02/go-engine/editor"
)

func main() {
	options := window.Options{Width: 1280, Height: 800, MinWidth: 900, MinHeight: 560}
	options.RegisterFlags(flag.CommandLine)
	play := flag.Bool("play", false, "start playing (for captures)")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: go-engine [flags] [game.json]")
		flag.PrintDefaults()
	}
	flag.Parse()
	path := flag.Arg(0)
	if path == "" {
		path = "game.json"
	}
	edited, err := editor.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *play {
		edited.Play()
	}
	view := editor.NewView(edited)
	if err := window.Run(view.Options(options), view.Build); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
