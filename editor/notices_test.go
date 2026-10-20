package editor

import (
	"testing"
	"time"
)

func TestANoticeThatInformsGoesByItselfAndAFailureStays(test *testing.T) {
	editor := &Editor{}
	editor.say("Saved the scene", false)
	editor.say("The build failed", true)
	editor.forgetOldNotices(time.Now().Add(noticeLifetime + time.Second))
	notices := editor.Notices()
	if len(notices) != 1 || !notices[0].Failed {
		test.Fatalf("after a while the notices are %+v", notices)
	}
	editor.CloseNotice(0)
	if len(editor.Notices()) != 0 {
		test.Fatalf("a failure closed is still shown: %+v", editor.Notices())
	}
}

func TestOnlyTheNewestNoticesAreKept(test *testing.T) {
	editor := &Editor{}
	for _, text := range []string{"one", "two", "three", "four"} {
		editor.say(text, true)
	}
	notices := editor.Notices()
	if len(notices) != maxNotices || notices[0].Text != "two" || notices[maxNotices-1].Text != "four" {
		test.Fatalf("the notices kept are %+v", notices)
	}
}
