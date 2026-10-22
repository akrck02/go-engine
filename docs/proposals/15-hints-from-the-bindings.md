# 15. Hints from the bindings

> Status: building · Author's approval: proposal ☑ (2026-10-22, "Vale, según el mando") plan ☑ (2026-10-22, "Si")

## Why

A game tells the player which button does what: "Tab menu", "Start to pause", "Enter OK · Esc
back". Today it writes those hints by hand, once for each device, beside the bindings it already
declared in engine.json. Two copies of one fact drift apart:

- **A player rebinds an action** (`Input.Rebind`, kept on the device) and the hint still names the
  old input.
- **The author changes engine.json** and forgets one of the hints.
- **The player picks up a pad** on a computer, and the hints still name keys.

The engine knows every action's inputs on the target played, rebinding included; it only does not
say them.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A game reads the inputs an action is bound to on the target played, as engine.json's words (`Input.ActionWords("menu")` gives `["Tab", "Pad Start"]`), rebinding included | tests: the words of a bound action, of one rebound, of one not bound (none) |
| R2 | A word has a short name a person reads on a key cap or a button (`InputLabel("Escape")` is "Esc", `"Left-"` is "Left", `"Touch A"` is "A", `"Alt+Enter"` stays as it is); a pad's buttons are named as the pad in the player's hands names them: the bottom face button is "A" on an Xbox pad, "✕" on a PlayStation pad and "B" on a Nintendo pad | a table test of every kind of word, and of each pad's layout |
| R3 | A game knows which kind of input the player used last (keyboard and mouse, pad, touch, a console's own buttons), so a hint names the inputs of the device in the player's hands | test: a key, then a pad's button, then a touch, each changes it |
| R4 | One call gives the hint for the device in use: the short name of the first input of that kind bound to the action (`Input.ActionHint("menu")` is "Tab" at the keyboard, "Start" with a pad) | tests on a computer with a keyboard and a pad, and on the 3DS |
| R5 | Reading words and hints allocates nothing each frame | `testing.AllocsPerRun` |
| R6 | The editor's Controls workspace and the guide to actions show how to write hints this way | the guide's example runs |

## Out of scope

- Pictures of buttons (glyphs of a pad's A, a 3DS's Start): text first; pictures need a set of
  icons of their own.
- Translating the names into other languages.
- Hints the engine draws by itself: the game decides where and how they appear.

## Decisions

1. **Names:** `Input.ActionWords`, `Input.ActionHint` and `InputLabel`, as proposed ("Vale").
2. **A pad's buttons:** named by the pad's own layout when the system tells which pad it is
   ("según el mando"); a pad it cannot tell is read as an Xbox pad, the layout most pads copy.

## Design

**Which pad it is (Go Renderer).** GLFW reads every pad by position: its *A* is always the bottom
face button. It also gives each pad's GUID, whose vendor says who made it (Sony, Nintendo,
Microsoft…). `input` gains `PadLayout` (`XboxLayout`, the default; `PlayStationLayout`;
`NintendoLayout`) and `State.PadLayout()`; the window loop sets it when a pad connects, from the
vendor in the GUID. The browser loop reads it from the Gamepad API's `id`, which names the vendor
too. Android and the 3DS keep the default for now: a 3DS's own buttons are already named by
Nintendo (`A`, `Start`), and Android's pads are a later step.

**An action's words (engine).** The action table keeps each action's words beside its inputs, set
by `BindActions` and by `Rebind`. `Input.ActionWords(name)` returns that slice: no copy, so it
allocates nothing; the comment says not to change it.

**Short names.** `InputLabel(word string, layout input.PadLayout) string` turns a word into what a
key cap or a button says: keys by a small table (`Escape` → `Esc`, `Page Up` → `PgUp`, arrows as
`↑`), a pad's face buttons by layout (bottom, right, left, top: `A B X Y` · `✕ ○ □ △` · `B A Y X`),
its other buttons by name (`Start`, `LB`), touch controls by their name, a direction's `-`/`+`
dropped, a combination kept whole. Strings are constants or pieces of the word: nothing allocated.

**The device in use.** `Input` remembers the kind of the last input that was pressed or moved past
a dead zone: keyboard and mouse, pad, touch, or a console's own buttons
(`Input.LastInputKind()`). It is worked out in the frame's update, from the state the engine
already reads.

**One hint.** `Input.ActionHint(name)` is `InputLabel` of the first word of the action whose kind
is the last kind used; when none is of that kind, the first word. On the 3DS every word is a
console's button, so it is that.

**The editor.** Each action's card in Controls shows its hint on the device shown, so the author
sees what a player will read; the actions guide gains a *Hints* section with the code.

**Alternatives not taken.** Reading a pad's layout from its name (GLFW's `GetGamepadName`): names
vary by driver, while the vendor in the GUID does not. Pictures of the buttons: out of scope.

## Tasks

- [x] T1 — Go Renderer: `input.PadLayout`, `State.PadLayout()`, set from the pad's vendor on the desktop and in the browser · `input`, `windowloop`, `browser` · R2 · table test of vendors to layouts
- [x] T2 — an action's words kept with it, through binding and rebinding; `Input.ActionWords` · `engine/actions.go`, `engine/rebind.go`, `engine/hints.go` · R1, R5 · tests: bound, rebound, unbound; no allocation
- [x] T3 — `InputLabel` for every kind of word and each pad layout · `engine/inputlabels.go` · R2, R5 · table test
- [x] T4 — the last kind of input used, `Input.LastInputKind` · `engine/hints.go` · R3 · test: a key, a pad's button, a touch
- [ ] T5 — `Input.ActionHint` · R4, R5 · tests on a computer with keyboard and pad, and on the 3DS; no allocation
- [ ] T6 — the hint on each action's card in Controls; a *Hints* section in the actions guide · `editor`, `docs` · R6 · capture; the guide's code compiled in a test

## Result

Not built yet: waiting for the author's approval of this proposal.
