# 12. Sound: effects and music from Go Sound

> Status: proposed · Author's approval: proposal ☐ plan ☐

## Why

Go Engine's games are silent. A jump with no sound feels half done, and io-world's Unity version
already had a jump and a calm piece of music (`jump.mp3`, `tranquil.mp3`). Go Sound, the author's
own library, plays WAV and Ogg Vorbis on the desktop, Android, the web and the 3DS. The engine
only has to use it.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A behaviour plays a sound by its file, once, with `context.Play("assets/sounds/jump.ogg")` | test `a played sound reaches the mixer once` |
| R2 | A scene can have music, which loops, set in its inspector, and keeps playing across a restart of the same scene | tests `a scene's music loops`, `restarting a scene keeps its music` |
| R3 | Sounds are read once, through the asset cache, and playing one allocates nothing a frame | the allocation test |
| R4 | Project shows sound files with a play button, and Play mutes and unmutes the game | capture; test `muting stops the game's sound` |
| R5 | The same game sounds on the desktop, Android and the 3DS | heard on the emulators (Android's reports it playing; Azahar plays it) |

## Out of scope

- MP3: Go Sound reads WAV and Ogg Vorbis. io-world's two MP3s are converted once, to Ogg.
- Sound placed in the world (louder when near): a later change, with a listener on the camera.
- An editor of sounds.

## Design

**Engine**: a `sounds` part of the runner owns one Go Sound mixer. `context.Play(file)` finds the
sound in the asset cache (reading it the first time) and starts a player. A scene's `music` is a
file and a volume, started when the scene starts. It is left alone when the next scene has the
same music, so a restart does not cut it.

**Scene files**: `Scene` gains `music: {file, volume}`.

**Editor**: Project lists `.wav` and `.ogg` files with a play button. The scene's inspector chooses
its music. Play has a mute switch, and Stop silences everything.

**Builds**: the 3DS build links Go Sound's 3DS output. Android needs nothing more than go-ui
already asks for.

**Alternative not taken**: a sound component on nodes (Unity's AudioSource). For one-off effects a
call from a behaviour is shorter, and music belongs to the scene, not to a node.

## Tasks

- [ ] T1 — `scenefile`: a scene's `music` · R2 · test: `a scene with music comes back as written`
- [ ] T2 — `engine`: the mixer, `Play`, music that loops and survives a restart · R1–R3 · tests
- [ ] T3 — builds: Go Sound on the 3DS and Android · R5
- [ ] T4 — editor: sounds in Project, music in the inspector, mute in Play · R4 · capture and test
- [ ] T5 — the quest: a coin's sound and the forest's music, in level 4 · R5

## Result

Not built yet.
