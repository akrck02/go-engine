# 5. The tools that build for consoles and phones, set from the editor

> Status: done · Author's approval: proposal ☑ plan ☑ (2026-10-18, "tienes que hacer que
> se pueda configurar tinygo y android desde ui": approved together)

## Why

Building for the 3DS and Android needs tools found outside the editor: TinyGo, devkitPro, a 3DS
emulator, the Android SDK and NDK. Today they are found through environment variables or the
usual folders, and when they are not, the only way out is a variable set before starting the
editor. A person should see what the editor found and point it at what it did not, in the editor.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | Where each tool is can be set in the editor, and is kept for every project on this computer (not in the project, which others share) | test `the tools' places are kept between sessions` |
| R2 | A place set in the editor comes before the environment and the usual folders; a place set where nothing is says so by its code | tests `a tool set in the editor is used first`, `a tool set where nothing is is said by its code` |
| R3 | The Build workspace shows, for each tool, where it was found or why not, and what it is for | capture |

## Out of scope

- Installing the tools (the panel links to how).
- The go command itself: the editor runs with one.

## Design

**Tools** (`internal/build`): a `Tools` value holds the place of each tool the person chose,
read from and saved to `go-engine/tools.json` in the user's configuration folder
(`os.UserConfigDir`). Each finder is a method of it (`tools.FindTinyGo()`,
`tools.FindAndroidSDK()`…): the place chosen first (a program or a folder, as the tool is), then
the variables, then the usual folders. `project.Open` reads it into `Project.Tools`, so the builds
use it, and keeps why it could not be read in `Project.ToolsUnread`.

**Panel**: under the Build cards, a "Tools on this computer" panel: a row per tool with its
name, what needs it, a field for its place, and where it was found or the error with its meaning.
Check looks where the places typed say; saving writes the file and looks again (not while the game
builds or plays: the build reads them). Nothing is looked for every frame.

## Tasks

- [x] T1 — Tools, kept in the user's folder, and the finders as its methods · `internal/build`,
  `project/` · R1, R2
- [x] T2 — the Tools panel in Build · `editor/` · R3

## Result

**Built.** `build.Tools`, kept in `go-engine/tools.json` of the user's configuration folder, with
a finder for each tool (TinyGo, devkitPro's 3dslink, a 3DS emulator, the Android SDK and its adb,
the NDK); a place chosen comes first, and one where the wrong thing is (nothing, a folder for a
program, a devkitPro with no 3dslink, an SDK with no adb) is said by `build.tool-place-wrong`.
Build's **Tools on this computer** shows each tool, what needs it, the place typed, and where it
was found or why not, with Check and Save; a file that could not be read is said there.

**Checked.** The tests of each requirement; captures of the panel with tools found and missing.

**Then, asked by the author.** A file picker, go-ui's own (`ui.FilePicker`, feature/3ds 3183ea9,
committed with the author's say): **Browse** beside each tool chooses a program or a folder, as
the tool needs, and types it in its field to check and save; Project's **Import model** chooses a
`.glb` and copies it into `assets/models`, which nothing in the editor offered before.

**For a person.** Browse with a real mouse: double clicks, and going through deep folders.
