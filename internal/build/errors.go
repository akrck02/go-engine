package build

import "github.com/akrck02/go-engine/failure"

// Error codes of the build package.
var (
	SourceNotFound          = failure.Define("build.source-not-found", "go-engine's source, which programs build against, is not where the editor was built")
	ToolFailed              = failure.Define("build.tool-failed", "a Go tool failed")
	TinyGoNotFound          = failure.Define("build.tinygo-not-found", "building for the 3DS needs TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds): install it or set its place in Build's tools, or $TINYGO")
	AndroidSDKNotFound      = failure.Define("build.android-sdk-not-found", "building for Android needs the Android SDK (Android Studio installs it): install it or set its place in Build's tools, or $ANDROID_HOME")
	AndroidNDKNotFound      = failure.Define("build.android-ndk-not-found", "building for Android needs the Android NDK, in the SDK's ndk folder (Android Studio's SDK Manager): install it or set its place in Build's tools, or $ANDROID_NDK_HOME")
	ADBNotFound             = failure.Define("build.adb-not-found", "playing on an Android device needs adb, the SDK's platform-tools")
	ThreeDSLinkNotFound     = failure.Define("build.3dslink-not-found", "sending to the 3DS needs devkitPro's 3dslink (the 3ds-dev package): install it or set its folder in Build's tools, or $DEVKITPRO")
	ConsoleEmulatorNotFound = failure.Define("build.console-emulator-not-found", "no 3DS emulator found (azahar, lime3ds, citra-qt or citra): install it or set its place in Build's tools, or put one on the path")
	JavaNotFound            = failure.Define("build.java-not-found", "Android's builds need a JDK (17 or newer, as Adoptium's Temurin, or Android Studio's): install one or set its folder in Build's tools, or $JAVA_HOME")
	TinyGoWithoutConsole    = failure.Define("build.tinygo-without-console", "the TinyGo found has no nintendo3ds target: it is not tinygo-3ds's (Install gets it)")
	DevkitProIncomplete     = failure.Define("build.devkitpro-incomplete", "devkitPro lacks parts of the 3DS's: install its 3ds-dev packages (Install runs devkitPro's installer)")
	AndroidSDKIncomplete    = failure.Define("build.android-sdk-incomplete", "the Android SDK lacks parts a build needs: install them with its SDK Manager, or Install")
	AndroidNDKIncomplete    = failure.Define("build.android-ndk-incomplete", "the Android NDK has no compilers for this system: install another version of it")
	JavaWithoutCompiler     = failure.Define("build.java-without-compiler", "the Java found is only a runtime, with no javac: Android's builds need a JDK")
	ToolPlaceWrong          = failure.Define("build.tool-place-wrong", "nothing is where the tool was set: change it in Build's tools, or empty it to look for the tool")
	ToolsNotRead            = failure.Define("build.tools-not-read", "the tools' places kept on this computer could not be read")
	ToolsNotSaved           = failure.Define("build.tools-not-saved", "the tools' places could not be kept on this computer")
)
