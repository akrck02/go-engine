# 14. Stable releases of the engine and its libraries

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

go-engine is tagged (v0.0.1 to v0.0.3), but it depends on loose commits of Go UI's and Go Renderer's main, not on releases. A game made with a tagged engine can therefore pull libraries nobody released, and nothing proves a version works on every device before it ships:

- **The libraries are not released.** Go Renderer's last tag is v0.2.0 and Go UI's v0.1.1; since then surfaces, strokes, fullscreen, `uitest`, `uidrive` and `DrawOver` have landed on their main.
- **`-update` follows a branch.** It moves a project to go-engine's main, not to its latest release; new projects pin whatever main held.
- **Only go-engine runs its tests on every push.** Go UI and Go Renderer only publish their docs; nothing builds the 3DS or Android on a push.
- **Devices are checked by hand.** Metal on a Mac, a real 3DS and a real phone are tried when someone remembers to.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | Go Renderer, Go UI, Go Sound, Go for 3DS and Go UI for Android are released with notes (Go Renderer v0.3.0, Go UI v0.2.0 first) before the engine that needs them | each repository's release page lists its notes; go-engine's go.mod names only tags |
| R2 | go-engine is released by tag, with notes that list what changed and what breaks, with what to do (`ShowInterface` went in proposal 11) | the release page of each version |
| R3 | `-update` and new projects take go-engine's latest release; main only when asked (`-engine main`) | test: `-update` with no `-engine` pins the latest tag |
| R4 | The editor and a built game say their version from the tag | test of the version read from the build |
| R5 | Every repository checks each push: format, vet and tests on Linux, macOS and Windows | each repository's workflow runs green on a push |
| R6 | go-engine's checks also build the samples, the 3DS with TinyGo and an Android APK | the workflow's jobs for each target |
| R7 | Before a release, scripts drive a sample on the Android emulator and in Azahar and report each step (`uidrive`), so every device is proven without hands | a release job runs them and keeps their reports |
| R8 | A short checklist for what only the author can try: Metal on a Mac, a real 3DS, a real phone | the checklist in `docs/8-contributing`, ticked in each release's notes |
| R9 | While versions are v0.x, the API may change, but never silently: a change that breaks a game is in the notes; v1.0 names what stays stable (`engine`, engine.json, the scene file) | the notes of each version; a page on what v1.0 promises |

## Out of scope

- Packaging the editor as an installer for each system: the editor is still installed with Go.
- Releasing games: projects already have their own release workflow.
- v1.0 itself: this sets the way to get there.

## Open questions for the author

1. **Who tags.** You, by hand after review, or a workflow when a tag is pushed from main?
2. **How often.** A release when a proposal is done, or on a calendar?
3. **The devices you test.** Which Mac, 3DS and phone count for the checklist?

## Result

Not built yet: waiting for the author's approval of this proposal.
