package engine

import (
	"encoding/json"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-ui/window"
)

// A player changes an action's inputs while the game plays (a settings screen of the game's own),
// and the change is kept on the device, among the window's preferences: the game's engine.json
// stays as its author wrote it. (On a 3DS, Go UI cannot keep preferences yet: a change lasts the
// run.)

// rebound are the actions the player changed on the target played, by name, and the game they are
// kept for.
var rebound struct {
	game    string
	target  string
	actions map[string][]string
}

// Rebind changes an action's inputs on the target played to some words (as engine.json's), and
// remembers the change for SaveBindings. Words the target does not know are left out, each
// reported.
func (keys *Input) Rebind(name string, words []string) []error {
	if rebound.actions == nil {
		rebound.actions = map[string][]string{}
	}
	rebound.actions[name] = append([]string(nil), words...)
	return bindOne(name, words, rebound.target)
}

// bindOne binds one action to some words on a target, adding it if it was not bound.
func bindOne(name string, words []string, target string) []error {
	var bound boundAction
	var problems []error
	for _, word := range words {
		input, err := readWord(word, target)
		if err != nil {
			problems = append(problems, err)
			continue
		}
		bound.inputs = append(bound.inputs, input)
	}
	if index, found := actionTable.index[name]; found {
		actionTable.actions[index].inputs = bound.inputs
		return problems
	}
	actionTable.index[name] = len(actionTable.actions)
	actionTable.actions = append(actionTable.actions, bound)
	return problems
}

// SaveBindings keeps the actions the player changed, for the next time the game plays here.
func SaveBindings() error {
	data, err := json.Marshal(rebound.actions)
	if err == nil {
		err = window.KeepPreference(bindingsPreference(), string(data))
	}
	if err != nil {
		return failure.Wrap(BindingsNotSaved, "", err)
	}
	return nil
}

// bindingsPreference is the preference a game's changed bindings are kept in, on one target.
func bindingsPreference() string {
	return "go-engine.bindings." + rebound.game + "." + rebound.target
}

// bindGame binds a game's actions on a target, then the changes its player kept here.
func bindGame(settings Settings, target string) []error {
	rebound.game, rebound.target, rebound.actions = settings.Name, target, nil
	problems := BindActions(settings.Actions, target)
	kept, found := window.Preference(bindingsPreference())
	if !found {
		return problems
	}
	var changed map[string][]string
	if err := json.Unmarshal([]byte(kept), &changed); err != nil {
		return append(problems, failure.Wrap(BindingsNotSaved, "the bindings kept could not be read", err))
	}
	for name, words := range changed {
		problems = append(problems, bindOne(name, words, target)...)
	}
	rebound.actions = changed
	return problems
}
