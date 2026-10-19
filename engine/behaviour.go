package engine

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/akrck02/go-engine/failure"
)

// A behaviour is the game's own code attached to a node: a struct of the game's, registered by
// name, whose exported fields the scene file fills (by their JSON names). It does what it does
// through the methods it has:
//
//	type Spin struct {
//		Speed float64 `json:"speed"` // degrees a second
//	}
//
//	func (spin *Spin) Update(context *engine.Context) {
//		context.Node.Transform.Turn(space.Up, spin.Speed*context.Delta)
//	}
//
// Start, when the behaviour has it, runs once before the first Update.
type (
	Starter interface{ Start(context *Context) }
	Updater interface{ Update(context *Context) }
)

// registry holds the behaviours the game registered, by name.
var registry = map[string]reflect.Type{}

// Register makes behaviours known by their type's name: Register(Spin{}) is "Spin".
func Register(behaviours ...any) {
	for _, behaviour := range behaviours {
		kind := structType(behaviour)
		RegisterAs(kind.Name(), behaviour)
	}
}

// RegisterAs makes a behaviour known by a name of its own.
func RegisterAs(name string, behaviour any) { registry[name] = structType(behaviour) }

// Registered are the names of the behaviours registered, in order.
func Registered() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func structType(behaviour any) reflect.Type {
	kind := reflect.TypeOf(behaviour)
	if kind.Kind() == reflect.Pointer {
		kind = kind.Elem()
	}
	return kind
}

// makeBehaviour makes a registered behaviour with its fields from the scene.
func makeBehaviour(name string, fields map[string]any) (any, error) {
	kind, found := registry[name]
	if !found {
		return nil, failure.Wrap(BehaviourUnregistered, name, nil)
	}
	made := reflect.New(kind).Interface()
	if len(fields) == 0 {
		return made, nil
	}
	data, err := json.Marshal(fields)
	if err == nil {
		err = json.Unmarshal(data, made)
	}
	if err != nil {
		return nil, failure.Wrap(BehaviourFields, name, err)
	}
	return made, nil
}

// Behaviour finds a node's behaviour of a type, while the game runs:
//
//	if spin, found := engine.Behaviour[Spin](node); found { spin.Speed *= 2 }
func Behaviour[Wanted any](node *Node) (*Wanted, bool) {
	for _, behaviour := range node.behaviours {
		if found, matches := behaviour.(*Wanted); matches {
			return found, true
		}
	}
	return nil, false
}

// MadeBehaviours are the node's behaviours as the game runs them, made from its Behaviours: what
// the editor watches while a scene plays. A scene not started has none.
func (node *Node) MadeBehaviours() []any { return node.behaviours }

// Attach gives a node a behaviour while the game runs, made by the game's code; it starts on the
// next frame.
func (node *Node) Attach(behaviour any) {
	node.behaviours = append(node.behaviours, behaviour)
	node.scene.pendingStarts = append(node.scene.pendingStarts, startOf{node, behaviour})
}

type startOf struct {
	node      *Node
	behaviour any
}
