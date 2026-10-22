# 🎮 Actions and controls

This guide is for whoever makes a game that plays on several devices. At the end you will read
what the player does by name, give it buttons on each device in the editor, and design the
sticks and buttons a phone shows.

## Read actions, not keys

A game that reads `Space` works on a keyboard and nowhere else. A game that reads **actions**
works anywhere: it asks for `jump`, and each device says which of its buttons jump.

```go
func (hero *Hero) Update(context *engine.Context) {
	if context.Input.Action("jump").Pressed {
		hero.Jump()
	}
	hero.Turn(context.Input.Action("turn").Value) // from -1 to 1
}
```

An action **presses** (`Down`, `Pressed`, `Released`) or moves an **axis** (`Value`, from -1 to 1).

## Give them buttons in the editor

Open the **Controls** workspace. The actions are cards on the left, each in its own colour; the
device is drawn in the middle, chosen in its tabs: **Gamepad**, **Keyboard and mouse**,
**Nintendo 3DS** and **Mobile**. A pad and a keyboard play the same game on a computer and in a
browser, so what you give them holds for both.

- **Click a control** of the device: a panel opens beside it. Pick an action, or type a new
  name and create it there. A button presses; a stick has two axes, across and up and down, each
  with its own action.
- **The cards** say which controls do each action; click one to see its controls stand out.
  The pencil renames it, the next icon switches it between pressing and an axis.
- **Every change is kept** in `engine.json` at once. Ctrl+Z undoes the last one.

## Write the hints from the bindings

A game tells the player what each button does: "Tab menu", "Start to pause". Don't write those
names by hand: the bindings already know them, and the player may have rebound them.

`context.Input.ActionWords("menu")` gives the words an action is bound to on the device played, as
`engine.json` writes them (`["Tab", "Pad Start"]`), with the player's rebinding. An action not bound
gives none. The slice is the engine's own, so reading it costs nothing each frame; don't change it.

A word is what an author writes; a player reads what the key cap or the button says.
`engine.InputLabel(word, layout)` turns one into the other:

| Word | Label |
|---|---|
| `Escape`, `Page Up`, `Control` | Esc, PgUp, Ctrl |
| `Left-`, `D+` | Left, D (a direction's sign dropped) |
| `Alt+Enter` | Alt+Enter (a combination stays whole) |
| `Mouse Left`, `Stick X` | Left click, Stick |
| `Touch Jump` | Jump |
| `Pad A` | A on an Xbox pad, ✕ on a PlayStation pad, B on a Nintendo pad |
| `Pad Start`, `Pad Left Shoulder` | Start, LB · Options, L1 · +, L |

The layout is the pad's, as Go Renderer tells it from the pad's maker
(`context.Input.State().PadLayout()`); a pad it cannot tell is read as an Xbox pad.

A computer may have a keyboard and a pad at once: the hint should name the one in the player's
hands. `context.Input.LastInputKind()` is the kind of the last input pressed or moved:
`KeyboardAndMouse`, `GamePad`, `TouchScreen` or `ConsoleButtons` (a 3DS's), and `NoInputYet`
before the first. It changes when the player picks up another device, not when they stop.

`context.Input.ActionHint` puts it all together: the label of the action's first input of the kind
in use, or of its first input when none is of that kind. With `menu` bound to `Tab` and `Pad Start`,
it is "Tab" at the keyboard, "Start" with an Xbox pad and "Options" with a PlayStation pad; on a
3DS, it names the 3DS's own button. It allocates nothing, so a game can ask for it every frame:

```go
func (keeper *Keeper) Update(context *engine.Context) {
	keeper.Prompt.Text = context.Input.ActionHint("menu") + " menu"
}
```

## Design a phone's controls

A phone has no buttons of its own, so the game draws them. Until you design some, it shows the
engine's four: the arrows, B and A. In **Mobile → Layout**:

1. **Design the game's own** starts from a stick and two buttons.
2. **Stick** and **Button** add more; drag each where the thumbs rest.
3. With one chosen, rename it, make it smaller or bigger, or remove it.

Back in **Actions**, the controls on the screen get their actions like any other. Places and
sizes are shares of the screen, so the layout fits every phone.

---

[← The second screen](8-the-second-screen.md) · [Surfaces on each screen →](10-surfaces-on-each-screen.md)
