package build

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTheNewestNDKIsChosenByItsNumbers(test *testing.T) {
	sdk := test.TempDir()
	for _, version := range []string{"27.0.9", "27.0.12077973", "26.3.11579264"} {
		if err := os.MkdirAll(filepath.Join(sdk, "ndk", version), 0o755); err != nil {
			test.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(sdk, "ndk", "zz-notes.txt"), nil, 0o644); err != nil {
		test.Fatal(err)
	}
	test.Setenv("ANDROID_NDK_HOME", "")
	if folder, err := AndroidNDK(sdk); err != nil || filepath.Base(folder) != "27.0.12077973" {
		test.Fatalf("the newest NDK is %q, %v", folder, err)
	}
}

func TestMissingToolsAreSaidByTheirCodes(test *testing.T) {
	empty := test.TempDir()
	for _, variable := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT", "ANDROID_NDK_HOME", "EMULATOR_3DS", "LOCALAPPDATA"} {
		test.Setenv(variable, "")
	}
	test.Setenv("HOME", empty)
	test.Setenv("PATH", empty)
	test.Setenv("DEVKITPRO", empty)
	cases := []struct {
		name   string
		find   func() (string, error)
		wanted error
	}{
		{"the Android SDK", AndroidSDK, AndroidSDKNotFound},
		{"the Android NDK", func() (string, error) { return AndroidNDK(empty) }, AndroidNDKNotFound},
		{"adb", ADBTool, ADBNotFound},
		{"3dslink", ThreeDSLinkTool, ThreeDSLinkNotFound},
		{"a 3DS emulator", ConsoleEmulator, ConsoleEmulatorNotFound},
	}
	for _, each := range cases {
		if _, err := each.find(); !errors.Is(err, each.wanted) {
			test.Errorf("looking for %s where there is none gives %v", each.name, err)
		}
	}
}
