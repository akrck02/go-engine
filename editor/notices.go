package editor

import (
	"strconv"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// What the editor says (a scene saved, a build that failed) is shown as notices floating over the
// bottom right corner, not in a bar: a bar would grow and shift with each message. A notice that
// only informs goes away by itself; one that says what failed stays until the person closes it.

// Notice is something the editor said, and when.
type Notice struct {
	Text   string
	Failed bool
	At     time.Time
}

// How notices are kept and shown.
const (
	noticeLifetime     = 5 * time.Second // how long a notice that only informs is shown
	maxNotices         = 3               // the oldest goes when another comes
	noticeWidth        = 360             // in pixels
	noticeIconSize     = 16              // in pixels
	noticeLeave        = 200 * time.Millisecond
	workspaceBarHeight = 52 // in pixels: the notices float above it
)

// say tells the person something: what went wrong when failed.
func (editor *Editor) say(message string, failed bool) {
	notice := Notice{Text: strings.Join(strings.Fields(message), " "), Failed: failed, At: time.Now()}
	editor.notices = append(editor.notices, notice)
	if len(editor.notices) > maxNotices {
		editor.notices = editor.notices[len(editor.notices)-maxNotices:]
	}
}

// Notices are what the editor said that is still shown, the oldest first.
func (editor *Editor) Notices() []Notice { return editor.notices }

// forgetOldNotices lets go of the notices that only inform once their time is over.
func (editor *Editor) forgetOldNotices(now time.Time) {
	kept := editor.notices[:0]
	for _, notice := range editor.notices {
		if notice.Failed || now.Sub(notice.At) < noticeLifetime {
			kept = append(kept, notice)
		}
	}
	editor.notices = kept
}

// CloseNotice lets go of a notice, by its place among Notices.
func (editor *Editor) CloseNotice(index int) {
	if index < 0 || index >= len(editor.notices) {
		return
	}
	editor.notices = append(editor.notices[:index], editor.notices[index+1:]...)
}

// notices shows what the editor said over the bottom right corner, and asks for a frame when the
// next one that only informs is due to go.
func (view *View) notices(context *ui.Context) {
	editor := view.Editor
	editor.forgetOldNotices(time.Now())
	if len(editor.notices) == 0 {
		return
	}
	options := ui.LayerOptions{AlignX: ui.End, AlignY: ui.End, Width: ui.Fixed(noticeWidth),
		Padding: ui.Padding{Right: 12, Bottom: workspaceBarHeight + 12}, Leave: ui.Leave{Duration: noticeLeave, Y: 8}}
	context.Overlay("Notices", options, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 8}, func() {
			for index, notice := range editor.notices {
				view.noticeCard(context, index, notice)
			}
		})
	})
	if wait, due := nextNoticeEnd(editor.notices, time.Now()); due {
		context.RequestFrame(wait)
	}
}

// noticeCard is one notice: an icon of its kind, its words, and a button that closes it.
func (view *View) noticeCard(context *ui.Context, index int, notice Notice) {
	icon, ink := iconNotice, context.Theme.Primary
	if notice.Failed {
		icon, ink = iconProblem, context.Theme.Danger
	}
	look := ui.Look{Fill: context.Theme.Surface, Radius: 10, Border: context.Theme.Line, BorderWidth: 1}
	context.Push(strconv.FormatInt(notice.At.UnixNano(), 10)) // its time names it: closing one moves the others
	defer context.Pop()
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.Pad(10), Gap: 8, AlignY: ui.Start}, look, func() {
		context.IconSized(icon, noticeIconSize, ink)
		wrapped(context, notice.Text, ui.TextStyle{Size: context.Theme.SmallTextSize, Weight: draw2d.Regular})
		if context.FlatIconButton(iconClose, "") {
			view.Editor.CloseNotice(index)
		}
		context.Tooltip("Close")
	})
}

// nextNoticeEnd is how long until the first notice that only informs goes, if one will.
func nextNoticeEnd(notices []Notice, now time.Time) (time.Duration, bool) {
	for _, notice := range notices {
		if !notice.Failed {
			return max(notice.At.Add(noticeLifetime).Sub(now), 0), true
		}
	}
	return 0, false
}
