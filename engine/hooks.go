package engine

import (
	"sync"

	"github.com/akrck02/go-renderer/input"
)

// A game's code that is no node's: once a frame before the behaviours (a network exchange, whose
// news the behaviours then read), and calls that must run between frames (a system keyboard that
// takes both screens of a console). Both run once a frame in PlayFrame, the main scene's frame,
// before any behaviour, in the editor's Play as in the game; a scene of the bottom screen does not
// run them again.

// frameHook runs once a frame before the behaviours, when the game sets one.
var frameHook func(context *Context)

// BeforeBehaviours sets code that runs once a frame, before every behaviour, with the frame's
// context (its Node is nil); nil removes it.
func BeforeBehaviours(hook func(context *Context)) { frameHook = hook }

// betweenFrames are the calls waiting for the next frame, and the list they move to while they
// run, kept so that a frame allocates nothing.
var betweenFrames struct {
	lock             sync.Mutex
	waiting, running []func()
}

// BetweenFrames asks for a call to run once, at the start of the next frame, before anything is
// drawn: on the 3DS, the system keyboard can take both screens there. It may be called from any
// goroutine.
func BetweenFrames(call func()) {
	betweenFrames.lock.Lock()
	defer betweenFrames.lock.Unlock()
	betweenFrames.waiting = append(betweenFrames.waiting, call)
}

// runBetweenFrames runs the calls asked for since the last frame, in the order they were asked.
func runBetweenFrames() {
	betweenFrames.lock.Lock()
	betweenFrames.waiting, betweenFrames.running = betweenFrames.running[:0], betweenFrames.waiting
	betweenFrames.lock.Unlock()
	for _, call := range betweenFrames.running {
		call()
	}
	clear(betweenFrames.running) // let go of what the calls held until the list is used again
}

// PlayFrame is a frame of the game's main scene: the calls asked for between frames run, the
// actions are read (opening and closing the surfaces they open, and the whole screen), the frame
// hook runs, then the
// scene moves on (Update).
func (owner *Scene) PlayFrame(seconds float64, keys *input.State) {
	runBetweenFrames()
	owner.beginFrame(seconds, keys)
	owner.input.updateActions()
	toggleSurfaces()
	toggleFullscreenIfAsked()
	owner.runFrameHook()
	owner.advance(seconds)
}

// runFrameHook runs the game's frame hook, if it set one.
func (owner *Scene) runFrameHook() {
	if frameHook == nil {
		return
	}
	owner.context.Node = nil
	frameHook(&owner.context)
}
