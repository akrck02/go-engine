//go:build unix

package editor

import (
	"os"
	"syscall"
)

// restart replaces the running editor with a program, keeping its arguments and environment.
func restart(program string) error {
	return syscall.Exec(program, os.Args, os.Environ())
}
