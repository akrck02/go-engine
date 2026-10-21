package build

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/akrck02/go-engine/failure"
)

// A tool found may still not do the job: a TinyGo without the 3DS, a devkitPro without its 3DS
// packages, an SDK whose install stopped halfway. Each check looks for the parts a build uses and
// says which are missing. They read folders, and CheckTinyGo runs TinyGo: not for every frame.

// CheckTinyGo tells whether a tinygo program has the nintendo3ds target, in the targets folder of
// its root (`tinygo targets` lists only the boards it can flash).
func CheckTinyGo(program string) error {
	root, err := Output("", []string{GoOnPath()}, program, "env", "TINYGOROOT") // TinyGo asks go where Go's root is
	if err != nil {
		return failure.Wrap(TinyGoWithout3DS, program, err)
	}
	if !FileExists(filepath.Join(strings.TrimSpace(root), "targets", "nintendo3ds.json")) {
		return failure.Wrap(TinyGoWithout3DS, program, nil)
	}
	return nil
}

// CheckDevkitPro tells whether a devkitPro folder has what a 3DS build links with: devkitARM's
// gcc, libctru and citro3d (which installs beside libctru).
func CheckDevkitPro(folder string) error {
	return partsIn(folder, DevkitProIncomplete,
		filepath.Join("devkitARM", "bin", "arm-none-eabi-gcc"+ExecutableSuffix()),
		filepath.Join("libctru", "lib", "libctru.a"),
		filepath.Join("libctru", "lib", "libcitro3d.a"))
}

// CheckAndroidSDK tells whether an SDK folder has what a build and a device use: adb, a version of
// the build-tools and a platform.
func CheckAndroidSDK(folder string) error {
	return partsIn(folder, AndroidSDKIncomplete,
		filepath.Join("platform-tools", "adb"+ExecutableSuffix()),
		filepath.Join("build-tools", "*", "apksigner*"),
		filepath.Join("platforms", "android-*", "android.jar"))
}

// CheckAndroidNDK tells whether an NDK folder has LLVM's compilers for this system, the ones
// go-ui-apk uses (macOS's are under darwin-x86_64 on every processor).
func CheckAndroidNDK(folder string) error {
	return partsIn(folder, AndroidNDKIncomplete,
		filepath.Join("toolchains", "llvm", "prebuilt", runtime.GOOS+"-x86_64", "bin"))
}

// CheckJava tells whether a JDK's folder has javac, which a runtime alone lacks.
func CheckJava(folder string) error {
	return partsIn(folder, JavaWithoutCompiler, filepath.Join("bin", "javac"+ExecutableSuffix()))
}

// partsIn is nil when a folder has every part (a path in it, which may hold * for any name), or an
// error with the code given that names the parts missing.
func partsIn(folder string, code *failure.Code, parts ...string) error {
	var missing []string
	for _, part := range parts {
		if found, _ := filepath.Glob(filepath.Join(folder, part)); len(found) == 0 { // the patterns are well formed
			missing = append(missing, filepath.ToSlash(part))
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return failure.Wrap(code, folder+": no "+strings.Join(missing, ", "), nil)
}
