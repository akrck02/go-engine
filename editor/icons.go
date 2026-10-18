package editor

// The editor's icons, from go-renderer's Lucide font, one for each meaning: an icon that stands
// for two things says neither (a test checks that no two of these are the same). An icon gets a
// word beside it unless everyone knows it already (play, save, search), and draws the result, not
// the mechanism: a window for "play in a window", not a link's arrow.
const (
	// the workspaces
	iconProject = "library-big"
	iconArt     = "palette"
	iconScene   = "map" // a scene, and the Level workspace that edits one
	iconCode    = "braces"
	iconPlay    = "play"
	iconBuild   = "hammer"

	// playing
	iconPause    = "pause"
	iconStep     = "step-forward"
	iconRestart  = "rotate-ccw"
	iconStop     = "circle-stop"
	iconInWindow = "app-window"

	// the project and its files
	iconOpenCode   = "square-code"
	iconScene2D    = "layers-2"
	iconScene3D    = "axis-3d"
	iconFolder     = "folder"
	iconOpenFolder = "folder-open"
	iconPicture    = "image"
	iconModel      = "package"
	iconSound      = "music"
	iconFile       = "file"
	iconReadAgain  = "refresh-cw"
	iconSettings   = "settings"
	iconSearch     = "search"
	iconComputer   = "monitor"
	iconConsole    = "gamepad-2"
	iconOutput     = "terminal"
	iconLog        = "scroll-text"
	iconWatch      = "activity"

	// nodes
	iconEmpty     = "circle-dashed"
	iconCube      = "box"
	iconRound     = "circle" // a sphere, or a circle in 2D
	iconFlat      = "square" // a plane, or a square in 2D
	iconCylinder  = "cylinder"
	iconCamera    = "video"
	iconTileMap   = "grid-3x3"
	iconAnimation = "film"
	iconBody      = "weight"

	// acting on nodes
	iconAdd         = "plus"
	iconDuplicate   = "copy"
	iconDelete      = "trash-2"
	iconRemove      = "circle-minus"
	iconClose       = "x"
	iconEarlier     = "arrow-up"
	iconLater       = "arrow-down"
	iconMakeChild   = "corner-down-right"
	iconMoveOut     = "corner-left-up"
	iconFrame       = "focus"
	iconLeaveCamera = "video-off"
	iconSave        = "save"

	// the Level workspace's tools and overlays
	iconSelect = "mouse-pointer-2"
	iconPaint  = "paintbrush"
	iconErase  = "eraser"
	iconFill   = "paint-bucket"
	iconHand   = "hand"
	iconCells  = "grid-2x2"

	// the editor's look
	iconLight = "sun"
	iconDark  = "moon"
)
