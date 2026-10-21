package engine

import "sync"

// A game's code that is no node's: once a frame before the behaviours (a network exchange, whose
// news the behaviours then read), and calls that must run between frames (a system keyboard that
// takes both screens of a console). Both run in Scene.Update, before any behaviour, so in the
// editor's Play as in the game.

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
}

// runFrameHook runs the game's frame hook, if it set one.
func (owner *Scene) runFrameHook() {
	if frameHook == nil {
		return
	}
	owner.context.Node = nil
	frameHook(&owner.context)
}
