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
