# 🎮 Building for the 3DS

Turn a game into a `.3dsx` for the Nintendo 3DS's Homebrew Launcher, try it in an emulator, and
send it to a console over Wi-Fi.

## The tools

Open **Build**, then its **Tools** tab, and press **Install** beside each:

| Tool | What it is for |
|---|---|
| TinyGo | Builds Go for the 3DS's processor: [TinyGo for the 3DS](https://github.com/akrck02/tinygo-3ds)'s release |
| devkitPro | Packs the program for the Homebrew Launcher, and sends it with `3dslink`; its installer runs in a terminal and asks for your password |
| 3DS emulator | Azahar, to try the game without a console |

Each tool can also be set by hand: type its place, or **Browse**, then **Check**. The places are
kept for every project in `go-engine/tools.json` of your configuration folder.

## Build and play

On the **Nintendo 3DS** card:

| Button | What it does |
|---|---|
| **Build** | Writes `build/<name>.3dsx`, without the editor, with the notices of the code it carries |
| **Emulator** | Builds, and opens the game in the emulator |
| **Send** | Builds, and sends the game to a console: open the Homebrew Launcher there and press Y first |
| **Log** | What building said |

The console's address is found on the network; if it is not, set it in **Project → Settings**.

<p align="center"><img alt="A game in Azahar: the forest at night on the 3DS's top screen, at 60 frames a second" src="../images/quest/3ds.png" width="560"></p>

## How the game behaves there

| | |
|---|---|
| Screens | The game is drawn on the top screen, 400 × 240; the bottom screen is free (see [the bottom screen](8-the-3ds-bottom-screen.md)) |
| Buttons | The cross and the circle pad are the arrows (`Input.Move()`), A is Space, B is Escape, X and Y their keys |
| Speed | 60 frames a second is the aim: lights, tiles and dozens of sprites keep it in the emulator |
| Files | The scenes and assets are inside the `.3dsx`; the SD card is not read |

## Variations

- **Without the editor.** In a terminal, in the project's folder:
  `tinygo build -target=nintendo3ds -o game.3dsx .` does what Build does, given TinyGo for the 3DS
  and devkitPro.

---

[← A camera that follows](5-a-camera-that-follows.md) · [Building for Android →](7-building-for-android.md)
