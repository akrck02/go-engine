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
	project.Settings.ConsoleAddress = "192.168.1.20"
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
