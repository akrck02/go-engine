package editor

import (
	"sort"
	"time"

	"github.com/akrck02/go-engine/project"
)

// The Code workspace's work, without its interface: which nodes of which scenes use a behaviour,
// and which behaviours the scenes use that the project's code does not define (yet).

// usesReadEvery is how often the scenes are read again to find which behaviours they use: the
// Code workspace asks on every frame it draws.
const usesReadEvery = 2 * time.Second

// behavioursUsed are the behaviours the saved scenes use, read again at most every couple of
// seconds, and at once after saving.
func (editor *Editor) behavioursUsed() map[string][]project.Use {
	if editor.used == nil || time.Since(editor.usedAt) > usesReadEvery {
		editor.used, editor.usedAt = editor.Project.BehavioursUsed(), time.Now()
	}
	return editor.used
}

// BehaviourUses are the nodes of the project's saved scenes that have a behaviour.
func (editor *Editor) BehaviourUses(name string) []project.Use {
	return editor.behavioursUsed()[name]
}

// MissingBehaviour is a behaviour the scenes use that the project's code does not define.
type MissingBehaviour struct {
	Name string
	Uses []project.Use
}

// MissingBehaviours are the behaviours the project's saved scenes use that its code does not
// define, by name: a typo in a scene, or code not written yet.
func (editor *Editor) MissingBehaviours() []MissingBehaviour {
	var missing []MissingBehaviour
	for name, uses := range editor.behavioursUsed() {
		if _, defined := editor.FindBehaviour(name); !defined {
			missing = append(missing, MissingBehaviour{Name: name, Uses: uses})
		}
	}
	sort.Slice(missing, func(first, second int) bool { return missing[first].Name < missing[second].Name })
	return missing
}
