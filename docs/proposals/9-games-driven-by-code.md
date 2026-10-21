# 9. 3D games driven by code

> Status: building · Author's approval: proposal ☑ (2026-10-21, "Okay") plan ☑ (2026-10-21, "Si")

## Why

The engine serves games whose scenes are arranged in the editor: models from `.glb` files,
cameras placed by hand, keys for input. An online 3D game made mostly in code needs more.

- Its vehicles are built from designs at run time.
- They move every frame.
- Its effects glow or fade.
- Its camera chases the player.
- On the 3DS it uses the slider's depth, both screens, and every control the console has.
- On the desktop it captures the mouse.
- It ships for Windows and macOS.

Without these, such a game has to skip the engine and draw with Go Renderer on its own. One game
is waiting on them now; every 3D game in code will want them.

## What

In the order that unblocks the most. Each requirement can come out in its own version.

| Id | Requirement | Check |
|---|---|---|
| R1 | A behaviour gives a node a mesh it built at run time, and gives it another later; the old one is let go of | test `a mesh set from code is drawn, and replacing it frees the old one` |
| R2 | A node is marked as moving every frame, so it is lit on the GPU and not baked again on each turn on the 3DS | test of the flag reaching Go Renderer; on the 3DS, a turning node at full frame rate |
| R3 | A mesh from code gets a material that is unlit, additive or translucent, and double-sided; its opacity changes from frame to frame | test per kind; capture of each on the desktop and in Azahar |
| R4 | The game draws a 2D layer each frame over the 3D view: over the window on the desktop, over the top screen on the 3DS, in both eyes when stereo is on | capture on the desktop; capture of both eyes in Azahar |
| R5 | A camera has its own far plane (up to thousands of metres) and field of view, set in the editor or from code | test `a camera's far plane is the one set` |
| R6 | On the 3DS, the top screen is drawn in stereo, its depth following the slider; slider down is flat | Azahar's stereo capture; by hand on a console |
| R7 | On the 3DS, the bottom screen shows a second 3D scene with its own camera, with a go-ui interface over it | capture in Azahar |
| R8 | `engine.Input` gives the 3DS's circle pad, C-stick, gyroscope, ZL/ZR, L/R, START and touches on the bottom screen; on the desktop, the mouse's movement with the pointer captured, and the keys; each pressed this frame as well as held | tests of each mapping with a made-up input state |
| R9 | A game runs code once a frame before the behaviours (a network exchange), and asks for something to run between frames (the 3DS's system keyboard, which takes both screens) | test `the frame's hook runs before every behaviour`; test of the call between frames |
| R10 | The 3DS loop yields once a frame, so Go Sound's buffers are filled on time | a sound plays without crackles in Azahar with a slow frame |
| R11 | A game's releases carry desktop programs for Windows, macOS and Linux, as well as the one Build makes for this computer | the workflow Build writes runs on GitHub and its programs start on each system |

## Out of scope

- Networking itself: the game brings its own (Go for 3DS has TCP and TLS).
- Sound in the engine: the game uses Go Sound directly. R10 only keeps the loop fair to it.
- Editing meshes from code in the editor: they exist only while the game plays.
- Signing and notarising macOS programs, and Windows installers.
- A text field of the engine's own: a game shows go-ui's over the 3D window (R4's layer, for
  go-ui).

## Design

### What is there already

Go UI's window on the 3DS already draws the top screen in stereo with the slider (R6), paints the
scene's own drawing (`World.Overlay.Drawing`) in each eye (R4 on the 3DS), and yields each frame
(`runtime.Gosched`, R10). Go Renderer's nodes have `Moving` (R2), and its materials are lit,
unlit, translucent and double-sided (most of R3). The engine only has to reach them.

### In the engine

- **Meshes from code** (R1, R2): `Node.SetMesh(*scene.Mesh)` replaces the node's geometry, and
  `Node.SetMoving(bool)` sets Go Renderer's flag on it, kept when a scene is copied for playing.
  Games use Go Renderer's `scene.Mesh` and `scene.Material` as they are: a type of the engine's
  own would only copy them.
- **Camera** (R5): `scenefile.Camera` gains `near` and `far` (metres; 0 keeps today's 0.05 and
  2000), shown in the inspector. `Node.Camera()` and `Node.SetCamera()` let code change them.
- **Over the 3D** (R4): `engine.SetOverlay(func(list *draw2d.List, size Size))` draws each frame
  over a 3D scene. On the 3DS it becomes `World.Overlay.Drawing`, as 2D scenes do today; on the
  desktop and phones Go UI paints it under the interface. `SetBottomScreen` becomes
  `SetInterface`: the bottom screen on the 3DS, over the window elsewhere (login fields).
- **Below, in 3D** (R7): `engine.SetBottomScene(scene, camera)` gives the 3DS's bottom screen a
  scene of its own, with the interface over it.
- **Input** (R8): `engine.Input` gains `Stick()`, `CStick()`, `Gyroscope()`, `Touches()` and
  `MouseMovement()`; L, R, ZL, ZR and START become keys with `Down` and `Pressed`.
  `SetPointerCaptured(bool)` hides and holds the desktop's pointer.
- **Hooks** (R9): `engine.BeforeBehaviours(func(context *Context))` runs first in each
  step, with the frame's context (its Node nil), in the game and in the editor's Play. `engine.BetweenFrames(func())` runs once at the start of the next step. On the 3DS
  that is outside the frame being drawn, where the system keyboard can take both screens.
- **Releases** (R11): Build gains "Release workflow". It writes `.github/workflows/release.yml`
  into the game: Go builds the game on GitHub's Windows, macOS and Linux machines and attaches
  each program to a release. Building for another system from Linux is not taken: GLFW needs
  cgo, and so a C compiler and the system's own libraries for each target.

### In the author's libraries (each on a `feature/` branch, with the author's permission)

- **Go Renderer**: `Material.Additive` (light adds to what is behind: beams, sparks) in OpenGL,
  Metal and ctr. The C-stick and the gyroscope in `ctr.Reading`, as Go for 3DS reads them.
- **Go UI**: `SceneView.Below` (a scene and camera for the bottom screen, the interface painted
  over it), `Options.KeepStart` (START reaches the game instead of quitting), and
  `Options.PointerCaptured` on the desktop (Go Renderer's window loop has it).

### Versions

Each version leaves the engine working and can be used by a game at once.

| Version | Requirements | Needs |
|---|---|---|
| v0.0.4 | R1, R2, R5, R9, R10 (checked) | the engine only |
| v0.0.5 | R3, R4 | Go Renderer's additive blend |
| v0.0.6 | R6 (checked), R7, R8 | Go UI and Go Renderer as above |
| v0.0.7 | R11 | the engine only |

## Tasks

- [x] T1 — `Node.SetMesh` and `Node.SetMoving` (set by the game while it plays, so not in scene files) · engine/node.go · R1, R2 · tests `a mesh set from code is drawn, and replacing it frees the old one`, `a moving node is lit on the GPU whatever it shows`
- [x] T2 — the camera's near and far, in files, the inspector and code · scenefile, engine/scene.go, editor/inspector.go · R5 · test `a camera's far plane is the one set`
- [x] T3 — `BeforeBehaviours` and `BetweenFrames` · engine/run.go, engine/play.go · R9 · tests `the frame's hook runs before every behaviour`, `a call between frames runs once`
- [ ] T4 — check the 3DS loop yields with a slow frame; a sample with Go Sound in Azahar · R10 · by hand in Azahar
- [ ] T5 — Go Renderer: `Material.Additive` in OpenGL, Metal and ctr · R3 · test of the blend chosen; captures
- [ ] T6 — materials from code: a sample with each kind and opacity changing per frame · examples · R3 · captures on the desktop and in Azahar
- [ ] T7 — `SetOverlay` and `SetInterface` on the desktop, phones and the 3DS · engine/screen*.go · R4 · captures, both eyes in Azahar
- [ ] T8 — Go UI: `SceneView.Below`, `KeepStart`, `PointerCaptured` · R7, R8 · Go UI's tests
- [ ] T9 — Go Renderer: the C-stick and the gyroscope in `ctr.Reading` · R8 · by hand in Azahar
- [ ] T10 — `SetBottomScene` · engine/screen_nintendo3ds.go · R7 · capture in Azahar
- [ ] T11 — `Input`'s sticks, gyroscope, touches, shoulder keys, START and mouse movement; `SetPointerCaptured` · engine/play.go, engine/console.go · R8 · tests of each mapping
- [ ] T12 — the release workflow from Build · project, editor/workspace_build.go · R11 · test of the file written; one run on GitHub

## Result

Not built yet.
