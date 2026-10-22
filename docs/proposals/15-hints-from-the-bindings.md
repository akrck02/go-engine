# 15. Hints from the bindings

> Status: proposed · Author's approval: proposal ☐ plan ☐

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
| R2 | A word has a short name a person reads on a key cap or a button (`InputLabel("Escape")` is "Esc", `"Pad Start"` is "Start", `"Left-"` is "Left", `"Touch A"` is "A", `"Alt+Enter"` stays as it is) | a table test of every kind of word |
| R3 | A game knows which kind of input the player used last (keyboard and mouse, pad, touch, a console's own buttons), so a hint names the inputs of the device in the player's hands | test: a key, then a pad's button, then a touch, each changes it |
| R4 | One call gives the hint for the device in use: the short name of the first input of that kind bound to the action (`Input.ActionHint("menu")` is "Tab" at the keyboard, "Start" with a pad) | tests on a computer with a keyboard and a pad, and on the 3DS |
| R5 | Reading words and hints allocates nothing each frame | `testing.AllocsPerRun` |
| R6 | The editor's Controls workspace and the guide to actions show how to write hints this way | the guide's example runs |

## Out of scope

- Pictures of buttons (glyphs of a pad's A, a 3DS's Start): text first; pictures need a set of
  icons of their own.
- Translating the names into other languages.
- Hints the engine draws by itself: the game decides where and how they appear.

## Open questions for the author

1. **Names.** `ActionWords`, `ActionHint` and `InputLabel`, or others you prefer?
2. **A pad's buttons.** Read as letters (A, B, X, Y) for every pad, or by the pad's own layout
   when the system tells it (an Xbox's A is a Nintendo pad's B)?

## Result

Not built yet: waiting for the author's approval of this proposal.
