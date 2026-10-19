package editor

import (
	"time"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// Undo and redo. After each change the editor keeps the scene as its file would hold it; undo
// builds the scene again from the one kept before, redo from the one after. Changes closer than
// a step apart are one step: a drag, or a number typed digit by digit, is undone at once.

// historyStep is the time two changes must be apart to be two steps of the history.
const historyStep = 500 * time.Millisecond

// historyLength is how many states the history keeps; the oldest are forgotten.
const historyLength = 200

// history is the scene as it was after each change, and where the editor is in it.
type history struct {
	states  [][]byte // the scene's file, after each change
	at      int      // the state the scene is in
	savedAt int      // the state the file holds, or -1 when none does
	lastAt  time.Time
}

// startHistory forgets the history and starts it from the scene as it is, saved.
func (editor *Editor) startHistory() {
	editor.history = history{at: 0, savedAt: 0}
	if state, err := editor.Scene.Describe().Marshal(); err == nil {
		editor.history.states = [][]byte{state}
	}
}

// remember keeps the scene as it is now, as a new step or as the last one still going on.
func (editor *Editor) remember() {
	state, err := editor.Scene.Describe().Marshal()
	if err != nil || len(editor.history.states) == 0 {
		return // a scene that cannot be written cannot be undone either
	}
	keeping := &editor.history
	now := editor.clock()
	going := now.Sub(keeping.lastAt) < historyStep && keeping.at > 0 && keeping.at == len(keeping.states)-1 && keeping.savedAt != keeping.at
	keeping.lastAt = now
	if going {
		keeping.states[keeping.at] = state
		return
	}
	keeping.states = append(keeping.states[:keeping.at+1], state)
	keeping.at++
	if len(keeping.states) > historyLength {
		keeping.states, keeping.at, keeping.savedAt = keeping.states[1:], keeping.at-1, keeping.savedAt-1
	}
}

// Undo takes the scene back one step.
func (editor *Editor) Undo() {
	if editor.history.at == 0 {
		editor.say("Nothing to undo", false)
		return
	}
	editor.restore(editor.history.at - 1)
}

// Redo takes the scene forward one step, after an undo.
func (editor *Editor) Redo() {
	if editor.history.at >= len(editor.history.states)-1 {
		editor.say("Nothing to redo", false)
		return
	}
	editor.restore(editor.history.at + 1)
}

// CanUndo and CanRedo tell whether there is a step to go back or forward to.
func (editor *Editor) CanUndo() bool { return editor.history.at > 0 }
func (editor *Editor) CanRedo() bool { return editor.history.at < len(editor.history.states)-1 }

// restore builds the scene again as it was at a step, keeping the nodes chosen by their paths.
func (editor *Editor) restore(at int) {
	description, err := scenefile.Read(editor.history.states[at])
	if err != nil {
		editor.say(err.Error(), true)
		return
	}
	built, err := engine.Build(editor.Project.Files(), description)
	if built == nil {
		editor.say(err.Error(), true)
		return
	}
	built.Editing = true
	built.Sync()
	chosen := editor.chosenPaths()
	editor.Scene, editor.history.at, editor.history.lastAt = built, at, time.Time{}
	editor.Saved = at == editor.history.savedAt
	editor.chooseAgain(chosen)
}

// chosenPaths are the paths of the nodes chosen, the inspector's first.
func (editor *Editor) chosenPaths() []string {
	var paths []string
	if editor.Selected != nil {
		paths = append(paths, editor.Selected.Path())
	}
	for _, node := range editor.Chosen {
		paths = append(paths, node.Path())
	}
	return paths
}

// chooseAgain chooses the nodes of the scene at paths, those still there.
func (editor *Editor) chooseAgain(paths []string) {
	editor.Selected, editor.Chosen = nil, nil
	for index, path := range paths {
		node, found := editor.Scene.Find(path)
		switch {
		case !found:
		case index == 0:
			editor.Selected = node
		default:
			editor.Chosen = append(editor.Chosen, node)
		}
	}
}

// clock is the editor's time: the wall's, or a test's.
func (editor *Editor) clock() time.Time {
	if editor.now != nil {
		return editor.now()
	}
	return time.Now()
}
