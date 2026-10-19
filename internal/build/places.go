package build

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/akrck02/go-engine/internal/install"
)

// Where tools are usually installed on each system, looked at after the places the person chose,
// the variables the tools read and the path: what Install left in go-engine's data folder,
// Flatpak's programs, ~/Applications, a Mac's /Applications and Windows' Program Files.

// installed are the tools Install left in go-engine's data folder.
func installed() install.Installed {
	folder, err := install.DataFolder()
	if err != nil {
		return install.Installed{} // without a data folder, Install has put nothing there
	}
	return install.InstalledIn(folder)
}

// azaharFlatpak is Azahar's name in Flathub, which Flatpak gives the program it exports.
const azaharFlatpak = "org.azahar_emu.Azahar"

// emulatorPlaces are where a 3DS emulator is usually installed on this system, the most usual
// first.
func emulatorPlaces() []string {
	home, _ := os.UserHomeDir() // without a home, its places are only not found
	switch runtime.GOOS {
	case "darwin":
		program := filepath.Join("Azahar.app", "Contents", "MacOS", "azahar")
		return []string{filepath.Join("/Applications", program), filepath.Join(home, "Applications", program)}
	case "windows":
		return []string{underVariable("ProgramFiles", "Azahar", "azahar.exe")}
	}
	places := []string{
		filepath.Join("/var/lib/flatpak/exports/bin", azaharFlatpak),
		filepath.Join(home, ".local", "share", "flatpak", "exports", "bin", azaharFlatpak),
	}
	images, _ := filepath.Glob(filepath.Join(home, "Applications", "[Aa]zahar*.AppImage")) // the pattern is well formed
	return append(places, images...)
}

// homebrewFolders are where Homebrew installs: on Apple's processors, then on Intel's and Linux's
// (where it is /home/linuxbrew/.linuxbrew).
var homebrewFolders = []string{"/opt/homebrew", "/usr/local", "/home/linuxbrew/.linuxbrew"}

// tinyGoPlaces are where TinyGo's program is installed by Homebrew, for an editor started without
// the shell's path.
func tinyGoPlaces() []string {
	places := make([]string, 0, len(homebrewFolders))
	for _, folder := range homebrewFolders {
		places = append(places, filepath.Join(folder, "bin", "tinygo"))
	}
	return places
}

// homebrewSDKPlaces are where Homebrew's android-commandlinetools keeps the SDK.
func homebrewSDKPlaces() []string {
	places := make([]string, 0, len(homebrewFolders))
	for _, folder := range homebrewFolders {
		places = append(places, filepath.Join(folder, "share", "android-commandlinetools"))
	}
	return places
}

// javaPlaces are where Android Studio keeps the JDK it brings (its "JetBrains Runtime") on this
// system.
func javaPlaces() []string {
	home, _ := os.UserHomeDir() // without a home, its places are only not found
	switch runtime.GOOS {
	case "darwin":
		studio := filepath.Join("Android Studio.app", "Contents", "jbr", "Contents", "Home")
		return []string{filepath.Join("/Applications", studio), filepath.Join(home, "Applications", studio)}
	case "windows":
		return []string{underVariable("ProgramFiles", "Android", "Android Studio", "jbr")}
	}
	return []string{
		filepath.Join(home, "android-studio", "jbr"),
		"/opt/android-studio/jbr",
		"/snap/android-studio/current/jbr",
		"/usr/local/android-studio/jbr",
	}
}

// firstExisting is the first of some places where something is, or "".
func firstExisting(places []string) string {
	for _, place := range places {
		if place != "" && FileExists(place) {
			return place
		}
	}
	return ""
}

// underVariable is a place in the folder a variable names, or "" when it names none (a place
// relative to wherever the editor runs would be found by chance).
func underVariable(variable string, parts ...string) string {
	folder := os.Getenv(variable)
	if folder == "" {
		return ""
	}
	return filepath.Join(append([]string{folder}, parts...)...)
}
