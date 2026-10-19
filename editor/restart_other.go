//go:build !unix

package editor

import (
	"os"
	"os/exec"
)

// restart starts a program in the running editor's place, with its arguments and environment,
// and ends this one.
func restart(program string) error {
	command := exec.Command(program, os.Args[1:]...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
