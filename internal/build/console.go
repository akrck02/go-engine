package build

import (
	"os"
	"os/exec"
	"path/filepath"
)

// The tools that take a game to the Nintendo 3DS: devkitPro's 3dslink, and an emulator.

// devkitProFolder is where devkitPro installs itself when $DEVKITPRO is not set.
const devkitProFolder = "/opt/devkitpro"

// ThreeDSLinkTool is devkitPro's 3dslink: in $DEVKITPRO's tools, devkitPro's usual folder, or on
// the path.
func ThreeDSLinkTool() (string, error) {
	devkitPro := os.Getenv("DEVKITPRO")
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

// ConsoleEmulator is a 3DS emulator: $EMULATOR_3DS, or the first of consoleEmulators on the path.
func ConsoleEmulator() (string, error) {
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
