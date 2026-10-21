package build

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/akrck02/go-engine/internal/install"
)

func TestToolsInstallLeftButWereNotSavedAreFound(test *testing.T) {
	data := withNoTools(test)
	if runtime.GOOS != "linux" {
		test.Skip("the data folder is $XDG_DATA_HOME on Linux only")
	}
	folder := filepath.Join(data, "go-engine")
	tinyGo := makeParts(test, filepath.Join(folder, "tinygo", install.TinyGoRelease), "tinygo/bin/tinygo")
	azahar := makeParts(test, filepath.Join(folder, "azahar", "2100.1"), "azahar.AppImage")
	sdk := makeParts(test, filepath.Join(folder, "android-sdk"), "platform-tools/adb")
	cases := []struct {
		name   string
		find   func() (string, error)
		wanted string
	}{
		{"TinyGo", Tools{}.FindTinyGo, filepath.Join(tinyGo, "tinygo", "bin", "tinygo")},
		{"Azahar", Tools{}.FindNintendo3DSEmulator, filepath.Join(azahar, "azahar.AppImage")},
		{"the Android SDK", Tools{}.FindAndroidSDK, sdk},
	}
	for _, each := range cases {
		if found, err := each.find(); err != nil || found != each.wanted {
			test.Errorf("%s Install left is found as %q, %v", each.name, found, err)
		}
	}
}

func TestAnEmulatorInTheHomesApplicationsIsFound(test *testing.T) {
	home := withNoTools(test)
	if runtime.GOOS != "linux" {
		test.Skip("AppImages are Linux's")
	}
	image := filepath.Join(makeParts(test, filepath.Join(home, "Applications"), "Azahar-2100.1.AppImage"), "Azahar-2100.1.AppImage")
	if found, err := (Tools{}).FindNintendo3DSEmulator(); err != nil || found != image {
		test.Fatalf("the AppImage in ~/Applications is found as %q, %v", found, err)
	}
}

func TestAJDKIsFoundFromTheJavacOnThePath(test *testing.T) {
	home := withNoTools(test)
	if runtime.GOOS == "windows" {
		test.Skip("links need permission on Windows")
	}
	jdk := makeParts(test, filepath.Join(home, "jdk-21"), "bin/javac")
	onPath := filepath.Join(home, "path")
	if err := os.MkdirAll(onPath, 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(jdk, "bin", "javac"), filepath.Join(onPath, "javac")); err != nil {
		test.Fatal(err)
	}
	test.Setenv("PATH", onPath)
	if found, err := (Tools{}).FindJava(); err != nil || found != jdk {
		test.Fatalf("the JDK of the javac on the path is found as %q, %v", found, err)
	}
}

func TestAndroidStudiosJDKIsFound(test *testing.T) {
	home := withNoTools(test)
	if runtime.GOOS != "linux" {
		test.Skip("Android Studio's place in the home is Linux's")
	}
	jdk := makeParts(test, filepath.Join(home, "android-studio", "jbr"), "bin/javac")
	if found, err := (Tools{}).FindJava(); err != nil || found != jdk {
		test.Fatalf("Android Studio's JDK is found as %q, %v", found, err)
	}
}

func TestAJavaHomeWithoutJavacIsPassedOver(test *testing.T) {
	home := withNoTools(test)
	if runtime.GOOS != "linux" {
		test.Skip("Android Studio's place in the home is Linux's")
	}
	test.Setenv("JAVA_HOME", makeParts(test, filepath.Join(home, ".jenv"), "version"))
	jdk := makeParts(test, filepath.Join(home, "android-studio", "jbr"), "bin/javac")
	if found, err := (Tools{}).FindJava(); err != nil || found != jdk {
		test.Fatalf("with $JAVA_HOME a version manager's folder, the JDK found is %q, %v", found, err)
	}
}

func TestAJDKIsFoundFromWhereJavaSaysItRuns(test *testing.T) {
	home := withNoTools(test)
	if runtime.GOOS == "windows" {
		test.Skip("the fake java is a shell script")
	}
	jdk := makeParts(test, filepath.Join(home, "versions", "21"), "bin/javac")
	script := "#!/bin/sh\necho '    java.home = " + jdk + "' >&2\n"
	shims := filepath.Join(home, "shims")
	if err := os.MkdirAll(shims, 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shims, "java"), []byte(script), 0o755); err != nil {
		test.Fatal(err)
	}
	test.Setenv("PATH", shims)
	if found, err := (Tools{}).FindJava(); err != nil || found != jdk {
		test.Fatalf("the JDK java runs from is found as %q, %v", found, err)
	}
}
