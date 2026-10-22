package engine

import (
	"errors"
	"testing"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/nintendo3ds"
)

func TestEngineJsonStartsAGameOnTheWholeScreenWhereItAsks(test *testing.T) {
	settings := Settings{Window: map[string]string{DesktopTarget: FullscreenWindow, WebTarget: Windowed}}
	for target, wanted := range map[string]bool{DesktopTarget: true, WebTarget: false, AndroidTarget: false} {
		if whole, err := startsFullscreen(settings, target); err != nil || whole != wanted {
			test.Fatalf("on %s the game starts on the whole screen: %v (%v)", target, whole, err)
		}
	}
	settings.Window[DesktopTarget] = "huge"
	if _, err := startsFullscreen(settings, DesktopTarget); !errors.Is(err, WindowUnknown) {
		test.Fatalf("an unknown window is %v", err)
	}
}

// pressedOnce works the actions out for a frame with some keys held, after one with none.
func pressedOnce(keys ...input.Key) {
	frameWith(input.NewState(), nintendo3ds.Reading{})
	state := input.NewState()
	for _, key := range keys {
		state.Feed().SetKey(key, true)
	}
	frameWith(state, nintendo3ds.Reading{})
}

func TestF11AndAltEnterPutTheGameOnTheWholeScreenAndBack(test *testing.T) {
	bindGame(Settings{Name: "Lighthouse"}, DesktopTarget)
	SetFullscreen(false)
	pressedOnce(input.KeyF11)
	toggleFullscreenIfAsked()
	if !Fullscreen() {
		test.Fatal("F11 does not put the game on the whole screen")
	}
	pressedOnce(input.KeyEnter)
	toggleFullscreenIfAsked()
	if !Fullscreen() {
		test.Fatal("Enter alone takes the game off the whole screen")
	}
	pressedOnce(input.KeyAlt, input.KeyEnter)
	toggleFullscreenIfAsked()
	if Fullscreen() {
		test.Fatal("Alt and Enter do not put the game back in its window")
	}
}

func TestAGameThatBindsFullscreenItselfKeepsF11(test *testing.T) {
	settings := Settings{Name: "Lighthouse", Actions: map[string]map[string][]string{FullscreenAction: {DesktopTarget: {"F"}}}}
	bindGame(settings, DesktopTarget)
	SetFullscreen(false)
	pressedOnce(input.KeyF11)
	toggleFullscreenIfAsked()
	if Fullscreen() {
		test.Fatal("F11 still toggles the whole screen")
	}
	pressedOnce(input.KeyF)
	toggleFullscreenIfAsked()
	if !Fullscreen() {
		test.Fatal("the game's own binding does not toggle it")
	}
}

func TestTheGamesRequestReachesTheWindowAndThePersonsComesBack(test *testing.T) {
	state := input.NewState()
	SetFullscreen(true)
	followFullscreen(state)
	if !state.Fullscreen() {
		test.Fatal("the game's request does not reach the window")
	}
	state.SetFullscreen(false) // the person left the whole screen (a browser's Escape)
	followFullscreen(state)
	if Fullscreen() || state.Fullscreen() {
		test.Fatal("the game still asks for the whole screen the person left")
	}
}
