# 💡 Lights

Light a scene, in 2D or 3D: an ambient for everything, lamps where they shine, and in 2D, pictures
that glow. To see it done step by step, follow [level 5 of the quest](../4-the-quest/5-night-in-the-forest.md).

<p align="center"><img alt="The editor's Level workspace on a 2D street at night: a street lamp chosen, its inner and outer circles drawn over the view, a window glowing on a house, and the inspector's Light section" src="../images/editor-lights.png" width="680"></p>

## A lamp

Press **+** and choose **Light**, or choose any node and press **Make it a light** in its
inspector. A **point** light shines all around its node: full within its **inner** radius, fading to
nothing at its **outer** one, in metres. Its **intensity** is how strong it is: 1 a plain lamp, up
to 2. It moves with its node, and grows with its scale.

## The ambient

Choose nothing, and in the scene's inspector tick **An ambient light of its own**:

| | 2D | 3D |
|---|---|---|
| What it is | The light that reaches everything, in a colour and an intensity | How strong the sky's light is |
| None | Daylight when the scene has lights, no lighting at all when it has none | The sky lights as it is |

White at 1 leaves the colours as they are; a dark blue at 0.4 is a night.

## A picture that glows (2D)

A **sprite light** shines through its node's picture, in its place, and the picture itself is not
drawn: a window, a sign, a coin's gleam. Put it on a child of the node it lights, with the same
sprite, so the node keeps its picture and gains the glow.

## In the editor

- **Lights**, over the 2D view, draws each light's two circles; in 3D, a chosen lamp shows two rings.
- **Lit**, over the 2D view, shows the scene lit as the game will be, or its colours unlit.

## From code

A behaviour can change a light while the game plays:

```go
func (alarm *Alarm) Update(context *engine.Context) {
	light, found := context.Node.Light()
	if !found {
		return
	}
	alarm.strength = 1 + math.Sin(context.Time*6) // a field of Alarm: the light keeps pointing at it
	light.Intensity = &alarm.strength
}
```

## How much it costs

Lights cost nothing in a scene that has none. In 2D, all of a frame's lights are drawn into one
picture at half the screen's resolution, then laid over the scene, so many lights cost little. In
3D, each thing drawn takes the four nearest lamps that reach it. On the 3DS, 3D lamps light at the
vertices: a floor shows a lamp's pool only where it has vertices.

---

[← Bodies and physics](3-bodies-and-physics.md) · [A camera that follows →](5-a-camera-that-follows.md)
