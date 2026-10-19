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
	if folder, err := (Tools{}).FindAndroidNDK(sdk); err != nil || filepath.Base(folder) != "27.0.12077973" {
		test.Fatalf("the newest NDK is %q, %v", folder, err)
	}
}

// withNoTools leaves a test with no tool to find: no variables, an empty path and home.
func withNoTools(test *testing.T) string {
	empty := test.TempDir()
	for _, variable := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT", "ANDROID_NDK_HOME", "EMULATOR_3DS", "LOCALAPPDATA", "TINYGO"} {
		test.Setenv(variable, "")
	}
	test.Setenv("HOME", empty)
	test.Setenv("PATH", empty)
	test.Setenv("DEVKITPRO", empty)
	return empty
}

func TestMissingToolsAreSaidByTheirCodes(test *testing.T) {
	empty := withNoTools(test)
	var tools Tools
	cases := []struct {
		name   string
		find   func() (string, error)
		wanted error
	}{
		{"TinyGo", tools.FindTinyGo, TinyGoNotFound},
		{"the Android SDK", tools.FindAndroidSDK, AndroidSDKNotFound},
		{"the Android NDK", func() (string, error) { return tools.FindAndroidNDK(empty) }, AndroidNDKNotFound},
		{"adb", tools.FindADB, ADBNotFound},
		{"3dslink", tools.FindThreeDSLink, ThreeDSLinkNotFound},
		{"a 3DS emulator", tools.FindConsoleEmulator, ConsoleEmulatorNotFound},
	}
	for _, each := range cases {
		if _, err := each.find(); !errors.Is(err, each.wanted) {
			test.Errorf("looking for %s where there is none gives %v", each.name, err)
		}
	}
}

func TestAToolSetInTheEditorIsUsedFirst(test *testing.T) {
	folder := withNoTools(test)
	program := filepath.Join(folder, "my-tinygo")
	if err := os.WriteFile(program, nil, 0o755); err != nil {
		test.Fatal(err)
	}
	test.Setenv("TINYGO", "/elsewhere/tinygo")
	if found, err := (Tools{TinyGo: program}).FindTinyGo(); err != nil || found != program {
		test.Fatalf("TinyGo set in the editor gives %q, %v", found, err)
	}
	sdk := filepath.Join(folder, "sdk")
	if err := os.MkdirAll(filepath.Join(sdk, "platform-tools"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sdk, "platform-tools", "adb"+ExecutableSuffix()), nil, 0o755); err != nil {
		test.Fatal(err)
	}
	if found, err := (Tools{AndroidSDK: sdk}).FindADB(); err != nil || found != filepath.Join(sdk, "platform-tools", "adb"+ExecutableSuffix()) {
		test.Fatalf("adb in the SDK set in the editor gives %q, %v", found, err)
	}
}

func TestAToolSetWhereNothingIsIsSaidByItsCode(test *testing.T) {
	withNoTools(test)
	tools := Tools{TinyGo: "/nowhere/tinygo", AndroidSDK: "/nowhere/sdk", AndroidNDK: "/nowhere/ndk", DevkitPro: "/nowhere/devkitpro", ConsoleEmulator: "/nowhere/azahar"}
	finders := []func() (string, error){tools.FindTinyGo, tools.FindAndroidSDK, tools.FindThreeDSLink, tools.FindConsoleEmulator, tools.FindADB,
		func() (string, error) { return tools.FindAndroidNDK("/nowhere/sdk") }}
	for _, find := range finders {
		if _, err := find(); !errors.Is(err, ToolPlaceWrong) {
			test.Errorf("a tool set where nothing is gives %v", err)
		}
	}
}

func TestAToolSetWhereTheWrongThingIsIsSaidByItsCode(test *testing.T) {
	folder := withNoTools(test)
	program := filepath.Join(folder, "program")
	if err := os.WriteFile(program, nil, 0o755); err != nil {
		test.Fatal(err)
	}
	tools := Tools{TinyGo: folder, ConsoleEmulator: folder, AndroidSDK: program, AndroidNDK: program, DevkitPro: folder}
	finders := map[string]func() (string, error){
		"a folder as TinyGo": tools.FindTinyGo, "a folder as the emulator": tools.FindConsoleEmulator,
		"a program as the SDK": tools.FindAndroidSDK, "a program as the NDK": func() (string, error) { return tools.FindAndroidNDK("") },
		"a devkitPro with no 3dslink": tools.FindThreeDSLink, "an SDK with no adb": Tools{AndroidSDK: folder}.FindADB,
	}
	for name, find := range finders {
		if _, err := find(); !errors.Is(err, ToolPlaceWrong) {
			test.Errorf("%s gives %v", name, err)
		}
	}
}

func TestTheToolsPlacesAreKeptBetweenSessions(test *testing.T) {
	useConfigurationFolder(test, test.TempDir())
	if read, err := ReadTools(); err != nil || read != (Tools{}) {
		test.Fatalf("with nothing saved the tools are %+v, %v", read, err)
	}
	saved := Tools{TinyGo: "/opt/tinygo/bin/tinygo", AndroidSDK: "/opt/android"}
	if err := saved.Save(); err != nil {
		test.Fatal(err)
	}
	if read, err := ReadTools(); err != nil || read != saved {
		test.Fatalf("the tools came back as %+v, %v", read, err)
	}
}

// useConfigurationFolder makes a test keep the tools' places in a folder of its own, not the
// person's, on every system: the variables os.UserConfigDir reads.
func useConfigurationFolder(test *testing.T, folder string) {
	for _, variable := range []string{"XDG_CONFIG_HOME", "HOME", "APPDATA", "LOCALAPPDATA"} {
		test.Setenv(variable, folder)
	}
}
