# 11. Each target's own screens and controls

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

A game made for several devices lays out its interface and reads its controls differently on
each, and today it has to do that by hand:

- pairs of files split by build tags (`controls.go` and `controls_nintendo3ds.go`);
- constants such as "has a touch screen";
- on the desktop, a picture of the 3DS's touch screen pasted into the middle of the window, with
  mouse clicks turned into touches.

That is one device's layout imposed on another. A game should say *what* it shows and *what* the
player does, and each target should say *where* and *with which buttons*.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A game reads **actions** by name (`Input.Pressed("menu")`, `Input.Down("fire")`, `Input.Axis("turn")`), with no build tags; engine.json binds each action to inputs per target (keys, mouse buttons, a pad's buttons and sticks, a 3DS's buttons, Android's back) | tests of each kind of binding with a made-up input; a sample game with no per-target file |
| R2 | The editor shows and edits the bindings, per target | capture of the bindings in the editor; a binding changed there reaches the game |
| R3 | A game can let the player rebind an action while it plays, and keep it on the device | test `a binding changed while playing is kept and read back` |
| R4 | A game paints named **surfaces** (`"main"`, `"panel"`, `"map"`…) with go-ui or draw2d, and each target places them: on a 3DS, some on the second screen and touchable; on a computer, in a corner or as a floating window opened and closed, with the mouse as the pointer; on Android, as a bottom sheet | captures of one sample on each target; the second screen becomes one way a target places a surface |
| R5 | The editor's Play previews each target's layout (today's `-3ds` view, made general) | captures in the editor for each target |
| R6 | On a computer and on the web, a game starts windowed or fullscreen as engine.json says per target, and toggles it (`engine.SetFullscreen`, `engine.Fullscreen()`); a default action (F11, Alt+Enter) does it. On the web it waits for the next click or key, as browsers ask | test of the setting; by hand on a desktop and in a browser |
| R7 | What exists keeps working: `SetInterface`, `SetSecondScreen` and the keys a game reads today | the platformer and lighthouse samples unchanged |

## Out of scope

- Touch gestures beyond the pointer (swipes, pinches) as actions.
- A full settings menu for players; the engine gives what one needs (R3), not the menu.
- Consoles other than the 3DS: the design makes room for them.

## Design

To fill once the author approves the proposal. Fullscreen will need Go Renderer and Go UI
(neither has it yet), on `feature/` branches with the author's permission. Godot's InputMap and
Unity's Input System will be looked at before choosing the bindings' shape.

## Tasks

To fill with the design.

## Result

Not built yet.
