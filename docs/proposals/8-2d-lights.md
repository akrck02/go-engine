# 8. Lights, in 2D and 3D

> Status: done, pending the author's review and a check of Metal on a Mac · Author's approval: proposal ☑ plan ☑ (2026-10-19, "Con metal": with
> Metal) · go-renderer's part, on its own `feature/lights` branch, is pushed only with the
> author's say (its CLAUDE.md)

## Why

io-world's start screen in Unity was lit: a global light at 72 % dimmed the whole scene, the lamp
cast a warm circle (`#faffa0`, 3.5 metres), the logo, "Press start" and the snail glowed through
their own pictures, and two thin lights lined "Demo". Made again in Go Engine (proposal 7), it
is flat: the engine draws 2D with no light at all. Light is what gives a factory its gloom and a
lamp its warmth; any 2D game will want it, on the desktop, on phones and on the 3DS.

Unity draws 2D lights in a pass of their own: it paints the lights into a light picture, cleared
to the global light, and then multiplies the sprites by it. go-renderer's 2D drawing only paints
one thing over another, with transparency: it has no way to add light nor to multiply by it.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | go-renderer's 2D lists draw a light layer: lights added into a light picture cleared to an ambient colour, which then multiplies what the layer drew before it | tests of `draw2d`'s commands; a test that renders a known list on OpenGL into pixels and compares |
| R2 | Three kinds of light: **global** (the ambient colour and intensity), **point** (a colour, intensity, inner and outer radius), **sprite** (a picture that glows as a light, tinted) | tests per kind |
| R3 | The light layer draws on OpenGL (desktop), OpenGL ES (Android, the web), Metal (macOS) and the 3DS (citro3d) | captures on each; the 3DS in Azahar; Metal by the author on a Mac |
| R4 | A 2D node can be a light (`light` in scene files: kind, colour, intensity, radii); a scene's ambient light is the scene's own setting; lights move with their nodes | tests in `engine` |
| R5 | Level shows and edits lights: an inspector for them, their radii drawn in the view, and the view lit as the game will be | captures |
| R6 | Lights cost little: a 2D frame with lights allocates nothing, and on the 3DS a screen of eight lights keeps 60 frames a second | the allocation test; a measure in Azahar and on the console |
| R7 | Io World's start screen has the Unity menu's seven lights, converted from `MainMenu.unity` by its script | capture; the author on the console |
| R8 | *(Added 2026-10-19, "debe funcionar en 2D y 3D".)* A 3D node can be a point light too, with the same `light` (colour, intensity, radii) and the same fade as in 2D; a 3D scene's `ambient` scales its sky light | tests in `sceneplan` and `engine`; captures on OpenGL, GL ES and the 3DS |
| R9 | 3D lights cost little: each draw takes at most four lights, the nearest that reach it, chosen once a frame on the processor for every backend; a frame allocates nothing for them | tests of the choice; the allocation test |
| R10 | A lighting inside a clip lights only the clip, so a lit view inside the editor leaves the panels around it as they are | a pixel test |

## Out of scope

- Shadows cast by sprites or by 3D lights, normal maps, spot lights (a later kind of the same
  `light`), and dynamic lights on the 3DS's GPU (its fragment lighting: 3D lights are baked into
  vertex colours there, which suits lights that seldom move).
- Blend modes as a general tool for programs (only what the light layer needs).

## Design

**go-renderer** (`draw2d`, `opengl`, `gles`, `ctr`), on `feature/lights`:

- `draw2d`: `list.PointLight(x, y, inner, outer, colour)` and `list.SpriteLight(picture, area,
  colour)` add lights (a colour's alpha is its intensity), and `list.Lighting(ambient)` lights
  everything drawn before it with the lights since the last one, inside the clip it is made in;
  what comes after (an interface) stays unlit. *(Built so instead of BeginLights/EndLights: one
  command marks where lighting happens, and nothing has to be closed.)*
- OpenGL and GL ES: a light target the size of the screen, at half its resolution (lights are
  soft), made like the backdrop's targets and kept between frames: cleared to the ambient
  colour, lights drawn into it additively (a point light is one quad whose shader makes its
  falloff), then drawn over the frame with a multiplying blend. One more draw call per light layer.
- 3DS (`ctr`): citro3d draws into a texture render target (`C3D_RenderTargetCreateFromTex`) at
  half the top screen's resolution, 256 by 128 (sides that are powers of two), and the PICA200
  adds and multiplies with its own blend modes (`GPU_BLEND_ADD`, `GPU_DST_COLOR`). With no
  fragment shader, a point light's fade is read from a texture of sixteen discs, and the ambient
  is painted in, since citro3d's clears run before a frame's draws.
- Metal: the same light target as a texture of its own, drawn with the same blends; written as
  the backdrop's Metal part was. It cannot be built nor run where this is written (Linux): the
  author checks it on a Mac.

**go-engine**:

- `scenefile`: a node's `light`: `{kind: point|sprite, colour, intensity, inner, outer}`; a
  sprite light uses the node's own picture, which it is drawn as instead of being shown. A
  scene's `ambient` colour and intensity is its global light; a scene with lights and no ambient
  is lit by daylight. *(Changed while building: no "global" kind on nodes, since the scene's
  ambient is the global light, and no falloff setting: every backend fades with the same smooth
  curve.)*
- `engine`: a 2D scene draws its sprites, then its lights (in their nodes' places and scales)
  into one light layer, then what goes above (the interface). Lights are found once a frame
  into a reused list, as the drawn nodes are.
- `editor`: a Light section in Level's inspector (kind, colour, intensity, radii), radii drawn
  as circles over the view when "Lights" is on, and Play drawn lit.

**Io World**: `tools/make_start_screen.py` reads the Unity menu's `Light2D` components (type,
colour, intensity, radii, the sprite a sprite light uses) into the start screen's nodes.

**3D (added with R8–R9).** go-renderer's `scene` gets a list of point lights (position, linear
colour times intensity, inner and outer radius). `sceneplan` chooses, for each draw, the four
nearest lights whose outer sphere meets its bounds (*built as "nearest", not "strongest": the
simpler rule Godot's mobile renderer uses, and stable while nothing moves*), listed by index, so
every backend gets the same short list.
OpenGL and GL ES add them in their lit shading next to the sun, as Lambert light times the same
smoothstep fade as 2D, from a fixed array of four uniforms (within OpenGL ES 2's uniform budget,
and no light loop beyond four). The 3DS bakes them into its vertex colours with the sun, and bakes
again when a light changes. Metal follows with T4b. The engine fills the list each frame from the
3D nodes that are lights, and a 3D scene's `ambient` multiplies its sky light. Researched against
Unity URP (Light2D, forward's per-object light list), Godot 4 (CanvasModulate, PointLight2D, 8+8
lights per mesh on mobile), Bevy and the PICA200's fragment lighting.

**Alternatives not taken**: drawing lights as translucent halos with no light picture (option A,
cheap, but light cannot darken what it does not cover nor multiply colours); a light per sprite
computed on the processor (does not scale, and the 3DS has a GPU for it).

## Tasks

- [x] T1 — `draw2d`: the light layer's commands and their tests · go-renderer · R1, R2
- [x] T2 — OpenGL: the light target and its blend; pixels compared · go-renderer · R1, R3
- [x] T3 — GL ES: the same for Android and the web · go-renderer · R3
- [x] T4 — 3DS: the light target in citro3d; Azahar captures, and a measure · go-renderer · R3, R6
- [x] T4b — Metal: the light target and its blends, unbuilt here · go-renderer · R3
- [x] T5 — `scenefile` and `engine`: lights on nodes, the ambient, drawn each frame without
  allocating · go-engine · R4, R6
- [x] T6 — Level: the inspector (2D and 3D), the ambient, the radii and a mark for each light, Play lit · go-engine · R5
- [x] T7 — Io World's start screen lit from Unity's lights · go-engine-io-world · R7
- [x] T8 — A lighting lights only its clip, on every backend; the 3DS paints its ambient instead
  of clearing it · go-renderer · R10
- [x] T9 — 3D point lights: `scene`, the choice in `sceneplan`, OpenGL and GL ES shading · go-renderer · R8, R9
- [x] T10 — 3D lights baked on the 3DS · go-renderer · R8
- [x] T11 — The engine's 3D lights and ambient; the audit's 2D fixes (culling, node scale, one
  screen rectangle) · go-engine · R8, R9

## Result

Built on go-renderer's `dev` (feature/lights merged with feature/3ds-moving-light), go-engine's
`feature/lights` and Io World's `main`. Each requirement's check, as run:

| Id | Check | Result |
|---|---|---|
| R1 | draw2d and painter tests; OpenGL pixel tests | pass |
| R2 | a test per kind; pixels for point and sprite lights | pass |
| R3 | OpenGL and WebGL (Chromium) pixels; Azahar captures; Metal | pass, but Metal: written, unbuilt (the author on a Mac) |
| R4 | `engine` tests: light on nodes, ambient, scale, hidden, out of sight | pass |
| R5 | editor tests; captures of Level in 2D and 3D | pass |
| R6 | allocation tests (2D frame, sync, choosing); eight lights at 60 FPS in Azahar | pass; the console is still to be measured |
| R7 | Io World's test; captures in the editor and in Azahar | pass; the author on the console |
| R8 | `sceneplan` and `engine` tests; OpenGL pixels; captures on WebGL and the 3DS | pass |
| R9 | the choice's tests, with instances; the allocation test | pass |
| R10 | a pixel test on OpenGL; a capture of both 3DS screens | pass |

What the author still has to review or check:

- **Metal** was written on Linux and never built: `go test ./metal` on a Mac compiles its shaders.
- **The console:** frame times with eight 2D lights, and a lamp moving over baked meshes.
- **On the 3DS**, a node marked Moving is not lit by lamps. A floor shows a lamp's pool only where
  it has vertices.
- **GL ES** lights the screen only, like its backdrops (no offscreen capture).
- **Changes beyond the proposal:**
  - a body's copy no longer shares its zone's pull (with a test);
  - the unknown body kind's message;
  - the broken GLB test deleted at the author's word;
  - a one-frame flicker of backdrops fixed in OpenGL.

Researched against Unity URP's Light2D, Godot 4 (CanvasModulate, PointLight2D, mobile's light
limits), Bevy and the PICA200's fragment lighting; audited by a sister session, whose findings
(clip, culling, node scale, banding, the order of lights) are in; reviewed after, and its
findings fixed but the GL ES framebuffer (noted above).

