# 🪟 Surfaces on each screen

This guide is for whoever makes a game that shows more than its scene (a map, a log, a score) on
several devices. At the end the game will paint each of them once, engine.json will say where
each goes on a 3DS, a computer and a phone, and whether a computer shows the game in a window or
on the whole screen.

## Paint a surface

A **surface** is something the game shows besides its scene, painted with
[Go UI](https://github.com/akrck02/go-ui) every frame. The game names it and paints it, and says
nothing of where it goes:

```go
func main() {
	engine.SetSurface("log", keeperLog)
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}

// keeperLog is the lighthouse keeper's notes.
func keeperLog(context *ui.Context) {
	context.Label("Fog at dawn. Two ships passed.")
}
```

With nothing else, the log goes where a game's interface always went: on the second screen of a
3DS, over the view on a computer and a phone. The game's interface (`engine.SetInterface`) is a
surface too, named `interface`.

## Place it on each target

engine.json's `surfaces` says, for each surface and each target, its **place**:

```json
"surfaces": {
	"log": {
		"nintendo3ds": {"place": "second-screen"},
		"desktop": {"place": "window", "align": "top-right", "width": 260, "action": "log"},
		"android": {"place": "sheet", "action": "log"}
	}
}
```

| Place | Where it goes |
|---|---|
| `over` | Over the view, at an edge or a corner (`align`); with no `align`, on the whole view |
| `window` | A floating window over the view, titled with the surface's name, opened and closed by its `action` |
| `second-screen` | The 3DS's bottom screen, touchable; several share it, one under another |
| `sheet` | A sheet from the bottom of a phone, opened by its `action`, closed by it or by the back button |
| `none` | Not shown on that target |

A place a target lacks becomes the nearest it has: a second screen on a computer is `over`, a
sheet on a 3DS is a `window`. On the 3DS's top screen no finger reaches, so what is placed there
is to be read, not touched. The [file reference](../6-reference/10-files.md) lists every key.

## See each target's layout in the editor

In the **Play** workspace, the tabs over the game choose the target it is shown at: **16:9** (a
computer), **3DS** (its two screens) and **Phone**. The game's surfaces go where engine.json places
them on that target, so one click shows the 3DS's second screen and the next the phone's sheet.

## In a window or on the whole screen

On a computer and in a browser, engine.json's `window` says how the game starts on each target:

```json
"window": {"desktop": "fullscreen", "web": "windowed"}
```

The player switches with **F11** or **Alt+Enter** (the action `fullscreen`, which the game may bind
to other keys), and the game with `engine.SetFullscreen(true)`; `engine.Fullscreen()` tells which.
A browser lets a page take the whole screen only when it answers a key or a click, so there the
request waits for the next one. Consoles and phones are always on their whole screen.

## A sample to read

[Night watch](../../examples/watch) plays the same on every device with no file of its own for
any: the keeper turns the lamp (`turn`) and marks ships in a log (`mark`). The lamp's heading sits
in a corner; the log is a window on a computer (Tab), the 3DS's second screen, and a sheet on a
phone (its Log button).

## Variations

- **Open with the game.** `"open": true` shows a window or a sheet when the game starts.
- **A fixed size.** `width` and `height` in pixels; 0 is the size of what it holds. On the second
  screen, a height gives a surface that much of it.
- **Read the scene menu.** `engine.SetInterface` gets the game's `Menu` (its scenes, the one
  playing); a surface that needs it is the interface.
- **The game's own fonts.** `engine.SetFonts(overlay.Fonts{Regular: pixelFont})` before `Run`
  draws all the game's text in them: its drawing over the scene, its surfaces and the second
  screen, on every target. `engine.TextMeasurer()` measures a line in those fonts, for a layout of
  the game's own.

---

[← Actions and controls](9-actions-and-controls.md) · [The reference →](../6-reference/0-packages.md)
