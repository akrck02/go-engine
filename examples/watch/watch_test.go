package main

import (
	"testing"

	"github.com/akrck02/go-engine/engine"
)

func TestEveryActionIsBoundToInputsItsTargetKnows(test *testing.T) {
	settings, err := engine.ReadSettings(files)
	if err != nil {
		test.Fatal(err)
	}
	for name, targets := range settings.Actions {
		for target, words := range targets {
			if problems := engine.CheckInputs(words, target); len(problems) != 0 {
				test.Errorf("%s on %s: %v", name, target, problems)
			}
		}
	}
	for _, target := range engine.InputTargets {
		if problems := engine.UseSurfaces(settings.Surfaces, target); len(problems) != 0 {
			test.Errorf("the surfaces on %s: %v", target, problems)
		}
	}
}

func TestTheLogKeepsItsNewestLines(test *testing.T) {
	keeper := keeperWatch{}
	for range longestLog + 2 {
		keeper.mark()
	}
	newest := keeper.lines[len(keeper.lines)-1]
	if len(keeper.lines) != longestLog || keeper.lines[0] != "Ship 3 sighted at 0°" || newest != "Ship 8 sighted at 0°" {
		test.Fatalf("the log holds %q", keeper.lines)
	}
}
