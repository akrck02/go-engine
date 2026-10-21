# 9. 3D games driven by code

> Status: proposed · Author's approval: proposal ☐ plan ☐

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
| R11 | Build makes desktop programs for Windows and macOS as well as for this computer | a Windows `.exe` and a macOS program built from Linux, started on each system |

## Out of scope

- Networking itself: the game brings its own (Go for 3DS has TCP and TLS).
- Sound in the engine: the game uses Go Sound directly. R10 only keeps the loop fair to it.
- Editing meshes from code in the editor: they exist only while the game plays.
- Signing and notarising macOS programs, and Windows installers.
- A text field of the engine's own: a game shows go-ui's over the 3D window (R4's layer, for
  go-ui).

## Design

To fill once the author approves the proposal. Some parts will be Go Renderer's or Go UI's, each
on a `feature/` branch with the author's permission: stereo, bottom 3D, materials, pointer capture.

## Tasks

To fill with the design.

## Result

Not built yet.
