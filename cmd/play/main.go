// Command play plays a game made with go-engine, from its JSON file, or the sample game.
//
//	go run ./cmd/play game.json
//	go run ./cmd/play            # the sample
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akrck02/go-engine/game"
	"github.com/akrck02/go-engine/player"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: play [game.json]")
		flag.PrintDefaults()
	}
	player.RegisterFlags(flag.CommandLine)
	flag.Parse()
	played := game.Sample()
	if path := flag.Arg(0); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			played, err = game.Read(data)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := player.Run(played); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
