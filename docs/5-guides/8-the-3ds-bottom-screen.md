# 🕹️ The 3DS's bottom screen

Use the 3DS's second screen: a ready-made menu of the game's scenes, or a screen drawn by the game
itself.

## A menu of the scenes

In **Project → Settings**, tick **List the scenes on the 3DS's bottom screen, to play any**. The
bottom screen then lists every scene; a touch plays it. It is handy while making a game, to jump
to any level.

## A screen of your own

`engine.SetBottomScreen` draws the bottom screen with [Go UI](https://github.com/akrck02/go-ui), in
the game's own look. Call it in `main.go`, before `engine.Run`:

```go
func main() {
	engine.Register(Player{})
	engine.SetBottomScreen(bottomScreen)
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}

// bottomScreen lists the levels, the one playing chosen; a touch plays another.
func bottomScreen(context *ui.Context, menu engine.Menu) {
	for _, scene := range menu.Scenes() {
		playing := menu.Playing() != nil && menu.Playing().Name == scene.Name
		if context.Selectable("map", scene.Name, playing) {
			menu.Play(scene.File)
		}
	}
}
```

| Call | What it gives |
|---|---|
| `menu.Scenes()` | The project's scenes, each with its file and name |
| `menu.Playing()` | The scene playing, or nil |
| `menu.Play(file)` | Plays another scene |

It needs `import ui "github.com/akrck02/go-ui"`, which the engine already brings. The function runs every frame the bottom screen is drawn, on the 3DS only: elsewhere there is no
bottom screen and it is never called. The screen is 320 × 240 pixels, and touch is its pointer.

## Variations

- **A map, a score, the controls.** Anything Go UI draws can be there: text, pictures, buttons.
- **Its own pictures.** Read a picture once (`draw2d.DecodeImage`) and paint it with
  `context.Paint`; the [Io World](https://github.com/akrck02/io-world/tree/main/IoWorldGo) game draws
  its menu of levels from its own bricks this way.

---

[← Building for Android](7-building-for-android.md) · [The reference →](../6-reference/0-packages.md)
