package engine

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

func TestAWordIsNamedAsItsKeyCapOrButtonSays(test *testing.T) {
	cases := map[string]string{
		"Escape": "Esc", "Page Up": "PgUp", "Page Down": "PgDn", "Control": "Ctrl", "Delete": "Del",
		"Space": "Space", "Tab": "Tab", "F1": "F1", "W": "W",
		"Left-": "Left", "D+": "D", "Pad Left-": "Left",
		"Alt+Enter": "Alt+Enter", "Control+S": "Control+S",
		"Mouse Left": "Left click", "Mouse Right": "Right click", "Mouse Middle": "Middle click",
		"Stick X": "Stick", "Right Stick Y": "Right stick",
		"Touch A": "A", "Touch Stick X": "Stick", "Back": "Back",
		"Pad Start": "Start", "Pad Up": "Up",
		"Start": "Start", "ZL": "ZL", // a 3DS's own buttons
		"": "",
	}
	for word, want := range cases {
		if got := InputLabel(word, input.XboxLayout); got != want {
			test.Errorf("InputLabel(%q) = %q, want %q", word, got, want)
		}
	}
}

func TestAPadsButtonsAreNamedAsThePadInThePlayersHandsNamesThem(test *testing.T) {
	words := []string{"Pad A", "Pad B", "Pad X", "Pad Y", "Pad Left Shoulder", "Pad Right Shoulder",
		"Pad Back", "Pad Start", "Pad Left Thumb", "Pad Right Thumb"}
	layouts := map[input.PadLayout][]string{
		input.XboxLayout:        {"A", "B", "X", "Y", "LB", "RB", "Back", "Start", "LS", "RS"},
		input.PlayStationLayout: {"✕", "○", "□", "△", "L1", "R1", "Share", "Options", "L3", "R3"},
		input.NintendoLayout:    {"B", "A", "Y", "X", "L", "R", "−", "+", "LS", "RS"},
	}
	for layout, labels := range layouts {
		for index, word := range words {
			if got := InputLabel(word, layout); got != labels[index] {
				test.Errorf("InputLabel(%q, %v) = %q, want %q", word, layout, got, labels[index])
			}
		}
	}
}

func TestNamingAWordAllocatesNothing(test *testing.T) {
	allocations := testing.AllocsPerRun(100, func() {
		_ = InputLabel("Pad A", input.PlayStationLayout)
		_ = InputLabel("Left-", input.XboxLayout)
		_ = InputLabel("Touch Stick X", input.XboxLayout)
	})
	if allocations != 0 {
		test.Fatalf("naming a word allocates %v times", allocations)
	}
}
