package editor

import (
	"sort"

	"github.com/akrck02/go-engine/project"
)

// The Code workspace's work, without its interface: which nodes of which scenes use a behaviour,
// and which behaviours the scenes use that the project's code does not define (yet).

// BehaviourUses are the nodes of the project's saved scenes that have a behaviour.
func (editor *Editor) BehaviourUses(name string) []project.Use {
	return editor.Project.BehavioursUsed()[name]
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
	for name, uses := range editor.Project.BehavioursUsed() {
		if _, defined := editor.FindBehaviour(name); !defined {
			missing = append(missing, MissingBehaviour{Name: name, Uses: uses})
		}
	}
	sort.Slice(missing, func(first, second int) bool { return missing[first].Name < missing[second].Name })
	return missing
}
