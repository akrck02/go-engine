package game

// World is a game being played: the scene now, where its things stand, and what the player has
// done in it. It plays on a grid, a step at a time, so every platform plays it the same way.
type World struct {
	game   *Game
	scene  int
	things []placed
	player int // index into things, or -1 when the scene has no player

	Steps     int // steps taken in this scene
	Collected int // things collected in this scene
	ToCollect int // how many the scene holds
	State     State
	Notice    string // a line for the player: why the goal does not open yet
}

// State is where a world stands.
type State int

const (
	Playing State = iota
	Won           // the last scene's goal was reached
)

// placed is a thing in play: its kind and its cell, and whether it is gone (collected).
type placed struct {
	kind *Kind
	x, y int
	gone bool
}

// Direction is a step on the grid.
type Direction struct{ X, Y int }

var (
	Up    = Direction{0, -1}
	Down  = Direction{0, 1}
	Left  = Direction{-1, 0}
	Right = Direction{1, 0}
)

// NewWorld starts a game at its first scene.
func NewWorld(game *Game) *World {
	world := &World{game: game}
	world.Start(0)
	return world
}

// Game is the game played.
func (world *World) Game() *Game { return world.game }

// Scene is the scene played now.
func (world *World) Scene() *Scene { return &world.game.Scenes[world.scene] }

// SceneNumber is the scene played now, from 0.
func (world *World) SceneNumber() int { return world.scene }

// Start plays a scene from its beginning.
func (world *World) Start(scene int) {
	world.scene = min(max(scene, 0), len(world.game.Scenes)-1)
	world.things = world.things[:0]
	world.player = -1
	world.Steps, world.Collected, world.ToCollect, world.Notice = 0, 0, 0, ""
	world.State = Playing
	for _, thing := range world.Scene().Things {
		kind, found := world.game.Kind(thing.Kind)
		if !found {
			continue
		}
		if kind.Player && world.player < 0 {
			world.player = len(world.things)
		}
		if kind.Collectable {
			world.ToCollect++
		}
		world.things = append(world.things, placed{kind: kind, x: thing.X, y: thing.Y})
	}
}

// Restart plays the scene now from its beginning.
func (world *World) Restart() { world.Start(world.scene) }

// Move takes a step with the player, pushing what can be pushed, and does what the cell it lands
// on says. A step into a wall, out of the scene or against a box that cannot move is not taken.
func (world *World) Move(direction Direction) {
	if world.State != Playing || world.player < 0 {
		return
	}
	player := &world.things[world.player]
	toX, toY := player.x+direction.X, player.y+direction.Y
	if !world.Scene().Inside(toX, toY) || world.solidAt(toX, toY) {
		return
	}
	if pushed := world.pushableAt(toX, toY); pushed >= 0 {
		beyondX, beyondY := toX+direction.X, toY+direction.Y
		if !world.Scene().Inside(beyondX, beyondY) || world.solidAt(beyondX, beyondY) || world.pushableAt(beyondX, beyondY) >= 0 {
			return
		}
		world.things[pushed].x, world.things[pushed].y = beyondX, beyondY
	}
	player.x, player.y = toX, toY
	world.Steps++
	world.Notice = ""
	world.arrive(toX, toY)
}

// arrive does what the cell the player stepped on says: collect, start again, or end the scene.
func (world *World) arrive(x, y int) {
	for index := range world.things {
		thing := &world.things[index]
		if thing.gone || index == world.player || thing.x != x || thing.y != y {
			continue
		}
		if thing.kind.Collectable {
			thing.gone = true
			world.Collected++
		}
	}
	for index := range world.things {
		thing := &world.things[index]
		if thing.gone || index == world.player || thing.x != x || thing.y != y {
			continue
		}
		if thing.kind.Hazard {
			world.Restart()
			return
		}
		if thing.kind.Goal {
			world.reachGoal()
			return
		}
	}
}

// reachGoal ends the scene if everything was collected, and goes on to the next one, or wins.
func (world *World) reachGoal() {
	if world.Collected < world.ToCollect {
		world.Notice = "Collect everything first"
		return
	}
	if world.scene+1 >= len(world.game.Scenes) {
		world.State = Won
		return
	}
	world.Start(world.scene + 1)
}

// solidAt tells whether a solid thing stands in a cell.
func (world *World) solidAt(x, y int) bool {
	for index, thing := range world.things {
		if !thing.gone && index != world.player && thing.kind.Solid && thing.x == x && thing.y == y {
			return true
		}
	}
	return false
}

// pushableAt finds a pushable thing in a cell, or -1.
func (world *World) pushableAt(x, y int) int {
	for index, thing := range world.things {
		if !thing.gone && index != world.player && thing.kind.Pushable && thing.x == x && thing.y == y {
			return index
		}
	}
	return -1
}

// Visit calls a function for each thing still in play, in the order they were placed (the player
// last, so it is drawn over what it stands on).
func (world *World) Visit(visit func(kind *Kind, x, y int)) {
	for index, thing := range world.things {
		if !thing.gone && index != world.player {
			visit(thing.kind, thing.x, thing.y)
		}
	}
	if world.player >= 0 {
		player := world.things[world.player]
		visit(player.kind, player.x, player.y)
	}
}
