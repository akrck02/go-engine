package engine

// A game runs on one target, named as engine.json names targets in its actions, surfaces and
// window: the screen files of each build set which (screen.go, screen_nintendo3ds.go,
// screen_android.go).

// The targets' names in engine.json.
const (
	DesktopTarget     = "desktop"
	Nintendo3DSTarget = "nintendo3ds"
	AndroidTarget     = "android"
	WebTarget         = "web"
)

// RunningOn is the name of the target the game runs on now.
func RunningOn() string { return runningTarget }
