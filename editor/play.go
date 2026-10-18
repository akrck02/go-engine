package editor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/akrck02/go-engine/engine"
)

// Playing in the view, beyond starting and stopping: pausing and stepping a frame at a time,
// starting again, the shape of the screen the game is shown at, what the chosen node is doing,
// and what the game and the engine say while it plays.

// oneFrame is the seconds a step moves the game on: a frame at sixty a second.
const oneFrame = 1.0 / 60

// StepFrame moves a paused game on by one frame, on the next frame the view draws.
func (editor *Editor) StepFrame() { editor.stepOnce = true }

// RestartPlay plays the scene again from how it is written in the editor.
func (editor *Editor) RestartPlay() {
	if editor.playingHere == nil {
		return
	}
	editor.PlayHere()
}

// PlayTarget is the shape of the screen the game is shown at while it plays in the view.
type PlayTarget int

const (
	// WideScreen is a computer's screen, sixteen wide for nine high.
	WideScreen PlayTarget = iota
	// ConsoleScreen is the Nintendo 3DS's top screen, 400 × 240 pixels.
	ConsoleScreen
)

// Aspect is the target's width over its height.
func (target PlayTarget) Aspect() float64 {
	if target == ConsoleScreen {
		return 400.0 / 240
	}
	return 16.0 / 9
}

// WatchLine is one thing the chosen node is doing while the game plays: what, and its value.
type WatchLine struct{ Name, Value string }

// Watch is what the chosen node of the scene played is doing: where it is, its body's velocity and
// whether it stands on the floor, the animation it plays, and its behaviours' exported fields.
func (editor *Editor) Watch() []WatchLine {
	node, found := editor.watchedNode()
	if !found {
		return nil
	}
	lines := []WatchLine{{"Position", formatVector(node.World().Position)}}
	if body, hasBody := node.Body(); hasBody {
		lines = append(lines, WatchLine{"Velocity", formatVector(body.Velocity)}, WatchLine{"On the floor", yesOrNo(body.OnFloor())})
	}
	if animation := node.Animation(); animation != "" {
		lines = append(lines, WatchLine{"Animation", animation + " · frame " + fmt.Sprint(node.Frame())})
	}
	for _, behaviour := range node.MadeBehaviours() {
		lines = append(lines, behaviourLines(behaviour)...)
	}
	return lines
}

// behaviourLines are a behaviour's exported fields, as "Behaviour · Field".
func behaviourLines(behaviour any) []WatchLine {
	value := reflect.Indirect(reflect.ValueOf(behaviour))
	if value.Kind() != reflect.Struct {
		return nil
	}
	var lines []WatchLine
	for index := 0; index < value.NumField(); index++ {
		field := value.Type().Field(index)
		if !field.IsExported() {
			continue
		}
		lines = append(lines, WatchLine{value.Type().Name() + " · " + field.Name, fmt.Sprint(value.Field(index).Interface())})
	}
	return lines
}

func yesOrNo(answer bool) string {
	if answer {
		return "yes"
	}
	return "no"
}

// PlayLog is what the game and the engine said while the scene played in the view, oldest first.
func (editor *Editor) PlayLog() []string { return editor.playLog.lines() }

// playLogLines is how many lines the play log keeps.
const playLogLines = 200

// playLog keeps the last lines logged while playing; the game may log from its own goroutines.
type playLog struct {
	mutex sync.Mutex
	kept  []string
}

func (log *playLog) add(line string) {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	log.kept = append(log.kept, line)
	if len(log.kept) > playLogLines {
		log.kept = log.kept[len(log.kept)-playLogLines:]
	}
}

func (log *playLog) lines() []string {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	return append([]string(nil), log.kept...)
}

func (log *playLog) clear() {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	log.kept = nil
}

// logHandler writes what is logged into the play log, and passes it on to the handler before.
type logHandler struct {
	log    *playLog
	before slog.Handler
}

func (handler logHandler) Enabled(context context.Context, level slog.Level) bool {
	return handler.before.Enabled(context, level)
}

func (handler logHandler) Handle(context context.Context, record slog.Record) error {
	line := strings.Builder{}
	line.WriteString(record.Level.String() + " " + record.Message)
	record.Attrs(func(attribute slog.Attr) bool {
		line.WriteString(" · " + attribute.Key + ": " + attribute.Value.String())
		return true
	})
	handler.log.add(line.String())
	return handler.before.Handle(context, record)
}

func (handler logHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	return logHandler{log: handler.log, before: handler.before.WithAttrs(attributes)}
}

func (handler logHandler) WithGroup(name string) slog.Handler {
	return logHandler{log: handler.log, before: handler.before.WithGroup(name)}
}

// listenToLog starts keeping what is logged in the play log, and gives back how to stop. What is
// logged also goes on to standard error through a handler of its own: the default one writes
// through the log package, which a new default sends back here, round and round.
func (editor *Editor) listenToLog() (stop func()) {
	before := slog.Default()
	editor.playLog.clear()
	slog.SetDefault(slog.New(logHandler{log: &editor.playLog, before: slog.NewTextHandler(os.Stderr, nil)}))
	return func() { slog.SetDefault(before) }
}

// watchedNode is the node of the played scene the chosen node stands for.
func (editor *Editor) watchedNode() (*engine.Node, bool) {
	played, playing := editor.PlayingHere()
	if !playing || editor.Selected == nil {
		return nil, false
	}
	return played.Find(editor.Selected.Path())
}
