# 5. The tools that build for consoles and phones, set from the editor

> Status: in progress · Author's approval: proposal ☑ plan ☑ (2026-10-18, "tienes que hacer que
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
(`os.UserConfigDir`). Each finder is a method of it (`tools.TinyGo()`, `tools.AndroidSDK()`…):
the place chosen first, then the variables, then the usual folders. `project.Open` reads it into
`Project.Tools`, so the builds use it.

**Panel**: under the Build cards, a "Tools on this computer" panel: a row per tool with its
name, what needs it, a field for its place, and where it was found or the error. Saving writes the
file and looks for them again; nothing is looked for every frame.

## Tasks

- [x] T1 — Tools, kept in the user's folder, and the finders as its methods · `internal/build`,
  `project/` · R1, R2
- [ ] T2 — the Tools panel in Build · `editor/` · R3
