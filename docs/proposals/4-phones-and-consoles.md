# 4. Phones and consoles: build, send and play from the editor

> Status: done · Author's approval: proposal ☑ plan ☑ (2026-10-18, "funciona en ds haz lo
> de ds y añade un export para android y un modo para jugar por adb": approved together)

## Why

The demo plays on the 3DS, but getting it there is work outside the editor: build, copy the
.3dsx, start the Homebrew Launcher. Android is not offered at all, though go-ui runs there
(go-ui-android). A game is tried where it will be played, so the editor should take it there.

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
version go-engine names. The build adds the tool, and the go-ui version go-engine uses, to the
game's go.mod when missing. `-tags release` leaves the editor out. The package carries arm64
(phones) and x86_64 (the emulator on a PC).

**Playing on a device**, through adb:

- `adb install -r` installs it.
- `adb shell monkey` starts the launcher activity.
- `adb logcat -s Go` streams into the card's log.
- `adb shell am force-stop` stops it.

**Touch buttons**: on Android, the interface that paints the scene paints four buttons over its
bottom corners. Each frame, the fingers over them set the keys, kept from frame to frame as on the
console. Where the buttons go is one function of the screen's size, used to draw and to read.

**3DS**: `3dslink` (devkitPro's tools) sends to the netloader (Y in the Homebrew Launcher). The
emulator is Azahar, or Citra, found on the path.

**go-renderer**: the GL ES shaders of sharp pictures did not link on Android, because a uniform
was declared with different precisions. The fix is go-renderer's, committed with the author's
say.

## Tasks

- [x] T1 — 2D on Android, the camera's width · `engine/`, `scenefile/`, `editor/` · R1
- [x] T2 — touch buttons · `engine/` · R2 · test `a finger on a touch button holds its key`
- [x] T3 — Android build and application id · `project/`, `editor/` · R3
- [x] T4 — play on a device through adb · `project/`, `editor/` · R4
- [x] T5 — send to the 3DS, open in an emulator · `project/`, `editor/` · R5
- [x] T6 — Play's two 3DS screens · `editor/` · R6

## Result

**Built.**

- A 2D game draws on Android. A 2D camera's width is the least it sees across. So an upright
  phone sees more height instead of losing the sides. The demo's 15.5 m keeps the 3DS's view as it
  was, and the Level view's camera frame follows the same rule.
- On a phone held upright, the game sits at the top at its camera's shape, with touch buttons
  under it (left, right, B, A). Held sideways, it fills the screen with the buttons over its
  corners. Every key of a plugged-in keyboard still reaches the game.
- The Build workspace builds an `.apk` without the editor, with go-ui-apk as a tool of the game's
  module. go-ui's replace is written again on every build, from the go-ui the game's go-engine
  asks for.
- It plays the `.apk` on the first device adb sees, with its log, and stops it there. The log
  ends when the game does.
- It sends a `.3dsx` with 3dslink, or opens it in Azahar or Citra.
- Play's 3DS shape shows both screens.
- The settings keep the application id (its form checked) and the 3DS's address.

**Checked.** Every task's tests pass. On the Android emulator, the package built, installed,
started, logged and stopped. Touch buttons were pressed with adb, walking and jumping. Play's two
screens and the Build cards were captured.

**Not done.**

- The touch buttons are painted, with nothing for TalkBack to read. A blind player cannot find
  them yet.
- A local Go Engine's commit does not reach the `.apk`: it cannot go in `GOFLAGS`.
- Choosing among several devices, and signing for a store, were out of scope.
- Sending to a real 3DS was not tried here, with no console on the network.

**Fixed in go-renderer.** Pictures drew nothing on Android. go-renderer's GL ES shaders for sharp
pictures (feature/3ds 3b08cd9) declared a uniform with different precisions in two stages, so the
program did not link. feature/3ds bb1c275 passes it as a varying, with the author's say.

**For a person.** Play on a real phone: the buttons' size and place under a thumb. Send to the
3DS over the network.
