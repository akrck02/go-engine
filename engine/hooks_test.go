package engine

import (
	"slices"
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// Recorder notes when it updates, for a test of what runs first.
type Recorder struct{ said *[]string }

func (recorder *Recorder) Update(context *Context) {
	*recorder.said = append(*recorder.said, "behaviour")
}

func TestTheFramesHookRunsBeforeEveryBehaviour(test *testing.T) {
	played := load(test)
	var said []string
	spinner, _ := played.Find("Spinner")
	spinner.Attach(&Recorder{said: &said})
	BeforeBehaviours(func(context *Context) {
		if context.Node != nil || context.Scene != played {
			test.Errorf("the hook's context is for %v in %v", context.Node, context.Scene)
		}
		said = append(said, "hook")
	})
	defer BeforeBehaviours(nil)
	played.PlayFrame(0.1, nil)
	if len(said) == 0 || said[0] != "hook" || !slices.Contains(said, "behaviour") {
		test.Fatalf("a frame ran %v", said)
	}
}

func TestACallBetweenFramesRunsOnceInOrder(test *testing.T) {
	played := load(test)
	var said []string
	BetweenFrames(func() { said = append(said, "first") })
	BetweenFrames(func() { said = append(said, "second") })
	played.PlayFrame(0.1, nil)
	played.PlayFrame(0.1, nil)
	if !slices.Equal(said, []string{"first", "second"}) {
		test.Fatalf("after two frames the calls ran as %v", said)
	}
}

func TestTheFramesHookRunsOnceWithABottomScene(test *testing.T) {
	game := &runner{played: load(test)}
	SetSecondScreen(load(test))
	defer SetSecondScreen(nil)
	runs := 0
	BeforeBehaviours(func(context *Context) {
		runs++
		if context.Scene != game.played {
			test.Errorf("the hook ran for %v", context.Scene)
		}
	})
	defer BeforeBehaviours(nil)
	game.step(input.NewState(), 1.0/60)
	if runs != 1 {
		test.Fatalf("with a bottom scene the hook ran %d times in a frame", runs)
	}
}
