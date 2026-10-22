package engine

// A game tells the player which input does what ("Tab menu", "Start to pause"). It reads the hint
// from the bindings, so a hint follows engine.json and the player's rebinding by itself, and names
// the device in the player's hands.

// ActionWords are the words an action is bound to on the target played, as engine.json's
// ("Tab", "Pad Start"), rebinding included; none for an action not bound. The slice is the
// engine's own, not a copy, so reading it allocates nothing: do not change it.
func (keys *Input) ActionWords(name string) []string {
	index, found := actionTable.index[name]
	if !found {
		return nil
	}
	return actionTable.actions[index].words
}
