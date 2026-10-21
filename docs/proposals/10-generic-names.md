# 10. Names for any device

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

The guidelines now ask for generic functions: named, and shaped, for what they do on any device,
not for the first one that needed them, as nobody knows how they will be used later. An audit of
the author's Go libraries found nine places that break it. Most of them are in go-engine's
building and running for each target, and in the word "console" standing for the 3DS. A second
console, or a game pad's right stick, would already have nowhere to go.

Nothing of go-engine's is used by others yet beyond its own sample and one game in development,
so this is the cheapest moment to rename.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | `Input.RightStick` reads the second stick of whatever device has one: a New 3DS's C-stick, and a game pad's right stick on a computer | tests with a made-up 3DS reading and a made-up game pad state |
| R2 | A project builds, sends, plays and stops a game for a target through one set of functions that take the target (`Build(target)`, `SendTo(target)`, `PlayOn(target)`, `StopOn(target)`), instead of one function per device | tests per target with the tools faked; the editor's Build cards call only these |
| R3 | A game has one application id (`applicationId` in engine.json), which each target turns into what it needs; `androidId` is still read, so older projects keep working | test `an old androidId is read as the application id` |
| R4 | Accepting a tool's licence is one set of functions that take the tool (`ReadLicence`, `Licence`, `AcceptLicence`, `DeclineLicence`), not Android's own | test of the licence flow through the generic functions |
| R5 | "Console" no longer stands for the 3DS: names that mean the 3DS say so (`Nintendo3DSTarget`, the 3DS's emulator), and names that are general become general (`DeviceAddress`, `SendToDevice`); engine.json's `consoleAddress` is still read | tests of the renamed settings; an old engine.json still opens |
| R6 | `SetBottomScreen` is removed, and `Settings.SceneMenu`'s comment describes a general scene menu | `go vet`; the reference shows no 3DS words in general names |
| R7 | Go Sound: a sound output that cannot start returns a general error with a code (`sound.output-unavailable`) on every system; what the 3DS's DSP lacks goes in its details | test in Go Sound that the code is defined everywhere; built for the 3DS |
| R8 | Go for 3DS: the `bottom` package becomes `secondscreen`, matching the engine | Go for 3DS's tests; the 3DS sample builds |
| R9 | The engine's code and docs are checked again with the audit's search, and nothing general is named after one device | the audit run again finds nothing in the engine |

## Out of scope

- Packages that exist for one platform only (Go Renderer's `ctr`, Go for 3DS's own) keep their
  platform's names inside.
- New targets: this only makes room for them.
- Keeping the old names as aliases. Nothing is released with them except go-engine v0.0.x, which
  is still experimental. The two settings in engine.json are read under both names, because they
  live in people's files.

## Design

To fill once the author approves the proposal. Go Sound and Go for 3DS change on `feature/`
branches of their own, with the author's permission.

## Tasks

To fill with the design.

## Result

Not built yet.
