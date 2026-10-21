package editor

import (
	"path"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-engine/project"
	"github.com/akrck02/go-engine/scenefile"
)

// The Project workspace: the project's material. Folders and scenes on the left; the assets of
// the folder chosen as thumbnails in the middle, filtered by kind or by what is searched; the
// one chosen on the right, large, with where it is used and what can be done with it; or the
// project's settings.

// The Project workspace's measures, in pixels.
const (
	projectListWidth  = 240
	projectSideWidth  = 320
	thumbnailSide     = 96
	thumbnailColumns  = 6
	previewLargest    = 280
	settingsFieldName = "Project name"
)

// assetFilters are the kinds of asset the grid can be narrowed to, in the order of its tabs.
var assetFilters = []string{"All", "Pictures", "Sprites", "Tilesets", "Models", "Sounds"}

// projectView is what the Project workspace remembers between frames.
type projectView struct {
	folder        string // the folder shown, "assets" or one under it
	filter        int    // the kind shown, by its place in assetFilters
	search        string
	chosen        string // the asset or scene chosen
	uses          []project.Use
	settings      bool // the right column shows the project's settings
	name          string
	editor        string
	applicationID string // the game's application id; empty makes one from the name
	deviceAddress string // the address of the device the game is sent to; empty looks for it
	menu          bool   // the game's interface lists the scenes
	start         int    // the start scene, by its place among the scenes
}

func (view *View) projectWorkspace(context *ui.Context) {
	view.readAssetsNow()
	if view.project.folder == "" {
		view.project.folder = project.AssetsFolder
	}
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		panel(context, "Folders and scenes", ui.Fixed(projectListWidth), func() { view.foldersAndScenes(context) })
		context.Separator()
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			view.projectBar(context)
			context.Separator()
			panel(context, "Assets grid", ui.Grow(), func() { view.assetGrid(context) })
		})
		context.Separator()
		panel(context, "Chosen asset", ui.Fixed(projectSideWidth), func() { view.projectSide(context) })
	})
}

// foldersAndScenes lists the assets' folders, to show one, and the scenes, to choose one.
func (view *View) foldersAndScenes(context *ui.Context) {
	context.Heading("Folders")
	view.folderRow(context, view.assetTree, 0)
	context.Separator()
	context.Heading("Scenes")
	for _, file := range view.Editor.Project.Scenes() {
		if context.Selectable(iconScene, sceneName(file), file == view.project.chosen) {
			view.choose(file)
		}
	}
}

// folderRow is a folder, and the folders under it.
func (view *View) folderRow(context *ui.Context, entry project.AssetEntry, depth int) {
	context.Push(entry.Path)
	defer context.Pop()
	icon := iconFolder
	if entry.Path == view.project.folder {
		icon = iconOpenFolder
	}
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: float32(depth) * 14}}, func() {
		if context.Selectable(icon, entry.Name, entry.Path == view.project.folder) {
			view.project.folder = entry.Path
		}
	})
	for _, child := range entry.Children {
		if child.Folder {
			view.folderRow(context, child, depth+1)
		}
	}
}

// projectBar is the strip over the grid: what to search, which kind, and making scenes.
func (view *View) projectBar(context *ui.Context) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(transportTall), Padding: ui.PadXY(10, 0), Gap: 8, AlignY: ui.Center}
	bar(context, layout, func() {
		context.IconSized(iconSearch, 16, context.Theme.InkSoft)
		context.Box(ui.Layout{Width: ui.Fixed(150)}, func() { context.TextField("Search", &view.project.search) })
		context.Box(ui.Layout{Width: ui.Fixed(430)}, func() { context.Tabs("Kind", &view.project.filter, assetFilters...) })
		context.Spacer()
		context.MenuButton(iconAdd, "New scene", func() {
			if context.MenuItem(iconScene2D, "2D scene", "") {
				view.newScene(scenefile.Mode2D)
			}
			if context.MenuItem(iconScene3D, "3D scene", "") {
				view.newScene(scenefile.Mode3D)
			}
		})
		if context.IconButton(iconImport, "Import model") {
			view.pickFile(ui.FilePickerOptions{Title: "Import a model", Suffixes: []string{".glb"}}, view.Editor.ImportModel)
		}
		context.Tooltip("Copy a .glb model into the project's assets/models")
		if context.IconButton(iconSettings, "Settings") {
			view.openSettings()
		}
	})
}

// shownAssets are the files the grid shows: those in the folder or under it, or anywhere while
// something is searched, of the kind chosen.
func (view *View) shownAssets() []string {
	search := strings.ToLower(strings.TrimSpace(view.project.search))
	var shown []string
	for _, file := range assetPaths(view.assetTree) {
		inFolder := strings.HasPrefix(file, view.project.folder+"/")
		if search != "" {
			inFolder = strings.Contains(strings.ToLower(file), search)
		}
		if inFolder && ofKind(file, assetFilters[view.project.filter]) {
			shown = append(shown, file)
		}
	}
	return shown
}

// ofKind tells whether a file is of a kind of the filter.
func ofKind(file, kind string) bool {
	switch kind {
	case "Pictures":
		return project.KindOf(file) == project.ImageAsset
	case "Sprites":
		return scenefile.IsSpriteFile(file)
	case "Tilesets":
		return scenefile.IsTilesetFile(file)
	case "Models":
		return project.KindOf(file) == project.ModelAsset
	case "Sounds":
		return project.KindOf(file) == project.SoundAsset
	}
	return true
}

// assetGrid is the assets shown, as thumbnails with their names.
func (view *View) assetGrid(context *ui.Context) {
	shown := view.shownAssets()
	if len(shown) == 0 {
		note(context, "Nothing here: copy files into "+view.Editor.Project.Path(view.project.folder)+", or choose another folder or kind.")
		return
	}
	context.Grid(thumbnailColumns, func() {
		for _, file := range shown {
			view.thumbnail(context, file)
		}
	})
}

// thumbnail is an asset's picture, or its kind's icon, over its name; a click chooses it.
func (view *View) thumbnail(context *ui.Context, file string) {
	look := ui.Look{Radius: 8}
	if file == view.project.chosen {
		look.Fill, look.Border, look.BorderWidth = context.Theme.Sunken, context.Theme.Primary, 2
	}
	hovered := look
	hovered.Fill = context.Theme.Sunken
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(6), Gap: 4, AlignX: ui.Center}
	context.Push(file)
	if context.Clickable("thumbnail", layout, ui.Looks{Normal: look, Hovered: hovered, Pressed: hovered}, func() {
		view.thumbnailPicture(context, file, thumbnailSide)
		context.Text(path.Base(file), ui.TextStyle{Size: context.Theme.SmallTextSize})
	}) {
		view.choose(file)
	}
	context.Pop()
}

// thumbnailPicture draws an asset as large as a side: a picture as it is, a sprite as its first
// frame, a tileset as its picture, anything else as its kind's icon.
func (view *View) thumbnailPicture(context *ui.Context, file string, side float32) {
	picture, found := view.assetPicture(file)
	if !found {
		context.Box(ui.Layout{Width: ui.Fixed(side), Height: ui.Fixed(side), AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.IconSized(assetIconOf(file), side/3, context.Theme.InkSoft)
		})
		return
	}
	width, height := picture.Size()
	scale := side / float32(max(width, height, 1))
	context.Box(ui.Layout{Width: ui.Fixed(side), Height: ui.Fixed(side), AlignX: ui.Center, AlignY: ui.Center}, func() {
		context.Image(picture, ui.Layout{Width: ui.Fixed(float32(width) * scale), Height: ui.Fixed(float32(height) * scale)}, 4)
	})
}

// assetPicture is the picture that stands for an asset, if it has one.
func (view *View) assetPicture(file string) (*draw2d.Image, bool) {
	scene := view.Editor.Scene
	switch {
	case project.KindOf(file) == project.ImageAsset:
		picture, err := scene.Picture(file)
		return picture, err == nil
	case scenefile.IsSpriteFile(file):
		return view.spriteThumbnail(file)
	case scenefile.IsTilesetFile(file):
		return view.tilesetThumbnail(file)
	}
	return nil, false
}

// assetIconOf is the icon of an asset's kind.
func assetIconOf(file string) string {
	switch {
	case scenefile.IsSpriteFile(file):
		return iconSprite
	case scenefile.IsTilesetFile(file):
		return iconTileset
	case strings.HasSuffix(file, ".scene.json"):
		return iconScene
	}
	return assetIcon(project.KindOf(file))
}

// choose chooses an asset or a scene, and finds where it is used.
func (view *View) choose(file string) {
	view.project.chosen, view.project.settings = file, false
	view.project.uses = nil
	if !strings.HasSuffix(file, ".scene.json") {
		view.project.uses = view.Editor.Project.UsesOf(file)
	}
}

// projectSide is the right column: the project's settings, or the asset or scene chosen.
func (view *View) projectSide(context *ui.Context) {
	switch {
	case view.project.settings:
		view.settingsPanel(context)
	case view.project.chosen == "":
		note(context, "Choose an asset to see it large, where it is used, and what can be done with it; or a scene to open it.")
	case strings.HasSuffix(view.project.chosen, ".scene.json"):
		view.scenePreview(context, view.project.chosen)
	default:
		view.assetPreview(context, view.project.chosen)
	}
}

// scenePreview is a scene chosen, with the button that opens it in Level.
func (view *View) scenePreview(context *ui.Context, file string) {
	context.Row(func() {
		context.IconSized(iconScene, 18, context.Theme.Primary)
		context.Heading(sceneName(file))
	})
	context.Caption(file)
	if context.IconButton(iconScene, "Open in Level") && view.openScene(file) {
		view.showWorkspace(LevelWorkspace)
	}
}

// assetPreview is an asset chosen, large, with where it is used and what can be done with it.
func (view *View) assetPreview(context *ui.Context, file string) {
	context.Row(func() {
		context.IconSized(assetIconOf(file), 18, context.Theme.Primary)
		context.Heading(path.Base(file))
	})
	context.Caption(file)
	view.thumbnailPicture(context, file, previewLargest)
	if picture, found := view.assetPicture(file); found && project.KindOf(file) == project.ImageAsset {
		width, height := picture.Size()
		context.Caption(FieldText(float64(width)) + " × " + FieldText(float64(height)) + " pixels")
	}
	view.assetActions(context, file)
	context.Separator()
	view.usesList(context)
}

// assetActions are what can be done with an asset: place it in the scene, or open it in Art.
func (view *View) assetActions(context *ui.Context, file string) {
	editor := view.Editor
	placeable := editor.Scene.Is2D() && (project.KindOf(file) == project.ImageAsset || scenefile.IsSpriteFile(file)) ||
		!editor.Scene.Is2D() && project.KindOf(file) == project.ModelAsset
	if placeable && context.IconButton(iconAdd, "Place in the scene") {
		view.place(file)
	}
	if (scenefile.IsSpriteFile(file) || scenefile.IsTilesetFile(file)) && context.IconButton(iconArt, "Open in Art") {
		view.openArt(file)
		view.showWorkspace(ArtWorkspace)
	}
	if project.KindOf(file) == project.ImageAsset && context.IconButton(iconArt, "Make it a sprite or a tileset") {
		view.choosePicture(file)
		view.showWorkspace(ArtWorkspace)
	}
}

// place adds a node showing an asset to the scene open, and goes to Level to place it.
func (view *View) place(file string) {
	if project.KindOf(file) == project.ModelAsset {
		view.Editor.AddModel(file)
	} else {
		view.Editor.AddSprite(file)
	}
	view.showWorkspace(LevelWorkspace)
}

// usesList is where the asset chosen is used.
func (view *View) usesList(context *ui.Context) {
	context.Heading("Used in")
	if len(view.project.uses) == 0 {
		note(context, "Nothing uses it yet.")
		return
	}
	for _, use := range view.project.uses {
		where := use.File
		if use.Node != "" {
			where = sceneName(use.File) + " · " + use.Node
		}
		context.Row(func() {
			context.IconSized(assetIconOf(use.File), 14, context.Theme.InkSoft)
			context.Label(where)
		})
	}
}

// openSettings shows the project's settings, ready to change.
func (view *View) openSettings() {
	settings := view.Editor.Project.Settings
	view.project.settings, view.project.name, view.project.editor = true, settings.Name, settings.CodeEditor
	view.project.applicationID, view.project.deviceAddress = settings.ApplicationID, settings.DeviceAddress
	view.project.menu = settings.SceneMenu
	view.project.start = max(0, indexOf(view.Editor.Project.Scenes(), settings.StartScene))
}

// settingsPanel changes the project's name, the scene it starts with, its code editor, its
// application id and the address of the device it is sent to.
func (view *View) settingsPanel(context *ui.Context) {
	editor := view.Editor
	settings := &editor.Project.Settings
	context.Row(func() {
		context.IconSized(iconSettings, 18, context.Theme.Primary)
		context.Heading("Settings")
	})
	context.Caption("Name")
	context.TextField(settingsFieldName, &view.project.name)
	scenes := editor.Project.Scenes()
	context.Caption("The scene the game starts with")
	context.Select("Start scene", &view.project.start, scenes)
	context.Caption("Code editor: a command, as zed, code or nvim; empty tries $VISUAL and $EDITOR")
	context.TextField("Code editor", &view.project.editor)
	context.Caption("Application id, as io.example.game; empty is " + editor.Project.ApplicationID())
	context.TextField("Application id", &view.project.applicationID)
	context.Caption("The address of the device the game is sent to (a 3DS); empty looks for it on the network")
	context.TextField("Device address", &view.project.deviceAddress)
	context.Checkbox("List the scenes in the game's interface, to play any", &view.project.menu)
	if context.PrimaryIconButton(iconSave, "Save settings") {
		settings.Name, settings.CodeEditor = strings.TrimSpace(view.project.name), strings.TrimSpace(view.project.editor)
		settings.ApplicationID, settings.DeviceAddress = strings.TrimSpace(view.project.applicationID), strings.TrimSpace(view.project.deviceAddress)
		settings.SceneMenu = view.project.menu
		if view.project.start < len(scenes) {
			settings.StartScene = scenes[view.project.start]
		}
		view.reportSettings(editor.Project.SaveSettings())
	}
}

func (view *View) reportSettings(err error) {
	if err != nil {
		view.Editor.say(err.Error(), true)
		return
	}
	view.Editor.say("Saved the project's settings", false)
}

// spriteThumbnail is a sprite's first frame.
func (view *View) spriteThumbnail(file string) (*draw2d.Image, bool) {
	sprite, err := view.Editor.Scene.SpriteFile(file)
	if err != nil {
		return nil, false
	}
	frames, err := view.Editor.Scene.Frames(sprite.Picture, &sprite.SpriteSheet)
	if err != nil || len(frames) == 0 {
		return nil, false
	}
	return frames[0], true
}

// tilesetThumbnail is a tileset's whole picture.
func (view *View) tilesetThumbnail(file string) (*draw2d.Image, bool) {
	tileset, err := view.Editor.Scene.TilesetFile(file)
	if err != nil {
		return nil, false
	}
	picture, err := view.Editor.Scene.Picture(tileset.Picture)
	return picture, err == nil
}
