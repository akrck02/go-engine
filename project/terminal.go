package project

import (
	"os/exec"
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
		command = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+strings.ReplaceAll(script, `"`, `\"`)+`"`)
	} else {
		terminal, found := findTerminal()
		if !found {
			return failure.Wrap(TerminalNotFound, strings.Join(words, " "), nil)
		}
		command = exec.Command(terminal, "-e", "sh", "-c", script)
	}
	if err := command.Start(); err != nil {
		return err
	}
	go command.Wait() // the terminal is the person's; its end tells nothing
	return nil
}

// shellQuoted is a command as a shell reads it, each word in single quotes.
func shellQuoted(words []string) string {
	quoted := make([]string, len(words))
	for index, word := range words {
		quoted[index] = "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}
