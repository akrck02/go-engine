package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
)

func TestSendingToTheConsoleRunsThreeDSLinkWithTheProgram(test *testing.T) {
	if runtime.GOOS == "windows" {
		test.Skip("the fake 3dslink is a shell script")
	}
	folder := test.TempDir()
	calls := filepath.Join(folder, "calls")
	link := filepath.Join(folder, "3dslink")
	if err := os.WriteFile(link, []byte("#!/bin/sh\necho \"$@\" >> "+calls+"\n"), 0o755); err != nil {
		test.Fatal(err)
	}
	project := &Project{Folder: folder, Settings: engine.Settings{Name: "Harbour Lights"}}
	if err := project.sendWith(link, "/games/harbour.3dsx"); err != nil {
		test.Fatal(err)
	}
	project.Settings.DeviceAddress = "192.168.1.20"
	if err := project.sendWith(link, "/games/harbour.3dsx"); err != nil {
		test.Fatal(err)
	}
	said, _ := os.ReadFile(calls)
	lines := strings.Split(strings.TrimSpace(string(said)), "\n")
	if len(lines) != 2 || lines[0] != "/games/harbour.3dsx" || lines[1] != "-a 192.168.1.20 /games/harbour.3dsx" {
		test.Errorf("3dslink was asked:\n%s", said)
	}
}

func TestACommandIsQuotedForTheShell(test *testing.T) {
	if quoted := shellQuoted([]string{"sudo", "/opt/dev kit/pacman", "it's"}); quoted != `'sudo' '/opt/dev kit/pacman' 'it'\''s'` {
		test.Fatalf("the command is quoted as %s", quoted)
	}
}

func TestEachTerminalIsToldTheProgramItsOwnWay(test *testing.T) {
	cases := map[string]string{"/usr/bin/gnome-terminal": "--", "/usr/bin/wezterm": "start --", "/usr/bin/kitty": "", "/usr/bin/xterm": "-e", "/usr/bin/konsole": "-e"}
	for terminal, wanted := range cases {
		if got := strings.Join(terminalArguments(terminal), " "); got != wanted {
			test.Errorf("%s is told %q, not %q", terminal, got, wanted)
		}
	}
}

func TestATextIsQuotedForAppleScript(test *testing.T) {
	if quoted := appleScriptQuoted(`'it'\''s' "lamp"`); quoted != `'it'\\''s' \"lamp\"` {
		test.Fatalf("the text is quoted as %s", quoted)
	}
}
