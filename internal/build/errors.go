package build

import "github.com/akrck02/go-engine/failure"

// Error codes of the build package.
var (
	SourceNotFound          = failure.Define("build.source-not-found", "go-engine's source, which programs build against, is not where the editor was built")
	ToolFailed              = failure.Define("build.tool-failed", "a Go tool failed")
	TinyGoNotFound          = failure.Define("build.tinygo-not-found", "building for the 3DS needs TinyGo with the nintendo3ds target (github.com/akrck02/tinygo-3ds): set $TINYGO")
	AndroidSDKNotFound      = failure.Define("build.android-sdk-not-found", "building for Android needs the Android SDK (Android Studio installs it): set $ANDROID_HOME")
	AndroidNDKNotFound      = failure.Define("build.android-ndk-not-found", "building for Android needs the Android NDK, in the SDK's ndk folder (Android Studio's SDK Manager): or set $ANDROID_NDK_HOME")
	ADBNotFound             = failure.Define("build.adb-not-found", "playing on an Android device needs adb, the SDK's platform-tools")
	ThreeDSLinkNotFound     = failure.Define("build.3dslink-not-found", "sending to the 3DS needs devkitPro's 3dslink (the 3ds-dev package): set $DEVKITPRO")
	ConsoleEmulatorNotFound = failure.Define("build.console-emulator-not-found", "no 3DS emulator found (azahar, lime3ds, citra-qt or citra): put one on the path or set $EMULATOR_3DS")
	ToolPlaceWrong          = failure.Define("build.tool-place-wrong", "nothing is where the tool was set: change it in Build's tools, or empty it to look for the tool")
	ToolsNotRead            = failure.Define("build.tools-not-read", "the tools' places kept on this computer could not be read")
	ToolsNotSaved           = failure.Define("build.tools-not-saved", "the tools' places could not be kept on this computer")
)
