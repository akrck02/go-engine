package build

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/akrck02/go-engine/failure"
)

// Tools are the places of the tools that build for consoles and phones, as the person chose them
// in the editor; empty is "look for it": in the variables the tools themselves read, on the path,
// then where they are usually installed (places.go). They belong to the computer, not to a project, and are kept in the
// user's configuration folder.
type Tools struct {
	TinyGo          string `json:"tinygo,omitempty"`          // the tinygo program, with the nintendo3ds target
	DevkitPro       string `json:"devkitpro,omitempty"`       // devkitPro's folder, which has tools/bin/3dslink
	ConsoleEmulator string `json:"consoleEmulator,omitempty"` // a 3DS emulator's program (Azahar, Citra)
	AndroidSDK      string `json:"androidSdk,omitempty"`      // the Android SDK's folder, which has platform-tools/adb
	AndroidNDK      string `json:"androidNdk,omitempty"`      // the Android NDK's folder (one version of it)
	Java            string `json:"java,omitempty"`            // a JDK's folder, what $JAVA_HOME names
}

// toolsFile is where the tools' places are kept, in the user's configuration folder.
var toolsFile = filepath.Join("go-engine", "tools.json")

// ReadTools reads the tools' places this computer keeps; none kept is no place chosen.
func ReadTools() (Tools, error) {
	var tools Tools
	file, err := toolsPath()
	if err != nil {
		return tools, failure.Wrap(ToolsNotRead, "", err)
	}
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return tools, nil
	}
	if err == nil {
		err = json.Unmarshal(data, &tools)
	}
	if err != nil {
		return Tools{}, failure.Wrap(ToolsNotRead, file, err)
	}
	return tools, nil
}

// Save keeps the tools' places on this computer.
func (tools Tools) Save() error {
	file, err := toolsPath()
	if err != nil {
		return failure.Wrap(ToolsNotSaved, "", err)
	}
	data, err := json.MarshalIndent(tools, "", "\t")
	if err == nil {
		err = os.MkdirAll(filepath.Dir(file), 0o755)
	}
	if err == nil {
		err = os.WriteFile(file, append(data, '\n'), 0o644)
	}
	if err != nil {
		return failure.Wrap(ToolsNotSaved, file, err)
	}
	return nil
}

// toolsPath is the tools' file.
func toolsPath() (string, error) {
	folder, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(folder, toolsFile), nil
}

// chosenProgram is a program the person chose, when they chose one: an error when no program is
// there (nothing, or a folder).
func chosenProgram(place string) (string, bool, error) {
	if place == "" {
		return "", false, nil
	}
	if details, err := os.Stat(place); err != nil || details.IsDir() {
		return "", true, failure.Wrap(ToolPlaceWrong, place, err)
	}
	return place, true, nil
}

// chosenFolder is a folder the person chose, when they chose one: an error when no folder is
// there.
func chosenFolder(place string) (string, bool, error) {
	if place == "" {
		return "", false, nil
	}
	if details, err := os.Stat(place); err != nil || !details.IsDir() {
		return "", true, failure.Wrap(ToolPlaceWrong, place, err)
	}
	return place, true, nil
}
