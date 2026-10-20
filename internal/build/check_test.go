package build

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// makeParts makes empty files in a folder, and the folders they are in.
func makeParts(test *testing.T, folder string, parts ...string) string {
	for _, part := range parts {
		path := filepath.Join(folder, part)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o755); err != nil {
			test.Fatal(err)
		}
	}
	return folder
}

// toolChecks are each folder tool's check, its parts, and the code it gives when one is missing.
var toolChecks = []struct {
	name   string
	check  func(folder string) error
	parts  []string
	wanted error
}{
	{"devkitPro", CheckDevkitPro, []string{"devkitARM/bin/arm-none-eabi-gcc" + ExecutableSuffix(), "libctru/lib/libctru.a", "libctru/lib/libcitro3d.a"}, DevkitProIncomplete},
	{"the Android SDK", CheckAndroidSDK, []string{"platform-tools/adb" + ExecutableSuffix(), "build-tools/34.0.0/apksigner", "platforms/android-34/android.jar"}, AndroidSDKIncomplete},
	{"the Android NDK", CheckAndroidNDK, []string{"toolchains/llvm/prebuilt/" + runtime.GOOS + "-x86_64/bin/clang"}, AndroidNDKIncomplete},
	{"a JDK", CheckJava, []string{"bin/javac" + ExecutableSuffix()}, JavaWithoutCompiler},
}

func TestAToolWithEveryPartPassesItsCheck(test *testing.T) {
	for _, tool := range toolChecks {
		if err := tool.check(makeParts(test, test.TempDir(), tool.parts...)); err != nil {
			test.Errorf("%s with every part gives %v", tool.name, err)
		}
	}
}

func TestEachToolMissingAPartSaysWhich(test *testing.T) {
	for _, tool := range toolChecks {
		for index, missing := range tool.parts {
			others := append(append([]string{}, tool.parts[:index]...), tool.parts[index+1:]...)
			err := tool.check(makeParts(test, test.TempDir(), others...))
			if !errors.Is(err, tool.wanted) {
				test.Errorf("%s without %s gives %v", tool.name, missing, err)
				continue
			}
			if !strings.Contains(err.Error(), ": no "+strings.Split(missing, "/")[0]) {
				test.Errorf("%s without %s does not say which: %v", tool.name, missing, err)
			}
		}
	}
}

// fakeTinyGo makes a tinygo program whose root has some targets, and says where the root is as
// `tinygo env TINYGOROOT` does.
func fakeTinyGo(test *testing.T, targets ...string) string {
	if runtime.GOOS == "windows" {
		test.Skip("the fake tinygo is a shell script")
	}
	root := test.TempDir()
	for _, target := range targets {
		makeParts(test, root, filepath.Join("targets", target+".json"))
	}
	program := filepath.Join(test.TempDir(), "tinygo")
	if err := os.WriteFile(program, []byte("#!/bin/sh\necho "+root+"\n"), 0o755); err != nil {
		test.Fatal(err)
	}
	return program
}

func TestOnlyATinyGoWithTheConsoleTargetPassesItsCheck(test *testing.T) {
	if err := CheckTinyGo(fakeTinyGo(test, "arduino", "nintendo3ds", "wasm")); err != nil {
		test.Errorf("a TinyGo with the 3DS gives %v", err)
	}
	if err := CheckTinyGo(fakeTinyGo(test, "arduino", "wasm")); !errors.Is(err, TinyGoWithoutConsole) {
		test.Errorf("a TinyGo without the 3DS gives %v", err)
	}
	if err := CheckTinyGo(filepath.Join(test.TempDir(), "nothing")); !errors.Is(err, TinyGoWithoutConsole) {
		test.Errorf("a TinyGo that does not run gives %v", err)
	}
}
