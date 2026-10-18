package project

import (
	"os/exec"

	"github.com/akrck02/go-engine/internal/build"
)

// A game built for the Nintendo 3DS is tried on the console itself, sent over the network to the
// Homebrew Launcher's netloader (Y opens it) with devkitPro's 3dslink, or in an emulator.

// SendTo3DS builds the game for the 3DS and sends it to the console's Homebrew Launcher, which
// starts it. The settings' ConsoleAddress is the console's; empty, 3dslink looks for it.
func (project *Project) SendTo3DS() error {
	built, err := project.Build3DS()
	if err != nil {
		return err
	}
	link, err := build.ThreeDSLinkTool()
	if err != nil {
		return err
	}
	return project.sendWith(link, built.Program)
}

// sendWith sends a program to the console with a 3dslink.
func (project *Project) sendWith(link, program string) error {
	var arguments []string
	if project.Settings.ConsoleAddress != "" {
		arguments = append(arguments, "-a", project.Settings.ConsoleAddress)
	}
	return build.Run(project.Folder, nil, link, append(arguments, program)...)
}

// EmulatorCommand builds the game for the 3DS and returns the command that opens it in an
// emulator (Azahar, or Citra), not started: the editor shows what it prints and can stop it.
func (project *Project) EmulatorCommand() (*exec.Cmd, error) {
	built, err := project.Build3DS()
	if err != nil {
		return nil, err
	}
	emulator, err := build.ConsoleEmulator()
	if err != nil {
		return nil, err
	}
	command := exec.Command(emulator, built.Program)
	command.Dir = project.Folder
	return command, nil
}
