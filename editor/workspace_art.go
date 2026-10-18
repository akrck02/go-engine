package editor

import (
	"math"
	"strconv"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/scenefile"
)

// The Art workspace: the project's sprites, tilesets and pictures on the left; the picture open,
// large, with its cut drawn over it, in the middle (a click adds a frame to the animation chosen,
// or turns a tile solid); the animation's frames and a preview under it; the cut, the animations
// and saving on the right.

// The Art workspace's measures, in pixels.
const (
	artListWidth    = 260
	artSideWidth    = 320
	artStripHeight  = 132
	stripFrameSide  = 56
	previewSide     = 96
	canvasMargin    = 24
	frameNumberSize = 11
)

// Colours drawn over the picture.
var (
	cutLine        = draw2d.Color{1, 1, 1, 0.55}
	cutShadow      = draw2d.Color{0, 0, 0, 0.35}
	animationFrame = draw2d.Color{0.95, 0.55, 0.1, 1}
	numberInk      = draw2d.Color{1, 1, 1, 0.9}
	canvasGround   = draw2d.Color{0.16, 0.17, 0.2, 1}
)

// artView is what the Art workspace remembers between frames.
type artView struct {
	picture   string      // a picture chosen that is not a sprite or a tileset yet
	animation int         // the animation chosen, by its place in the sheet
	canvas    draw2d.Rect // where the picture was drawn, and how large
	scale     float32
	pressing  bool // the button was down over the canvas on the frame before
	since     time.Time
}

func (view *View) artWorkspace(context *ui.Context) {
	view.readAssetsNow()
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Art files", ui.Fixed(artListWidth), func() { view.artList(context) })
		context.Separator()
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.artCanvas(context)
			if _, isSprite, _ := view.Editor.ArtOpen(); isSprite {
				context.Separator()
				view.animationStrip(context)
			}
		})
		context.Separator()
		panel(context, "Art settings", ui.Fixed(artSideWidth), func() { view.artSettings(context) })
	})
}

// artList is the project's sprites, tilesets and pictures, and the chosen node's own sheet.
func (view *View) artList(context *ui.Context) {
	view.ownSheetCard(context)
	open, _, _ := view.Editor.ArtOpen()
	paths := assetPaths(view.assetTree)
	view.artGroup(context, "Sprites", iconAnimation, paths, scenefile.IsSpriteFile, open)
	view.artGroup(context, "Tilesets", iconTileMap, paths, scenefile.IsTilesetFile, open)
	context.Heading("Pictures")
	for _, picture := range pictures(view.assetTree) {
		if context.Selectable(iconPicture, picture, picture == view.art.picture && open == "") {
			view.Editor.CloseArt()
			view.art.picture = picture
		}
	}
}

// artGroup lists the files of one kind, to open one.
func (view *View) artGroup(context *ui.Context, heading, icon string, paths []string, belongs func(string) bool, open string) {
	context.Heading(heading)
	shown := 0
	for _, path := range paths {
		if !belongs(path) {
			continue
		}
		shown++
		if context.Selectable(icon, path, path == open) {
			view.openArt(path)
		}
	}
	if shown == 0 {
		note(context, "None yet: choose a picture below and make one.")
	}
}

func (view *View) openArt(path string) {
	if err := view.Editor.OpenArt(path); err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.art.picture, view.art.animation = "", 0
	view.forgetAnimationTexts()
}

// ownSheetCard offers to move the chosen node's own sheet into a sprite's file.
func (view *View) ownSheetCard(context *ui.Context) {
	node := view.Editor.Selected
	if node == nil {
		return
	}
	if _, own := node.Sheet(); !own {
		return
	}
	note(context, node.Name+" cuts its picture into frames itself, as scenes did before sprites had files.")
	if context.IconButton(iconAnimation, "Make it a sprite file") {
		if _, err := view.Editor.MakeSpriteOfNode(); err != nil {
			view.Editor.say(err.Error(), true)
		}
		view.art.animation = 0
	}
	context.Separator()
}

// artPicture is the picture the canvas shows: the open file's, or the picture chosen.
func (view *View) artPicture() string {
	switch {
	case view.Editor.art.sprite != nil:
		return view.Editor.art.sprite.Picture
	case view.Editor.art.tileset != nil:
		return view.Editor.art.tileset.Picture
	}
	return view.art.picture
}

// artCanvas is the picture, as large as fits in whole pixels, its cut drawn over it.
func (view *View) artCanvas(context *ui.Context) {
	file := view.artPicture()
	if file == "" {
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(24)}, ui.Look{Fill: context.Theme.Background}, func() {
			note(context, "Choose a sprite, a tileset or a picture on the left.")
		})
		return
	}
	picture, err := view.Editor.Scene.Picture(file)
	if err != nil {
		context.Text(err.Error(), ui.TextStyle{Color: context.Theme.Warning})
		return
	}
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func(area draw2d.Rect, list *draw2d.List) {
		list.Rectangle(area, 0, canvasGround, draw2d.Color{}, 0)
		view.art.canvas, view.art.scale = fitWhole(picture, area)
		list.Image(picture, view.art.canvas, 0, draw2d.Color{1, 1, 1, 1})
		view.drawCut(list)
	})
	view.clickCanvas(context)
}

// fitWhole is where a picture is drawn in an area, centred, scaled by a whole number so its
// pixels stay square, and the scale.
func fitWhole(picture *draw2d.Image, area draw2d.Rect) (draw2d.Rect, float32) {
	width, height := picture.Size()
	room := math.Min(float64(area.Width-2*canvasMargin)/float64(max(width, 1)), float64(area.Height-2*canvasMargin)/float64(max(height, 1)))
	scale := float32(max(math.Floor(room), 1))
	if room < 1 {
		scale = float32(room)
	}
	shown := draw2d.Rect{Width: float32(width) * scale, Height: float32(height) * scale}
	shown.X, shown.Y = area.X+(area.Width-shown.Width)/2, area.Y+(area.Height-shown.Height)/2
	return shown, scale
}

// artCells is how many columns and rows the open file cuts its picture into.
func (view *View) artCells() (columns, rows int) {
	editor := view.Editor
	switch {
	case editor.art.sprite != nil:
		return editor.art.sprite.Columns, editor.art.sprite.Rows
	case editor.art.tileset != nil:
		picture, err := editor.Scene.Picture(editor.art.tileset.Picture)
		if err != nil || editor.art.tileset.TileSize <= 0 {
			return 0, 0
		}
		width, height := picture.Size()
		return width / editor.art.tileset.TileSize, height / editor.art.tileset.TileSize
	}
	return 0, 0
}

// drawCut draws the cut over the picture: each cell outlined and numbered; the frames of the
// animation chosen, or the solid tiles, stand out.
func (view *View) drawCut(list *draw2d.List) {
	columns, rows := view.artCells()
	if columns <= 0 || rows <= 0 {
		return
	}
	cellWidth, cellHeight := view.art.canvas.Width/float32(columns), view.art.canvas.Height/float32(rows)
	for cell := 0; cell < columns*rows; cell++ {
		rect := draw2d.Rect{X: view.art.canvas.X + float32(cell%columns)*cellWidth, Y: view.art.canvas.Y + float32(cell/columns)*cellHeight, Width: cellWidth, Height: cellHeight}
		list.Rectangle(rect, 0, view.cellFill(cell), cutShadow, 1)
		list.Rectangle(rect, 0, draw2d.Color{}, view.cellLine(cell), view.cellLineWidth(cell))
		if cellWidth >= 3*frameNumberSize {
			list.Text(strconv.Itoa(cell), rect.X+3, rect.Y+2, frameNumberSize, draw2d.Bold, numberInk)
		}
	}
}

// cellFill tints a solid tile of the open tileset.
func (view *View) cellFill(cell int) draw2d.Color {
	if tileset := view.Editor.art.tileset; tileset != nil && tileset.IsSolid(cell) {
		return solidFill
	}
	return draw2d.Color{}
}

// cellLine outlines the frames of the animation chosen.
func (view *View) cellLine(cell int) draw2d.Color {
	if view.inChosenAnimation(cell) {
		return animationFrame
	}
	return cutLine
}

func (view *View) cellLineWidth(cell int) float32 {
	if view.inChosenAnimation(cell) {
		return 3
	}
	return 1
}

// inChosenAnimation tells whether a frame is in the animation chosen.
func (view *View) inChosenAnimation(frame int) bool {
	animation, _, found := view.Editor.selectedAnimation(view.art.animation)
	if !found || view.Editor.art.sprite == nil {
		return false
	}
	for _, each := range animation.Frames {
		if each == frame {
			return true
		}
	}
	return false
}

// clickCanvas adds the frame clicked to the animation chosen, or turns the tile clicked solid.
func (view *View) clickCanvas(context *ui.Context) {
	pressing := context.Pressing()
	clicked := pressing && !view.art.pressing
	view.art.pressing = pressing
	if !clicked {
		return
	}
	x, y := context.Pointer()
	cell, inside := view.cellAt(x, y)
	if !inside {
		return
	}
	if view.Editor.art.tileset != nil {
		view.Editor.ToggleSolidTile(cell)
		return
	}
	view.Editor.AddFrame(view.art.animation, cell)
}

// cellAt is the cell of the open file's cut under a point of the window.
func (view *View) cellAt(x, y float32) (int, bool) {
	columns, rows := view.artCells()
	canvas := view.art.canvas
	if columns <= 0 || rows <= 0 || x < canvas.X || y < canvas.Y || x >= canvas.X+canvas.Width || y >= canvas.Y+canvas.Height {
		return 0, false
	}
	column := int((x - canvas.X) / (canvas.Width / float32(columns)))
	row := int((y - canvas.Y) / (canvas.Height / float32(rows)))
	return row*columns + column, true
}

// artSettings is the right column: what the open file is cut into, and saving it; or making a
// sprite or a tileset of the picture chosen.
func (view *View) artSettings(context *ui.Context) {
	editor := view.Editor
	open, isSprite, isTileset := editor.ArtOpen()
	switch {
	case isSprite:
		view.spriteSettings(context, open)
	case isTileset:
		view.tilesetSettings(context, open)
	case view.art.picture != "":
		view.makeArtButtons(context)
	default:
		note(context, "A picture becomes a sprite (frames and animations) or a tileset (tiles to paint levels with), in a file of its own beside it that every scene shares.")
	}
}

// makeArtButtons make the picture chosen a sprite or a tileset.
func (view *View) makeArtButtons(context *ui.Context) {
	editor := view.Editor
	context.Heading("Make it…")
	context.Caption(view.art.picture)
	if context.IconButton(iconAnimation, "A sprite") {
		view.reportArt(editor.MakeSprite(view.art.picture, 1, 1))
	}
	context.Tooltip("Cut it into frames and make animations of them")
	if context.IconButton(iconTileMap, "A tileset") {
		view.reportArt(editor.MakeTileset(view.art.picture, defaultTileSize))
	}
	context.Tooltip("Cut it into tiles to paint levels with")
}

// defaultTileSize is a new tileset's tiles, in pixels a side.
const defaultTileSize = 16

func (view *View) reportArt(file string, err error) {
	if err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.art.picture, view.art.animation = "", 0
	view.Editor.say("Made "+file, false)
}

// spriteSettings are the sprite's cut, its animations and saving it.
func (view *View) spriteSettings(context *ui.Context, file string) {
	editor := view.Editor
	view.artFileHeading(context, iconAnimation, file)
	sprite := editor.art.sprite
	context.Caption("Frames across and down")
	columns, rows := float64(sprite.Columns), float64(sprite.Rows)
	context.Push("Cut")
	context.Grid(2, func() {
		if numberField(context, "Columns", &columns, sheetCellLimits, "%.0f across") ||
			numberField(context, "Rows", &rows, sheetCellLimits, "%.0f down") {
			editor.ResizeSheet(int(columns), int(rows))
		}
	})
	context.Pop()
	context.Separator()
	view.animationList(context, &sprite.SpriteSheet)
}

// tilesetSettings are the tileset's tile size, and saving it.
func (view *View) tilesetSettings(context *ui.Context, file string) {
	editor := view.Editor
	view.artFileHeading(context, iconTileMap, file)
	size := float64(editor.art.tileset.TileSize)
	context.Caption("Pixels a tile, a side")
	if numberField(context, "Tile size", &size, tileSizeLimits, "%.0f px") {
		editor.SetArtTileSize(int(size))
	}
	note(context, "Click a tile on the picture to make bodies stand on it; click it again to let them through. Solid tiles: "+solidList(editor.art.tileset.Solid))
}

func solidList(solid []int) string {
	if len(solid) == 0 {
		return "none."
	}
	return writtenFrames(solid) + "."
}

// artFileHeading names the open file, with its saved state and the button that saves it.
func (view *View) artFileHeading(context *ui.Context, icon, file string) {
	editor := view.Editor
	context.Row(func() {
		context.IconSized(icon, 18, context.Theme.Primary)
		context.Heading(file)
	})
	context.Row(func() {
		if context.IconButton(iconSave, "Save") {
			if err := editor.SaveArt(); err != nil {
				editor.say(err.Error(), true)
			}
		}
		if editor.ArtSaved() {
			context.Text("Saved", ui.TextStyle{Color: context.Theme.Positive, Size: context.Theme.SmallTextSize})
		} else {
			context.Text("Unsaved changes", ui.TextStyle{Color: context.Theme.Warning, Size: context.Theme.SmallTextSize})
		}
	})
	context.Separator()
}

// animationList chooses an animation of the sheet, adds one, and edits the one chosen.
func (view *View) animationList(context *ui.Context, sheet *scenefile.SpriteSheet) {
	editor := view.Editor
	context.Row(func() {
		context.Heading("Animations")
		context.Spacer()
		if context.IconButton(iconAdd, "Add") {
			view.art.animation = editor.AddAnimation()
		}
	})
	for index, animation := range sheet.Animations {
		label := animation.Name
		if sheet.Start == animation.Name {
			label += " · starts"
		}
		if context.Selectable(iconAnimation, label, index == view.art.animation) {
			view.art.animation = index
		}
	}
	if view.art.animation < len(sheet.Animations) {
		view.chosenAnimation(context, sheet, view.art.animation)
	}
}

// chosenAnimation edits the animation chosen: its name, pace, looping, and whether it starts.
func (view *View) chosenAnimation(context *ui.Context, sheet *scenefile.SpriteSheet, index int) {
	editor := view.Editor
	animation := &sheet.Animations[index]
	context.Separator()
	view.animationNameField(context, index, animation.Name)
	if numberField(context, "Pace", &animation.FramesPerSecond, paceLimits, "%.0f frames a second") {
		editor.sheetChanged()
	}
	if context.Checkbox("Loops", &animation.Loop) {
		editor.sheetChanged()
	}
	if sheet.Start != animation.Name && context.Button("Play it when the scene starts") {
		editor.StartWith(index)
	}
	if context.IconButton(iconRemove, "Remove the animation") {
		editor.RemoveAnimation(index)
		view.forgetAnimationTexts()
		view.art.animation = 0
	}
}

// animationStrip is the chosen animation's frames, in order (a click takes one out), and a preview
// that plays them at their pace.
func (view *View) animationStrip(context *ui.Context) {
	editor := view.Editor
	sprite := editor.art.sprite
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(artStripHeight), Padding: ui.PadXY(12, 8), Gap: 12, AlignY: ui.Center}
	bar(context, layout, func() {
		animation, _, found := editor.selectedAnimation(view.art.animation)
		if !found {
			note(context, "Add an animation on the right, then click frames on the picture to put them in it.")
			return
		}
		frames, err := editor.Scene.Frames(sprite.Picture, &sprite.SpriteSheet)
		if err != nil {
			context.Text(err.Error(), ui.TextStyle{Color: context.Theme.Warning})
			return
		}
		view.preview(context, animation, frames)
		context.Separator()
		view.frameStrip(context, animation, frames)
	})
}

// preview plays the animation at its pace.
func (view *View) preview(context *ui.Context, animation *scenefile.Animation, frames []*draw2d.Image) {
	if view.art.since.IsZero() {
		view.art.since = time.Now()
	}
	shown := previewFrame(animation, time.Since(view.art.since).Seconds())
	context.Box(ui.Layout{Direction: ui.Vertical, Gap: 4, AlignX: ui.Center}, func() {
		if shown >= 0 && shown < len(frames) {
			context.Image(frames[shown], ui.Layout{Width: ui.Fixed(previewSide), Height: ui.Fixed(previewSide)}, 4)
		}
		context.Caption("Preview")
	})
	context.RequestFrame(time.Second / 30)
}

// previewFrame is the frame an animation shows after some seconds, looping or held on its last.
func previewFrame(animation *scenefile.Animation, seconds float64) int {
	if len(animation.Frames) == 0 {
		return -1
	}
	step := int(seconds * animation.FramesPerSecond)
	if animation.Loop {
		return animation.Frames[step%len(animation.Frames)]
	}
	return animation.Frames[min(step, len(animation.Frames)-1)]
}

// frameStrip is the animation's frames in order; a click on one takes it out.
func (view *View) frameStrip(context *ui.Context, animation *scenefile.Animation, frames []*draw2d.Image) {
	editor := view.Editor
	inside := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), AlignY: ui.Center}
	context.ScrollArea("Frames", inside, func() {
		// the scroll area stacks what it holds: a row of its own keeps the frames side by side
		context.Box(ui.Layout{Direction: ui.Horizontal, Gap: 6, AlignY: ui.Center}, func() {
			for place, frame := range animation.Frames {
				if frame < len(frames) {
					view.stripFrame(context, frames[frame], frame, place)
				}
			}
			if context.IconButton(iconClose, "Clear") {
				editor.ClearFrames(view.art.animation)
			}
			context.Tooltip("Take every frame out, to click them again")
		})
	})
}

// stripFrame is one frame of the animation in the strip; a click takes it out.
func (view *View) stripFrame(context *ui.Context, picture *draw2d.Image, frame, place int) {
	context.Push(strconv.Itoa(place))
	defer context.Pop()
	look := ui.Look{Fill: canvasGround, Radius: 6, Border: context.Theme.ControlLine, BorderWidth: 1}
	layout := ui.Layout{Width: ui.Fixed(stripFrameSide), Height: ui.Fixed(stripFrameSide), Padding: ui.Pad(3)}
	if context.Clickable("frame", layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		context.Image(picture, ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, 3)
	}) {
		view.Editor.RemoveFrameAt(view.art.animation, place)
	}
	context.Tooltip("Frame " + strconv.Itoa(frame) + ": click to take it out")
}
