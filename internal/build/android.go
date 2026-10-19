package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// The tools that build a game for Android and play it on a device are the Android SDK's: found in
// the variables Android's own tools read, or where Android Studio puts them.

// FindAndroidSDK finds the Android SDK's folder: the one chosen, $ANDROID_HOME,
// $ANDROID_SDK_ROOT, or Android Studio's.
func (tools Tools) FindAndroidSDK() (string, error) {
	if place, isChosen, err := chosenFolder(tools.AndroidSDK); isChosen {
		return place, err
	}
	for _, variable := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if folder := os.Getenv(variable); folder != "" && FileExists(folder) {
			return folder, nil
		}
	}
	if folder := androidStudioSDK(); folder != "" && FileExists(folder) {
		return folder, nil
	}
	return "", AndroidSDKNotFound
}

// androidStudioSDK is where Android Studio installs the SDK on this system.
func androidStudioSDK() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Android", "sdk")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Android", "Sdk")
	}
	return filepath.Join(home, "Android", "Sdk")
}

// FindAndroidNDK finds the Android NDK's folder: the one chosen, $ANDROID_NDK_HOME, or the
// newest in the SDK.
func (tools Tools) FindAndroidNDK(sdk string) (string, error) {
	if place, isChosen, err := chosenFolder(tools.AndroidNDK); isChosen {
		return place, err
	}
	if folder := os.Getenv("ANDROID_NDK_HOME"); folder != "" && FileExists(folder) {
		return folder, nil
	}
	entries, err := os.ReadDir(filepath.Join(sdk, "ndk"))
	if err != nil {
		return "", AndroidNDKNotFound
	}
	newest := ""
	for _, entry := range entries {
		if entry.IsDir() && (newest == "" || newerVersion(entry.Name(), newest)) {
			newest = entry.Name()
		}
	}
	if newest == "" {
		return "", AndroidNDKNotFound
	}
	return filepath.Join(sdk, "ndk", newest), nil
}

// newerVersion tells whether a version written as numbers joined by dots (28.2.13676358) is
// newer than another, comparing number by number.
func newerVersion(version, other string) bool {
	parts, others := strings.Split(version, "."), strings.Split(other, ".")
	for index := range min(len(parts), len(others)) {
		number, _ := strconv.Atoi(parts[index]) // a part that is no number counts as 0
		otherNumber, _ := strconv.Atoi(others[index])
		if number != otherNumber {
			return number > otherNumber
		}
	}
	return len(parts) > len(others)
}

// FindADB finds the Android Debug Bridge: the SDK's, or adb on the path; with an SDK chosen,
// only the SDK's.
func (tools Tools) FindADB() (string, error) {
	sdk, err := tools.FindAndroidSDK()
	if tools.AndroidSDK != "" {
		return adbIn(sdk, err)
	}
	if err == nil {
		if tool := filepath.Join(sdk, "platform-tools", "adb"+ExecutableSuffix()); FileExists(tool) {
			return tool, nil
		}
	}
	if tool, err := exec.LookPath("adb"); err == nil {
		return tool, nil
	}
	return "", ADBNotFound
}

// adbIn is the adb of an SDK chosen, or why there is none.
func adbIn(sdk string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if tool := filepath.Join(sdk, "platform-tools", "adb"+ExecutableSuffix()); FileExists(tool) {
		return tool, nil
	}
	return "", failure.Wrap(ToolPlaceWrong, sdk+": no platform-tools/adb", nil)
}
