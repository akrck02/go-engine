package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// devkitPro is installed by its own installer, not downloaded here: dkp-pacman, its package
// manager, installs the 3DS's packages (3ds-dev: devkitARM, libctru, citro3d, the 3DS tools) and
// asks for the system's password to write /opt/devkitpro. Where there is no dkp-pacman yet, the
// person follows devkitPro's guide to get it.

// DevkitProGuide is devkitPro's guide to installing it.
const DevkitProGuide = "https://devkitpro.org/wiki/Getting_Started"

// devkitProPackages is the group of devkitPro's packages for the 3DS.
const devkitProPackages = "3ds-dev"

// DevkitProCommand is the command that installs devkitPro's 3DS packages, to run where the person
// can type their password; false when there is no dkp-pacman to run it (DevkitProGuide says how
// to get one).
func DevkitProCommand() ([]string, bool) {
	manager, found := devkitProManager()
	if !found {
		return nil, false
	}
	return []string{"sudo", manager, "-S", "--needed", devkitProPackages}, true
}

// devkitProManager is dkp-pacman: on the path (Linux), or in devkitPro's folder (a Mac).
func devkitProManager() (string, bool) {
	if runtime.GOOS == "windows" {
		return "", false // devkitPro's Windows installer is a program of its own
	}
	if manager, err := exec.LookPath("dkp-pacman"); err == nil {
		return manager, true
	}
	manager := filepath.Join("/opt", "devkitpro", "pacman", "bin", "pacman")
	if _, err := os.Stat(manager); err == nil {
		return manager, true
	}
	return "", false
}
