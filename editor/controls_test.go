package editor

import (
	"reflect"
	"testing"
)

func TestTheControlsTypedAreTheBindingsKept(test *testing.T) {
	actions := map[string]map[string][]string{"fire": {"desktop": {"Space", "Pad A"}, "nintendo3ds": {"A"}}}
	typed := typedFrom(actions)
	if typed["fire"]["desktop"] != "Space, Pad A" {
		test.Fatalf("the desktop's inputs are typed as %q", typed["fire"]["desktop"])
	}
	typed["fire"]["android"] = " Back ,  "
	typed["fire"]["web"] = ""
	wanted := map[string]map[string][]string{"fire": {"desktop": {"Space", "Pad A"}, "nintendo3ds": {"A"}, "android": {"Back"}}}
	if kept := typed.bindings(); !reflect.DeepEqual(kept, wanted) {
		test.Fatalf("the bindings kept are %v", kept)
	}
}
