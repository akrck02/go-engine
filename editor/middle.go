package editor

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
)

// frame turns the view to the chosen node.
func (view *View) frame() {
	if view.Editor.Scene.Is2D() {
		view.viewport.flat.frame(view.Editor.Selected)
		return
	}
	view.viewport.frame(view.Editor.Selected)
}

// middle is the view of the scene: played, or edited with the tool rail beside it and, while a
// tile map is chosen, the palette under it.
func (view *View) middle(context *ui.Context) {
	column := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(8)}
	if played, playing := view.Editor.PlayingHere(); playing {
		context.Box(column, func() { view.playingView(context, played) })
		return
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.toolRail(context)
			context.Box(column, func() {
				view.viewStrip(context)
				view.sceneView(context)
			})
		})
		view.paletteUnderView(context)
	})
}

// paletteUnderView shows the palette while a tile map of a 2D scene is chosen.
func (view *View) paletteUnderView(context *ui.Context) {
	chosen := view.Editor.Selected
	if chosen == nil || !view.Editor.Scene.Is2D() {
		return
	}
	if tiles, isTileMap := chosen.Tiles(); isTileMap {
		context.Separator()
		view.paletteStrip(context, chosen, tiles)
	}
}

// toolRail is the column of tools beside the 2D view.
func (view *View) toolRail(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(48), Height: ui.Grow(), Padding: ui.PadXY(6, 10), Gap: 4, AlignX: ui.Center}
	bar(context, layout, func() {
		for _, tool := range ToolsFor(view.Editor.Scene.Is2D()) {
			view.toolButton(context, tool)
		}
	})
	context.Separator()
}

// toolButton takes a tool; the one taken stands out.
func (view *View) toolButton(context *ui.Context, tool Tool) {
	taken := view.Editor.Tool == tool
	ink, look := context.Theme.InkSoft, ui.Look{Radius: 8}
	if taken {
		ink, look.Fill = context.Theme.Primary, context.Theme.Sunken
	}
	hovered := look
	hovered.Fill = context.Theme.Sunken
	layout := ui.Layout{Width: ui.Fixed(36), Height: ui.Fixed(36), AlignX: ui.Center, AlignY: ui.Center}
	if context.Clickable(tool.Name(), layout, ui.Looks{Normal: look, Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized(tool.Icon(), 18, ink)
	}) {
		view.Editor.Tool = tool
	}
	context.Tooltip(tool.Name() + " (" + tool.Letter() + "): " + tool.Does())
}

// floatingLook is the look of what floats over the view: a card with a shadow.
func floatingLook(context *ui.Context) ui.Look {
	return ui.Look{Fill: context.Theme.Surface, Radius: 8, Shadow: true}
}

// playingView says the scene is playing and, in 2D, paints it (in 3D the window shows the scene
// behind the interface).
func (view *View) playingView(context *ui.Context, played *engine.Scene) {
	context.Frame(ui.Layout{Padding: ui.PadXY(10, 6), Gap: 6, AlignY: ui.Center}, floatingLook(context), func() {
		context.IconSized(iconPlay, 16, context.Theme.Positive)
		context.Text("Playing", ui.TextStyle{Weight: draw2d.Medium})
		context.Caption("Escape stops")
	})
	if !played.Is2D() {
		return
	}
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
		paintPlaying(played, area, list)
	})
}

// viewStrip is the strip over the view: in 2D what to draw over the scene, in 3D the camera's
// projection; framing the chosen node, and leaving a camera looked through.
func (view *View) viewStrip(context *ui.Context) {
	context.Frame(ui.Layout{Padding: ui.PadXY(6, 4), Gap: 6, AlignY: ui.Center}, floatingLook(context), func() {
		if view.Editor.Scene.Is2D() {
			view.overlayToggles(context)
		} else {
			view.projectionSwitch(context)
		}
		if context.FlatIconButton(iconFrame, "Frame") {
			view.frame()
		}
		context.Tooltip("Centre the view on the chosen node (F)")
		view.leaveCameraButton(context)
	})
}

// overlayToggles switch what the 2D view draws over the scene.
func (view *View) overlayToggles(context *ui.Context) {
	flat := view.viewport.flat
	if context.Selectable(iconBody, "Bodies", flat.showBodies) {
		flat.showBodies = !flat.showBodies
	}
	context.Tooltip("Outline bodies and fill solid tiles")
	if context.Selectable(iconCamera, "Camera bounds", flat.showBounds) {
		flat.showBounds = !flat.showBounds
	}
	context.Tooltip("Draw the box a following camera keeps its view in")
	if context.Selectable(iconLamp, "Lights", flat.showLights) {
		flat.showLights = !flat.showLights
	}
	context.Tooltip("Draw how far each light reaches")
	if context.Selectable(iconLit, "Lit", flat.showLit) {
		flat.showLit = !flat.showLit
	}
	context.Tooltip("Light the scene as the game will; off shows its colours unlit")
}

// dimensionsSwitch makes the scene 3D or 2D.
func (view *View) dimensionsSwitch(context *ui.Context) {
	dimensions := 0
	if view.Editor.Scene.Is2D() {
		dimensions = 1
	}
	context.Box(ui.Layout{Width: ui.Fixed(110)}, func() {
		if context.Tabs("Dimensions", &dimensions, "3D", "2D") {
			view.Editor.SetMode([]scenefile.Mode{scenefile.Mode3D, scenefile.Mode2D}[dimensions])
		}
	})
	context.Tooltip("Whether the scene is 3D (shapes and models) or 2D (sprites and flat shapes)")
}

// projectionSwitch chooses how the editor's camera shows a 3D scene.
func (view *View) projectionSwitch(context *ui.Context) {
	mode := int(view.viewport.mode)
	context.Box(ui.Layout{Width: ui.Fixed(330)}, func() {
		if context.Tabs("View", &mode, "Perspective", "Orthographic", "Isometric") {
			view.viewport.setMode(ViewMode(mode))
		}
	})
}

// leaveCameraButton goes back to the editor's camera after looking through one of the scene's.
func (view *View) leaveCameraButton(context *ui.Context) {
	if view.viewport.through == nil {
		return
	}
	if context.FlatIconButton(iconLeaveCamera, "") {
		view.viewport.through = nil
	}
	context.Tooltip("Back to the editor's camera")
}

// sceneView is the scene being edited: painted here in 2D, behind the interface in 3D, with a line
// of how to move around it.
func (view *View) sceneView(context *ui.Context) {
	if view.Editor.Scene.Is2D() {
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
			view.viewport.flat.paint(view.Editor, area, list)
		})
		context.Caption("Middle-drag or Hand (H) moves the view · scroll zooms · Select (V) chooses and drags · Paint (B), Erase (E), Fill (G) work on the chosen tile map")
		return
	}
	context.Spacer()
	context.Text(view.viewHint(), ui.TextStyle{Color: context.Theme.InkQuiet, Size: context.Theme.SmallTextSize})
}

// viewHint is how to move around the 3D view, as it is.
func (view *View) viewHint() string {
	if view.viewport.mode == ViewIsometric {
		return "Isometric is a fixed angle: drag to pan · scroll to zoom · click to choose · drag the chosen to move it"
	}
	return "Drag to turn · right-drag or Hand (H) to pan · scroll to zoom · click to choose (Shift adds) · drag the chosen to move it"
}
