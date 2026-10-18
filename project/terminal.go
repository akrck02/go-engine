package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// RunInTerminal runs a command in a new terminal window, left open until the person presses
// Enter: its questions (a password) and what it says are theirs to answer and read.
func RunInTerminal(words []string) error {
	script := shellQuoted(words) + "; echo; printf 'Press Enter to close. '; read answer"
	var command *exec.Cmd
	if runtime.GOOS == "darwin" {
		command = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+appleScriptQuoted(script)+`"`)
	} else {
		terminal, found := findTerminal()
		if !found {
			return failure.Wrap(TerminalNotFound, strings.Join(words, " "), nil)
		}
		arguments := append(terminalArguments(terminal), "sh", "-c", script)
		command = exec.Command(terminal, arguments...)
	}
	if err := command.Start(); err != nil {
		return err
	}
	go command.Wait() // the terminal is the person's; its end tells nothing
	return nil
}

// terminals are the terminals looked for, the person's own ($TERMINAL) first.
var terminals = []string{"x-terminal-emulator", "gnome-terminal", "konsole", "kitty", "alacritty", "wezterm", "xterm"}

// findTerminal finds a terminal to open a window of.
func findTerminal() (string, bool) {
	for _, terminal := range append([]string{os.Getenv("TERMINAL")}, terminals...) {
		if terminal == "" {
			continue
		}
		if found, err := exec.LookPath(terminal); err == nil {
			return found, true
		}
	}
	return "", false
}

// terminalArguments are what a terminal wants before the program it runs and its arguments: each
// terminal says it its own way.
func terminalArguments(terminal string) []string {
	switch filepath.Base(terminal) {
	case "gnome-terminal":
		return []string{"--"}
	case "wezterm":
		return []string{"start", "--"}
	case "kitty":
		return nil
	}
	return []string{"-e"} // xterm, konsole, alacritty and x-terminal-emulator's
}

// shellQuoted is a command as a shell reads it, each word in single quotes.
func shellQuoted(words []string) string {
	quoted := make([]string, len(words))
	for index, word := range words {
		quoted[index] = "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}

// appleScriptQuoted is a text as an AppleScript string holds it: backslashes and quotes escaped.
func appleScriptQuoted(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, `\`, `\\`), `"`, `\"`)
}
