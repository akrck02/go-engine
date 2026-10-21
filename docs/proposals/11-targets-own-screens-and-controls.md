# 11. Each target's own screens and controls

> Status: building · Author's approval: proposal ☑ (2026-10-21, "dale") plan ☑ (2026-10-21, "si")

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
| R1 | A game reads **actions** by name (`Input.Action("menu").Pressed`, `Input.Action("turn").Value`), with no build tags; engine.json binds each action to inputs per target (keys, mouse buttons, a pad's buttons and sticks, a 3DS's buttons, Android's back) | tests of each kind of binding with a made-up input; a sample game with no per-target file |
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

### Targets at run time

The engine knows which target it runs on (a constant per build: `desktop`, `nintendo3ds`,
`android`, `web`), named as engine.json names them. `project.Target` gains `Web`'s name for the
settings, though building for the web stays out of this change.

### Actions (R1–R3)

engine.json gains `actions`. Each action lists, per target, the inputs that press it, as words a
person can read and type:

```json
"actions": {
  "menu": {"desktop": ["Tab", "P", "Pad Start"], "nintendo3ds": ["Start"], "android": ["Back"]},
  "fire": {"desktop": ["Space", "Mouse Left", "Pad A"], "nintendo3ds": ["R", "A"]},
  "turn": {"desktop": ["A-", "D+", "Left-", "Right+", "Stick X"], "nintendo3ds": ["Circle X", "Left-", "Right+"]}
}
```

- A **button** input presses the action: a key, a mouse button, a pad's button, a 3DS's button,
  Android's back.
- An **axis** input moves it from -1 to 1: a stick's axis, or a button with `-` or `+` after it.
- `Input.Action(name)` gives the action's state this frame: `Down`, `Pressed`, `Released` and its
  `Value` (-1 to 1; 1 while a button holds it). One call, one value returned: Go has no
  overloading, so `Input.Down` keeps taking keys, which games read today (R7).
- The bindings are read once into a table of the engine (`engine/actions.go`), so reading an
  action each frame is a map lookup and some booleans, with nothing allocated.
- **Rebinding** (R3): `Input.Rebind(action, inputs)` changes the current target's inputs, and
  `engine.SaveBindings` keeps them in the user's configuration folder (on a 3DS, on the SD card).
  The game's own engine.json stays as its author wrote it.
- **The editor** (R2): Project → Settings gains a Controls table, one row per action, one column
  per target; a cell is the inputs, typed or caught by pressing them.

### Surfaces (R4, R5)

`engine.SetSurface(name, paint func(context *ui.Context))` names what the game shows. engine.json's
`surfaces` says, per target, where each goes:

| Place | What it is |
|---|---|
| `over` | over the game's view, aligned to an edge or a corner, with a size |
| `window` | a floating window over the view, opened and closed by an action |
| `second-screen` | the second screen (a 3DS's), touchable; several share it in a column |
| `sheet` | a sheet from the bottom (Android), opened and closed by an action |
| `none` | not shown on that target |

The engine lays each surface out with what Go UI has (`Within`, `Overlay`, `PanelWith`, `Sheet`).
`SetInterface` becomes the surface `interface`, placed on the second screen on the 3DS and over the
view elsewhere, as today (R7). `SetSecondScreen`'s 3D scene stays as it is. The editor's Play gains
a target to preview beside the shapes it has (R5): each target's places for its surfaces.

### Fullscreen (R6)

- **Go Renderer**: `input.State.SetFullscreen` and `Fullscreen`, requested like the captured
  pointer and applied by the window loop with GLFW (`SetMonitor`, back to the window's last place
  and size). The browser loop calls the Fullscreen API on the next key or click. The function keys
  F1 to F12, and a game pad's buttons, join `input.Key` and the state.
- **Go UI**: a scene's request reaches the window, as its captured pointer does.
- **go-engine**: `engine.SetFullscreen(bool)` and `engine.Fullscreen()`; engine.json's `window`
  (`windowed` or `fullscreen`) per target; an action `fullscreen`, bound to F11 and Alt+Enter on
  the desktop unless the game binds it otherwise.

### Alternatives not taken

- Bindings in code: engine.json lets the editor show them and a player change them without a
  build.
- One layout file per target: a table per surface keeps a surface's places side by side.

## Tasks

- [x] T1 — Go Renderer: function keys, a pad's buttons in `input.State`, read from GLFW · R1 · tests of the mapping
- [x] T2 — the target at run time · engine · R1, R4 · test per build
- [x] T3 — actions: engine.json's `actions`, the inputs' words, the table, `Input.Action(name)` · engine/actions.go · R1 · tests of each kind of input, and `reading an action allocates nothing`
- [ ] T4 — rebinding and keeping it · R3 · test `a binding changed while playing is kept and read back`
- [ ] T5 — the editor's Controls table · editor · R2 · capture; a binding changed reaches the game
- [ ] T6 — surfaces: `SetSurface`, engine.json's `surfaces`, the five places on the desktop and the 3DS · engine · R4 · captures in the editor and Azahar
- [ ] T7 — `sheet` on Android, and `SetInterface` as the `interface` surface · R4, R7 · test; the samples unchanged
- [ ] T8 — the editor previews each target's surfaces · editor · R5 · captures
- [ ] T9 — fullscreen in Go Renderer (desktop and web) and Go UI · R6 · test of the request; by hand
- [ ] T10 — `SetFullscreen`, `window` per target, the `fullscreen` action · engine · R6 · tests
- [ ] T11 — a sample game with actions and surfaces and no per-target file; docs: a guide to actions and surfaces · examples, docs · R1, R4 · captures on the desktop and in Azahar

## Result

Not built yet.
