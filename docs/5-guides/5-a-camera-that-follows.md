# 🎥 A camera that follows

Make a 2D camera follow the player, smoothly, without showing past the level's edges.

## The shortest way

Choose the **Camera**. Under **Follows**, choose the node to follow (*Player*), and set how many
seconds it takes to catch up: 0 stays right on it, 0.2 feels smooth. That is all.

## What it sees

| Setting | What it means |
|---|---|
| Metres seen, top to bottom | How much of the world fits on the screen's height |
| Metres seen across, at least | A width it never shows less of: on a phone held upright, the camera sees more height instead of losing the sides |
| The scene is seen through this camera | The main camera; a scene with several plays through this one |

## Keeping inside the level

**Its view keeps inside** chooses the limits:

- **The tile maps** (the default): the camera stops where the tiles end, at the sides and the
  bottom. The sky stays open, so it rises with a jump. Tick **The top of the tile maps is a
  ceiling** for a cave.
- **A box of its own**: four numbers, the lowest and highest X and Y it may show.

**Camera bounds**, over the view, draws the box a following camera keeps its view in.

## Variations

- **Following a moving platform.** Follow the platform's node instead: anything can be followed.
- **Look through it while editing.** A camera's inspector has **Look through it**, to see the scene
  from it in the editor's view.

---

[← Lights](4-lights.md) · [Building for the 3DS →](6-building-for-the-3ds.md)
