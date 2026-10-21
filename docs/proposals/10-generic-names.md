# 10. Names for any device

> Status: building · Author's approval: proposal ☑ (2026-10-21, "Si") plan ☑ (2026-10-21, "Si")

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
| R7 | Go Sound: a sound output that cannot start returns one general error, `ErrOutputUnavailable`, defined on every system; on the 3DS it is returned when the DSP cannot start | test in Go Sound that a program checking it builds everywhere; built for the 3DS |
| R8 | Go for 3DS: the `bottom` package becomes `secondscreen`, matching the engine | Go for 3DS's tests; the 3DS sample builds |
| R9 | The engine's code and docs are checked again with the audit's search, and nothing general is named after one device | the audit run again finds nothing in the engine |

## Out of scope

- Packages that exist for one platform only (Go Renderer's `ctr`, Go for 3DS's own) keep their
  platform's names inside.
- New targets: this only makes room for them.
- Error codes across Go Sound: it uses plain sentinel errors throughout, which is its own change.
- Keeping the old names as aliases. Nothing is released with them except go-engine v0.0.x, which
  is still experimental. The two settings in engine.json are read under both names, because they
  live in people's files.

## Design

### Go Renderer: a game pad on a computer (R1)

`input.State` gains `RightStickX` and `RightStickY`, as `StickX` and `StickY` are for the main
stick. The desktop's window loop reads the first game pad GLFW knows (`GetGamepadState`, which maps
every pad to one layout) into both sticks, each with a dead zone. The 3DS's `ctr.Controls` puts the
C-stick into the right stick once the motion controls are on. Go Renderer's `feature/` branch, as
before.

### go-engine

- **Sticks** (R1): `Input.RightStick` reads `RightStickX` and `RightStickY`, whatever filled them;
  the engine's own reading of the C-stick goes.
- **Targets** (R2, R5): the `project` package gets a `Target` type (`Desktop`, `Nintendo3DS`,
  `Android`) with `Build`, `PlayOn`, `PlayInEmulator` and `StopOn`, each choosing the target's way
  inside: today's `Build3DS`, `SendTo3DS`, `BuildAndroid`, `PlayOnAndroid`, `StopOnAndroid`,
  `EmulatorCommand` and `PlayCommand` become their private parts. Sending to the 3DS is playing
  there, as its Homebrew Launcher starts what it is sent (built: the plan's separate `SendTo` was
  one verb too many). A target without something (an emulator of its own) is refused with a code.
  The editor's `BuildTarget` becomes `project.Target`, and `ConsoleTarget` becomes `Nintendo3DS`;
  the editor's `PlayOn(target)` and `PlayInEmulator(target)` replace `PlayInWindow`, `PlayOnDevice`,
  `SendToConsole` and the old `PlayInEmulator`.
- **Settings** (R3, R5): `Settings.ApplicationID` (`applicationId`) and `Settings.DeviceAddress`
  (`deviceAddress`). Reading engine.json takes `androidId` and `consoleAddress` when the new names
  are missing, and saving writes only the new ones.
- **Tools** (R5): `build.Tools.ConsoleEmulator` becomes `Nintendo3DSEmulator` (`nintendo3dsEmulator`
  in tools.json, the old name still read); and the Play workspace's `ConsoleScreen` shape becomes
  `Nintendo3DSScreens`.
- **Licences** (R4): the editor's licence sheet takes a tool's name; Android's SDK is today's only
  tool with one.
- **Leftovers** (R6): `SetBottomScreen` goes; `SceneMenu`'s comment is made general.

### Go for 3DS and Go Sound (R7, R8)

Each on a `feature/` branch: Go for 3DS's `bottom` package is renamed `secondscreen`, its sample
with it; Go Sound defines `ErrOutputUnavailable` in `sound.go`, and the 3DS's output returns it.

### Alternatives not taken

- Aliases for the old function names: they would keep the old names alive in new code.
- One `Do(target, action)`: four verbs read better, and each has its own errors.

## Tasks

- [x] T1 — Go Renderer: `RightStickX`/`RightStickY`, the desktop's game pad into both sticks, the C-stick into the right one · R1 · tests of the dead zone and the mapping with a made-up pad state
- [x] T2 — `Input.RightStick` from the right stick · engine/controls.go · R1 · test with a made-up state
- [x] T3 — `project.Target` with `Build`, `PlayOn`, `PlayInEmulator`, `StopOn`; the old functions made private · project · R2 · tests per target; `a target that cannot be sent to says so by its code`
- [x] T4 — the editor on `project.Target`: `PlayOn(target)`, `PlayInEmulator(target)`, `Nintendo3DSScreens` · editor · R2, R5 · the editor's tests
- [ ] T5 — `applicationId`, `deviceAddress`, `nintendo3dsEmulator`, the old names still read · engine/run.go, internal/build/tools.go · R3, R5 · tests `an old androidId is read as the application id`, and the same for the others
- [ ] T6 — generic licences · editor/install.go · R4 · test of the licence flow
- [ ] T7 — `SetBottomScreen` removed, `SceneMenu`'s comment · engine · R6 · `go vet`
- [ ] T8 — Go Sound `ErrOutputUnavailable` · R7 · test; built for the 3DS
- [ ] T9 — Go for 3DS `bottom` → `secondscreen` · R8 · its tests; its sample built
- [ ] T10 — the audit again, and the reference and docs regenerated · R9 · the audit's search finds nothing in the engine

## Result

Not built yet.
