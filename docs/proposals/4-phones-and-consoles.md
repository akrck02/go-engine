# 4. Phones and consoles: build, send and play from the editor

> Status: in progress · Author's approval: proposal ☑ plan ☑ (2026-10-18, "funciona en ds haz lo
> de ds y añade un export para android y un modo para jugar por adb": approved together)

## Why

The demo plays on the 3DS, but getting it there is work outside the editor: build, copy the
.3dsx, start the Homebrew Launcher. Android is not offered at all, though go-ui runs there
(go-ui-android). A game is tried where it will be played; the editor should take it there.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A 2D game draws on Android, and a phone held upright sees the whole width the camera asks for | test `a narrow screen sees more height to keep the camera's width`; capture |
| R2 | On a touch screen the game shows buttons (left, right, A, B) that hold the keys the console's do | test `a finger on a touch button holds its key` |
| R3 | The Build workspace builds an Android package (.apk) of the game, without the editor, under an application id the project's settings keep | tests `an application id is made from the game's name`, `an Android build asks for the package` |
| R4 | The Android card plays the game on the device adb sees: installs it, starts it, shows its log and stops it | test `playing on a device installs, starts and reads the log`; the emulator |
| R5 | The 3DS card sends the .3dsx to a console's Homebrew Launcher (3dslink) and opens it in an emulator when one is installed | test `sending to the console runs 3dslink with the program` |
| R6 | Play's 3DS shape shows both screens: the game on the top one, the bottom one under it | capture |

## Out of scope

- Signing for a store (go-ui-apk's `-release`): a debug package is for trying.
- Touch controls a game designs itself; the buttons are the console's, as keys.
- Choosing among several devices: the first one adb lists.

## Design

**Android build**: go-ui-apk, run as a tool of the game's module (`go tool go-ui-apk`), at a
version go-engine names; the build adds the tool and the go-ui version go-engine uses to the
game's go.mod when missing. `-tags release` leaves the editor out. The package carries arm64 (phones)
and x86_64 (the emulator on a PC).

**Playing on a device**: `adb install -r`, `adb shell monkey` to start the launcher activity,
`adb logcat -s Go` streamed into the card's log, `adb shell am force-stop` to stop.

**Touch buttons**: on Android, the interface that paints the scene paints four buttons over its
bottom corners; each frame the fingers over them set the keys, kept from frame to frame as on the
console. Where the buttons go is one function of the screen's size, used to draw and to read.

**3DS**: `3dslink` (devkitPro's tools) sends to the netloader (Y in the Homebrew Launcher); the
emulator is Azahar, or Citra, found on the path.

**go-renderer**: the GL ES shaders of sharp pictures did not link on Android (a uniform declared
with different precisions); the fix is go-renderer's, committed with the author's say.

## Tasks

- [x] T1 — 2D on Android, the camera's width · `engine/`, `scenefile/`, `editor/` · R1
- [x] T2 — touch buttons · `engine/` · R2 · test `a finger on a touch button holds its key`
- [ ] T3 — Android build and application id · `project/`, `editor/` · R3
- [ ] T4 — play on a device through adb · `project/`, `editor/` · R4
- [ ] T5 — send to the 3DS, open in an emulator · `project/`, `editor/` · R5
- [ ] T6 — Play's two 3DS screens · `editor/` · R6

## Result
