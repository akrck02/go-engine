package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
)

// The tools that build a game for Android and play it on a device are the Android SDK's: found in
// the variables Android's own tools read, or where Android Studio puts them.

// AndroidSDK is the Android SDK's folder: $ANDROID_HOME, $ANDROID_SDK_ROOT, or Android Studio's.
func AndroidSDK() (string, error) {
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

// AndroidNDK is the Android NDK's folder: $ANDROID_NDK_HOME, or the newest in the SDK.
func AndroidNDK(sdk string) (string, error) {
	if folder := os.Getenv("ANDROID_NDK_HOME"); folder != "" && FileExists(folder) {
		return folder, nil
	}
	versions, _ := filepath.Glob(filepath.Join(sdk, "ndk", "*")) // a bad pattern is impossible: it is fixed
	if len(versions) == 0 {
		return "", AndroidNDKNotFound
	}
	sort.Strings(versions)
	return versions[len(versions)-1], nil
}

// ADBTool is the Android Debug Bridge: the SDK's, or adb on the path.
func ADBTool() (string, error) {
	if sdk, err := AndroidSDK(); err == nil {
		if tool := filepath.Join(sdk, "platform-tools", "adb"+ExecutableSuffix()); FileExists(tool) {
			return tool, nil
		}
	}
	if tool, err := exec.LookPath("adb"); err == nil {
		return tool, nil
	}
	return "", ADBNotFound
}
