# 6. Install the tools from the editor

> Status: in progress · Author's approval: proposal ☑ plan ☑ (2026-10-18: "Todas"; tinygo-3ds commit, push and release allowed)

## Why

Proposal 5 lets a person say where TinyGo, devkitPro, a 3DS emulator and Android's SDK and NDK
are; they still have to find, download and install each by hand, and TinyGo for the 3DS is a
fork with no binaries for anyone but its author. A person who has the editor should get a game
onto their console or phone without leaving it.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | tinygo-3ds publishes releases with a TinyGo that needs nothing installed (LLVM inside), for Linux, and macOS and Windows when their builds pass | the release's files; one run on a machine with no LLVM |
| R2 | tinygo-3ds finds devkitPro in `$DEVKITPRO` (as devkitPro's own tools do), `/opt/devkitpro` when unset | test in tinygo-3ds; a 3DS build with devkitPro elsewhere |
| R3 | **Install** beside TinyGo downloads the release for this system into the user's data folder, checks it against its checksum, and sets TinyGo's place | tests with a fake server: `a release is downloaded, checked and set`, `a release that does not match its checksum is refused` |
| R4 | **Install** beside devkitPro runs devkitPro's own installer for the 3DS (`dkp-pacman -S 3ds-dev`) in a terminal, as it asks for the system's password; without dkp-pacman, it opens devkitPro's install guide | test of the command asked for; by hand on Linux |
| R5 | **Install** beside the Android SDK downloads Google's command-line tools and, once the person has read and accepted Android's licences in the panel, installs platform-tools, build-tools, a platform and the NDK | tests with a fake server and a fake sdkmanager |
| R6 | **Install** beside the 3DS emulator downloads Azahar's release for this system (Linux AppImage; the page on the others) | test with a fake server |
| R7 | The panel shows each install's progress and what it said; installs run one at a time, and the builds wait | capture |

## Out of scope

- Installing Go or a JDK (the editor runs with Go; Android's tools need Java: the panel says so).
- Updating what is installed (Install again does it).
- devkitPro's packages downloaded one by one: their servers and packages (zstd) are devkitPro's
  installer's to handle.

## Design

**tinygo-3ds**: a workflow on `main` that runs TinyGo's own release build (`make release
STATIC=1`) on Linux, macOS and Windows and attaches the archives and their SHA-256 to a release
(`v0.42.0-3ds.N`). The target's devkitPro paths become `{devkitpro}`, which the builder fills
from `$DEVKITPRO`.

**go-engine** `internal/install`: a download (HTTP, with progress), a checksum, and unpacking of
`.tar.gz` and `.zip` (standard library only), into `go-engine/` of `os.UserCacheDir`'s sibling
data folder. One function per tool returns the place to set. The editor runs installs as it runs
builds (`Editor.Run`), shows progress under the tool, and saves the place found.

## Tasks

- [x] T1 — tinygo-3ds: `{devkitpro}` from `$DEVKITPRO`; the release workflow · R1, R2
- [x] T2 — `internal/install`: download, checksum, unpack · R3
- [x] T3 — TinyGo's Install · R3, R7
- [x] T4 — devkitPro's Install · R4
- [ ] T5 — Android's Install, with the licences · R5
- [ ] T6 — Azahar's Install · R6
