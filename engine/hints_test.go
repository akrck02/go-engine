package engine

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"
)

func TestAnActionGivesTheWordsItIsBoundToOnTheTargetPlayed(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"Space", "Mouse Left", "Pad A"}) {
		test.Fatalf("fire's words on a computer are %q", words)
	}
	BindActions(gameActions, Nintendo3DSTarget)
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"A", "R"}) {
		test.Fatalf("fire's words on a 3DS are %q", words)
	}
}

func TestAReboundActionGivesItsNewWords(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	keys.Rebind("fire", []string{"Enter", "Not A Key", "Pad B"})
	if words := keys.ActionWords("fire"); !slices.Equal(words, []string{"Enter", "Pad B"}) {
		test.Fatalf("fire rebound gives %q", words)
	}
}

func TestAnActionNotBoundGivesNoWords(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	if words := keys.ActionWords("jump"); len(words) != 0 {
		test.Fatalf("an action not bound gives %q", words)
	}
	if words := keys.ActionWords("menu"); !slices.Equal(words, []string{"Tab", "F1"}) {
		test.Fatalf("menu gives %q", words)
	}
}

func TestReadingAnActionsWordsAllocatesNothing(test *testing.T) {
	BindActions(gameActions, DesktopTarget)
	keys := &Input{state: input.NewState()}
	allocations := testing.AllocsPerRun(100, func() { _ = keys.ActionWords("fire") })
	if allocations != 0 {
		test.Fatalf("reading an action's words allocates %v times", allocations)
	}
}
