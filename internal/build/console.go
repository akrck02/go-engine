package build

import (
	"os"
	"os/exec"
	"path/filepath"
)

// The tools that take a game to the Nintendo 3DS: devkitPro's 3dslink, and an emulator.

// devkitProFolder is where devkitPro installs itself when $DEVKITPRO is not set.
const devkitProFolder = "/opt/devkitpro"

// FindThreeDSLink finds devkitPro's 3dslink: in the tools of the devkitPro chosen, $DEVKITPRO's,
// devkitPro's usual folder's, or on the path.
func (tools Tools) FindThreeDSLink() (string, error) {
	if _, isChosen, err := chosen(tools.DevkitPro); isChosen && err != nil {
		return "", err
	}
	devkitPro := firstOf(tools.DevkitPro, os.Getenv("DEVKITPRO"))
	if devkitPro == "" {
		devkitPro = devkitProFolder
	}
	if tool := filepath.Join(devkitPro, "tools", "bin", "3dslink"+ExecutableSuffix()); FileExists(tool) {
		return tool, nil
	}
	if tool, err := exec.LookPath("3dslink"); err == nil {
		return tool, nil
	}
	return "", ThreeDSLinkNotFound
}

// consoleEmulators are the 3DS emulators looked for on the path, the most kept up first.
var consoleEmulators = []string{"azahar", "lime3ds", "citra-qt", "citra"}

// FindConsoleEmulator finds a 3DS emulator: the one chosen, $EMULATOR_3DS, or the first of
// consoleEmulators on the path.
func (tools Tools) FindConsoleEmulator() (string, error) {
	if place, isChosen, err := chosen(tools.ConsoleEmulator); isChosen {
		return place, err
	}
	if tool := os.Getenv("EMULATOR_3DS"); tool != "" {
		return tool, nil
	}
	for _, name := range consoleEmulators {
		if tool, err := exec.LookPath(name); err == nil {
			return tool, nil
		}
	}
	return "", ConsoleEmulatorNotFound
}

// firstOf is the first value that is not empty.
func firstOf(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
