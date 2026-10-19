package install

import (
	"os"
	"path/filepath"
	"slices"
)

// What Install left in its folder, for the builds to find when the place it gave was not saved:
// the person may have closed the editor before saving, or emptied the place to look again.

// Installed are the places of the tools installed in a folder; empty is not installed there.
type Installed struct {
	TinyGo     string // the tinygo program
	Azahar     string // Azahar's program
	AndroidSDK string // the SDK's folder
}

// InstalledIn finds the tools installed in a folder (DataFolder, or a test's own), for this
// system.
func InstalledIn(folder string) Installed {
	installer := Installer{Folder: folder}
	return Installed{
		TinyGo:     installer.installedTinyGo(),
		Azahar:     installer.installedAzahar(),
		AndroidSDK: installer.installedAndroidSDK(),
	}
}

// installedTinyGo is the tinygo program of the release go-engine installs, else of the newest
// other release installed.
func (installer Installer) installedTinyGo() string {
	for _, release := range releasesFirstWanted(filepath.Join(installer.Folder, "tinygo"), TinyGoRelease) {
		if program := tinyGoProgram(release, installer.system()); fileExists(program) {
			return program
		}
	}
	return ""
}

// installedAzahar is the program of the newest Azahar installed.
func (installer Installer) installedAzahar() string {
	for _, release := range releasesFirstWanted(filepath.Join(installer.Folder, "azahar"), "") {
		if image := filepath.Join(release, "azahar.AppImage"); fileExists(image) {
			return image
		}
		if program, err := findProgram(release, azaharProgram(installer.system())); err == nil {
			return program
		}
	}
	return ""
}

// installedAndroidSDK is the SDK's folder, once Install has put platform-tools in it (the
// command-line tools alone are an install that stopped halfway).
func (installer Installer) installedAndroidSDK() string {
	sdk := filepath.Join(installer.Folder, androidSDKFolder)
	if !fileExists(filepath.Join(sdk, "platform-tools")) {
		return ""
	}
	return sdk
}

// releasesFirstWanted are the folders of the releases in a folder: the wanted one first, then the
// others by name from the last down (the newest, while their numbers have the same digits).
func releasesFirstWanted(folder, wanted string) []string {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil // nothing installed
	}
	var releases []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != wanted {
			releases = append(releases, filepath.Join(folder, entry.Name()))
		}
	}
	slices.Sort(releases)
	slices.Reverse(releases)
	if wanted != "" && fileExists(filepath.Join(folder, wanted)) {
		releases = slices.Insert(releases, 0, filepath.Join(folder, wanted))
	}
	return releases
}
