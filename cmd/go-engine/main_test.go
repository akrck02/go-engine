package main

import (
	"flag"
	"slices"
	"testing"
)

func TestFlagsAreReadAfterTheFolderToo(test *testing.T) {
	flags := flag.NewFlagSet("go-engine", flag.ContinueOnError)
	update := flags.Bool("update", false, "")
	engine := flags.String("engine", "main", "")
	passed, positional := parseAnywhere(flags, []string{"../game", "-update", "-engine", "v0.0.2"})
	if !*update || *engine != "v0.0.2" {
		test.Fatalf("after the folder, -update is %v and -engine %q", *update, *engine)
	}
	if !slices.Equal(positional, []string{"../game"}) || !slices.Equal(passed, []string{"-update", "-engine", "v0.0.2"}) {
		test.Fatalf("the arguments are %q and the flags %q", positional, passed)
	}
}

func TestTheEditorIsNotGivenTheFlagsOnlyTheCommandReads(test *testing.T) {
	passed := []string{"-open=true", "-width", "800", "--open", "-select", "Player"}
	if kept := withoutFlag(passed, "open"); !slices.Equal(kept, []string{"-width", "800", "-select", "Player"}) {
		test.Fatalf("the editor is given %q", kept)
	}
}
